import type {
  Snapshot,
  Tuple,
  Txn,
  TxnConfig,
  TxnEvent,
  TxnEventType,
  TxnScenario,
  TxnSpec,
  TxnWorld,
  WalRecord,
  XactStatus,
} from './types.ts'
import { DEFAULT_TXN_CONFIG } from './types.ts'

export interface Ctx {
  world: TxnWorld
  events: TxnEvent[]
  scenario: TxnScenario
}

/** xid, которым «созданы» исходные строки таблицы. Он закоммичен задолго до начала прогона. */
export const BOOT_XID = 99

export function resolveConfig(partial: Partial<TxnConfig> | undefined): TxnConfig {
  return { ...DEFAULT_TXN_CONFIG, ...(partial ?? {}) }
}

export function newTxn(
  spec: TxnSpec,
  idx: number,
  config: TxnConfig,
  at: number,
  n: { run: number; iter: number; retries: number; repeatsLeft: number },
): Txn {
  return {
    spec: idx,
    name: spec.name,
    run: n.run,
    iter: n.iter,
    looped: (spec.repeat ?? 1) > 1,
    retries: n.retries,
    isolation: spec.isolation ?? config.isolation,
    state: 'idle',
    xid: null,
    opIdx: 0,
    nextAt: at,
    snapshot: null,
    holdsSnapshot: false,
    reads: {},
    seen: {},
    scans: [],
    wait: null,
    rwOut: [],
    rwIn: [],
    sireads: [],
    effects: [],
    wrote: [],
    startTick: at,
    endTick: null,
    commitLsn: null,
    acked: false,
    ackTick: null,
    error: null,
    lost: false,
    repeatsLeft: n.repeatsLeft,
    restartAt: null,
    retryNext: false,
  }
}

export function createWorld(scenario: TxnScenario): TxnWorld {
  const config = resolveConfig(scenario.config)
  const tuples: Tuple[] = scenario.rows.map((r, i) => ({
    id: i + 1,
    key: r.key,
    value: r.value,
    xmin: BOOT_XID,
    xmax: null,
    lockOnly: false,
    next: null,
    page: Math.floor(i / config.pageSlots),
    slot: i % config.pageSlots,
    lsnIn: 0,
    lsnOut: null,
  }))
  const expected: Record<string, number | null> = {}
  for (const r of scenario.rows) expected[r.key] = r.value

  return {
    tick: 0,
    config,
    tuples,
    nextTupleId: tuples.length + 1,
    txns: scenario.sessions.map((s, i) =>
      newTxn(s, i, config, Math.max(1, s.at), { run: 0, iter: 0, retries: 0, repeatsLeft: Math.max(0, (s.repeat ?? 1) - 1) }),
    ),
    history: [],
    nextXid: BOOT_XID + 1,
    xact: { [BOOT_XID]: 'committed' },
    wal: [],
    nextLsn: 1,
    flushedLsn: 0,
    flushing: null,
    lastFlushTick: 0,
    checkpointLsn: 0,
    downUntil: 0,
    crashed: false,
    expected,
    invariantBroken: false,
    marks: scenario.sessions.map(() => null),
    stats: {
      commits: 0,
      aborts: 0,
      rollbacks: 0,
      serializationFailures: 0,
      deadlocks: 0,
      retries: 0,
      waits: 0,
      waitTicks: 0,
      commitWaitTicks: 0,
      flushes: 0,
      anomalies: 0,
      vacuumRuns: 0,
      vacuumRemoved: 0,
      lostCommits: 0,
      maxDead: 0,
    },
    finished: false,
  }
}

export function emit(
  ctx: Ctx,
  type: TxnEventType,
  actors: TxnEvent['actors'] = {},
  payload: Record<string, unknown> = {},
): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

/* ───────────────────────────── имена ───────────────────────────── */

/** Метка попытки для SSI и подписей: номер сессии и номер прогона. */
export const tagOf = (t: Txn) => `${t.spec}:${t.run}`

/** T1, T1′, T1″ — повтор после ошибки помечается штрихами; круги сессии с repeat — номером: W#3. */
export function labelOf(t: Txn): string {
  const iter = t.looped ? `#${t.iter + 1}` : ''
  const primes = t.retries === 0 ? '' : t.retries < 4 ? '′'.repeat(t.retries) : `′×${t.retries}`
  return `${t.name}${iter}${primes}`
}

export function allTxns(w: TxnWorld): Txn[] {
  return [...w.history, ...w.txns]
}

export function txnByTag(w: TxnWorld, tag: string): Txn | undefined {
  return allTxns(w).find((t) => tagOf(t) === tag)
}

export function txnByXid(w: TxnWorld, xid: number): Txn | undefined {
  return allTxns(w).find((t) => t.xid === xid)
}

/** Подпись xid: «101 (T2)». */
export function xidLabel(w: TxnWorld, xid: number | null): string {
  if (xid === null) return '—'
  if (xid === BOOT_XID) return `${xid}`
  const t = txnByXid(w, xid)
  return t ? `${xid} (${labelOf(t)})` : `${xid}`
}

/* ─────────────────────────── состояние xid ─────────────────────────── */

export function statusOf(w: TxnWorld, xid: number): XactStatus {
  return w.xact[xid] ?? 'aborted'
}

export const isLive = (t: Txn) => t.state === 'active' || t.state === 'waiting' || t.state === 'committing'

/**
 * Снимок: запоминаем, какие транзакции ещё идут. Всё, что начнётся позже
 * xmax или идёт сейчас (xip), для этого снимка не существует — даже после коммита.
 */
export function takeSnapshot(w: TxnWorld, own: number | null): Snapshot {
  const xip = w.txns
    .filter((t) => isLive(t) && t.xid !== null && t.xid !== own && statusOf(w, t.xid) === 'in-progress')
    .map((t) => t.xid!)
    .sort((a, b) => a - b)
  const xmax = w.nextXid
  return { xmin: Math.min(xmax, ...xip), xmax, xip, tick: w.tick }
}

/** Закоммичена ли транзакция ДО снимка — с его точки зрения. */
export function committedBefore(w: TxnWorld, xid: number, snap: Snapshot): boolean {
  if (statusOf(w, xid) !== 'committed') return false
  return xid < snap.xmax && !snap.xip.includes(xid)
}

/**
 * Видна ли версия строки — правило HeapTupleSatisfiesMVCC.
 *
 * `dirty` — учебный READ UNCOMMITTED: видна самая свежая версия, если её
 * создатель не откатился, закоммичен он или нет.
 */
export function visibleTo(w: TxnWorld, t: Tuple, snap: Snapshot, own: number | null, dirty = false): boolean {
  if (dirty) {
    if (statusOf(w, t.xmin) === 'aborted') return false
    if (t.xmax !== null && !t.lockOnly && statusOf(w, t.xmax) !== 'aborted') return false
    return true
  }
  if (own !== null && t.xmin === own) return !(t.xmax === own && !t.lockOnly)
  if (!committedBefore(w, t.xmin, snap)) return false
  if (t.xmax === null || t.lockOnly) return true
  if (own !== null && t.xmax === own) return false
  if (statusOf(w, t.xmax) === 'aborted') return true
  return !committedBefore(w, t.xmax, snap)
}

/**
 * Как транзакция видит таблицу прямо сейчас — для схемы и проверок.
 * У REPEATABLE READ и SERIALIZABLE это их снимок. У READ COMMITTED снимок живёт
 * только внутри оператора, поэтому между операторами показываем тот, что взял бы
 * следующий, а посреди ожидания — снимок оператора, который ждёт.
 */
export function viewOf(w: TxnWorld, t: Txn): Snapshot | null {
  if (!isLive(t) && t.state !== 'idle') return null
  const perStatement = t.isolation === 'read-committed' || t.isolation === 'read-uncommitted'
  if (t.snapshot && (!perStatement || t.state === 'waiting')) return t.snapshot
  return takeSnapshot(w, t.xid)
}

/** Снимок «прямо сейчас, без своей транзакции» — так видит таблицу новый клиент. */
export function nowSnapshot(w: TxnWorld): Snapshot {
  return takeSnapshot(w, null)
}

/** Закоммиченное значение по ключу. null — строки нет. */
export function committedValue(w: TxnWorld, key: string): number | null {
  const snap = nowSnapshot(w)
  const t = w.tuples.find((x) => x.key === key && visibleTo(w, x, snap, null))
  return t ? t.value : null
}

/** Все ключи, у которых есть хоть одна версия, — в порядке появления. */
export function keysOf(w: TxnWorld): string[] {
  return [...new Set(w.tuples.map((t) => t.key))]
}

/**
 * Мёртвая версия: её заменила или удалила закоммиченная транзакция, или её
 * создатель откатился. Убрать её можно, только если она мертва для ВСЕХ снимков.
 */
export function isDead(w: TxnWorld, t: Tuple): boolean {
  if (statusOf(w, t.xmin) === 'aborted') return true
  return t.xmax !== null && !t.lockOnly && statusOf(w, t.xmax) === 'committed'
}

/**
 * Горизонт очистки: самый старый xmin среди живых транзакций. Версию, умершую
 * раньше горизонта, не видит уже никто.
 */
export function horizon(w: TxnWorld): { xid: number; holder: Txn | null } {
  let best = w.nextXid
  let holder: Txn | null = null
  for (const t of w.txns) {
    if (!isLive(t)) continue
    const cands: number[] = []
    if (t.xid !== null) cands.push(t.xid)
    if (t.holdsSnapshot && t.snapshot) cands.push(t.snapshot.xmin)
    for (const c of cands) {
      if (c < best) {
        best = c
        holder = t
      }
    }
  }
  return { xid: best, holder }
}

/* ─────────────────────────────── куча ─────────────────────────────── */

/** Первый свободный слот. Освобождённое вакуумом место занимается снова — так куча перестаёт расти. */
export function allocSlot(w: TxnWorld): { page: number; slot: number } {
  const used = new Set(w.tuples.map((t) => t.page * 1000 + t.slot))
  for (let i = 0; ; i++) {
    const page = Math.floor(i / w.config.pageSlots)
    const slot = i % w.config.pageSlots
    if (!used.has(page * 1000 + slot)) return { page, slot }
  }
}

export function pagesOf(w: TxnWorld): number {
  return w.tuples.reduce((m, t) => Math.max(m, t.page + 1), 0)
}

export function tupleById(w: TxnWorld, id: number): Tuple | undefined {
  return w.tuples.find((t) => t.id === id)
}

/* ─────────────────────────────── WAL ─────────────────────────────── */

export function appendWal(w: TxnWorld, rec: Omit<WalRecord, 'lsn' | 'tick'>): number {
  const lsn = w.nextLsn++
  w.wal.push({ ...rec, lsn, tick: w.tick })
  return lsn
}

export function assignXid(ctx: Ctx, t: Txn): number {
  const w = ctx.world
  if (t.xid !== null) return t.xid
  t.xid = w.nextXid++
  w.xact[t.xid] = 'in-progress'
  emit(ctx, 'txn.xid', { txn: [t.spec] }, { txn: labelOf(t), xid: t.xid })
  return t.xid
}
