import type { Client, ClientOp, Mark, NodeState, Query, ReplWorld } from './types.ts'
import type { Ctx } from './world.ts'
import { emit, lagOf, lastLsn, primaryOf, recordAt, replicasOf, retainedOf } from './world.ts'

/**
 * Один шаг всей системы.
 *
 * Порядок фиксирован: сбои по расписанию, переключение на реплику, сеть
 * доставляет сообщения, реплики сбрасывают и проигрывают WAL, ведущий
 * сбрасывает свой WAL и досылает новое, клиенты делают следующий шаг, в конце —
 * вакуум, хранение WAL и итог тика.
 */
export function tick(ctx: Ctx): void {
  const w = ctx.world
  if (w.finished) return
  w.tick++
  w.marks = w.clients.map(() => null)

  faults(ctx)
  failover(ctx)
  replicaFlush(ctx)
  deliver(ctx)
  replay(ctx)
  replies(ctx)
  primaryFlush(ctx)
  send(ctx)
  for (const c of w.clients) stepClient(ctx, c)
  if (w.config.vacuumEvery > 0 && w.tick % w.config.vacuumEvery === 0) vacuum(ctx)
  retention(ctx)
  account(ctx)
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

function faults(ctx: Ctx): void {
  const w = ctx.world
  for (const f of ctx.scenario.faults) {
    if (f.kind === 'primary-down' && f.at === w.tick) {
      const p = primaryOf(w)
      if (!p.up) continue
      p.up = false
      w.primaryDownSince = w.tick
      // То, что ещё не ушло с упавшей машины, не уйдёт никогда.
      const inflight = new Set(w.net.filter((m) => m.from === p.id && m.kind === 'wal').map((m) => (m.kind === 'wal' ? m.lsn : 0))).size
      w.net = w.net.filter((m) => m.from !== p.id)
      const waiting = w.clients.filter((c) => c.state === 'commit-wait' && c.waitNode === p.id)
      emit(ctx, 'primary.down', { node: [p.id], client: waiting.map((c) => c.id) }, {
        node: p.name,
        waiting: waiting.length,
        inflight,
        failover: w.config.failover,
        detect: w.config.detectAfter,
      })
      // Клиент, ждавший подтверждения коммита, получил обрыв соединения.
      // Закоммичено или нет — он не знает: запись в WAL уже была.
      for (const c of waiting) {
        w.stats.unknown++
        emit(ctx, 'commit.unknown', { client: [c.id], lsn: [c.waitLsn!] }, { client: c.name, lsn: c.waitLsn, waited: w.tick - (c.waitSince ?? w.tick) })
        c.errors++
        finishOp(ctx, c, 1)
        w.marks[c.id] = { kind: 'error' }
      }
    }
    if (f.kind === 'replica-down') {
      const n = w.nodes[f.replica]
      if (!n || n.id === w.primary) continue
      if (f.at === w.tick && n.up) {
        n.up = false
        n.queries = []
        n.conflictSince = null
        emit(ctx, 'replica.down', { node: [n.id] }, { node: n.name, until: f.until ?? null, sync: isSyncCandidate(w, n) })
      }
      if (f.until !== undefined && f.until === w.tick && !n.up) {
        n.up = true
        const v = w.view[n.id]
        if (v) v.sent = n.writeLsn
        emit(ctx, 'replica.up', { node: [n.id] }, { node: n.name, from: n.writeLsn, behind: lastLsn(w) - n.writeLsn, oldest: w.walOldest })
      }
    }
    if (f.kind === 'replica-slow') {
      const n = w.nodes[f.replica]
      if (!n || n.id === w.primary) continue
      if (f.at === w.tick) {
        n.slowUntil = f.until
        n.slowRate = f.rate
        emit(ctx, 'replica.slow', { node: [n.id] }, { node: n.name, rate: f.rate, normal: n.replayRate, until: f.until })
      }
    }
  }
}

/**
 * Переключение: менеджер кластера заметил смерть ведущего и делает ведущим
 * самую свежую из живых реплик. Всё, что она не успела получить, пропало.
 */
function failover(ctx: Ctx): void {
  const w = ctx.world
  if (w.primaryDownSince === null || !w.config.failover) return
  if (w.tick < w.primaryDownSince + w.config.detectAfter) return
  const cands = replicasOf(w).filter((n) => n.up && !n.broken)
  if (cands.length === 0) return
  const best = cands.sort((a, b) => b.flushLsn - a.flushLsn || a.id - b.id)[0]!
  const old = primaryOf(w)
  const end = best.flushLsn

  // Реплика проигрывает всё, что успела получить, и становится ведущей.
  for (let lsn = best.replayLsn + 1; lsn <= end; lsn++) applyRecord(w, best, lsn)
  best.replayLsn = end
  best.writeLsn = end
  best.role = 'primary'
  best.promotedAt = w.tick
  best.queries = []
  best.conflictSince = null
  old.role = 'replica'
  old.broken = true

  const lostRecs = w.wal.filter((r) => r.lsn > end)
  w.wal = w.wal.filter((r) => r.lsn <= end)
  w.nextLsn = end + 1
  w.walOldest = Math.min(w.walOldest, end + 1)
  w.net = []
  w.dead = w.dead.filter((l) => l <= end)
  const lost = w.ackedLsns.filter((a) => a.lsn > end && !a.lost)
  for (const a of lost) a.lost = true
  w.stats.lost += lost.length
  w.primary = best.id
  const downFor = w.tick - w.primaryDownSince
  w.primaryDownSince = null

  // Остальные реплики теперь следуют за новым ведущим.
  w.view = {}
  for (const n of w.nodes) {
    if (n.id === best.id) continue
    n.writeLsn = Math.min(n.writeLsn, end)
    n.flushLsn = Math.min(n.flushLsn, end)
    w.view[n.id] = { sent: n.writeLsn, write: n.writeLsn, flush: n.flushLsn, replay: n.replayLsn, feedback: null, lastReply: w.tick }
  }
  emit(ctx, 'failover.promote', { node: [best.id, old.id], client: lost.map((a) => a.client) }, {
    node: best.name,
    old: old.name,
    end,
    discarded: lostRecs.filter((r) => r.kind === 'commit').length,
    lost: lost.length,
    downFor,
  })
  for (const a of lost) {
    const c = w.clients[a.client]!
    emit(ctx, 'commit.lost', { client: [a.client], lsn: [a.lsn] }, { client: c.name, lsn: a.lsn, key: a.key, end, node: best.name })
  }
}

/* ─────────────────────────────── реплики ─────────────────────────────── */

const replicaReady = (w: ReplWorld, n: NodeState) => n.id !== w.primary && n.up && !n.broken

/** walreceiver сбрасывает на диск то, что записал за прошлый тик. */
function replicaFlush(ctx: Ctx): void {
  const w = ctx.world
  for (const n of w.nodes) if (replicaReady(w, n)) n.flushLsn = n.writeLsn
}

function deliver(ctx: Ctx): void {
  const w = ctx.world
  const now = w.net.filter((m) => m.arriveAt <= w.tick)
  w.net = w.net.filter((m) => m.arriveAt > w.tick)
  const got = new Map<number, { from: number; to: number }>()
  for (const m of now) {
    const to = w.nodes[m.to]
    if (!to || !to.up) continue
    if (m.kind === 'wal') {
      if (m.from !== w.primary || !replicaReady(w, to)) continue
      if (m.lsn !== to.writeLsn + 1) continue
      to.writeLsn = m.lsn
      const g = got.get(to.id)
      got.set(to.id, { from: g?.from ?? m.lsn, to: m.lsn })
    } else {
      if (m.to !== w.primary) continue
      const v = w.view[m.from]
      if (!v) continue
      v.write = Math.max(v.write, m.write)
      v.flush = Math.max(v.flush, m.flush)
      v.replay = Math.max(v.replay, m.replay)
      v.feedback = m.feedback
      v.lastReply = w.tick
    }
  }
  for (const [id, r] of got) {
    emit(ctx, 'replica.receive', { node: [id] }, { node: w.nodes[id]!.name, from: r.from, to: r.to, latency: w.nodes[id]!.latency })
  }
}

function applyRecord(w: ReplWorld, n: NodeState, lsn: number): void {
  const r = recordAt(w, lsn)
  if (r && r.kind === 'commit' && r.key !== undefined) n.kv[r.key] = { value: r.value ?? 0, lsn }
}

/**
 * Проигрывание WAL — startup-процесс реплики. Запись очистки, которая удаляет
 * версии, нужные идущему запросу, останавливает проигрывание: реплика ждёт
 * запрос до max_standby_streaming_delay, потом отменяет его.
 */
function replay(ctx: Ctx): void {
  const w = ctx.world
  for (const n of w.nodes) {
    if (!replicaReady(w, n)) continue
    // Кончившиеся запросы освобождают снимок.
    const rate = w.tick < n.slowUntil ? n.slowRate : n.replayRate
    let applied = 0
    const from = n.replayLsn + 1
    while (applied < rate && n.replayLsn < n.flushLsn) {
      const lsn = n.replayLsn + 1
      const r = recordAt(w, lsn)
      if (r && r.kind === 'cleanup') {
        const blocking = n.queries.filter((q) => q.snapshot < (r.maxSuperseded ?? 0))
        if (blocking.length > 0) {
          if (n.conflictSince === null) {
            n.conflictSince = w.tick
            emit(ctx, 'conflict.wait', { node: [n.id], client: blocking.map((q) => q.client), lsn: [lsn] }, {
              node: n.name,
              lsn,
              removed: r.removed ?? 0,
              delay: w.config.maxStandbyDelay,
              clients: blocking.map((q) => w.clients[q.client]!.name).join(', '),
            })
          }
          const limit = w.config.maxStandbyDelay
          if (limit < 0 || w.tick - n.conflictSince < limit) break
          // Время вышло: запросы отменяются, проигрывание продолжается.
          for (const q of blocking) cancelQuery(ctx, n, q, lsn)
          n.conflictSince = null
        } else if (n.conflictSince !== null) {
          n.conflictSince = null
        }
      }
      applyRecord(w, n, lsn)
      n.replayLsn = lsn
      applied++
    }
    if (applied > 0) {
      emit(ctx, 'replica.replay', { node: [n.id] }, { node: n.name, from, to: n.replayLsn, rate, slow: w.tick < n.slowUntil, behind: lastLsn(w) - n.replayLsn })
    }
  }
}

function cancelQuery(ctx: Ctx, n: NodeState, q: Query, lsn: number): void {
  const w = ctx.world
  n.queries = n.queries.filter((x) => x !== q)
  const c = w.clients[q.client]!
  w.stats.cancels++
  c.errors++
  c.query = null
  emit(ctx, 'conflict.cancel', { node: [n.id], client: [c.id], lsn: [lsn] }, {
    node: n.name,
    client: c.name,
    ran: w.tick - q.startTick,
    planned: q.endTick - q.startTick,
    waited: n.conflictSince === null ? 0 : w.tick - n.conflictSince,
  })
  finishOp(ctx, c, 1)
  w.marks[c.id] = { kind: 'error' }
}

/** Каждая реплика сообщает ведущему, докуда дошла, а с hot_standby_feedback — и свой самый старый снимок. */
function replies(ctx: Ctx): void {
  const w = ctx.world
  for (const n of w.nodes) {
    if (!replicaReady(w, n)) continue
    const feedback = w.config.hotStandbyFeedback && n.queries.length > 0 ? Math.min(...n.queries.map((q) => q.snapshot)) : null
    w.net.push({
      kind: 'reply',
      from: n.id,
      to: w.primary,
      write: n.writeLsn,
      flush: n.flushLsn,
      replay: n.replayLsn,
      feedback,
      arriveAt: w.tick + Math.max(1, n.latency),
    })
  }
}

/* ─────────────────────────────── ведущий ─────────────────────────────── */

function primaryFlush(ctx: Ctx): void {
  const w = ctx.world
  const p = primaryOf(w)
  if (!p.up) return
  p.flushLsn = p.writeLsn
}

/**
 * walsender: каждой реплике — всё, что ведущий уже сбросил на диск; той, что
 * догоняет, — не больше catchupRate за тик.
 */
function send(ctx: Ctx): void {
  const w = ctx.world
  const p = primaryOf(w)
  if (!p.up) return
  const last = p.flushLsn
  for (const n of replicasOf(w)) {
    const v = w.view[n.id]
    if (!v || !n.up || n.broken) continue
    if (v.sent < last && v.sent + 1 < w.walOldest) {
      n.broken = true
      emit(ctx, 'replica.broken', { node: [n.id] }, { node: n.name, need: v.sent + 1, oldest: w.walOldest, slots: w.config.slots })
      continue
    }
    let count = 0
    while (v.sent < last && count < w.config.catchupRate) {
      v.sent++
      count++
      w.net.push({ kind: 'wal', from: w.primary, to: n.id, lsn: v.sent, arriveAt: w.tick + Math.max(1, n.latency) })
    }
  }
}

/* ─────────────────────────────── клиенты ─────────────────────────────── */

const LEVEL_FIELD = { remote_write: 'write', on: 'flush', remote_apply: 'replay' } as const

export function isSyncCandidate(w: ReplWorld, n: NodeState): boolean {
  if (w.config.standbys === 'none' || n.id === w.primary) return false
  if (w.config.standbys === 'any-1') return true
  return n.id === 1
}

/** Готов ли коммит к подтверждению — SyncRepWaitForLSN в зависимости от synchronous_commit. */
export function commitReady(w: ReplWorld, lsn: number): { ok: boolean; waitingFor: string } {
  const c = w.config
  const p = primaryOf(w)
  if (c.syncCommit === 'off') return { ok: true, waitingFor: '' }
  if (p.flushLsn < lsn) return { ok: false, waitingFor: 'fsync на ведущем' }
  if (c.syncCommit === 'local' || c.standbys === 'none') return { ok: true, waitingFor: '' }
  const field = LEVEL_FIELD[c.syncCommit]
  const cands = replicasOf(w).filter((n) => isSyncCandidate(w, n))
  const done = cands.some((n) => (w.view[n.id]?.[field] ?? 0) >= lsn)
  const names = cands.map((n) => n.name).join(c.standbys === 'any-1' ? ' или ' : ', ')
  return { ok: done, waitingFor: `${names || 'синхронной реплики'}: ${field}` }
}

function finishOp(ctx: Ctx, c: Client, delay: number): void {
  const w = ctx.world
  const spec = ctx.scenario.clients[c.id]!
  c.state = 'idle'
  c.waitLsn = null
  c.waitSince = null
  c.waitNode = null
  c.query = null
  c.opIdx++
  if (c.opIdx >= spec.ops.length) {
    if (c.repeatsLeft > 0) {
      c.repeatsLeft--
      c.iter++
      c.opIdx = 0
      c.nextAt = w.tick + Math.max(1, spec.every ?? 1)
      return
    }
    c.state = 'done'
    return
  }
  c.nextAt = w.tick + Math.max(delay, spec.ops[c.opIdx]!.after ?? 1)
}

function stepClient(ctx: Ctx, c: Client): void {
  const w = ctx.world
  const spec = ctx.scenario.clients[c.id]!
  if (c.state === 'done') return

  if (c.state === 'commit-wait') {
    tryAck(ctx, c)
    return
  }
  if (c.state === 'reading' && c.query) {
    const n = w.nodes.find((x) => x.queries.includes(c.query!))
    if (!n) {
      // Реплика упала вместе с запросом.
      readError(ctx, c, w.nodes.find((x) => x.id !== w.primary && !x.up))
      return
    }
    if (w.tick >= c.query.endTick) {
      n.queries = n.queries.filter((q) => q !== c.query)
      emit(ctx, 'query.done', { client: [c.id], node: [n.id] }, { client: c.name, node: n.name, ticks: c.query.endTick - c.query.startTick })
      finishOp(ctx, c, 1)
    }
    w.marks[c.id] = { kind: 'query', node: n.id }
    return
  }
  if (w.tick < c.nextAt) return
  const op = spec.ops[c.opIdx]
  if (!op) return
  execOp(ctx, c, op)
}

function execOp(ctx: Ctx, c: Client, op: ClientOp): void {
  const w = ctx.world
  switch (op.kind) {
    case 'write': {
      const p = primaryOf(w)
      if (!p.up) {
        w.stats.writeErrors++
        c.errors++
        emit(ctx, 'write.error', { client: [c.id], node: [p.id] }, {
          client: c.name,
          node: p.name,
          since: w.primaryDownSince,
          failover: w.config.failover,
        })
        // Приложение повторит ту же запись через пару тиков.
        c.nextAt = w.tick + 2
        w.marks[c.id] = { kind: 'error' }
        return
      }
      const prev = [...w.wal].reverse().find((r) => r.kind === 'commit' && r.key === op.key)
      const value = (prev?.value ?? p.kv[op.key]?.value ?? 0) + 1
      const lsn = w.nextLsn++
      w.wal.push({ lsn, kind: 'commit', key: op.key, value, client: c.id, tick: w.tick })
      p.writeLsn = lsn
      // Прежняя версия строки стала мёртвой — её уберёт вакуум.
      w.dead.push(lsn)
      w.stats.commits++
      c.state = 'commit-wait'
      c.waitLsn = lsn
      c.waitSince = w.tick
      c.waitNode = p.id
      emit(ctx, 'wal.write', { client: [c.id], node: [p.id], lsn: [lsn] }, { client: c.name, key: op.key, value, lsn })
      w.marks[c.id] = { kind: 'write' }
      if (!tryAck(ctx, c, true)) {
        const r = commitReady(w, lsn)
        emit(ctx, 'commit.wait', { client: [c.id], lsn: [lsn] }, {
          client: c.name,
          lsn,
          level: w.config.syncCommit,
          standbys: w.config.standbys,
          waitingFor: r.waitingFor,
        })
      }
      return
    }
    case 'read':
      return doRead(ctx, c, op)
    case 'query': {
      const n = w.nodes[op.from]
      if (!n || !n.up || n.broken || n.id === w.primary) {
        readError(ctx, c, n)
        return
      }
      const q = { client: c.id, key: op.key, snapshot: n.replayLsn, startTick: w.tick, endTick: w.tick + op.ticks }
      n.queries.push(q)
      c.query = q
      c.state = 'reading'
      emit(ctx, 'query.start', { client: [c.id], node: [n.id] }, { client: c.name, node: n.name, snapshot: q.snapshot, ticks: op.ticks, feedback: w.config.hotStandbyFeedback })
      w.marks[c.id] = { kind: 'query', node: n.id }
      return
    }
  }
}

/** Подтвердить коммит, если всё, чего он ждёт, уже случилось. */
function tryAck(ctx: Ctx, c: Client, first = false): boolean {
  const w = ctx.world
  const lsn = c.waitLsn!
  const r = commitReady(w, lsn)
  if (!r.ok) {
    w.marks[c.id] = first ? { kind: 'write' } : { kind: 'wait' }
    const waited = w.tick - (c.waitSince ?? w.tick)
    if (waited === 8) {
      emit(ctx, 'commit.hang', { client: [c.id], lsn: [lsn] }, { client: c.name, lsn, waited, waitingFor: r.waitingFor, standbys: w.config.standbys })
    }
    return false
  }
  const rec = recordAt(w, lsn)!
  const p = primaryOf(w)
  p.kv[rec.key!] = { value: rec.value ?? 0, lsn }
  p.replayLsn = Math.max(p.replayLsn, lsn)
  const waited = w.tick - (c.waitSince ?? w.tick)
  w.stats.acked++
  w.stats.maxCommitWait = Math.max(w.stats.maxCommitWait, waited)
  c.acked++
  c.ownAcked[rec.key!] = lsn
  c.seen[rec.key!] = Math.max(c.seen[rec.key!] ?? 0, lsn)
  w.ackedLsns.push({ lsn, client: c.id, key: rec.key!, lost: false })
  if (!first || waited > 0) {
    emit(ctx, 'commit.ack', { client: [c.id], lsn: [lsn] }, { client: c.name, lsn, waited, level: w.config.syncCommit, standbys: w.config.standbys })
  } else {
    emit(ctx, 'commit.ack', { client: [c.id], lsn: [lsn] }, { client: c.name, lsn, waited: 0, level: w.config.syncCommit, standbys: w.config.standbys })
  }
  if (!first) w.marks[c.id] = { kind: 'ack' }
  finishOp(ctx, c, 1)
  return true
}

function readError(ctx: Ctx, c: Client, n: NodeState | undefined): void {
  const w = ctx.world
  w.stats.readErrors++
  c.errors++
  emit(ctx, 'read.error', { client: [c.id], node: n ? [n.id] : [] }, {
    client: c.name,
    node: n?.name ?? '?',
    why: !n ? 'нет узла' : n.broken ? 'broken' : !n.up ? 'down' : 'primary',
  })
  w.marks[c.id] = { kind: 'error' }
  finishOp(ctx, c, 1)
}

function doRead(ctx: Ctx, c: Client, op: Extract<ClientOp, { kind: 'read' }>): void {
  const w = ctx.world
  let n: NodeState | undefined
  if (op.from === 'primary') n = primaryOf(w)
  else if (op.from === 'any') {
    const pool = replicasOf(w).filter((x) => x.up && !x.broken)
    n = pool.length > 0 ? pool[c.rr++ % pool.length] : undefined
  } else n = w.nodes[op.from]
  if (!n || !n.up || n.broken) {
    readError(ctx, c, n)
    return
  }
  const got = n.kv[op.key]
  const value = got?.value ?? null
  const lsn = got?.lsn ?? 0
  const own = c.ownAcked[op.key]
  const seen = c.seen[op.key] ?? 0
  const fresh = primaryOf(w).kv[op.key]?.lsn ?? 0
  w.stats.reads++
  const base = { client: c.name, node: n.name, key: op.key, value, lsn, isPrimary: n.id === w.primary, lag: lagOf(w, n) }
  let bad = false
  if (own !== undefined && lsn < own) {
    bad = true
    w.stats.ownStale++
    w.stats.staleReads++
    const ownRec = recordAt(w, own)
    emit(ctx, 'read.own-stale', { client: [c.id], node: [n.id], lsn: [own] }, { ...base, own, ownValue: ownRec?.value ?? null })
  } else if (lsn < seen) {
    bad = true
    w.stats.backwards++
    w.stats.staleReads++
    const seenRec = recordAt(w, seen)
    emit(ctx, 'read.backwards', { client: [c.id], node: [n.id], lsn: [seen] }, { ...base, seen, seenValue: seenRec?.value ?? null, prevNode: c.lastRead ? w.nodes[c.lastRead.node]?.name : null })
  } else if (lsn < fresh) {
    w.stats.staleReads++
    emit(ctx, 'read.stale', { client: [c.id], node: [n.id] }, { ...base, fresh, freshValue: primaryOf(w).kv[op.key]?.value ?? null })
  } else {
    emit(ctx, 'read.ok', { client: [c.id], node: [n.id] }, base)
  }
  c.seen[op.key] = Math.max(seen, lsn)
  c.lastRead = { key: op.key, value, node: n.id, tick: w.tick }
  w.marks[c.id] = { kind: 'read', node: n.id, bad }
  finishOp(ctx, c, 1)
}

/* ─────────────────────────── вакуум и хранение WAL ─────────────────────────── */

/**
 * Автовакуум на ведущем убирает версии, заменённые раньше горизонта, и пишет
 * об этом запись в WAL. С hot_standby_feedback горизонт учитывает снимки
 * запросов на репликах — и мёртвые версии копятся уже на ведущем.
 */
function vacuum(ctx: Ctx): void {
  const w = ctx.world
  if (!primaryOf(w).up) return
  const own = w.nextLsn
  const feedbacks = w.config.hotStandbyFeedback
    ? Object.values(w.view)
        .map((v) => v.feedback)
        .filter((f): f is number => f !== null)
    : []
  const hz = Math.min(own, ...feedbacks.map((f) => f + 1))
  const removable = w.dead.filter((l) => l < hz)
  const kept = w.dead.length - removable.length
  if (removable.length > 0) {
    w.dead = w.dead.filter((l) => l >= hz)
    const lsn = w.nextLsn++
    w.wal.push({ lsn, kind: 'cleanup', tick: w.tick, maxSuperseded: Math.max(...removable), removed: removable.length })
    primaryOf(w).writeLsn = lsn
  }
  emit(ctx, 'vacuum.run', {}, { removed: removable.length, kept, horizon: hz, feedback: feedbacks.length > 0 })
  if (kept > 0 && feedbacks.length > 0) {
    const holder = replicasOf(w).find((n) => w.view[n.id]?.feedback !== null && w.view[n.id]?.feedback !== undefined)
    emit(ctx, 'feedback.hold', { node: holder ? [holder.id] : [] }, { kept, node: holder?.name ?? '?', horizon: hz })
  }
}

/**
 * Сколько WAL хранить. Со слотами — всё, что ещё не забрала самая отстающая
 * реплика, даже выключенная. Без слотов — только wal_keep_size последних записей.
 */
function retention(ctx: Ctx): void {
  const w = ctx.world
  if (!primaryOf(w).up) return
  const last = lastLsn(w)
  const before = w.walOldest
  let oldest: number
  if (w.config.slots) {
    const needs = replicasOf(w)
      .filter((n) => !n.broken)
      .map((n) => (w.view[n.id]?.flush ?? 0) + 1)
    oldest = Math.min(last + 1, ...needs)
  } else {
    oldest = last + 1 - w.config.walKeep
  }
  w.walOldest = Math.max(w.walOldest, oldest)
  const kept = retainedOf(w)
  const prevMax = w.stats.maxRetained
  w.stats.maxRetained = Math.max(prevMax, kept)
  if (w.config.slots) {
    const threshold = w.config.walKeep
    if (kept > threshold && Math.floor((kept - 1) / threshold) > Math.floor((Math.max(1, prevMax) - 1) / threshold)) {
      const holder = replicasOf(w)
        .filter((n) => !n.broken)
        .sort((a, b) => (w.view[a.id]?.flush ?? 0) - (w.view[b.id]?.flush ?? 0))[0]
      emit(ctx, 'wal.retained', { node: holder ? [holder.id] : [] }, { kept, node: holder?.name ?? '?', up: holder?.up ?? false })
    }
  } else if (w.walOldest > before) {
    for (const n of replicasOf(w)) {
      if (n.broken) continue
      const need = (w.view[n.id]?.sent ?? 0) + 1
      if (need >= before && need < w.walOldest && need <= last) {
        emit(ctx, 'wal.removed', { node: [n.id] }, { node: n.name, need, oldest: w.walOldest, keep: w.config.walKeep, up: n.up })
      }
    }
  }
}

/* ─────────────────────────────── итог тика ─────────────────────────────── */

function account(ctx: Ctx): void {
  const w = ctx.world
  for (const c of w.clients) if (c.state === 'commit-wait') w.stats.commitWaitTicks++
  if (!primaryOf(w).up) w.stats.downTicks++
  for (const n of replicasOf(w)) if (n.up && !n.broken) w.stats.maxLag = Math.max(w.stats.maxLag, lagOf(w, n))
  w.stats.maxDead = Math.max(w.stats.maxDead, w.dead.length)
  const done = w.clients.every((c) => c.state === 'done')
  const caughtUp = replicasOf(w).every((n) => !n.up || n.broken || n.replayLsn >= lastLsn(w)) || !primaryOf(w).up
  if (done && caughtUp && w.tick >= (ctx.scenario.minTicks ?? 0)) {
    w.finished = true
    w.finishReason = 'all-done'
    return
  }
  if (ctx.scenario.stopAfter && w.tick >= ctx.scenario.stopAfter) {
    w.finished = true
    w.finishReason = 'stop-after'
  }
}

export type { Mark }
