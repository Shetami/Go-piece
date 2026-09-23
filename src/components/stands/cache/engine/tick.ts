import type { Query, Sample, Waiter } from './types.ts'
import type { Ctx } from './world.ts'
import { FIRST_MILESTONE, STAMPEDE, cachedCount, dbCapacity, emit, emptySample, geometric, multiplierAt, poisson, random, sampleKey } from './world.ts'

/**
 * Один тик. Порядок фаз:
 *
 *   1. сбои по расписанию: сброс или прогрев кэша;
 *   2. база отдаёт готовые ответы — чтения кладутся в кэш, записи его инвалидируют;
 *   3. приходят запросы: попадание отвечает сразу, промах идёт в базу;
 *   4. свободные воркеры базы берут запросы из очереди;
 *   5. тик пишется в журнал.
 */
export function tick(ctx: Ctx): void {
  const w = ctx.world
  w.tick++
  const x = multiplierAt(ctx.scenario, w.tick)
  if (x !== w.x) {
    emit(ctx, 'load.change', {}, { from: w.x, to: x, rate: w.config.rate * x })
    w.x = x
  }
  const sample = emptySample(x)
  faults(ctx)
  complete(ctx)
  arrive(ctx, sample)
  dispatch(ctx, sample)
  account(ctx, sample)
}

/* ─────────────────────────────── кэш ─────────────────────────────── */

function setEntry(ctx: Ctx, key: number, version: number): void {
  const w = ctx.world
  const c = w.config
  if (c.cacheSize <= 0) return
  const jitter = c.ttl > 0 ? Math.floor(random(w, 'ttl') * ((c.ttl * c.ttlJitter) / 100 + 1)) : 0
  w.cache[key] = { version, expiresAt: c.ttl > 0 ? w.tick + c.ttl + jitter : Number.POSITIVE_INFINITY, lastUsed: w.tick, setAt: w.tick }
  // LRU: вытесняем ключ, к которому дольше всего не обращались.
  while (cachedCount(w) > c.cacheSize) {
    let victim = -1
    let oldest = Number.POSITIVE_INFINITY
    for (const [k, e] of Object.entries(w.cache)) {
      if (e.lastUsed < oldest) {
        oldest = e.lastUsed
        victim = Number(k)
      }
    }
    delete w.cache[victim]
  }
}

function faults(ctx: Ctx): void {
  const w = ctx.world
  for (const f of ctx.scenario.faults) {
    if (f.at !== w.tick) continue
    if (f.kind === 'flush') {
      const entries = cachedCount(w)
      w.cache = {}
      emit(ctx, 'cache.flush', {}, { entries })
    } else {
      const n = Math.min(f.count, w.config.cacheSize, w.config.keys)
      for (let k = 0; k < n; k++) setEntry(ctx, k, w.db[k]!)
      emit(ctx, 'cache.warm', {}, { count: n, ttl: w.config.ttl, jitter: w.config.ttlJitter })
    }
  }
}

/* ─────────────────────────────── база ─────────────────────────────── */

function respond(ctx: Ctx, waiters: Waiter[], kind: 'miss' | 'write'): void {
  const w = ctx.world
  for (const x of waiters) ctx.log.done.push({ t: w.tick, lat: w.tick - x.bornAt, kind })
}

function complete(ctx: Ctx): void {
  const w = ctx.world
  const c = w.config
  w.workers.forEach((q, i) => {
    if (!q || q.finishAt !== w.tick) return
    w.workers[i] = null
    if (q.kind === 'write') {
      w.db[q.key]!++
      if (c.invalidation === 'delete') delete w.cache[q.key]
      else if (c.invalidation === 'update' && w.cache[q.key]) setEntry(ctx, q.key, w.db[q.key]!)
      respond(ctx, q.waiters, 'write')
      emit(ctx, 'write.done', { key: [q.key], query: [q.id] }, { key: q.key, version: w.db[q.key], invalidation: c.invalidation, cached: w.cache[q.key] !== undefined })
      return
    }
    w.inflight[q.key] = (w.inflight[q.key] ?? 1) - 1
    if (w.inflight[q.key]! <= 0) {
      delete w.inflight[q.key]
      w.stampedeKeys = w.stampedeKeys.filter((k) => k !== q.key)
    }
    const version = q.version ?? 0
    setEntry(ctx, q.key, version)
    // Гонка cache-aside: чтение взяло значение до записи, а положило его в кэш после инвалидации.
    if (c.cacheSize > 0 && version < w.db[q.key]!) {
      emit(ctx, 'cache.stale-set', { key: [q.key], query: [q.id] }, { key: q.key, cached: version, db: w.db[q.key], ttl: c.ttl, invalidation: c.invalidation })
    }
    respond(ctx, q.waiters, 'miss')
    emit(ctx, 'db.done', { key: [q.key], query: [q.id] }, { key: q.key, waiters: q.waiters.length, service: (q.finishAt ?? w.tick) - (q.startAt ?? w.tick), refresh: q.kind === 'refresh' })
  })
}

/* ─────────────────────────────── запросы ─────────────────────────────── */

function send(ctx: Ctx, kind: Query['kind'], key: number, waiters: Waiter[]): Query {
  const w = ctx.world
  const q: Query = { id: w.nextId++, kind, key, waiters, arriveAt: w.tick, startAt: null, finishAt: null, version: null }
  w.queue.push(q)
  if (kind === 'write') {
    w.stats.dbWrites++
    return q
  }
  w.stats.dbReads++
  const n = (w.inflight[key] = (w.inflight[key] ?? 0) + 1)
  w.stats.maxSameKey = Math.max(w.stats.maxSameKey, n)
  if (n >= STAMPEDE && !w.stampedeKeys.includes(key)) {
    w.stampedeKeys.push(key)
    emit(ctx, 'db.stampede', { key: [key] }, { key, inflight: n, coalesce: w.config.coalesce, ttl: w.config.ttl })
  }
  return q
}

/** Запрос в базу за этим ключом, к которому можно присоединиться. */
const pendingRead = (ctx: Ctx, key: number) =>
  [...ctx.world.queue, ...ctx.world.workers].find((q): q is Query => q !== null && q.key === key && q.kind !== 'write')

function arrive(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const c = w.config
  const n = poisson(w, c.rate * w.x)
  for (let i = 0; i < n; i++) {
    const key = sampleKey(w)
    // Бросок делается всегда: доля записей не сдвигает, какие ключи спрашивают.
    const isWrite = random(w, 'key') < c.writeShare / 100
    const me: Waiter = { id: w.nextId++, bornAt: w.tick }
    sample.arrived++
    if (isWrite) {
      w.stats.writes++
      send(ctx, 'write', key, [me])
      continue
    }
    w.stats.reads++
    const e = w.cache[key]
    if (e && e.expiresAt > w.tick) {
      e.lastUsed = w.tick
      w.stats.hits++
      sample.hits++
      ctx.log.done.push({ t: w.tick, lat: 1, kind: 'hit' })
      if (e.version < w.db[key]!) {
        w.stats.stale++
        sample.stale++
        emit(ctx, 'req.stale', { key: [key] }, { key, cached: e.version, db: w.db[key], age: w.tick - e.setAt, invalidation: c.invalidation })
      } else {
        emit(ctx, 'req.hit', { key: [key] }, { key })
      }
      continue
    }
    if (e) {
      if (c.staleWhileRevalidate) {
        e.lastUsed = w.tick
        w.stats.swr++
        sample.hits++
        ctx.log.done.push({ t: w.tick, lat: 1, kind: 'swr' })
        const refreshing = pendingRead(ctx, key) !== undefined
        if (!refreshing) {
          send(ctx, 'refresh', key, [])
          emit(ctx, 'cache.expired', { key: [key] }, { key, age: w.tick - e.setAt, swr: true, coalesce: c.coalesce })
        }
        emit(ctx, 'req.swr', { key: [key] }, { key, age: w.tick - e.setAt, refreshing })
        continue
      }
      delete w.cache[key]
      emit(ctx, 'cache.expired', { key: [key] }, { key, age: w.tick - e.setAt, swr: false, coalesce: c.coalesce })
    }
    w.stats.misses++
    sample.misses++
    const pending = c.coalesce ? pendingRead(ctx, key) : undefined
    if (pending) {
      pending.waiters.push(me)
      w.stats.joins++
      emit(ctx, 'req.join', { key: [key], query: [pending.id] }, { key, waiters: pending.waiters.length })
    } else {
      const q = send(ctx, 'read', key, [me])
      emit(ctx, 'req.miss', { key: [key], query: [q.id] }, { key, cached: c.cacheSize > 0, queue: w.queue.length })
    }
  }
}

function dispatch(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  w.workers.forEach((busy, i) => {
    if (busy) return
    const q = w.queue.shift()
    if (!q) return
    q.startAt = w.tick
    q.finishAt = w.tick + geometric(w, q.key === 0 && w.config.hotTime > 0 ? w.config.hotTime : w.config.dbTime)
    // Чтение видит значение на момент начала: всё, что запишут позже, оно уже не узнает.
    if (q.kind !== 'write') q.version = w.db[q.key]!
    w.workers[i] = q
    sample.dbStarted++
  })
}

/* ─────────────────────────────── учёт ─────────────────────────────── */

function account(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const q = w.queue.length
  sample.dbQueue = q
  sample.dbBusy = w.workers.filter((x) => x !== null).length
  sample.cached = cachedCount(w)
  w.stats.maxDbQueue = Math.max(w.stats.maxDbQueue, q)
  ctx.log.series.push(sample)
  if (q >= w.qMilestone) {
    emit(ctx, 'db.queue', {}, { len: q, capacity: dbCapacity(w.config), cached: sample.cached })
    while (w.qMilestone <= q) w.qMilestone *= 2
  } else if (q === 0 && w.qMilestone > FIRST_MILESTONE) {
    emit(ctx, 'db.drained', {}, { peak: w.qMilestone / 2, cached: sample.cached })
    w.qMilestone = FIRST_MILESTONE
  }
}
