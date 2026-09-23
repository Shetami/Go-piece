import type { Replica, Request, Sample } from './types.ts'
import { OUTLIER_EJECT, OUTLIER_ERRORS, SCALE_EVERY } from './types.ts'
import type { Ctx } from './world.ts'
import { FIRST_MILESTONE, busyOf, emit, emptySample, geometric, multiplierAt, newReplica, poisson, random, routable } from './world.ts'

/**
 * Один тик. Порядок фаз:
 *
 *   1. сбои по расписанию ломают и чинят реплики;
 *   2. запустившиеся реплики входят в ротацию;
 *   3. воркеры отдают готовые ответы;
 *   4. клиенты, чей дедлайн прошёл, бросают запросы;
 *   5. балансировщик проверяет здоровье и возвращает исключённые реплики;
 *   6. приходят запросы и повторы, балансировщик выбирает реплику;
 *   7. свободные воркеры берут запросы из очередей;
 *   8. автомасштабирование принимает решение, тик пишется в журнал.
 */
export function tick(ctx: Ctx): void {
  const w = ctx.world
  w.tick++
  const x = multiplierAt(ctx.scenario, w.tick)
  if (x !== w.x) {
    emit(ctx, 'load.change', {}, { from: w.x, to: x, rate: w.config.rate * x, capacity: capacityNow(ctx) })
    w.x = x
  }
  const sample = emptySample(x, w.replicas.length)

  applyFaults(ctx, sample)
  boot(ctx)
  complete(ctx, sample)
  expire(ctx, sample)
  healthCheck(ctx)
  arrive(ctx, sample)
  for (const r of w.replicas) dispatch(ctx, r)
  autoscale(ctx)
  account(ctx, sample)
}

const capacityNow = (ctx: Ctx) => {
  const w = ctx.world
  return w.replicas.filter((r) => r.state === 'up' || r.state === 'slow').reduce((a, r) => a + w.config.workers / (w.config.serviceTime * r.factor), 0)
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

function applyFaults(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  for (const f of ctx.scenario.faults) {
    const r = w.replicas[f.replica]
    if (!r) continue
    if (f.at === w.tick) {
      r.state = f.kind
      r.factor = f.kind === 'slow' ? f.factor : 1
      // Упавшая реплика рвёт соединения: всё, что в ней было, получает ошибку.
      if (f.kind === 'crash' || f.kind === 'errors') dropAll(ctx, r, sample)
      emit(ctx, 'replica.fault', { replica: [r.id] }, { replica: r.name, kind: f.kind, factor: r.factor, inRotation: r.inRotation, algo: w.config.algo })
    }
    if (f.until !== undefined && f.until === w.tick && r.state !== 'booting') {
      const was = r.state
      // Зависшую реплику перезапускают: брошенные в ней запросы пропадают, остальные получают ошибку.
      if (was === 'hang') dropAll(ctx, r, sample)
      r.state = 'up'
      r.factor = 1
      emit(ctx, 'replica.recover', { replica: [r.id] }, { replica: r.name, was, inRotation: r.inRotation })
    }
  }
}

/** Реплика потеряла все запросы: живые клиенты получают ошибку соединения. */
function dropAll(ctx: Ctx, r: Replica, sample: Sample): void {
  const all = [...r.queue, ...r.workers.filter((x): x is Request => x !== null)]
  r.queue = []
  r.workers = r.workers.map(() => null)
  for (const q of all) {
    if (q.abandoned) continue
    r.outstanding--
    fail(ctx, r, q.origin, q.attempt, q.bornAt, sample, 'reset')
  }
}

function boot(ctx: Ctx): void {
  const w = ctx.world
  for (const r of w.replicas) {
    if (r.state !== 'booting' || r.bootUntil > w.tick) continue
    r.state = 'up'
    r.inRotation = true
    emit(ctx, 'scale.ready', { replica: [r.id] }, { replica: r.name, boot: w.config.bootTime, total: w.replicas.filter((x) => x.state !== 'booting').length })
  }
}

/* ─────────────────────────────── ответы ─────────────────────────────── */

function complete(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  for (const r of w.replicas) {
    if (r.state === 'hang') continue
    r.workers.forEach((q, i) => {
      if (!q || q.finishAt !== w.tick) return
      r.workers[i] = null
      if (q.abandoned) {
        w.stats.wasted++
        ctx.log.done.push({ t: w.tick, lat: w.tick - q.bornAt, replica: r.id, outcome: 'wasted' })
        return
      }
      r.outstanding--
      r.served++
      r.errorsInRow = 0
      w.stats.ok++
      sample.ok++
      ctx.log.done.push({ t: w.tick, lat: w.tick - q.bornAt, replica: r.id, outcome: 'ok' })
      emit(ctx, 'req.done', { replica: [r.id], req: [q.id] }, { replica: r.name, lat: w.tick - q.bornAt, wait: (q.startAt ?? w.tick) - q.arriveAt, service: q.service, attempt: q.attempt })
    })
  }
}

function expire(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  if (w.config.timeout <= 0) return
  for (const r of w.replicas) {
    for (const q of [...r.queue, ...r.workers.filter((x): x is Request => x !== null)]) {
      if (q.abandoned || q.deadline === null || w.tick < q.deadline) continue
      q.abandoned = true
      r.outstanding--
      w.stats.timeouts++
      sample.timeouts++
      emit(ctx, 'req.timeout', { replica: [r.id], req: [q.id] }, {
        replica: r.name,
        state: r.state,
        where: q.startAt === null ? 'queue' : 'work',
        timeout: w.config.timeout,
        retriesLeft: Math.max(0, w.config.retries - q.attempt),
      })
      retryOrFail(ctx, r.id, q.origin, q.attempt, q.bornAt, 'timeout')
    }
  }
}

/** Попытка кончилась ошибкой на реплике r. */
function fail(ctx: Ctx, r: Replica, origin: number, attempt: number, bornAt: number, sample: Sample, reason: 'errors' | 'crash' | 'reset'): void {
  const w = ctx.world
  r.errors++
  r.errorsInRow++
  w.stats.errors++
  sample.errors++
  ctx.log.done.push({ t: w.tick, lat: w.tick - bornAt, replica: r.id, outcome: 'error' })
  emit(ctx, 'req.error', { replica: [r.id] }, { replica: r.name, reason, inRow: r.errorsInRow, algo: w.config.algo, outstanding: r.outstanding, retriesLeft: Math.max(0, w.config.retries - attempt) })
  if (w.config.outlier && r.errorsInRow >= OUTLIER_ERRORS && r.ejectedUntil <= w.tick) {
    r.ejectedUntil = w.tick + OUTLIER_EJECT
    r.errorsInRow = 0
    emit(ctx, 'outlier.eject', { replica: [r.id] }, { replica: r.name, errors: OUTLIER_ERRORS, until: r.ejectedUntil, forTicks: OUTLIER_EJECT })
  }
  retryOrFail(ctx, r.id, origin, attempt, bornAt, 'error')
}

function retryOrFail(ctx: Ctx, avoid: number, origin: number, attempt: number, bornAt: number, reason: 'error' | 'timeout' | 'nobackend'): void {
  const w = ctx.world
  if (attempt < w.config.retries) {
    w.retryQ.push({ origin, attempt: attempt + 1, bornAt, at: w.tick + 1, avoid })
    return
  }
  w.stats.failed++
  emit(ctx, 'req.fail', {}, { attempts: attempt + 1, waited: w.tick - bornAt, reason })
}

/* ─────────────────────────────── здоровье ─────────────────────────────── */

function healthCheck(ctx: Ctx): void {
  const w = ctx.world
  const c = w.config
  for (const r of w.replicas) {
    if (r.ejectedUntil === w.tick) emit(ctx, 'outlier.return', { replica: [r.id] }, { replica: r.name, state: r.state })
  }
  if (c.healthEvery <= 0 || w.tick % c.healthEvery !== 0) return
  for (const r of w.replicas) {
    if (r.state === 'booting') continue
    // Проверка — лёгкий запрос /healthz. Её проходит и медленная реплика, и та, что отвечает
    // ошибками: процесс жив и на /healthz отвечает, а в базу за ответом не ходит.
    const pass = r.state === 'up' || r.state === 'slow' || r.state === 'errors'
    if (pass) {
      r.failedChecks = 0
      if (!r.inRotation && ++r.passedChecks >= c.healthFails) {
        r.inRotation = true
        r.passedChecks = 0
        emit(ctx, 'health.restore', { replica: [r.id] }, { replica: r.name, checks: c.healthFails, every: c.healthEvery })
      }
      continue
    }
    r.passedChecks = 0
    r.failedChecks++
    if (!r.inRotation) continue
    emit(ctx, 'health.fail', { replica: [r.id] }, { replica: r.name, state: r.state, failed: r.failedChecks, need: c.healthFails })
    if (r.failedChecks >= c.healthFails) {
      r.inRotation = false
      emit(ctx, 'health.eject', { replica: [r.id] }, { replica: r.name, state: r.state, checks: c.healthFails, every: c.healthEvery })
    }
  }
}

/* ─────────────────────────────── приход ─────────────────────────────── */

function arrive(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const due = w.retryQ.filter((p) => p.at <= w.tick)
  if (due.length > 0) w.retryQ = w.retryQ.filter((p) => p.at > w.tick)
  for (const p of due) {
    w.stats.retries++
    emit(ctx, 'req.retry', {}, { attempt: p.attempt + 1, avoid: w.replicas[p.avoid]?.name ?? '', waited: w.tick - p.bornAt })
    route(ctx, sample, p.origin, p.attempt, p.bornAt, p.avoid)
  }
  const n = poisson(w, w.config.rate * w.x)
  for (let i = 0; i < n; i++) {
    w.stats.logical++
    route(ctx, sample, w.nextOrigin++, 0, w.tick, -1)
  }
}

/** Выбор реплики. Балансировщик знает только, сколько своих запросов он ждёт от каждой. */
export function pick(ctx: Ctx, avoid: number): Replica | null {
  const w = ctx.world
  let cand = w.replicas.filter((r) => routable(w, r) && r.id !== avoid)
  if (cand.length === 0) cand = w.replicas.filter((r) => routable(w, r))
  if (cand.length === 0) return null
  const rnd = () => random(w, 'route')
  switch (w.config.algo) {
    case 'round-robin': {
      const n = w.replicas.length
      for (let k = 0; k < n; k++) {
        const r = w.replicas[(w.rr + k) % n]!
        if (cand.includes(r)) {
          w.rr = (r.id + 1) % n
          return r
        }
      }
      return cand[0]!
    }
    case 'random':
      return cand[Math.floor(rnd() * cand.length)]!
    case 'least-conn': {
      const min = Math.min(...cand.map((r) => r.outstanding))
      const best = cand.filter((r) => r.outstanding === min)
      return best[Math.floor(rnd() * best.length)]!
    }
    case 'p2c': {
      if (cand.length === 1) return cand[0]!
      const a = Math.floor(rnd() * cand.length)
      let b = Math.floor(rnd() * (cand.length - 1))
      if (b >= a) b++
      const ra = cand[a]!
      const rb = cand[b]!
      return rb.outstanding < ra.outstanding ? rb : ra
    }
  }
}

function route(ctx: Ctx, sample: Sample, origin: number, attempt: number, bornAt: number, avoid: number): void {
  const w = ctx.world
  sample.arrived++
  w.stats.attempts++
  const r = pick(ctx, avoid)
  if (!r) {
    w.stats.noBackend++
    sample.errors++
    emit(ctx, 'req.nobackend', {}, { replicas: w.replicas.length })
    retryOrFail(ctx, -1, origin, attempt, bornAt, 'nobackend')
    return
  }
  r.routed++
  if (r.state === 'errors' || r.state === 'crash') {
    fail(ctx, r, origin, attempt, bornAt, sample, r.state)
    return
  }
  const q: Request = {
    id: w.nextId++,
    origin,
    attempt,
    bornAt,
    arriveAt: w.tick,
    deadline: w.config.timeout > 0 ? w.tick + w.config.timeout : null,
    service: 0,
    startAt: null,
    finishAt: null,
    abandoned: false,
  }
  r.queue.push(q)
  r.outstanding++
  emit(ctx, 'req.route', { replica: [r.id], req: [q.id] }, { replica: r.name, algo: w.config.algo, outstanding: r.outstanding, attempt })
}

/* ─────────────────────────────── работа ─────────────────────────────── */

function dispatch(ctx: Ctx, r: Replica): void {
  const w = ctx.world
  if (r.state === 'hang' || r.state === 'errors' || r.state === 'crash' || r.state === 'booting') return
  r.workers.forEach((busy, i) => {
    if (busy) return
    const q = r.queue.shift()
    if (!q) return
    q.startAt = w.tick
    q.service = Math.max(1, Math.ceil(geometric(w, w.config.serviceTime) * r.factor))
    q.finishAt = w.tick + q.service
    r.workers[i] = q
  })
}

/* ─────────────────────────────── масштаб ─────────────────────────────── */

function autoscale(ctx: Ctx): void {
  const w = ctx.world
  const c = w.config
  for (const r of w.replicas) {
    if (!routable(w, r)) continue
    w.busyAcc += busyOf(r)
    w.capAcc += r.workers.length
  }
  if (w.tick % SCALE_EVERY !== 0) return
  const util = w.capAcc === 0 ? 0 : w.busyAcc / w.capAcc
  w.busyAcc = 0
  w.capAcc = 0
  if (!c.autoscale) return
  const active = w.replicas.filter((r) => routable(w, r)).length
  const booting = w.replicas.filter((r) => r.state === 'booting').length
  // Как у Kubernetes HPA: нужно столько реплик, чтобы загрузка стала целевой.
  const desired = Math.ceil((active * util * 100) / c.scaleTarget)
  const adding = Math.min(desired - active - booting, c.maxReplicas - w.replicas.length)
  if (adding <= 0) return
  const ids: number[] = []
  for (let k = 0; k < adding; k++) {
    const r = newReplica(w.replicas.length, c.workers, true, w.tick + c.bootTime)
    w.replicas.push(r)
    w.qMilestone.push(FIRST_MILESTONE)
    ids.push(r.id)
  }
  emit(ctx, 'scale.decide', { replica: ids }, { util, target: c.scaleTarget / 100, active, desired, adding, boot: c.bootTime, max: c.maxReplicas })
}

/* ─────────────────────────────── учёт ─────────────────────────────── */

function account(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  sample.load = w.replicas.map((r) => r.queue.length + busyOf(r))
  sample.inRotation = w.replicas.filter((r) => routable(w, r)).length
  sample.replicas = w.replicas.length
  ctx.log.series.push(sample)
  w.replicas.forEach((r, i) => {
    const q = r.queue.length
    w.stats.maxQueue = Math.max(w.stats.maxQueue, q)
    if (q === 0) w.qMilestone[i] = FIRST_MILESTONE
    else if (q >= (w.qMilestone[i] ?? FIRST_MILESTONE)) {
      emit(ctx, 'queue.grow', { replica: [r.id] }, { replica: r.name, len: q, state: r.state, factor: r.factor, algo: w.config.algo })
      while ((w.qMilestone[i] ?? FIRST_MILESTONE) <= q) w.qMilestone[i] = (w.qMilestone[i] ?? FIRST_MILESTONE) * 2
    }
  })
}
