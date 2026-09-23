import type { EventKind, Message, Order, Sample, Work } from './types.ts'
import type { Ctx } from './world.ts'
import { emit, emptySample, geometric, inFlight, paidNotShipped, poisson, random } from './world.ts'

/**
 * Один тик. Порядок фаз:
 *
 *   1. сбои по расписанию: падение сервиса, недоступный брокер, дубли доставки;
 *   2. события доходят до потребителей и попадают в обработку;
 *   3. потребители заканчивают обработку и меняют состояние заказов;
 *   4. повторы неудачных обработок возвращаются в очередь;
 *   5. неопубликованные события уходят в брокер — напрямую или из outbox;
 *   6. приходят новые заказы;
 *   7. тик пишется в журнал.
 */
export function tick(ctx: Ctx): void {
  const w = ctx.world
  w.tick++
  const sample = emptySample(w.x)
  faults(ctx)
  deliver(ctx)
  complete(ctx, sample)
  retryDue(ctx)
  publishPending(ctx)
  create(ctx, sample)
  account(ctx, sample)
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

function faults(ctx: Ctx): void {
  const w = ctx.world
  for (const f of ctx.scenario.faults) {
    if (f.kind === 'crash' && f.at === w.tick) {
      // Сервис заказов перезапустился. Всё, что он собирался отправить, но ещё не отправил,
      // существует только в его памяти — и пропадает. Outbox лежит в базе и переживает падение.
      const lost = w.config.publish === 'dual-write' ? w.orders.filter((o) => !o.published && o.state === 'created') : []
      emit(ctx, 'svc.crash', {}, { publish: w.config.publish, lost: lost.length, outbox: w.outbox.filter((r) => r.sentAt === null).length })
      for (const o of lost) loseEvent(ctx, o, 'crash')
    }
    if (f.kind === 'broker-down') {
      if (f.at === w.tick) {
        w.brokerUp = false
        emit(ctx, 'broker.down', {}, { until: f.until, publish: w.config.publish })
      }
      if (f.until === w.tick) {
        w.brokerUp = true
        emit(ctx, 'broker.up', {}, { outbox: w.outbox.filter((r) => r.sentAt === null).length })
      }
    }
    if (f.kind === 'duplicates') {
      if (f.at === w.tick) w.duplicating = true
      if (f.until === w.tick) w.duplicating = false
    }
  }
}

function loseEvent(ctx: Ctx, o: Order, reason: 'crash' | 'broker'): void {
  const w = ctx.world
  w.stats.lost++
  w.stats.stuck++
  emit(ctx, 'event.lost', { order: [o.id] }, { order: o.id, reason, publish: w.config.publish })
  emit(ctx, 'saga.stuck', { order: [o.id] }, { order: o.id, state: o.state, reason: 'lost' })
}

/* ─────────────────────────────── брокер ─────────────────────────────── */

function push(ctx: Ctx, kind: EventKind, order: number, duplicate = false, attempt = 0): void {
  const w = ctx.world
  const msg: Message = {
    id: w.nextMsg++,
    kind,
    order,
    // Ключ идемпотентности один и тот же у оригинала и у его дубля.
    key: `${kind}:${order}`,
    publishedAt: w.tick,
    arriveAt: w.tick + w.config.brokerDelay,
    attempt,
    duplicate,
  }
  w.broker.push(msg)
}

/** Отправка того, что накопилось: напрямую или отправщиком outbox. */
function publishPending(ctx: Ctx): void {
  const w = ctx.world
  const c = w.config
  if (c.publish === 'dual-write') {
    for (const o of w.orders) {
      if (o.published || o.state !== 'created') continue
      if (!w.brokerUp) {
        // Событие некуда записать: брокер недоступен, а второй записи в базе нет.
        loseEvent(ctx, o, 'broker')
        o.published = true
        continue
      }
      o.published = true
      push(ctx, 'order.created', o.id)
      emit(ctx, 'event.publish', { order: [o.id] }, { order: o.id, publish: 'dual-write' })
    }
    return
  }
  if (w.tick % Math.max(1, c.relayEvery) !== 0) return
  const rows = w.outbox.filter((r) => r.sentAt === null)
  if (rows.length === 0) return
  if (!w.brokerUp) return // строки остаются в базе и уйдут, когда брокер вернётся
  for (const r of rows) {
    r.sentAt = w.tick
    const o = w.orders.find((x) => x.id === r.order)
    if (o) o.published = true
    push(ctx, r.kind, r.order)
  }
  emit(ctx, 'relay.send', {}, { rows: rows.length, every: c.relayEvery, waited: w.tick - Math.min(...rows.map((r) => r.writtenAt)) })
}

function deliver(ctx: Ctx): void {
  const w = ctx.world
  const c = w.config
  const due = w.broker.filter((m) => m.arriveAt <= w.tick)
  if (due.length === 0) return
  for (const m of due) {
    const consumer: Work['consumer'] = m.kind === 'order.paid' ? 'ship' : 'pay'
    const busy = w.work.filter((x) => x.consumer === consumer).length
    if (busy >= c.workers) continue // потребитель занят, сообщение подождёт в брокере
    w.broker = w.broker.filter((x) => x !== m)
    // Доставка «хотя бы один раз»: иногда то же сообщение приходит дважды.
    if (w.duplicating && !m.duplicate) {
      w.stats.duplicates++
      push(ctx, m.kind, m.order, true, m.attempt)
      emit(ctx, 'event.duplicate', { order: [m.order], msg: [m.id] }, { order: m.order, kind: m.kind, idempotent: c.idempotent })
    }
    // Ключ захватывается в начале обработки, как уникальная строка в базе:
    // иначе дубль, пришедший пока обрабатывается оригинал, проскочит мимо проверки.
    const busyKey = w.work.some((x) => x.msg.key === m.key)
    if (c.idempotent && (w.seen.includes(m.key) || busyKey)) {
      emit(ctx, 'event.deliver', { order: [m.order], msg: [m.id] }, { order: m.order, kind: m.kind, skipped: true, duplicate: m.duplicate })
      continue
    }
    w.work.push({ msg: m, startAt: w.tick, finishAt: w.tick + geometric(w, c.workTime), consumer })
    emit(ctx, 'event.deliver', { order: [m.order], msg: [m.id] }, { order: m.order, kind: m.kind, skipped: false, duplicate: m.duplicate })
  }
}

/* ─────────────────────────────── обработка ─────────────────────────────── */

function settle(ctx: Ctx, o: Order, kind: 'shipped' | 'cancelled'): void {
  const w = ctx.world
  o.settledAt = w.tick
  ctx.log.done.push({ t: w.tick, lat: w.tick - o.createdAt, kind })
}

function complete(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const c = w.config
  for (const job of [...w.work]) {
    if (job.finishAt > w.tick) continue
    w.work = w.work.filter((x) => x !== job)
    const m = job.msg
    const o = w.orders.find((x) => x.id === m.order)
    if (!o) continue
    // Обработанным ключ считается только после успеха: неудачу нужно уметь повторить.
    const done = () => {
      if (c.idempotent) w.seen.push(m.key)
    }

    if (m.kind === 'order.created') {
      if (random(w, 'fail') < c.payFail / 100) {
        w.stats.payFails++
        emit(ctx, 'pay.fail', { order: [o.id] }, { order: o.id, attempt: m.attempt + 1, retriesLeft: Math.max(0, c.retries - m.attempt) })
        retryOrGiveUp(ctx, m, o)
        continue
      }
      done()
      o.charges++
      if (o.charges > 1) {
        w.stats.doubleCharges++
        emit(ctx, 'pay.double', { order: [o.id] }, { order: o.id, charges: o.charges, idempotent: c.idempotent, duplicate: m.duplicate })
      } else {
        emit(ctx, 'pay.ok', { order: [o.id] }, { order: o.id, waited: w.tick - o.createdAt })
      }
      if (o.state === 'created') {
        o.state = 'paid'
        w.stats.paid++
        // Платежи рассказывают о списании тем же путём, что и заказы.
        if (c.publish === 'outbox') w.outbox.push({ id: w.nextMsg++, order: o.id, kind: 'order.paid', writtenAt: w.tick, sentAt: null })
        else push(ctx, 'order.paid', o.id)
      }
      continue
    }

    if (m.kind === 'order.paid') {
      if (random(w, 'fail') < c.shipFail / 100) {
        w.stats.shipFails++
        emit(ctx, 'ship.fail', { order: [o.id] }, { order: o.id, attempt: m.attempt + 1, retriesLeft: Math.max(0, c.retries - m.attempt), compensate: c.compensate })
        retryOrGiveUp(ctx, m, o)
        continue
      }
      done()
      if (o.state === 'paid') {
        o.state = 'shipped'
        w.stats.shipped++
        sample.shipped++
        settle(ctx, o, 'shipped')
      }
      emit(ctx, 'ship.ok', { order: [o.id] }, { order: o.id, waited: w.tick - o.createdAt })
      continue
    }

    // payment.refund — компенсирующее действие саги.
    done()
    if (!o.refunded) {
      o.refunded = true
      o.state = 'cancelled'
      w.stats.cancelled++
      sample.cancelled++
      settle(ctx, o, 'cancelled')
    }
  }
}

function retryOrGiveUp(ctx: Ctx, m: Message, o: Order): void {
  const w = ctx.world
  const c = w.config
  if (m.attempt < c.retries) {
    w.stats.retries++
    w.retryQ.push({ msg: { ...m, attempt: m.attempt + 1 }, at: w.tick + c.retryAfter })
    emit(ctx, 'pay.retry', { order: [o.id] }, { order: o.id, kind: m.kind, attempt: m.attempt + 2, after: c.retryAfter })
    return
  }
  w.stats.giveups++
  emit(ctx, 'pay.giveup', { order: [o.id] }, { order: o.id, kind: m.kind, attempts: m.attempt + 1 })
  if (m.kind === 'order.paid' && c.compensate) {
    w.stats.compensations++
    emit(ctx, 'saga.compensate', { order: [o.id] }, { order: o.id, step: 'доставка', action: 'вернуть деньги' })
    if (c.publish === 'outbox') w.outbox.push({ id: w.nextMsg++, order: o.id, kind: 'payment.refund', writtenAt: w.tick, sentAt: null })
    else push(ctx, 'payment.refund', o.id)
    return
  }
  w.stats.stuck++
  emit(ctx, 'saga.stuck', { order: [o.id] }, { order: o.id, state: o.state, reason: 'giveup' })
}

function retryDue(ctx: Ctx): void {
  const w = ctx.world
  const due = w.retryQ.filter((r) => r.at <= w.tick)
  if (due.length === 0) return
  w.retryQ = w.retryQ.filter((r) => r.at > w.tick)
  for (const r of due) w.broker.push({ ...r.msg, arriveAt: w.tick })
}

/* ─────────────────────────────── заказы ─────────────────────────────── */

function create(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const c = w.config
  const n = poisson(w, c.rate * w.x)
  for (let i = 0; i < n; i++) {
    const o: Order = { id: w.nextId++, createdAt: w.tick, state: 'created', published: false, charges: 0, refunded: false, settledAt: null }
    w.orders.push(o)
    w.stats.orders++
    sample.created++
    emit(ctx, 'order.created', { order: [o.id] }, { order: o.id, publish: c.publish })
    if (c.publish === 'outbox') {
      w.outbox.push({ id: w.nextMsg++, order: o.id, kind: 'order.created', writtenAt: w.tick, sentAt: null })
      emit(ctx, 'outbox.write', { order: [o.id] }, { order: o.id, rows: w.outbox.filter((r) => r.sentAt === null).length })
    }
  }
}

/* ─────────────────────────────── учёт ─────────────────────────────── */

function account(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  sample.inFlight = inFlight(w)
  sample.inBroker = w.broker.length + w.work.length + w.retryQ.length
  sample.outbox = w.outbox.filter((r) => r.sentAt === null).length
  w.stats.inconsistentTicks += paidNotShipped(w)
  ctx.log.series.push(sample)
}
