import type { Expr, Invariant, Op, RmwStyle, Snapshot, Tuple, Txn, TxnEventType, TxnWorld } from './types.ts'
import type { Ctx } from './world.ts'
import {
  allTxns,
  allocSlot,
  appendWal,
  assignXid,
  committedBefore,
  committedValue,
  emit,
  horizon,
  isDead,
  isLive,
  keysOf,
  labelOf,
  newTxn,
  pagesOf,
  statusOf,
  tagOf,
  takeSnapshot,
  txnByTag,
  txnByXid,
  visibleTo,
  xidLabel,
} from './world.ts'

/**
 * Один шаг всей системы.
 *
 * Порядок фиксирован: падение сервера, потом завершение сброса WAL (ждущие
 * коммиты становятся видны), потом сессии по порядку, потом фоновые процессы —
 * новый сброс WAL, поиск дедлоков, автовакуум, контрольная точка — и итог тика.
 */
export function tick(ctx: Ctx): void {
  const w = ctx.world
  if (w.finished) return
  w.tick++
  w.marks = ctx.scenario.sessions.map(() => null)

  if (w.config.crashAt > 0 && w.tick === w.config.crashAt && !w.crashed) crash(ctx)

  if (w.tick < w.downUntil) {
    w.marks = w.marks.map(() => ({ kind: 'down' }))
    account(ctx)
    return
  }
  if (w.crashed && w.tick === w.downUntil) recovered(ctx)

  finishFlush(ctx)
  for (const t of [...w.txns]) stepSession(ctx, t)
  startFlush(ctx)
  deadlockCheck(ctx)
  if (w.config.autovacuum > 0 && w.tick % w.config.autovacuum === 0) vacuum(ctx)
  if (w.config.checkpointEvery > 0 && w.tick % w.config.checkpointEvery === 0) checkpoint(ctx)
  account(ctx)
}

/* ────────────────────────────── сессии ────────────────────────────── */

const RETRYABLE = new Set(['serialization', 'ssi', 'deadlock', 'crash'])
const MAX_RETRIES = 5

function stepSession(ctx: Ctx, t: Txn): void {
  const w = ctx.world
  const spec = ctx.scenario.sessions[t.spec]!

  // Транзакция закончилась — пора ли начинать следующую?
  if (t.state === 'committed' || t.state === 'aborted') {
    if (t.restartAt === null || w.tick < t.restartAt) return
    const retry = t.retryNext
    const next = newTxn(spec, t.spec, w.config, w.tick, {
      run: t.run + 1,
      iter: retry ? t.iter : t.iter + 1,
      retries: retry ? t.retries + 1 : 0,
      repeatsLeft: retry ? t.repeatsLeft : t.repeatsLeft - 1,
    })
    w.history.push(t)
    w.txns[t.spec] = next
    if (retry) {
      w.stats.retries++
      emit(ctx, 'txn.retry', { txn: [t.spec] }, { txn: labelOf(next), prev: labelOf(t), reason: t.error, attempt: next.retries + 1 })
    }
    t = next
  }

  if (t.state === 'done') return
  if (t.state === 'committing') {
    w.marks[t.spec] = { kind: 'commit-wait' }
    return
  }
  if (t.state === 'waiting') {
    execOp(ctx, t, true)
    return
  }
  if (w.tick < t.nextAt) return
  if (t.opIdx >= spec.ops.length) return // «idle in transaction»: операции кончились, а COMMIT так и не пришёл
  execOp(ctx, t, false)
}

/** Операция в том виде, в котором её выполнит база, с учётом стиля «прочитать — изменить — записать». */
export function effectiveOp(style: RmwStyle, op: Op): Op {
  if (op.kind === 'read' && op.rmw && style === 'for-update') return { ...op, forUpdate: true }
  if (op.kind === 'update' && op.set.kind === 'read' && style === 'atomic') {
    return { ...op, set: { kind: 'delta', d: op.set.d } }
  }
  return op
}

/**
 * Снимок для оператора. READ COMMITTED берёт новый на каждый оператор,
 * REPEATABLE READ и SERIALIZABLE — один на всю транзакцию, при первом операторе.
 * Операция, проснувшаяся после ожидания блокировки, продолжает со своим старым
 * снимком: это всё ещё тот же оператор.
 */
function stmtSnapshot(ctx: Ctx, t: Txn, resuming: boolean): Snapshot {
  const w = ctx.world
  const perStatement = t.isolation === 'read-committed' || t.isolation === 'read-uncommitted'
  if (t.snapshot && (resuming || !perStatement)) return t.snapshot
  t.snapshot = takeSnapshot(w, t.xid)
  t.holdsSnapshot = t.isolation !== 'read-uncommitted'
  emit(ctx, 'snap.take', { txn: [t.spec] }, {
    txn: labelOf(t),
    xmin: t.snapshot.xmin,
    xmax: t.snapshot.xmax,
    xip: t.snapshot.xip.map((x) => xidLabel(w, x)).join(', '),
    perStatement,
    isolation: t.isolation,
  })
  return t.snapshot
}

type OpResult = 'done' | 'wait' | 'abort'

function execOp(ctx: Ctx, t: Txn, resuming: boolean): void {
  const w = ctx.world
  const spec = ctx.scenario.sessions[t.spec]!
  const raw = spec.ops[t.opIdx]
  if (!raw) return
  const op = effectiveOp(w.config.rmw, raw)

  if (t.state === 'idle') {
    t.state = 'active'
    t.startTick = w.tick
    emit(ctx, 'txn.begin', { txn: [t.spec] }, { txn: labelOf(t), isolation: t.isolation, run: t.run, retries: t.retries })
  }

  let res: OpResult
  switch (op.kind) {
    case 'read':
      res = doRead(ctx, t, op, resuming)
      break
    case 'scan':
      res = doScan(ctx, t, op.min ?? Number.NEGATIVE_INFINITY)
      break
    case 'update':
      res = doUpdate(ctx, t, op.key, op.set, op.when, resuming)
      break
    case 'insert':
      res = doInsert(ctx, t, op.key, op.value)
      break
    case 'delete':
      res = doDelete(ctx, t, op.key, resuming)
      break
    case 'commit':
      res = doCommit(ctx, t)
      break
    case 'rollback':
      res = doRollback(ctx, t)
      break
  }

  if (res === 'wait') {
    w.marks[t.spec] = { kind: 'wait' }
    return
  }
  if (res === 'abort') {
    w.marks[t.spec] = { kind: 'abort' }
    return
  }

  w.marks[t.spec] = { kind: 'op', op: op.kind, key: 'key' in op ? op.key : undefined, bad: false }
  if (t.state === 'waiting') {
    t.state = 'active'
    emit(ctx, 'lock.granted', { txn: [t.spec], key: t.wait ? [t.wait.key] : [] }, {
      txn: labelOf(t),
      key: t.wait?.key,
      waited: t.wait ? w.tick - t.wait.since : 0,
    })
    t.wait = null
  }
  // У READ COMMITTED снимок живёт, пока идёт оператор, — горизонт очистки он держит только это время.
  if (t.isolation === 'read-committed' || t.isolation === 'read-uncommitted') t.holdsSnapshot = false

  if (op.kind === 'commit' || op.kind === 'rollback') return
  t.opIdx++
  const next = spec.ops[t.opIdx]
  t.nextAt = w.tick + Math.max(1, next?.after ?? 1)
}

/* ──────────────────────────── блокировки ──────────────────────────── */

type Acquire = { status: 'ok'; tuple: Tuple } | { status: 'wait' } | { status: 'abort' } | { status: 'gone' }

function startWait(ctx: Ctx, t: Txn, holder: Txn, key: string, tuple: Tuple, behind: boolean): void {
  const w = ctx.world
  const on = tagOf(holder)
  if (t.state === 'waiting' && t.wait) {
    // Уже ждём эту строку — просто сменился тот, кого ждём: очередь продвинулась.
    if (t.wait.on !== on) {
      t.wait = { ...t.wait, on, xid: holder.xid, behind, timerFrom: w.tick, checked: false }
    }
    return
  }
  t.state = 'waiting'
  t.wait = { on, xid: holder.xid, key, since: w.tick, timerFrom: w.tick, behind, checked: false }
  w.stats.waits++
  emit(ctx, 'lock.wait', { txn: [t.spec, holder.spec], tuple: [tuple.id], key: [key] }, {
    txn: labelOf(t),
    holder: labelOf(holder),
    xid: holder.xid,
    key,
    lockOnly: tuple.lockOnly && !behind,
    behind,
  })
}

/** Кто уже стоит в очереди к этой строке раньше нас. Очередь честная: пришедший позже не проскочит вперёд. */
function aheadOf(ctx: Ctx, t: Txn, key: string): Txn | undefined {
  const mine = t.state === 'waiting' && t.wait?.key === key ? t.wait.since : Number.POSITIVE_INFINITY
  return ctx.world.txns
    .filter((o) => o !== t && o.state === 'waiting' && o.wait?.key === key)
    .filter((o) => o.wait!.since < mine || (o.wait!.since === mine && o.spec < t.spec))
    .sort((a, b) => a.wait!.since - b.wait!.since || a.spec - b.spec)[0]
}

function waitOnXid(ctx: Ctx, t: Txn, xid: number, key: string, tuple: Tuple): void {
  const holder = txnByXid(ctx.world, xid)
  if (!holder) throw new Error(`нет транзакции с xid ${xid}`)
  startWait(ctx, t, holder, key, tuple, false)
}

/**
 * Взять строку для изменения — heap_update/heap_lock_tuple и их ожидание.
 *
 * Строку, которую меняет другая идущая транзакция, выдаёт её xmax: ждём, пока
 * та закончится. Если она закоммитилась, READ COMMITTED идёт по цепочке версий к
 * самой свежей и перепроверяет её (EvalPlanQual), а REPEATABLE READ и
 * SERIALIZABLE сдаются: их снимок эту версию не видит, а менять невидимое нельзя.
 */
function acquire(ctx: Ctx, t: Txn, start: Tuple): Acquire {
  const w = ctx.world
  // Перепроверки копятся и попадают в ленту, только если строку в итоге взяли:
  // пока транзакция стоит в очереди, она каждый тик заново проходит ту же цепочку.
  const rechecks: { from: Tuple; to: Tuple; by: Txn | undefined }[] = []
  const flush = () => {
    for (const r of rechecks) {
      emit(ctx, 'row.recheck', { txn: r.by ? [t.spec, r.by.spec] : [t.spec], tuple: [r.from.id, r.to.id], key: [r.from.key] }, {
        txn: labelOf(t),
        key: r.from.key,
        by: r.by ? labelOf(r.by) : String(r.from.xmax),
        from: r.from.value,
        to: r.to.value,
      })
    }
  }
  let cur = start
  for (let guard = 0; guard < 64; guard++) {
    // Учебный READ UNCOMMITTED мог прочитать версию, чей создатель ещё идёт. Менять её до его конца нельзя.
    if (cur.xmin !== t.xid && statusOf(w, cur.xmin) === 'in-progress') {
      waitOnXid(ctx, t, cur.xmin, cur.key, cur)
      return { status: 'wait' }
    }
    if (t.xid !== null && cur.xmax === t.xid) {
      flush()
      return { status: 'ok', tuple: cur }
    }
    const st = cur.xmax === null ? null : statusOf(w, cur.xmax)
    if (st === 'in-progress') {
      waitOnXid(ctx, t, cur.xmax!, cur.key, cur)
      return { status: 'wait' }
    }
    if (st === null || st === 'aborted' || cur.lockOnly) {
      // Строка свободна, но к ней уже стоит очередь — встаём в хвост.
      const ahead = aheadOf(ctx, t, cur.key)
      if (ahead) {
        startWait(ctx, t, ahead, cur.key, cur, true)
        return { status: 'wait' }
      }
      flush()
      return { status: 'ok', tuple: cur }
    }

    // Версию заменила закоммиченная транзакция.
    const by = txnByXid(w, cur.xmax!)
    if (t.isolation === 'read-committed' || t.isolation === 'read-uncommitted') {
      const next = cur.next === null ? undefined : w.tuples.find((x) => x.id === cur.next)
      if (!next) {
        flush()
        return { status: 'gone' }
      }
      rechecks.push({ from: cur, to: next, by })
      cur = next
      continue
    }
    abort(ctx, t, 'serialization', {
      key: cur.key,
      by: by ? labelOf(by) : String(cur.xmax),
      message: 'could not serialize access due to concurrent update',
    })
    return { status: 'abort' }
  }
  throw new Error('слишком длинная цепочка версий')
}

function findVisible(ctx: Ctx, t: Txn, key: string, snap: Snapshot): Tuple | undefined {
  const w = ctx.world
  return w.tuples.find((x) => x.key === key && visibleTo(w, x, snap, t.xid, t.isolation === 'read-uncommitted'))
}

/* ─────────────────────────────── SSI ─────────────────────────────── */

function addRw(ctx: Ctx, reader: Txn, writer: Txn, key: string): void {
  if (reader === writer || tagOf(reader) === tagOf(writer)) return
  if (reader.isolation !== 'serializable' || writer.isolation !== 'serializable') return
  if (reader.state === 'aborted' || writer.state === 'aborted') return
  const wt = tagOf(writer)
  const rt = tagOf(reader)
  if (reader.rwOut.includes(wt)) return
  reader.rwOut.push(wt)
  writer.rwIn.push(rt)
  emit(ctx, 'ssi.conflict', { txn: [reader.spec, writer.spec], key: key === '*' ? [] : [key] }, {
    reader: labelOf(reader),
    writer: labelOf(writer),
    key,
  })
}

/** Читатель не увидел чужую запись: для каждой невидимой ему версии ключа — rw-зависимость читатель → писатель. */
function ssiOnRead(ctx: Ctx, r: Txn, key: string | '*', snap: Snapshot): void {
  if (r.isolation !== 'serializable') return
  const w = ctx.world
  if (!r.sireads.includes(key)) r.sireads.push(key)
  for (const x of w.tuples) {
    if (key !== '*' && x.key !== key) continue
    const writers: number[] = []
    if (x.xmin !== r.xid && statusOf(w, x.xmin) !== 'aborted' && !committedBefore(w, x.xmin, snap)) writers.push(x.xmin)
    if (
      x.xmax !== null &&
      !x.lockOnly &&
      x.xmax !== r.xid &&
      statusOf(w, x.xmax) !== 'aborted' &&
      !committedBefore(w, x.xmax, snap) &&
      visibleTo(w, x, snap, r.xid)
    ) {
      writers.push(x.xmax)
    }
    for (const xid of writers) {
      const wr = txnByXid(w, xid)
      if (wr) addRw(ctx, r, wr, x.key)
    }
  }
}

/** Писатель меняет то, что уже прочитал кто-то параллельный: rw-зависимость читатель → писатель. */
function ssiOnWrite(ctx: Ctx, wr: Txn, key: string): void {
  if (wr.isolation !== 'serializable') return
  const w = ctx.world
  const since = wr.snapshot?.tick ?? w.tick
  for (const r of allTxns(w)) {
    if (r === wr || r.isolation !== 'serializable' || r.state === 'aborted') continue
    if (!r.sireads.includes(key) && !r.sireads.includes('*')) continue
    const concurrent = isLive(r) || (r.state === 'committed' && (r.endTick ?? 0) >= since)
    if (concurrent) addRw(ctx, r, wr, key)
  }
}

/* ───────────────────────────── операторы ───────────────────────────── */

function doRead(ctx: Ctx, t: Txn, op: Extract<Op, { kind: 'read' }>, resuming: boolean): OpResult {
  const w = ctx.world
  const snap = stmtSnapshot(ctx, t, resuming)
  let tup = findVisible(ctx, t, op.key, snap)

  if (tup && op.forUpdate) {
    const got = acquire(ctx, t, tup)
    if (got.status === 'wait') return 'wait'
    if (got.status === 'abort') return 'abort'
    if (got.status === 'gone') tup = undefined
    else {
      tup = got.tuple
      if (t.xid === null || tup.xmax !== t.xid) {
        const xid = assignXid(ctx, t)
        tup.xmax = xid
        tup.lockOnly = true
        tup.lsnOut = appendWal(w, { xid, kind: 'lock', key: tup.key })
      }
      emit(ctx, 'row.lock', { txn: [t.spec], tuple: [tup.id], key: [op.key] }, { txn: labelOf(t), key: op.key, value: tup.value })
    }
  }

  const value = tup ? tup.value : null
  const before = t.seen[op.key] ?? []
  const prev = before.length > 0 ? before[before.length - 1] : undefined
  t.seen[op.key] = [...before, value]
  t.reads[op.key] = value
  ssiOnRead(ctx, t, op.key, snap)

  const creator = tup ? txnByXid(w, tup.xmin) : undefined
  emit(ctx, 'row.read', { txn: [t.spec], tuple: tup ? [tup.id] : [], key: [op.key] }, {
    txn: labelOf(t),
    key: op.key,
    value,
    xmin: tup ? xidLabel(w, tup.xmin) : null,
    own: tup ? tup.xmin === t.xid : false,
    forUpdate: op.forUpdate === true,
    isolation: t.isolation,
    perStatement: t.isolation === 'read-committed',
  })

  if (tup && creator && tup.xmin !== t.xid && statusOf(w, tup.xmin) === 'in-progress') {
    anomaly(ctx, t, 'anomaly.dirty', { key: op.key, value, by: labelOf(creator) }, [creator.spec])
  }
  if (prev !== undefined && prev !== value && !t.wrote.includes(op.key)) {
    anomaly(ctx, t, 'anomaly.nonrepeatable', { key: op.key, before: prev, after: value, isolation: t.isolation })
  }
  return 'done'
}

function doScan(ctx: Ctx, t: Txn, min: number): OpResult {
  const w = ctx.world
  const snap = stmtSnapshot(ctx, t, false)
  const dirty = t.isolation === 'read-uncommitted'
  const rows = w.tuples.filter((x) => x.value >= min && visibleTo(w, x, snap, t.xid, dirty))
  const count = rows.length
  const sum = rows.reduce((s, x) => s + x.value, 0)
  const prev = [...t.scans].reverse().find((s) => s.min === min)
  t.scans.push({ count, sum, min })
  ssiOnRead(ctx, t, '*', snap)

  emit(ctx, 'row.scan', { txn: [t.spec], tuple: rows.map((x) => x.id) }, {
    txn: labelOf(t),
    min: Number.isFinite(min) ? min : null,
    count,
    sum,
    keys: rows.map((x) => x.key).join(', '),
    isolation: t.isolation,
  })
  if (prev && prev.count !== count) {
    anomaly(ctx, t, 'anomaly.phantom', { before: prev.count, after: count, min: Number.isFinite(min) ? min : null, isolation: t.isolation })
  } else if (prev && prev.sum !== sum) {
    anomaly(ctx, t, 'anomaly.nonrepeatable', { key: 'sum(v)', before: prev.sum, after: sum, isolation: t.isolation })
  }
  return 'done'
}

function evalExpr(t: Txn, key: string, e: Expr, cur: number): number {
  switch (e.kind) {
    case 'const':
      return e.value
    case 'delta':
      return cur + e.d
    case 'read':
      return (t.reads[key] ?? cur) + e.d
  }
}

function doUpdate(ctx: Ctx, t: Txn, key: string, set: Expr, when: Extract<Op, { kind: 'update' }>['when'], resuming: boolean): OpResult {
  const w = ctx.world
  if (when && !resuming) {
    const sum = when.sumOf.reduce((s, k) => s + (t.reads[k] ?? 0), 0)
    if (sum < when.gte) {
      emit(ctx, 'stmt.skip', { txn: [t.spec], key: [key] }, { txn: labelOf(t), key, why: 'cond', sum, gte: when.gte, of: when.sumOf.join(' + ') })
      return 'done'
    }
  }
  const snap = stmtSnapshot(ctx, t, resuming)
  const start = findVisible(ctx, t, key, snap)
  if (!start) {
    emit(ctx, 'stmt.skip', { txn: [t.spec], key: [key] }, { txn: labelOf(t), key, why: 'no-row' })
    return 'done'
  }
  const got = acquire(ctx, t, start)
  if (got.status === 'wait') return 'wait'
  if (got.status === 'abort') return 'abort'
  if (got.status === 'gone') {
    emit(ctx, 'stmt.skip', { txn: [t.spec], key: [key] }, { txn: labelOf(t), key, why: 'deleted' })
    return 'done'
  }
  const target = got.tuple
  const value = evalExpr(t, key, set, target.value)
  const read = t.reads[key]

  if (set.kind === 'read' && read !== undefined && read !== null && read !== target.value) {
    const by = txnByXid(w, target.xmin)
    anomaly(ctx, t, 'anomaly.lost', { key, read, actual: target.value, wrote: value, by: by ? labelOf(by) : xidLabel(w, target.xmin) }, by ? [by.spec] : [])
  }

  const xid = assignXid(ctx, t)
  const lsn = appendWal(w, { xid, kind: 'update', key })
  const { page, slot } = allocSlot(w)
  const fresh: Tuple = {
    id: w.nextTupleId++,
    key,
    value,
    xmin: xid,
    xmax: null,
    lockOnly: false,
    next: null,
    page,
    slot,
    lsnIn: lsn,
    lsnOut: null,
  }
  target.xmax = xid
  target.lockOnly = false
  target.lsnOut = lsn
  target.next = fresh.id
  w.tuples.push(fresh)
  if (!t.wrote.includes(key)) t.wrote.push(key)
  t.effects.push(set.kind === 'const' ? { key, kind: 'set', value } : { key, kind: 'add', value: set.d })
  ssiOnWrite(ctx, t, key)

  emit(ctx, 'row.update', { txn: [t.spec], tuple: [target.id, fresh.id], key: [key] }, {
    txn: labelOf(t),
    key,
    from: target.value,
    to: value,
    xid,
    expr: set.kind,
    d: set.kind === 'const' ? null : set.d,
    oldCtid: `(${target.page},${target.slot})`,
    newCtid: `(${page},${slot})`,
    samePage: target.page === page,
  })
  return 'done'
}

function doInsert(ctx: Ctx, t: Txn, key: string, value: number): OpResult {
  const w = ctx.world
  // Первичный ключ: живая или ещё не решённая версия с тем же ключом мешает вставке.
  for (const x of w.tuples) {
    if (x.key !== key || x.next !== null || statusOf(w, x.xmin) === 'aborted') continue
    const deleted = x.xmax !== null && !x.lockOnly && statusOf(w, x.xmax) !== 'aborted'
    if (!deleted) {
      if (x.xmin !== t.xid && statusOf(w, x.xmin) === 'in-progress') {
        waitOnXid(ctx, t, x.xmin, key, x)
        return 'wait'
      }
      abort(ctx, t, 'unique', { key, message: `duplicate key value violates unique constraint "t_pkey"` })
      return 'abort'
    }
    if (x.xmax !== t.xid && statusOf(w, x.xmax!) === 'in-progress') {
      waitOnXid(ctx, t, x.xmax!, key, x)
      return 'wait'
    }
  }
  stmtSnapshot(ctx, t, false)
  const xid = assignXid(ctx, t)
  const lsn = appendWal(w, { xid, kind: 'insert', key })
  const { page, slot } = allocSlot(w)
  const fresh: Tuple = { id: w.nextTupleId++, key, value, xmin: xid, xmax: null, lockOnly: false, next: null, page, slot, lsnIn: lsn, lsnOut: null }
  w.tuples.push(fresh)
  if (!t.wrote.includes(key)) t.wrote.push(key)
  t.effects.push({ key, kind: 'insert', value })
  ssiOnWrite(ctx, t, key)
  emit(ctx, 'row.insert', { txn: [t.spec], tuple: [fresh.id], key: [key] }, { txn: labelOf(t), key, value, xid, ctid: `(${page},${slot})` })
  return 'done'
}

function doDelete(ctx: Ctx, t: Txn, key: string, resuming: boolean): OpResult {
  const w = ctx.world
  const snap = stmtSnapshot(ctx, t, resuming)
  const start = findVisible(ctx, t, key, snap)
  if (!start) {
    emit(ctx, 'stmt.skip', { txn: [t.spec], key: [key] }, { txn: labelOf(t), key, why: 'no-row' })
    return 'done'
  }
  const got = acquire(ctx, t, start)
  if (got.status === 'wait') return 'wait'
  if (got.status === 'abort') return 'abort'
  if (got.status === 'gone') {
    emit(ctx, 'stmt.skip', { txn: [t.spec], key: [key] }, { txn: labelOf(t), key, why: 'deleted' })
    return 'done'
  }
  const target = got.tuple
  const xid = assignXid(ctx, t)
  target.xmax = xid
  target.lockOnly = false
  target.lsnOut = appendWal(w, { xid, kind: 'delete', key })
  if (!t.wrote.includes(key)) t.wrote.push(key)
  t.effects.push({ key, kind: 'delete', value: 0 })
  ssiOnWrite(ctx, t, key)
  emit(ctx, 'row.delete', { txn: [t.spec], tuple: [target.id], key: [key] }, { txn: labelOf(t), key, value: target.value, xid })
  return 'done'
}

function doCommit(ctx: Ctx, t: Txn): OpResult {
  const w = ctx.world
  // SSI: транзакция — «шарнир» опасной структуры, и та, кого она не увидела, уже закоммичена.
  if (t.isolation === 'serializable' && t.rwIn.length > 0) {
    // Коммитящаяся считается закоммиченной: её проверка уже пройдена, назад она не повернёт.
    const out = t.rwOut.map((tag) => txnByTag(w, tag)).find((x) => x && (x.state === 'committed' || x.state === 'committing'))
    if (out) {
      const inTxn = txnByTag(w, t.rwIn[0]!)
      abort(ctx, t, 'ssi', {
        out: labelOf(out),
        in: inTxn ? labelOf(inTxn) : '?',
        message: 'could not serialize access due to read/write dependencies among transactions',
      })
      return 'abort'
    }
  }

  if (t.xid === null) {
    // Ничего не писала — и в WAL писать нечего: коммит только-читающей транзакции бесплатен.
    finalizeCommit(ctx, t, true)
    return 'done'
  }
  t.commitLsn = appendWal(w, { xid: t.xid, kind: 'commit' })
  if (w.config.syncCommit) {
    t.state = 'committing'
    emit(ctx, 'commit.wait', { txn: [t.spec] }, { txn: labelOf(t), lsn: t.commitLsn, flushed: w.flushedLsn, flushing: w.flushing?.upTo ?? null })
    return 'done'
  }
  finalizeCommit(ctx, t, false)
  return 'done'
}

function finalizeCommit(ctx: Ctx, t: Txn, readOnly: boolean): void {
  const w = ctx.world
  if (t.xid !== null) w.xact[t.xid] = 'committed'
  t.state = 'committed'
  t.endTick = w.tick
  t.acked = true
  t.ackTick = w.tick
  t.holdsSnapshot = false
  t.wait = null
  w.stats.commits++
  for (const e of t.effects) {
    const cur = w.expected[e.key] ?? 0
    if (e.kind === 'set' || e.kind === 'insert') w.expected[e.key] = e.value
    else if (e.kind === 'add') w.expected[e.key] = cur + e.value
    else w.expected[e.key] = null
  }
  const flushed = t.commitLsn === null || t.commitLsn <= w.flushedLsn
  emit(ctx, 'txn.commit', { txn: [t.spec], key: t.wrote }, {
    txn: labelOf(t),
    xid: t.xid,
    readOnly,
    sync: w.config.syncCommit,
    flushed,
    lsn: t.commitLsn,
    waited: w.config.syncCommit && !readOnly ? w.tick - (w.wal.find((r) => r.lsn === t.commitLsn)?.tick ?? w.tick) : 0,
    took: w.tick - t.startTick,
  })
  scheduleNext(ctx, t, false)
  checkInvariant(ctx, t)
}

function doRollback(ctx: Ctx, t: Txn): OpResult {
  const w = ctx.world
  if (t.xid !== null) {
    w.xact[t.xid] = 'aborted'
    appendWal(w, { xid: t.xid, kind: 'abort' })
  }
  t.state = 'aborted'
  t.endTick = w.tick
  t.holdsSnapshot = false
  w.stats.rollbacks++
  emit(ctx, 'txn.rollback', { txn: [t.spec], key: t.wrote }, { txn: labelOf(t), xid: t.xid, wrote: t.wrote.join(', ') })
  scheduleNext(ctx, t, false)
  return 'done'
}

/** Ошибка: транзакция прервана, всё, что она успела записать, стало мёртвыми версиями. */
function abort(ctx: Ctx, t: Txn, reason: string, payload: Record<string, unknown>): void {
  const w = ctx.world
  if (t.xid !== null) {
    w.xact[t.xid] = 'aborted'
    // После падения писать некуда: транзакцию откатит само восстановление.
    if (reason !== 'crash') appendWal(w, { xid: t.xid, kind: 'abort' })
  }
  t.state = 'aborted'
  t.endTick = w.tick
  t.error = reason
  t.holdsSnapshot = false
  t.wait = null
  w.stats.aborts++
  if (reason === 'serialization' || reason === 'ssi') w.stats.serializationFailures++
  const willRetry = w.config.retry && RETRYABLE.has(reason) && t.retries < MAX_RETRIES
  emit(ctx, 'txn.abort', { txn: [t.spec], key: t.wrote }, {
    txn: labelOf(t),
    xid: t.xid,
    reason,
    isolation: t.isolation,
    willRetry,
    wrote: t.wrote.join(', '),
    ...payload,
  })
  scheduleNext(ctx, t, willRetry)
}

function scheduleNext(ctx: Ctx, t: Txn, retry: boolean): void {
  const w = ctx.world
  const spec = ctx.scenario.sessions[t.spec]!
  if (retry) {
    t.retryNext = true
    t.restartAt = Math.max(w.tick + 1, w.downUntil)
  } else if (t.repeatsLeft > 0) {
    t.retryNext = false
    t.restartAt = Math.max(w.tick + Math.max(1, spec.every ?? 1), w.downUntil)
  } else {
    t.restartAt = null
  }
}

function anomaly(ctx: Ctx, t: Txn, type: TxnEventType, payload: Record<string, unknown>, others: number[] = []): void {
  ctx.world.stats.anomalies++
  const key = typeof payload.key === 'string' && payload.key !== 'sum(v)' ? [payload.key] : []
  emit(ctx, type, { txn: [t.spec, ...others], key }, { txn: labelOf(t), ...payload })
}

/* ─────────────────────────────── инвариант ─────────────────────────────── */

export function invariantState(w: TxnWorld, inv: Invariant | undefined): { ok: boolean; detail: string } | null {
  if (!inv) return null
  if (inv.kind === 'min-sum') {
    const vals = inv.keys.map((k) => committedValue(w, k) ?? 0)
    const sum = vals.reduce((a, b) => a + b, 0)
    return { ok: sum >= inv.min, detail: `${inv.keys.join(' + ')} = ${sum}, нужно ≥ ${inv.min}` }
  }
  const bad = inv.keys
    .map((k) => ({ k, have: committedValue(w, k), want: w.expected[k] ?? null }))
    .filter((x) => x.have !== x.want)
  if (bad.length === 0) return { ok: true, detail: inv.keys.map((k) => `${k} = ${committedValue(w, k) ?? '—'}`).join(', ') }
  return {
    ok: false,
    detail: bad.map((x) => `${x.k} = ${x.have ?? '—'}, а подтверждено ${x.want ?? '—'}`).join('; '),
  }
}

function checkInvariant(ctx: Ctx, t: Txn | null): void {
  const w = ctx.world
  const st = invariantState(w, ctx.scenario.invariant)
  if (!st) return
  if (!st.ok && !w.invariantBroken) {
    w.invariantBroken = true
    w.stats.anomalies++
    emit(ctx, 'invariant.broken', { txn: t ? [t.spec] : [] }, {
      txn: t ? labelOf(t) : null,
      label: ctx.scenario.invariant!.label,
      kind: ctx.scenario.invariant!.kind,
      detail: st.detail,
    })
  } else if (st.ok) {
    w.invariantBroken = false
  }
}

/* ─────────────────────────────── WAL ─────────────────────────────── */

/**
 * Сброс закончился: всё до upTo на диске. Ждущие коммиты подтверждаются разом —
 * это групповой коммит. Их запись могла попасть на диск и с контрольной точкой:
 * такие подтверждаются здесь же, без отдельного сброса.
 */
function finishFlush(ctx: Ctx): void {
  const w = ctx.world
  let upTo: number | null = null
  if (w.flushing && w.tick >= w.flushing.doneAt) {
    upTo = w.flushing.upTo
    w.flushedLsn = Math.max(w.flushedLsn, upTo)
    w.flushing = null
    w.lastFlushTick = w.tick
    w.stats.flushes++
  }
  const ready = w.txns.filter((t) => t.state === 'committing' && t.commitLsn !== null && t.commitLsn <= w.flushedLsn)
  if (upTo !== null) {
    emit(ctx, 'wal.flush', { txn: ready.map((t) => t.spec) }, {
      upTo,
      commits: ready.length,
      names: ready.map(labelOf).join(', '),
      background: ready.length === 0,
    })
  }
  for (const t of ready) finalizeCommit(ctx, t, false)
}

/**
 * Начать сброс. Коммит, которому нужен диск, запускает его сразу, если диск
 * свободен; иначе ждёт и уедет следующим сбросом вместе со всеми, кто успел
 * подойти. Без ждущих коммитов WAL сбрасывает фоновый walwriter по таймеру.
 */
function startFlush(ctx: Ctx): void {
  const w = ctx.world
  if (w.flushing) return
  const last = w.nextLsn - 1
  if (last <= w.flushedLsn) return
  const needed = w.txns.some((t) => t.state === 'committing')
  const timer = w.tick - w.lastFlushTick >= w.config.walWriterDelay
  if (!needed && !timer) return
  w.flushing = { upTo: last, doneAt: w.tick + Math.max(1, w.config.fsyncTicks) }
}

function checkpoint(ctx: Ctx): void {
  const w = ctx.world
  const lsn = appendWal(w, { xid: null, kind: 'checkpoint' })
  const replayWas = w.wal.filter((r) => r.lsn > w.checkpointLsn).length
  w.checkpointLsn = lsn
  w.flushedLsn = Math.max(w.flushedLsn, lsn)
  emit(ctx, 'checkpoint', {}, { lsn, pages: pagesOf(w), replayWas })
}

/* ────────────────────────────── дедлоки ────────────────────────────── */

/**
 * Поиск дедлока — DeadLockCheck. Каждая ждущая транзакция делает его один раз,
 * через deadlock_timeout после начала ожидания. Нашла цикл — прерывает себя.
 */
function deadlockCheck(ctx: Ctx): void {
  const w = ctx.world
  for (const t of w.txns) {
    if (t.state !== 'waiting' || !t.wait || t.wait.checked) continue
    if (w.tick - t.wait.timerFrom < w.config.deadlockTimeout) continue
    t.wait.checked = true
    const cycle: Txn[] = [t]
    let cur: Txn | undefined = txnByTag(w, t.wait.on)
    while (cur && cur.state === 'waiting' && cur.wait && cur !== t && cycle.length < 16) {
      cycle.push(cur)
      cur = txnByTag(w, cur.wait.on)
    }
    if (cur !== t) continue
    w.stats.deadlocks++
    emit(ctx, 'deadlock.found', { txn: cycle.map((x) => x.spec), key: cycle.map((x) => x.wait!.key) }, {
      txn: labelOf(t),
      cycle: cycle.map((x) => `${labelOf(x)} ждёт ${x.wait!.key}`).join(' → '),
      size: cycle.length,
      waited: w.tick - t.wait.timerFrom,
    })
    abort(ctx, t, 'deadlock', { message: 'deadlock detected', key: t.wait?.key })
    w.marks[t.spec] = { kind: 'abort' }
  }
}

/* ────────────────────────────── очистка ────────────────────────────── */

/**
 * Можно ли убрать мёртвую версию прямо сейчас. Мало умереть раньше горизонта:
 * коммит, который её убил, должен уже лежать в WAL на диске. В настоящем
 * PostgreSQL очистка сама пишется в WAL после этого коммита, и падение откатило
 * бы её вместе с ним; модель просто не трогает такие версии до сброса.
 */
function removableNow(w: TxnWorld, x: Tuple, hz: number): boolean {
  if (statusOf(w, x.xmin) === 'aborted') return true
  if (x.xmax! >= hz) return false
  const killer = txnByXid(w, x.xmax!)
  return !killer || killer.commitLsn === null || killer.commitLsn <= w.flushedLsn
}

export function vacuum(ctx: Ctx): void {
  const w = ctx.world
  const hz = horizon(w)
  const removable = new Set<number>()
  let recentlyDead = 0
  for (const x of w.tuples) {
    if (!isDead(w, x)) continue
    if (removableNow(w, x, hz.xid)) removable.add(x.id)
    else recentlyDead++
  }
  w.tuples = w.tuples.filter((x) => !removable.has(x.id))
  for (const x of w.tuples) if (x.next !== null && removable.has(x.next)) x.next = null
  w.stats.vacuumRuns++
  w.stats.vacuumRemoved += removable.size
  emit(ctx, 'vacuum.run', { txn: hz.holder && recentlyDead > 0 ? [hz.holder.spec] : [] }, {
    removed: removable.size,
    kept: recentlyDead,
    horizon: hz.xid,
    holder: hz.holder ? labelOf(hz.holder) : null,
    live: w.tuples.length - recentlyDead,
    pages: pagesOf(w),
  })
  if (recentlyDead > 0 && hz.holder) {
    emit(ctx, 'vacuum.blocked', { txn: [hz.holder.spec] }, {
      kept: recentlyDead,
      holder: labelOf(hz.holder),
      horizon: hz.xid,
      isolation: hz.holder.isolation,
      since: hz.holder.startTick,
      age: w.tick - hz.holder.startTick,
      idle: hz.holder.opIdx >= (ctx.scenario.sessions[hz.holder.spec]?.ops.length ?? 0),
    })
  }
}

/* ─────────────────────────── падение сервера ─────────────────────────── */

/**
 * Сервер упал. Выживает только то, что успело попасть в WAL на диске: страницы
 * данных в памяти пропали, а на диск они не могли попасть раньше своих записей
 * WAL. Всё, что после flushedLsn, как будто не происходило.
 */
function crash(ctx: Ctx): void {
  const w = ctx.world
  const flushed = w.flushedLsn
  const lostRecords = w.wal.filter((r) => r.lsn > flushed).length
  const live = w.txns.filter(isLive)
  w.crashed = true
  w.flushing = null
  const replay = w.wal.filter((r) => r.lsn > w.checkpointLsn && r.lsn <= flushed).length
  w.downUntil = w.tick + 1 + Math.ceil(replay / Math.max(1, w.config.replayRate))
  emit(ctx, 'db.crash', { txn: live.map((t) => t.spec) }, {
    flushed,
    lostRecords,
    active: live.length,
    replay,
    down: w.downUntil - w.tick,
    sync: w.config.syncCommit,
  })

  const survived = new Set(w.wal.filter((r) => r.kind === 'commit' && r.lsn <= flushed).map((r) => r.xid))
  w.wal = w.wal.filter((r) => r.lsn <= flushed)

  for (const t of allTxns(w)) {
    if (t.state === 'committed' && t.xid !== null && !survived.has(t.xid)) {
      t.lost = true
      w.xact[t.xid] = 'aborted'
      w.stats.lostCommits++
      emit(ctx, 'commit.lost', { txn: [t.spec], key: t.wrote }, {
        txn: labelOf(t),
        xid: t.xid,
        ackTick: t.ackTick,
        lsn: t.commitLsn,
        flushed,
        wrote: t.wrote.join(', '),
      })
    }
  }
  for (const t of live) {
    abort(ctx, t, 'crash', { message: 'server closed the connection unexpectedly' })
  }
  for (const [xid, st] of Object.entries(w.xact)) {
    if (st === 'in-progress') w.xact[Number(xid)] = 'aborted'
  }

  w.tuples = w.tuples.filter((x) => x.lsnIn <= flushed)
  const ids = new Set(w.tuples.map((x) => x.id))
  for (const x of w.tuples) {
    if (x.lsnOut !== null && x.lsnOut > flushed) {
      x.xmax = null
      x.lockOnly = false
      x.lsnOut = null
    }
    if (x.next !== null && !ids.has(x.next)) x.next = null
  }
  for (const t of w.txns) {
    if (t.state === 'idle') t.nextAt = Math.max(t.nextAt, w.downUntil)
    if (t.restartAt !== null) t.restartAt = Math.max(t.restartAt, w.downUntil)
  }
  w.checkpointLsn = flushed
  w.lastFlushTick = w.downUntil
  w.marks = w.marks.map(() => ({ kind: 'down' }))
}

function recovered(ctx: Ctx): void {
  const w = ctx.world
  emit(ctx, 'db.recovered', {}, {
    lost: w.stats.lostCommits,
    rows: keysOf(w).map((k) => `${k} = ${committedValue(w, k) ?? '—'}`).join(', '),
  })
  checkInvariant(ctx, null)
}

/* ─────────────────────────────── итог тика ─────────────────────────────── */

function account(ctx: Ctx): void {
  const w = ctx.world
  for (const t of w.txns) {
    if (t.state === 'waiting') w.stats.waitTicks++
    if (t.state === 'committing') w.stats.commitWaitTicks++
  }
  // Операция, на которой случилась аномалия, помечается в расписании.
  for (const e of ctx.events) {
    if (!e.type.startsWith('anomaly.') && e.type !== 'invariant.broken') continue
    const s = e.actors.txn?.[0]
    const m = s === undefined ? null : w.marks[s]
    if (m && m.kind === 'op') m.bad = true
  }
  const dead = w.tuples.filter((x) => isDead(w, x)).length
  if (w.tuples.length > 0) w.stats.maxDead = Math.max(w.stats.maxDead, dead / w.tuples.length)

  const settled = w.txns.every(
    (t) => t.state === 'done' || ((t.state === 'committed' || t.state === 'aborted') && t.restartAt === null),
  )
  if (settled && !w.flushing && w.tick >= w.downUntil && w.tick >= (ctx.scenario.minTicks ?? 0)) {
    w.finished = true
    w.finishReason = 'all-done'
    return
  }
  const limit = ctx.scenario.stopAfter
  if (limit && w.tick >= limit) {
    w.finished = true
    w.finishReason = 'stop-after'
  }
}

/** Для UI: сколько версий мертво и сколько из них уже можно убрать. */
export function heapStats(w: TxnWorld): { total: number; dead: number; removable: number; pages: number } {
  const hz = horizon(w)
  let dead = 0
  let removable = 0
  for (const x of w.tuples) {
    if (!isDead(w, x)) continue
    dead++
    if (removableNow(w, x, hz.xid)) removable++
  }
  return { total: w.tuples.length, dead, removable, pages: pagesOf(w) }
}
