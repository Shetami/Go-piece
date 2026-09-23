import type { DepCall, Request, Sample } from './types.ts'
import type { Ctx } from './world.ts'
import { FIRST_MILESTONE, busyOf, emit, emptySample, geometric, multiplierAt, poisson, random, waitingOf } from './world.ts'

/**
 * Один тик. Порядок фаз:
 *
 *   1. сбои по расписанию: B тормозит, отвечает ошибками или зависает;
 *   2. предохранитель переходит из разомкнутого в полуоткрытое состояние;
 *   3. B отдаёт ответы, A завершает свои шаги, срабатывают таймауты;
 *   4. приходят запросы — через ограничитель частоты;
 *   5. свободные воркеры A и B берут работу из очередей;
 *   6. тик пишется в журнал.
 */
export function tick(ctx: Ctx): void {
  const w = ctx.world
  w.tick++
  const x = multiplierAt(ctx.scenario, w.tick)
  if (x !== w.x) {
    emit(ctx, 'load.change', {}, { from: w.x, to: x, rate: w.config.rate * x, workers: w.config.workers })
    w.x = x
  }
  const sample = emptySample(x)
  faults(ctx)
  breakerTick(ctx)
  depComplete(ctx, sample)
  ownComplete(ctx, sample)
  expire(ctx, sample)
  arrive(ctx, sample)
  dispatch(ctx)
  depDispatch(ctx)
  account(ctx, sample)
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

function faults(ctx: Ctx): void {
  const w = ctx.world
  for (const f of ctx.scenario.faults) {
    if (f.at === w.tick) {
      w.dep = f.kind
      w.depFactor = f.kind === 'slow' ? f.factor : 1
      emit(ctx, 'dep.fault', {}, { kind: f.kind, factor: w.depFactor, timeout: w.config.timeout, breaker: w.config.breaker })
    }
    if (f.until !== undefined && f.until === w.tick) {
      const was = w.dep
      w.dep = 'ok'
      w.depFactor = 1
      emit(ctx, 'dep.recover', {}, { was, breaker: w.breaker })
    }
  }
}

/* ───────────────────────────── предохранитель ───────────────────────────── */

function breakerTick(ctx: Ctx): void {
  const w = ctx.world
  if (w.breaker === 'open' && w.tick >= w.breakerUntil) {
    w.breaker = 'half-open'
    w.probing = false
    emit(ctx, 'breaker.half', {}, { after: w.config.breakerOpen })
  }
}

function breakerFail(ctx: Ctx): void {
  const w = ctx.world
  if (!w.config.breaker) return
  if (w.breaker === 'half-open') {
    w.breaker = 'open'
    w.breakerUntil = w.tick + w.config.breakerOpen
    w.probing = false
    emit(ctx, 'breaker.open', {}, { fails: w.config.breakerFails, openFor: w.config.breakerOpen, probe: true })
    return
  }
  w.breakerFails++
  if (w.breakerFails >= w.config.breakerFails) {
    w.breaker = 'open'
    w.breakerUntil = w.tick + w.config.breakerOpen
    w.breakerFails = 0
    emit(ctx, 'breaker.open', {}, { fails: w.config.breakerFails, openFor: w.config.breakerOpen, probe: false })
  }
}

function breakerOk(ctx: Ctx): void {
  const w = ctx.world
  if (!w.config.breaker) return
  if (w.breaker === 'half-open') {
    w.breaker = 'closed'
    w.probing = false
    emit(ctx, 'breaker.close', {}, { openFor: w.config.breakerOpen })
  }
  w.breakerFails = 0
}

/* ─────────────────────────────── ответы ─────────────────────────────── */

function finish(ctx: Ctx, r: Request, kind: 'ok' | 'degraded' | 'error', sample: Sample): void {
  const w = ctx.world
  if (r.worker !== null) w.workers[r.worker] = null
  r.stage = 'done'
  r.worker = null
  const lat = w.tick - r.bornAt
  ctx.log.done.push({ t: w.tick, lat, kind, needsDep: r.needsDep })
  if (kind === 'ok') {
    w.stats.ok++
    sample.ok++
    if (!r.needsDep && lat > w.config.ownTime * 3) w.stats.localSlow++
    emit(ctx, 'req.done', { req: [r.id] }, { id: r.id, lat, needsDep: r.needsDep, attempt: r.attempt })
  } else if (kind === 'degraded') {
    w.stats.degraded++
    sample.degraded++
    emit(ctx, 'req.fallback', { req: [r.id] }, { id: r.id, lat, breaker: w.breaker, dep: w.dep })
  } else {
    w.stats.errors++
    sample.errors++
    emit(ctx, 'req.fail', { req: [r.id] }, { id: r.id, lat, attempts: r.attempt + 1, dep: w.dep, breaker: w.breaker })
  }
  delete w.requests[r.id]
}

/** Вызов B не удался: повторить, ответить заглушкой или ошибкой. */
function afterFailure(ctx: Ctx, r: Request, sample: Sample, reason: 'timeout' | 'error' | 'open' | 'bulkhead'): void {
  const w = ctx.world
  if ((reason === 'timeout' || reason === 'error') && r.attempt < w.config.retries) {
    r.attempt++
    w.stats.retries++
    emit(ctx, 'dep.retry', { req: [r.id] }, { id: r.id, attempt: r.attempt + 1, reason })
    callDep(ctx, r, sample)
    return
  }
  finish(ctx, r, w.config.fallback ? 'degraded' : 'error', sample)
}

function depComplete(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  if (w.dep === 'hang') return
  w.depWorkers.forEach((call, i) => {
    if (!call || call.finishAt !== w.tick) return
    w.depWorkers[i] = null
    const r = w.requests[call.req]
    if (!r) return
    r.inDep = false
    if (w.dep === 'errors') {
      w.stats.depErrors++
      emit(ctx, 'dep.error', { req: [r.id] }, { id: r.id, waited: w.tick - (r.calledAt ?? w.tick), breaker: w.breaker })
      breakerFail(ctx)
      afterFailure(ctx, r, sample, 'error')
      return
    }
    breakerOk(ctx)
    emit(ctx, 'dep.ok', { req: [r.id] }, { id: r.id, waited: w.tick - (r.calledAt ?? w.tick) })
    finish(ctx, r, 'ok', sample)
  })
}

/** A закончил свою часть работы: либо готов ответ, либо пора звать B. */
function ownComplete(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  for (const id of [...w.workers]) {
    if (id === null) continue
    const r = w.requests[id]
    if (!r || r.stage !== 'own' || r.until > w.tick) continue
    if (!r.needsDep) {
      finish(ctx, r, 'ok', sample)
      continue
    }
    callDep(ctx, r, sample)
  }
}

function callDep(ctx: Ctx, r: Request, sample: Sample): void {
  const w = ctx.world
  const c = w.config
  // Предохранитель разомкнут: в B не ходим вовсе — отвечаем сразу.
  if (c.breaker && w.breaker === 'open') {
    finish(ctx, r, c.fallback ? 'degraded' : 'error', sample)
    return
  }
  if (c.breaker && w.breaker === 'half-open') {
    if (w.probing) {
      finish(ctx, r, c.fallback ? 'degraded' : 'error', sample)
      return
    }
    w.probing = true
  }
  // Переборка: больше стольких воркеров A одновременно ждать B не могут.
  if (c.bulkhead > 0 && waitingOf(w) >= c.bulkhead) {
    w.stats.blocked++
    emit(ctx, 'bulkhead.block', { req: [r.id] }, { id: r.id, limit: c.bulkhead, workers: c.workers })
    afterFailure(ctx, r, sample, 'bulkhead')
    return
  }
  r.stage = 'dep'
  r.calledAt = w.tick
  r.inDep = true
  w.stats.depCalls++
  w.depQueue.push({ id: w.nextId++, req: r.id, arriveAt: w.tick, startAt: null, finishAt: null })
  emit(ctx, 'dep.call', { req: [r.id], worker: r.worker === null ? [] : [r.worker] }, { id: r.id, attempt: r.attempt + 1, depQueue: w.depQueue.length, timeout: c.timeout })
}

function expire(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const c = w.config
  if (c.timeout <= 0) return
  for (const id of [...w.workers]) {
    if (id === null) continue
    const r = w.requests[id]
    if (!r || r.stage !== 'dep' || r.calledAt === null || w.tick - r.calledAt < c.timeout) continue
    // Таймаут с отменой: вызов B снимается и не занимает его воркеры дальше.
    w.depQueue = w.depQueue.filter((q) => q.req !== r.id)
    w.depWorkers.forEach((q, i) => {
      if (q?.req === r.id) w.depWorkers[i] = null
    })
    r.inDep = false
    w.stats.depTimeouts++
    emit(ctx, 'dep.timeout', { req: [r.id] }, { id: r.id, timeout: c.timeout, dep: w.dep, retriesLeft: Math.max(0, c.retries - r.attempt) })
    breakerFail(ctx)
    afterFailure(ctx, r, sample, 'timeout')
  }
}

/* ─────────────────────────────── приход ─────────────────────────────── */

function arrive(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const c = w.config
  if (c.rateLimit > 0) w.tokens = Math.min(c.burst, w.tokens + c.rateLimit)
  const n = poisson(w, c.rate * w.x)
  for (let i = 0; i < n; i++) {
    const needsDep = random(w, 'kind') < c.depShare / 100
    w.stats.arrived++
    sample.arrived++
    if (c.rateLimit > 0) {
      if (w.tokens < 1) {
        w.stats.rejected++
        sample.rejected++
        ctx.log.done.push({ t: w.tick, lat: 1, kind: 'rejected', needsDep })
        emit(ctx, 'req.reject', {}, { limit: c.rateLimit, burst: c.burst, queue: w.queue.length })
        continue
      }
      w.tokens--
    }
    const r: Request = { id: w.nextId++, needsDep, bornAt: w.tick, stage: 'queue', until: 0, worker: null, attempt: 0, calledAt: null, inDep: false }
    w.requests[r.id] = r
    w.queue.push(r.id)
    emit(ctx, 'req.accept', { req: [r.id] }, { id: r.id, needsDep, queue: w.queue.length })
  }
}

function dispatch(ctx: Ctx): void {
  const w = ctx.world
  w.workers.forEach((busy, i) => {
    if (busy !== null) return
    const id = w.queue.shift()
    if (id === undefined) return
    const r = w.requests[id]
    if (!r) return
    r.stage = 'own'
    r.worker = i
    r.until = w.tick + geometric(w, w.config.ownTime)
    w.workers[i] = id
  })
}

function depDispatch(ctx: Ctx): void {
  const w = ctx.world
  if (w.dep === 'hang') return
  w.depWorkers.forEach((busy, i) => {
    if (busy) return
    const call = w.depQueue.shift()
    if (!call) return
    call.startAt = w.tick
    call.finishAt = w.tick + Math.max(1, Math.round(geometric(w, w.config.depTime) * w.depFactor))
    w.depWorkers[i] = call as DepCall
  })
}

/* ─────────────────────────────── учёт ─────────────────────────────── */

function account(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const busy = busyOf(w)
  const waiting = waitingOf(w)
  sample.busy = busy
  sample.waiting = waiting
  sample.queue = w.queue.length
  sample.depQueue = w.depQueue.length + w.depWorkers.filter((x) => x !== null).length
  w.stats.busyTicks += busy
  w.stats.waitTicks += waiting
  w.stats.maxQueue = Math.max(w.stats.maxQueue, w.queue.length)
  ctx.log.series.push(sample)

  if (!w.saturated && busy === w.config.workers && waiting >= Math.ceil(w.config.workers * 0.75)) {
    w.saturated = true
    emit(ctx, 'pool.saturated', {}, { workers: w.config.workers, waiting, queue: w.queue.length, dep: w.dep, timeout: w.config.timeout })
  } else if (w.saturated && waiting <= w.config.workers / 2) {
    w.saturated = false
  }
  const q = w.queue.length
  if (q === 0) w.qMilestone = FIRST_MILESTONE
  else if (q >= w.qMilestone) {
    emit(ctx, 'queue.grow', {}, { len: q, waiting, dep: w.dep, breaker: w.breaker })
    while (w.qMilestone <= q) w.qMilestone *= 2
  }
}
