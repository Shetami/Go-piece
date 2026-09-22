import type { Request, Sample } from './types.ts'
import type { Ctx } from './world.ts'
import { FIRST_MILESTONE, capacityOf, emit, emptySample, geometric, meanService, multiplierAt, offeredLoad, poisson, random } from './world.ts'

/**
 * Один тик сервиса. Порядок фаз важен и повторяет то, что видит клиент:
 *
 *   1. воркеры отдают готовые ответы — ответ, пришедший ровно к дедлайну, успел;
 *   2. клиенты, чей дедлайн прошёл, бросают запросы и планируют повторы;
 *   3. приходят новые запросы и повторы, лишние получают отказ;
 *   4. освободившиеся воркеры берут запросы из головы очереди;
 *   5. тик записывается в журнал.
 */
export function tick(ctx: Ctx): void {
  const w = ctx.world
  w.tick++

  const x = multiplierAt(ctx.scenario, w.tick)
  if (x !== w.x) {
    emit(ctx, 'load.change', {}, { from: w.x, to: x, rate: w.config.rate * x, rho: offeredLoad(w.config, x) })
    w.x = x
  }
  const sample = emptySample(x)

  complete(ctx, sample)
  expire(ctx, sample)
  arrive(ctx, sample)
  dispatch(ctx)
  account(ctx, sample)
}

/* ─────────────────────────────── уход ─────────────────────────────── */

function depart(ctx: Ctx, r: Request, outcome: 'ok' | 'wasted' | 'cancelled'): void {
  const w = ctx.world
  const s = w.stats
  s.departures++
  s.sojournSum += w.tick - r.arriveAt
  s[outcome]++
  ctx.log.done.push({
    t: w.tick,
    id: r.id,
    lat: w.tick - r.bornAt,
    sojourn: w.tick - r.arriveAt,
    wait: (r.startAt ?? w.tick) - r.arriveAt,
    outcome,
    slow: r.slow,
  })
}

/** Воркер доделал запрос — клиенту или в пустоту. */
function complete(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  w.workers.forEach((r, i) => {
    if (!r || r.finishAt !== w.tick) return
    w.workers[i] = null
    if (r.abandoned) {
      depart(ctx, r, 'wasted')
      sample.wasted++
      emit(ctx, 'req.wasted', { req: [r.id], worker: [i] }, { id: r.id, worker: i, service: r.service, late: w.tick - (r.deadline ?? w.tick) })
    } else {
      depart(ctx, r, 'ok')
      sample.ok++
      emit(ctx, 'req.done', { req: [r.id], worker: [i] }, {
        id: r.id,
        worker: i,
        lat: w.tick - r.bornAt,
        wait: (r.startAt ?? w.tick) - r.arriveAt,
        service: r.service,
        attempt: r.attempt,
        slow: r.slow,
      })
    }
    w.recent.push(r.abandoned ? 'wasted' : 'ok')
    if (w.recent.length > RECENT) w.recent.shift()
  })
  watchCollapse(ctx)
}

/** Сколько последних работ воркеров смотреть, чтобы понять, работает ли сервис впустую. */
const RECENT = 10

function watchCollapse(ctx: Ctx): void {
  const w = ctx.world
  if (w.recent.length < RECENT) return
  const wasted = w.recent.filter((o) => o === 'wasted').length
  if (w.collapsedAt === null && wasted >= RECENT - 2) {
    w.collapsedAt = w.tick
    emit(ctx, 'overload.collapse', {}, { wasted, of: RECENT, queue: w.queue.length, timeouts: w.stats.timeouts, retries: w.stats.retries })
  } else if (w.collapsedAt !== null && wasted === 0) {
    emit(ctx, 'overload.recover', {}, { took: w.tick - w.collapsedAt, queue: w.queue.length })
    w.collapsedAt = null
  }
}

/* ─────────────────────────────── таймауты ─────────────────────────────── */

function expire(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const c = w.config
  if (c.timeout <= 0) return
  const late = (r: Request) => !r.abandoned && r.deadline !== null && w.tick >= r.deadline

  const onTimeout = (r: Request, where: 'queue' | 'work', worker?: number) => {
    r.abandoned = true
    w.stats.timeouts++
    sample.timeouts++
    emit(ctx, 'req.timeout', { req: [r.id], worker: worker === undefined ? [] : [worker] }, {
      id: r.id,
      attempt: r.attempt,
      where,
      cancelled: c.cancelOnTimeout,
      waited: w.tick - r.arriveAt,
      timeout: c.timeout,
      retriesLeft: Math.max(0, c.retries - r.attempt),
    })
    if (c.cancelOnTimeout) depart(ctx, r, 'cancelled')
    giveUp(ctx, r.origin, r.attempt, r.bornAt, 'timeout')
  }

  for (const r of w.queue) if (late(r)) onTimeout(r, 'queue')
  w.workers.forEach((r, i) => {
    if (r && late(r)) {
      onTimeout(r, 'work', i)
      if (c.cancelOnTimeout) w.workers[i] = null
    }
  })
  if (c.cancelOnTimeout) w.queue = w.queue.filter((r) => !r.abandoned)
}

/** Попытка не удалась: клиент повторит её или сдастся. */
function giveUp(ctx: Ctx, origin: number, attempt: number, bornAt: number, reason: 'timeout' | 'reject'): void {
  const w = ctx.world
  const c = w.config
  if (attempt < c.retries) {
    // Полный разброс (full jitter): пауза случайна от 1 до 2^attempt·BACKOFF_BASE тиков.
    const delay = c.backoff ? 1 + Math.floor(random(w, 'retry') * Math.min(BACKOFF_CAP, BACKOFF_BASE * 2 ** attempt)) : 1
    w.retryQ.push({ origin, attempt: attempt + 1, bornAt, at: w.tick + delay })
    return
  }
  w.stats.failed++
  emit(ctx, 'req.fail', {}, { origin, attempts: attempt + 1, waited: w.tick - bornAt, reason })
}

const BACKOFF_BASE = 4
const BACKOFF_CAP = 32

/* ─────────────────────────────── приход ─────────────────────────────── */

function arrive(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const c = w.config

  const due = w.retryQ.filter((p) => p.at <= w.tick)
  if (due.length > 0) w.retryQ = w.retryQ.filter((p) => p.at > w.tick)
  for (const p of due) admit(ctx, sample, p.origin, p.attempt, p.bornAt)

  const lambda = c.rate * w.x
  let n: number
  if (c.arrivals === 'even') {
    w.carry += lambda
    n = Math.floor(w.carry + 1e-9)
    w.carry -= n
  } else {
    n = poisson(w, lambda)
  }
  for (let i = 0; i < n; i++) {
    w.stats.logical++
    admit(ctx, sample, w.nextOrigin++, 0, w.tick)
  }
}

function admit(ctx: Ctx, sample: Sample, origin: number, attempt: number, bornAt: number): void {
  const w = ctx.world
  const c = w.config
  const id = w.nextId++
  w.stats.arrivals++
  sample.arrived++
  if (attempt > 0) {
    w.stats.retries++
    sample.retried++
  }

  if (c.queueLimit > 0 && w.queue.length >= c.queueLimit) {
    w.stats.rejected++
    sample.rejected++
    emit(ctx, 'req.reject', { req: [id] }, { id, attempt, limit: c.queueLimit, retriesLeft: Math.max(0, c.retries - attempt) })
    giveUp(ctx, origin, attempt, bornAt, 'reject')
    return
  }

  // Бросок делается всегда: тогда при любой доле медленных запросов обычные получают те же времена.
  const slow = random(w, 'work') < c.slowShare / 100
  const base = c.serviceDist === 'fixed' ? c.serviceTime : geometric(w, c.serviceTime)
  const r: Request = {
    id,
    origin,
    attempt,
    bornAt,
    arriveAt: w.tick,
    deadline: c.timeout > 0 ? w.tick + c.timeout : null,
    service: slow ? base * c.slowFactor : base,
    slow,
    startAt: null,
    finishAt: null,
    abandoned: false,
  }
  w.queue.push(r)
  w.stats.admitted++
  if (attempt > 0) {
    emit(ctx, 'req.retry', { req: [id] }, { id, origin, attempt, retries: c.retries, waited: w.tick - bornAt, queue: w.queue.length, backoff: c.backoff })
  } else {
    emit(ctx, 'req.arrive', { req: [id] }, { id, queue: w.queue.length })
  }
}

/* ─────────────────────────────── работа ─────────────────────────────── */

function dispatch(ctx: Ctx): void {
  const w = ctx.world
  w.workers.forEach((busy, i) => {
    if (busy) return
    const r = w.queue.shift()
    if (!r) return
    r.startAt = w.tick
    r.finishAt = w.tick + r.service
    w.workers[i] = r
    emit(ctx, r.slow ? 'req.slow' : 'req.start', { req: [r.id], worker: [i] }, {
      id: r.id,
      worker: i,
      wait: w.tick - r.arriveAt,
      service: r.service,
      abandoned: r.abandoned,
      behind: w.queue.length,
      normal: w.config.serviceTime,
    })
  })
}

/* ─────────────────────────────── учёт ─────────────────────────────── */

function account(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const s = w.stats
  const q = w.queue.length
  const busy = w.workers.filter((r) => r !== null).length
  sample.queue = q
  sample.busy = busy
  s.busyTicks += busy
  s.inSystemSum += q + busy
  s.maxQueue = Math.max(s.maxQueue, q)
  ctx.log.series.push(sample)

  if (q > 0 && w.qSince === null) w.qSince = w.tick
  w.qPeak = Math.max(w.qPeak, q)
  if (q >= w.qMilestone) {
    const zombies = w.queue.filter((r) => r.abandoned).length
    // Новичок в хвосте очереди ждёт, пока воркеры разберут всех впереди.
    const wait = Math.round((q * meanService(w.config)) / w.config.workers)
    emit(ctx, 'queue.grow', {}, { len: q, wait, zombies, since: w.qSince ?? w.tick, rho: offeredLoad(w.config, w.x), capacity: capacityOf(w.config) })
    while (w.qMilestone <= q) w.qMilestone *= 2
  }
  if (q === 0) {
    if (w.qPeak >= FIRST_MILESTONE) {
      emit(ctx, 'queue.drained', {}, { peak: w.qPeak, took: w.tick - (w.qSince ?? w.tick), rho: offeredLoad(w.config, w.x) })
    }
    w.qPeak = 0
    w.qSince = null
    w.qMilestone = FIRST_MILESTONE
  }
}
