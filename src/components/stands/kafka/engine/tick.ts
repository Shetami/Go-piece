import type { Batch, Consumer, Group, KafkaWorld, Msg, Partition, Producer, Rec } from './types.ts'
import type { Ctx } from './world.ts'
import { commonPrefix, emit, getRec, leaderLog, leo, partitionForKey, send, trace } from './world.ts'
import type { Rng } from './rng.ts'

/**
 * Один шаг всего кластера.
 *
 * Порядок фиксирован: сначала сбои из сценария, потом доставка сообщений,
 * пришедших по сети, потом контроллер, продюсеры, репликация и потребители.
 * Всё, что отправлено в этом тике, приедет не раньше следующего: сеть в модели
 * не бывает мгновенной, и именно поэтому у каждого шага пути есть свой тик.
 */
export function tick(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  if (w.finished) return
  w.tick++

  phaseFaults(ctx)
  phaseNetwork(ctx)
  phaseController(ctx)
  for (const p of w.producers) phaseProducer(ctx, p, rng)
  for (const p of w.partitions) phaseReplication(ctx, p)
  phaseLossCheck(ctx)
  for (const g of w.groups) phaseGroup(ctx, g)
  for (const c of w.consumers) phaseConsumer(ctx, c, rng)
  phaseFinish(ctx)
}

const RETRY_BACKOFF = 2
/** Сколько записей лидер отдаёт фолловеру за один запрос. */
const REPL_MAX = 8

/* ─────────────────────────────── сбои ─────────────────────────────── */

function phaseFaults(ctx: Ctx): void {
  const w = ctx.world
  for (const f of ctx.scenario.faults) {
    if (f.at !== w.tick) continue
    switch (f.kind) {
      case 'broker.down': {
        const b = w.brokers[f.broker]
        if (!b || !b.alive) break
        b.alive = false
        b.detectAt = w.tick + w.config.electionDelay
        const leads = w.partitions.filter((p) => p.leader === b.id).map((p) => p.id)
        for (const p of w.partitions) {
          if (p.leader === b.id) p.pending = []
        }
        emit(ctx, 'broker.down', { broker: [b.id], partition: leads }, {
          broker: b.id,
          leads,
          detectIn: w.config.electionDelay,
        })
        break
      }
      case 'broker.up': {
        const b = w.brokers[f.broker]
        if (!b || b.alive) break
        brokerUp(ctx, b.id)
        break
      }
      case 'broker.slow': {
        const b = w.brokers[f.broker]
        if (!b) break
        b.slow = Math.max(1, f.factor)
        const follows = w.partitions.filter((p) => p.replicas.includes(b.id) && p.leader !== b.id).map((p) => p.id)
        emit(ctx, 'broker.slow', { broker: [b.id], partition: follows }, { broker: b.id, factor: b.slow, follows })
        break
      }
      case 'consumer.crash': {
        const c = w.consumers.find((x) => x.name === f.consumer)
        if (!c || c.state === 'dead') break
        const lostInMemory = c.buffer.length + (c.processing ? 1 : 0)
        const wasActive = c.state === 'active'
        c.state = 'dead'
        c.deadDetectAt = wasActive ? w.tick + w.config.sessionTimeout : w.tick
        const partitions = [...c.assigned]
        const unprocessed = [...(c.processing ? [c.processing.rec] : []), ...c.buffer.map((x) => x.rec)]
        c.buffer = []
        c.processing = null
        c.fetching = []
        emit(ctx, 'consumer.crash', { consumer: [c.id], partition: partitions, rec: unprocessed }, {
          consumer: c.name,
          group: c.group,
          partitions,
          lostInMemory,
          detectIn: w.config.sessionTimeout,
          mode: w.config.commitMode,
        })
        break
      }
      case 'drop.response': {
        const p = w.producers.find((x) => x.name === f.producer)
        if (p) p.dropNext++
        break
      }
    }
  }
}

function brokerUp(ctx: Ctx, id: number): void {
  const w = ctx.world
  const b = w.brokers[id]!
  b.alive = true
  b.slow = 1
  const detected = b.detectAt === null || b.detectAt <= w.tick
  b.detectAt = null

  const truncated: { partition: number; count: number }[] = []
  const led: number[] = []
  for (const p of w.partitions) {
    if (!p.replicas.includes(id)) continue
    p.fetching[id] = null
    if (!detected) {
      // Брокер вернулся раньше, чем контроллер успел заметить: для кластера он и не пропадал.
      if (p.electAt !== null && p.leader === id) p.electAt = null
      continue
    }
    if (p.leader === null && p.isr.includes(id)) {
      // Партиция была без лидера и ждала именно эту реплику — последнюю из ISR.
      elect(ctx, p, id, false, null)
      led.push(p.id)
      continue
    }
    if (p.leader !== null && p.leader !== id) {
      const n = truncateTo(ctx, p, id, leaderLog(p)!)
      if (n > 0) truncated.push({ partition: p.id, count: n })
      p.followerLeo[id] = leo(p, id)
    }
  }
  emit(ctx, 'broker.up', { broker: [id], partition: [...truncated.map((t) => t.partition), ...led] }, {
    broker: id,
    truncated,
    led,
  })
}

/**
 * Реплика обрезает свой лог до общего префикса с лидером. Так Kafka по эпохам
 * лидера находит место, где истории разошлись: всё после него — записи, которых
 * у нового лидера нет, и их больше не существует.
 */
function truncateTo(ctx: Ctx, p: Partition, broker: number, leader: { rec: number; epoch: number }[]): number {
  const rep = p.logs[broker]
  if (!rep) return 0
  const keep = commonPrefix(rep.log, leader)
  const cut = rep.log.slice(keep)
  if (cut.length === 0) return 0
  rep.log = rep.log.slice(0, keep)
  rep.hw = Math.min(rep.hw, keep)
  p.stats.truncated += cut.length
  for (const e of cut) {
    const r = getRec(ctx.world, e.rec)
    trace(ctx, r, 'truncated', { broker, partition: p.id, offset: keep + cut.indexOf(e) })
  }
  emit(ctx, 'log.truncate', { broker: [broker], partition: [p.id], rec: cut.map((e) => e.rec) }, {
    broker,
    partition: p.id,
    from: keep,
    count: cut.length,
    recs: cut.map((e) => e.rec),
    leader: p.leader,
  })
  return cut.length
}

/* ─────────────────────────────── сеть ─────────────────────────────── */

function phaseNetwork(ctx: Ctx): void {
  const w = ctx.world
  const due = w.net.filter((m) => m.deliverAt <= w.tick).sort((a, b) => a.id - b.id)
  if (due.length === 0) return
  w.net = w.net.filter((m) => m.deliverAt > w.tick)
  for (const m of due) deliver(ctx, m)
}

function deliver(ctx: Ctx, m: Msg): void {
  const w = ctx.world
  switch (m.kind) {
    case 'produce':
      // Мёртвый брокер запрос не получит и не ответит: продюсер узнает об этом только по таймауту.
      if (w.brokers[m.broker]?.alive) handleProduce(ctx, m)
      else if (w.config.acks === 0) {
        // Ответа никто не ждёт, но соединение оборвалось — это продюсер заметит
        // и обновит метаданные. Сообщения, ушедшие в оборванное соединение, пропали.
        loseUnacknowledged(ctx, m.batch, 'брокер недоступен')
        w.producers[m.producer]!.metadata[m.partition] = null
      }
      break
    case 'produce-resp':
      handleProduceResp(ctx, m)
      break
    case 'repl-fetch':
      if (w.brokers[m.broker]?.alive) handleReplFetch(ctx, m)
      break
    case 'repl-resp':
      handleReplResp(ctx, m)
      break
    case 'fetch':
      if (w.brokers[m.broker]?.alive) handleFetch(ctx, m)
      break
    case 'fetch-resp':
      handleFetchResp(ctx, m)
      break
    case 'commit':
      handleCommit(ctx, m)
      break
  }
}

/* ───────────────────────── брокер: запись ───────────────────────── */

function handleProduce(ctx: Ctx, m: Extract<Msg, { kind: 'produce' }>): void {
  const w = ctx.world
  const cfg = w.config
  const p = w.partitions[m.partition]!
  const batch = w.batches[m.batch - 1]!
  const reply = (error: string | null, baseOffset: number | null) => {
    if (cfg.acks === 0) return
    send(ctx, {
      kind: 'produce-resp',
      producer: m.producer,
      broker: m.broker,
      batch: m.batch,
      partition: p.id,
      req: m.id,
      error,
      baseOffset,
    }, cfg.netLatency)
  }

  if (p.leader !== m.broker) {
    if (cfg.acks === 0) loseUnacknowledged(ctx, batch.id, `B${m.broker} не лидер партиции`)
    reply('NOT_LEADER_OR_FOLLOWER', null)
    return
  }
  if (cfg.acks === 'all' && p.isr.length < cfg.minInsyncReplicas) {
    reply('NOT_ENOUGH_REPLICAS', null)
    return
  }

  const log = p.logs[m.broker]!.log
  const already = batch.recs.every((id) => log.some((e) => e.rec === id))

  if (already && cfg.idempotence) {
    // Брокер помнит последний номер последовательности этого продюсера в партиции:
    // пакет с тем же номером — повтор, писать его второй раз нельзя.
    const base = log.findIndex((e) => e.rec === batch.recs[0])
    const last = base + batch.recs.length - 1
    w.stats.dedups++
    for (const id of batch.recs) trace(ctx, getRec(w, id), 'dedup', { broker: m.broker, offset: base + batch.recs.indexOf(id), attempt: batch.attempts })
    emit(ctx, 'produce.dedup', { broker: [m.broker], partition: [p.id], producer: [m.producer], rec: batch.recs }, {
      broker: m.broker,
      partition: p.id,
      batch: batch.id,
      recs: batch.recs,
      baseOffset: base,
      attempt: batch.attempts,
    })
    if (cfg.acks === 'all' && p.hw <= last) {
      p.pending.push({ producer: m.producer, batch: batch.id, req: m.id, lastOffset: last, baseOffset: base, dedup: true })
    } else reply(null, base)
    return
  }

  const base = log.length
  for (const id of batch.recs) {
    const r = getRec(w, id)
    const offset = log.length
    log.push({ rec: id, epoch: p.epoch })
    p.stats.appended++
    if (already) {
      p.stats.duplicates++
      w.stats.duplicates++
      trace(ctx, r, 'duplicate', { broker: m.broker, partition: p.id, offset, first: r.offset })
    } else {
      r.appended = true
      r.offset = offset
      trace(ctx, r, 'append', { broker: m.broker, partition: p.id, offset, epoch: p.epoch, attempt: batch.attempts })
    }
  }
  if (!already) p.caughtUpAt[m.broker] = w.tick
  const last = base + batch.recs.length - 1
  emit(ctx, already ? 'produce.duplicate' : 'produce.append', {
    broker: [m.broker],
    partition: [p.id],
    producer: [m.producer],
    rec: batch.recs,
  }, {
    broker: m.broker,
    partition: p.id,
    batch: batch.id,
    recs: batch.recs,
    from: base,
    to: last,
    epoch: p.epoch,
    acks: cfg.acks,
    attempt: batch.attempts,
    isr: [...p.isr],
    first: already ? getRec(w, batch.recs[0]!).offset : null,
  })

  if (cfg.acks === 1) reply(null, base)
  else if (cfg.acks === 'all') {
    p.pending.push({ producer: m.producer, batch: batch.id, req: m.id, lastOffset: last, baseOffset: base, dedup: false })
  }
}

/** acks=0: запрос не дошёл, а продюсер давно считает сообщения отправленными. */
function loseUnacknowledged(ctx: Ctx, batchId: number, why: string): void {
  const w = ctx.world
  const batch = w.batches[batchId - 1]!
  const lost: number[] = []
  for (const id of batch.recs) {
    const r = getRec(w, id)
    if (r.appended || r.state === 'lost') continue
    r.state = 'lost'
    w.stats.lost++
    lost.push(id)
    trace(ctx, r, 'lost', { why, acks: 0 })
  }
  if (lost.length > 0) {
    emit(ctx, 'rec.lost', { rec: lost, partition: [batch.partition] }, {
      partition: batch.partition,
      recs: lost,
      why,
      acks: 0,
    })
  }
}

/* ─────────────────────── продюсер: ответы ─────────────────────── */

function handleProduceResp(ctx: Ctx, m: Extract<Msg, { kind: 'produce-resp' }>): void {
  const w = ctx.world
  const pr = w.producers[m.producer]!
  const inf = pr.inFlight.find((x) => x.req === m.req)
  // Ответ на запрос, который продюсер уже списал по таймауту: соединение закрыто, ответ выброшен.
  if (!inf) return
  const batch = w.batches[m.batch - 1]!

  if (pr.dropNext > 0) {
    pr.dropNext--
    for (const id of batch.recs) trace(ctx, getRec(w, id), 'resp-lost', { broker: m.broker, error: m.error })
    emit(ctx, 'resp.lost', { producer: [pr.id], broker: [m.broker], partition: [m.partition], rec: batch.recs }, {
      producer: pr.name,
      broker: m.broker,
      partition: m.partition,
      batch: batch.id,
      recs: batch.recs,
      wasError: m.error,
      timeoutIn: inf.sentTick + w.config.requestTimeout - w.tick,
    })
    return
  }

  pr.inFlight = pr.inFlight.filter((x) => x !== inf)

  if (m.error === null) {
    batch.state = 'done'
    const latencies: number[] = []
    for (const id of batch.recs) {
      const r = getRec(w, id)
      if (r.state === 'lost') continue
      r.state = 'acked'
      r.ackedTick = w.tick
      latencies.push(w.tick - r.createdTick)
      w.stats.acked++
      w.stats.ackLatency += w.tick - r.createdTick
      w.stats.ackCount++
      pr.stats.acked++
      trace(ctx, r, 'ack', { broker: m.broker, offset: r.offset, latency: w.tick - r.createdTick, acks: w.config.acks, attempt: batch.attempts })
    }
    emit(ctx, 'produce.ack', { producer: [pr.id], partition: [m.partition], rec: batch.recs, broker: [m.broker] }, {
      producer: pr.name,
      partition: m.partition,
      batch: batch.id,
      recs: batch.recs,
      baseOffset: m.baseOffset,
      acks: w.config.acks,
      latency: Math.max(0, ...latencies),
      attempt: batch.attempts,
    })
    return
  }

  retry(ctx, pr, batch, m.error, m.broker)
}

function retry(ctx: Ctx, pr: Producer, batch: Batch, error: string, broker: number): void {
  const w = ctx.world
  const p = w.partitions[batch.partition]!
  pr.metadata[batch.partition] = p.leader
  batch.state = 'ready'
  batch.retryAt = w.tick + RETRY_BACKOFF
  // Повтор встаёт в голову очереди: пакеты одной партиции должны уйти по порядку.
  pr.ready = [batch.id, ...pr.ready.filter((x) => x !== batch.id)].sort((a, b) => a - b)
  pr.stats.retries++
  w.stats.retries++
  for (const id of batch.recs) {
    const r = getRec(w, id)
    r.state = 'batched'
    trace(ctx, r, 'retry', { error, broker, attempt: batch.attempts, leader: p.leader })
  }
  emit(ctx, 'produce.error', { producer: [pr.id], broker: [broker], partition: [batch.partition], rec: batch.recs }, {
    producer: pr.name,
    broker,
    partition: batch.partition,
    batch: batch.id,
    recs: batch.recs,
    error,
    attempt: batch.attempts,
    leader: p.leader,
    isr: [...p.isr],
    minIsr: w.config.minInsyncReplicas,
    age: w.tick - batch.createdTick,
    deliveryTimeout: w.config.deliveryTimeout,
  })
}

/* ─────────────────────────── продюсер ─────────────────────────── */

function phaseProducer(ctx: Ctx, pr: Producer, rng: Rng): void {
  const w = ctx.world
  const cfg = w.config
  const spec = ctx.scenario.producers[pr.spec]!

  // 1. Таймауты запросов: ответа нет — считаем запрос проваленным и повторяем.
  for (const inf of [...pr.inFlight]) {
    if (w.tick - inf.sentTick < cfg.requestTimeout) continue
    pr.inFlight = pr.inFlight.filter((x) => x !== inf)
    retry(ctx, pr, w.batches[inf.batch - 1]!, 'REQUEST_TIMEOUT', inf.broker)
  }

  // 2. delivery.timeout.ms: пакет, который так и не удалось доставить, отдаётся приложению как ошибка.
  for (const id of [...pr.ready, ...Object.values(pr.open)]) {
    const b = w.batches[id - 1]!
    if (w.tick - b.createdTick < cfg.deliveryTimeout) continue
    pr.ready = pr.ready.filter((x) => x !== id)
    if (pr.open[b.partition] === id) delete pr.open[b.partition]
    b.state = 'failed'
    pr.stats.failed += b.recs.length
    for (const rid of b.recs) {
      const r = getRec(w, rid)
      r.state = 'failed'
      w.stats.failed++
      trace(ctx, r, 'failed', { attempts: b.attempts, age: w.tick - b.createdTick })
    }
    emit(ctx, 'produce.failed', { producer: [pr.id], partition: [b.partition], rec: b.recs }, {
      producer: pr.name,
      partition: b.partition,
      batch: b.id,
      recs: b.recs,
      attempts: b.attempts,
      age: w.tick - b.createdTick,
    })
  }

  // 3. Приложение вызывает send().
  while (pr.created < spec.messages && w.tick >= pr.nextSendAt) {
    createRecord(ctx, pr, spec.keys ?? [], spec.weights, rng)
    pr.nextSendAt += rng.duration(spec.every)
  }

  // 4. Sender. Сначала повторы и полные пакеты, потом открытые, чей linger истёк.
  //    Открытый пакет принимает новые записи до той самой секунды, пока Sender его
  //    не заберёт, — поэтому под нагрузкой пакеты растут даже при linger.ms=0.
  const lingered = Object.values(pr.open)
    .map((id) => w.batches[id - 1]!)
    .filter((b) => w.tick - b.createdTick >= cfg.lingerTicks)
    .map((b) => b.id)
  const blocked = new Set<number>()
  for (const id of [...pr.ready, ...lingered]) {
    const b = w.batches[id - 1]!
    if (b.state === 'done' || b.state === 'failed' || b.state === 'in-flight') continue
    // С идемпотентностью пакеты одной партиции уходят строго по порядку: ждущий повтора держит остальных.
    if (cfg.idempotence && blocked.has(b.partition)) continue
    if (b.retryAt > w.tick) {
      blocked.add(b.partition)
      continue
    }
    let leader = pr.metadata[b.partition] ?? null
    if (leader === null) {
      // Метаданные пусты: спрашиваем кластер заново. Пока лидера нет, отправлять некуда.
      leader = w.partitions[b.partition]!.leader
      pr.metadata[b.partition] = leader
      if (leader === null) {
        blocked.add(b.partition)
        continue
      }
    }
    const busy = pr.inFlight.filter((x) => x.broker === leader).length
    if (busy >= cfg.maxInFlight) {
      blocked.add(b.partition)
      continue
    }

    if (b.state === 'open') seal(ctx, pr, b, 'linger')
    pr.ready = pr.ready.filter((x) => x !== id)
    b.attempts++
    b.state = 'in-flight'
    pr.stats.requests++
    w.stats.requests++
    w.stats.recsSent += b.recs.length
    send(ctx, { kind: 'produce', producer: pr.id, broker: leader, batch: b.id, partition: b.partition }, cfg.netLatency)
    const req = w.nextMsgId - 1
    for (const rid of b.recs) {
      const r = getRec(w, rid)
      r.state = 'in-flight'
      trace(ctx, r, 'request', { broker: leader, attempt: b.attempts, size: b.recs.length, acks: cfg.acks, inFlight: busy + 1 })
    }
    emit(ctx, 'produce.request', { producer: [pr.id], broker: [leader], partition: [b.partition], rec: b.recs }, {
      producer: pr.name,
      broker: leader,
      partition: b.partition,
      batch: b.id,
      recs: b.recs,
      attempt: b.attempts,
      acks: cfg.acks,
      inFlight: busy + 1,
      maxInFlight: cfg.maxInFlight,
    })

    if (cfg.acks === 0) {
      // Никто ничего не подтверждает: продюсер считает сообщения доставленными в момент отправки.
      b.state = 'done'
      for (const rid of b.recs) {
        const r = getRec(w, rid)
        r.state = 'acked'
        r.ackedTick = w.tick
        w.stats.acked++
        pr.stats.acked++
        trace(ctx, r, 'ack', { acks: 0, latency: w.tick - r.createdTick, offset: null })
      }
      emit(ctx, 'produce.ack', { producer: [pr.id], partition: [b.partition], rec: b.recs }, {
        producer: pr.name,
        partition: b.partition,
        batch: b.id,
        recs: b.recs,
        baseOffset: null,
        acks: 0,
        latency: 0,
        attempt: 1,
      })
    } else {
      pr.inFlight.push({ batch: b.id, broker: leader, req, sentTick: w.tick })
    }
  }
}

function createRecord(ctx: Ctx, pr: Producer, keys: string[], weights: number[] | undefined, rng: Rng): void {
  const w = ctx.world
  const cfg = w.config
  const key = keys.length > 0 ? keys[rng.weighted(keys.length, weights)]! : null
  const partition = key === null ? pr.sticky : partitionForKey(key, cfg.partitions)

  let bid = pr.open[partition]
  let batch = bid !== undefined ? w.batches[bid - 1] : undefined
  if (!batch) {
    batch = {
      id: w.batches.length + 1,
      producer: pr.id,
      partition,
      recs: [],
      createdTick: w.tick,
      sealedTick: null,
      sealReason: null,
      attempts: 0,
      retryAt: 0,
      state: 'open',
    }
    w.batches.push(batch)
    pr.open[partition] = batch.id
    bid = batch.id
  }

  const rec: Rec = {
    id: w.recs.length + 1,
    producer: pr.id,
    key,
    partition,
    batch: batch.id,
    createdTick: w.tick,
    state: 'batched',
    appended: false,
    offset: null,
    committedTick: null,
    ackedTick: null,
    processed: {},
    skipped: [],
    trace: [],
  }
  w.recs.push(rec)
  batch.recs.push(rec.id)
  pr.created++
  w.stats.produced++

  trace(ctx, rec, 'send', { producer: pr.name, key })
  trace(ctx, rec, 'partition', { partition, key, how: key === null ? 'sticky' : 'hash', partitions: cfg.partitions })
  trace(ctx, rec, 'batch', { batch: batch.id, size: batch.recs.length, batchSize: cfg.batchSize })
  emit(ctx, 'rec.send', { rec: [rec.id], producer: [pr.id], partition: [partition] }, {
    rec: rec.id,
    producer: pr.name,
    key,
    partition,
    how: key === null ? 'sticky' : 'hash',
    batch: batch.id,
    inBatch: batch.recs.length,
  })

  if (batch.recs.length >= cfg.batchSize) seal(ctx, pr, batch, 'size')
}

function seal(ctx: Ctx, pr: Producer, b: Batch, reason: 'size' | 'linger'): void {
  const w = ctx.world
  b.state = 'ready'
  b.sealedTick = w.tick
  b.sealReason = reason
  delete pr.open[b.partition]
  pr.ready.push(b.id)
  pr.ready.sort((x, y) => x - y)
  // Липкий партиционер: пакет закрыт — следующие сообщения без ключа пойдут в другую партицию.
  if (b.partition === pr.sticky) pr.sticky = (pr.sticky + 1) % w.config.partitions
  for (const id of b.recs) trace(ctx, getRec(w, id), 'seal', { batch: b.id, reason, size: b.recs.length, waited: w.tick - b.createdTick })
  emit(ctx, 'batch.seal', { producer: [pr.id], partition: [b.partition], rec: b.recs }, {
    producer: pr.name,
    partition: b.partition,
    batch: b.id,
    recs: b.recs,
    size: b.recs.length,
    reason,
    waited: w.tick - b.createdTick,
    batchSize: w.config.batchSize,
    linger: w.config.lingerTicks,
  })
}

/* ─────────────────────────── контроллер ─────────────────────────── */

function phaseController(ctx: Ctx): void {
  const w = ctx.world
  for (const b of w.brokers) {
    if (b.alive || b.detectAt !== w.tick) continue
    // Контроллер заметил смерть брокера: вычеркнуть из ISR и выбрать новых лидеров.
    for (const p of w.partitions) {
      if (!p.replicas.includes(b.id)) continue
      if (p.leader === b.id) {
        chooseLeader(ctx, p, b.id)
      } else if (p.isr.includes(b.id)) {
        p.isr = p.isr.filter((x) => x !== b.id)
        emit(ctx, 'isr.shrink', { broker: [b.id], partition: [p.id] }, {
          partition: p.id,
          broker: b.id,
          reason: 'dead',
          isr: [...p.isr],
          minIsr: w.config.minInsyncReplicas,
        })
        updateHw(ctx, p)
      }
    }
  }
}

function chooseLeader(ctx: Ctx, p: Partition, dead: number): void {
  const w = ctx.world
  const alive = (x: number) => w.brokers[x]!.alive && x !== dead
  const clean = p.replicas.find((x) => alive(x) && p.isr.includes(x))
  if (clean !== undefined) {
    elect(ctx, p, clean, false, dead)
    return
  }
  const any = p.replicas.find(alive)
  if (any !== undefined && w.config.uncleanElection) {
    elect(ctx, p, any, true, dead)
    return
  }
  // Выбрать некого: партиция недоступна, пока не вернётся кто-то из ISR.
  p.leader = null
  p.pending = []
  // Последнего члена ISR Kafka запоминает, чтобы знать, кого ждать.
  emit(ctx, 'leader.elect', { partition: [p.id], broker: [dead] }, {
    partition: p.id,
    from: dead,
    to: null,
    epoch: p.epoch,
    unclean: false,
    isr: [...p.isr],
    candidates: p.replicas.filter(alive),
  })
}

function elect(ctx: Ctx, p: Partition, to: number, unclean: boolean, from: number | null): void {
  const w = ctx.world
  const old = from !== null ? (p.logs[from]?.log ?? []) : []
  p.leader = to
  p.epoch++
  p.electAt = null
  p.pending = []
  if (from !== null) p.isr = p.isr.filter((x) => x !== from)
  if (unclean) p.isr = [to]
  if (!p.isr.includes(to)) p.isr = [to, ...p.isr]
  const rep = p.logs[to]!
  p.hw = Math.min(rep.hw, rep.log.length)
  w.stats.elections++

  const missing = old.slice(commonPrefix(old, rep.log)).map((e) => e.rec)
  emit(ctx, 'leader.elect', { partition: [p.id], broker: [to, ...(from !== null ? [from] : [])], rec: missing }, {
    partition: p.id,
    from,
    to,
    epoch: p.epoch,
    unclean,
    isr: [...p.isr],
    leo: rep.log.length,
    hw: p.hw,
    missing,
  })

  // Живые фолловеры обрезают то, чего нет у нового лидера.
  for (const f of p.replicas) {
    if (f === to) continue
    p.fetching[f] = null
    if (!w.brokers[f]!.alive) continue
    truncateTo(ctx, p, f, rep.log)
    p.followerLeo[f] = leo(p, f)
    p.caughtUpAt[f] = w.tick
    p.lastFetchLeo[f] = leo(p, f)
    p.lastFetchAt[f] = w.tick
  }
}

/* ─────────────────────────── репликация ─────────────────────────── */

function phaseReplication(ctx: Ctx, p: Partition): void {
  const w = ctx.world
  const cfg = w.config
  const L = p.leader
  if (L === null || !w.brokers[L]!.alive) return

  // Фолловеры тянут данные сами — лидер ничего не рассылает.
  for (const f of p.replicas) {
    if (f === L || !w.brokers[f]!.alive) continue
    const target: number | null | undefined = p.fetching[f]
    if (target !== null && target !== undefined) {
      if (target === L && w.brokers[target]!.alive) continue
      p.fetching[f] = null
    }
    p.fetching[f] = L
    send(ctx, { kind: 'repl-fetch', from: f, broker: L, partition: p.id, fetchOffset: leo(p, f), epoch: p.epoch }, cfg.netLatency * w.brokers[f]!.slow)
  }

  // Фолловер, который давно не догонял лидера, вылетает из ISR.
  const leaderLeo = leo(p, L)
  for (const f of [...p.isr]) {
    if (f === L || !w.brokers[f]!.alive) continue
    const behind = leaderLeo - (p.followerLeo[f] ?? 0)
    if (behind > 0 && w.tick - (p.caughtUpAt[f] ?? 0) > cfg.replicaLagMax) {
      p.isr = p.isr.filter((x) => x !== f)
      emit(ctx, 'isr.shrink', { broker: [f], partition: [p.id] }, {
        partition: p.id,
        broker: f,
        reason: 'lag',
        behind,
        lagged: w.tick - (p.caughtUpAt[f] ?? 0),
        isr: [...p.isr],
        minIsr: cfg.minInsyncReplicas,
      })
    }
  }

  updateHw(ctx, p)
}

function handleReplFetch(ctx: Ctx, m: Extract<Msg, { kind: 'repl-fetch' }>): void {
  const w = ctx.world
  const p = w.partitions[m.partition]!
  if (p.leader !== m.broker || m.epoch !== p.epoch) return
  const log = p.logs[m.broker]!.log
  const f = m.from
  p.followerLeo[f] = m.fetchOffset
  // Правило Kafka (Replica.updateFetchState): фолловер «догнал», если дочитал
  // до конца лога прямо сейчас — или хотя бы до того конца, что был у лидера
  // при его прошлом запросе. Иначе под постоянной записью не догонял бы никто.
  if (m.fetchOffset >= log.length) p.caughtUpAt[f] = w.tick
  else if (m.fetchOffset >= (p.lastFetchLeo[f] ?? 0)) p.caughtUpAt[f] = Math.max(p.caughtUpAt[f] ?? 0, p.lastFetchAt[f] ?? 0)
  p.lastFetchLeo[f] = log.length
  p.lastFetchAt[f] = w.tick

  if (!p.isr.includes(f) && w.brokers[f]!.alive && m.fetchOffset >= p.hw && m.fetchOffset >= log.length) {
    p.isr = [...p.isr, f].sort((a, b) => p.replicas.indexOf(a) - p.replicas.indexOf(b))
    emit(ctx, 'isr.expand', { broker: [f], partition: [p.id] }, {
      partition: p.id,
      broker: f,
      isr: [...p.isr],
      leo: m.fetchOffset,
    })
  }
  updateHw(ctx, p)

  send(ctx, {
    kind: 'repl-resp',
    from: m.broker,
    broker: f,
    partition: p.id,
    fetchOffset: m.fetchOffset,
    entries: log.slice(m.fetchOffset, m.fetchOffset + REPL_MAX).map((e) => ({ ...e })),
    hw: p.hw,
    epoch: p.epoch,
  }, w.config.netLatency * w.brokers[f]!.slow)
}

function handleReplResp(ctx: Ctx, m: Extract<Msg, { kind: 'repl-resp' }>): void {
  const w = ctx.world
  const p = w.partitions[m.partition]!
  const f = m.broker
  if (p.fetching[f] === m.from) p.fetching[f] = null
  if (!w.brokers[f]!.alive || p.leader !== m.from || m.epoch !== p.epoch) return
  const rep = p.logs[f]!
  if (rep.log.length !== m.fetchOffset) return

  for (const e of m.entries) {
    const offset = rep.log.length
    rep.log.push({ ...e })
    const r = getRec(w, e.rec)
    trace(ctx, r, 'replicate', { broker: f, partition: p.id, offset, leader: m.from })
  }
  rep.hw = Math.min(m.hw, rep.log.length)
  if (m.entries.length > 0) {
    emit(ctx, 'repl.fetch', { broker: [f, m.from], partition: [p.id], rec: m.entries.map((e) => e.rec) }, {
      partition: p.id,
      follower: f,
      leader: m.from,
      from: m.fetchOffset,
      count: m.entries.length,
      recs: m.entries.map((e) => e.rec),
      leo: rep.log.length,
      slow: w.brokers[f]!.slow,
    })
  }
}

/**
 * High watermark = минимальный LEO среди реплик ISR, как его знает лидер.
 * Всё ниже HW есть на всех синхронных репликах: это и есть «закоммичено».
 */
function updateHw(ctx: Ctx, p: Partition): void {
  const w = ctx.world
  const L = p.leader
  if (L === null) return
  const log = p.logs[L]!.log
  let hw = log.length
  for (const f of p.isr) {
    if (f === L) continue
    hw = Math.min(hw, p.followerLeo[f] ?? 0)
  }
  if (hw > p.hw) {
    const from = p.hw
    p.hw = hw
    p.logs[L]!.hw = hw
    const recs: number[] = []
    for (let o = from; o < hw; o++) {
      const e = log[o]!
      const r = getRec(w, e.rec)
      recs.push(e.rec)
      if (r.committedTick === null && r.offset === o) {
        r.committedTick = w.tick
        w.stats.committed++
        trace(ctx, r, 'commit', { partition: p.id, offset: o, hw, isr: [...p.isr], leader: L })
      }
    }
    emit(ctx, 'hw.advance', { partition: [p.id], broker: [L], rec: recs }, {
      partition: p.id,
      leader: L,
      from,
      to: hw,
      recs,
      isr: [...p.isr],
    })
  }

  // acks=all: запросы в чистилище ждут, пока HW не перейдёт их последнюю запись.
  if (p.pending.length === 0) return
  const done = p.pending.filter((x) => p.hw > x.lastOffset)
  if (done.length === 0) return
  p.pending = p.pending.filter((x) => p.hw <= x.lastOffset)
  // HW перешёл запись, но ISR к этому моменту уже меньше min.insync.replicas:
  // запись лежит в логе, а продюсер всё равно получает ошибку и повторяет.
  const short = w.config.acks === 'all' && p.isr.length < w.config.minInsyncReplicas
  for (const x of done) {
    send(ctx, {
      kind: 'produce-resp',
      producer: x.producer,
      broker: L,
      batch: x.batch,
      partition: p.id,
      req: x.req,
      error: short ? 'NOT_ENOUGH_REPLICAS_AFTER_APPEND' : null,
      baseOffset: x.baseOffset,
    }, w.config.netLatency)
  }
}

/* ────────────────────────── проверка потерь ────────────────────────── */

/**
 * Подтверждённое сообщение, которого больше нет в логе лидера, — потеряно.
 * Продюсер получил «ок» и никогда не узнает об обратном.
 */
function phaseLossCheck(ctx: Ctx): void {
  const w = ctx.world
  const byPart = new Map<number, number[]>()
  for (const r of w.recs) {
    if (r.state !== 'acked' || !r.appended) continue
    const p = w.partitions[r.partition]!
    const log = leaderLog(p)
    if (!log || log.some((e) => e.rec === r.id)) continue
    r.state = 'lost'
    w.stats.lost++
    trace(ctx, r, 'lost', { partition: p.id, leader: p.leader, acks: w.config.acks, ackedTick: r.ackedTick, offset: r.offset })
    byPart.set(p.id, [...(byPart.get(p.id) ?? []), r.id])
  }
  for (const [part, recs] of byPart) {
    emit(ctx, 'rec.lost', { partition: [part], rec: recs }, {
      partition: part,
      recs,
      acks: w.config.acks,
      leader: w.partitions[part]!.leader,
      why: 'нового лидера',
    })
  }
}

/* ─────────────────────────── группы ─────────────────────────── */

function phaseGroup(ctx: Ctx, g: Group): void {
  const w = ctx.world
  const members = w.consumers.filter((c) => c.group === g.name)

  // Новые участники приходят к координатору и запускают ребалансировку.
  const joined: string[] = []
  for (const c of members) {
    if (c.state === 'waiting' && joinTick(ctx, c) === w.tick) {
      joined.push(c.name)
      emit(ctx, 'consumer.join', { consumer: [c.id] }, { consumer: c.name, group: g.name, members: members.filter((x) => x.state !== 'dead').length })
    }
  }
  // Упавший перестал слать heartbeat — через session.timeout координатор его вычёркивает.
  const expired = members.filter((c) => c.state === 'dead' && c.deadDetectAt === w.tick).map((c) => c.name)

  if (joined.length > 0 || expired.length > 0) {
    startRebalance(ctx, g, joined.length > 0 ? `вступил ${joined.join(', ')}` : `истекла сессия ${expired.join(', ')}`)
  }

  if (g.rebalanceUntil === w.tick) finishRebalance(ctx, g)
}

/** Тик, на котором потребитель подключается к группе. */
function joinTick(ctx: Ctx, c: Consumer): number {
  return Math.max(1, ctx.scenario.consumers[c.spec]?.joinAt ?? 1)
}

function startRebalance(ctx: Ctx, g: Group, reason: string): void {
  const w = ctx.world
  g.generation++
  g.rebalanceUntil = w.tick + w.config.rebalanceTicks
  g.rebalanceReason = reason
  // Eager-протокол: все живые участники отдают все партиции. Перед этим
  // коммитят свои оффсеты — это onPartitionsRevoked. Автокоммит и здесь
  // фиксирует позицию, а не обработанное.
  for (const c of w.consumers) {
    if (c.group !== g.name || c.state !== 'active') continue
    const src = w.config.commitMode === 'auto' ? c.positions : c.processedNext
    const offsets: Record<number, number> = {}
    for (const p of c.assigned) {
      const next = src[p]
      if (next !== undefined && next > (g.committed[p] ?? 0)) offsets[p] = next
    }
    applyCommit(ctx, g, c, offsets, 'revoke')
    c.assigned = []
    c.buffer = []
    c.processing = null
    c.fetching = []
    c.state = 'waiting'
  }
}

function finishRebalance(ctx: Ctx, g: Group): void {
  const w = ctx.world
  g.rebalanceUntil = null
  const members = w.consumers.filter((c) => c.group === g.name && c.state === 'waiting' && joinTick(ctx, c) <= w.tick)
  const n = members.length
  const parts = w.partitions.map((p) => p.id)
  const assignment: Record<string, number[]> = {}

  // RangeAssignor для одного топика: партиции подряд, первым участникам — на одну больше.
  const per = n === 0 ? 0 : Math.floor(parts.length / n)
  const extra = n === 0 ? 0 : parts.length % n
  let at = 0
  members.forEach((c, i) => {
    const take = per + (i < extra ? 1 : 0)
    c.assigned = parts.slice(at, at + take)
    at += take
    c.state = 'active'
    c.buffer = []
    c.processing = null
    c.fetching = []
    c.lastAutoCommit = w.tick
    for (const p of c.assigned) {
      c.positions[p] = g.committed[p] ?? 0
      c.processedNext[p] = g.committed[p] ?? 0
    }
    assignment[c.name] = [...c.assigned]
  })
  g.stats.rebalances++
  w.stats.rebalances++

  // Всё, что ниже закоммиченного оффсета и не обработано этой группой, не будет обработано никогда.
  for (const p of w.partitions) {
    const log = leaderLog(p)
    if (!log) continue
    const upto = Math.min(g.committed[p.id] ?? 0, log.length)
    const skipped: number[] = []
    for (let o = 0; o < upto; o++) {
      const r = getRec(w, log[o]!.rec)
      if ((r.processed[g.name] ?? 0) > 0 || r.skipped.includes(g.name)) continue
      r.skipped.push(g.name)
      g.stats.skipped++
      w.stats.skipped++
      skipped.push(r.id)
      trace(ctx, r, 'skipped', { group: g.name, partition: p.id, committed: g.committed[p.id] })
    }
    if (skipped.length > 0) {
      emit(ctx, 'rec.skipped', { partition: [p.id], rec: skipped }, {
        group: g.name,
        partition: p.id,
        recs: skipped,
        committed: g.committed[p.id],
      })
    }
  }

  emit(ctx, 'group.rebalance', { consumer: members.map((c) => c.id), partition: parts }, {
    group: g.name,
    generation: g.generation,
    reason: g.rebalanceReason,
    assignment,
    idle: members.filter((c) => c.assigned.length === 0).map((c) => c.name),
    positions: Object.fromEntries(parts.map((p) => [p, g.committed[p] ?? 0])),
    paused: w.config.rebalanceTicks,
    first: g.stats.rebalances === 1,
  })
}

/* ─────────────────────────── потребители ─────────────────────────── */

function phaseConsumer(ctx: Ctx, c: Consumer, rng: Rng): void {
  const w = ctx.world
  const cfg = w.config
  if (c.state !== 'active') return
  const g = w.groups.find((x) => x.name === c.group)!

  // Обработка: одно сообщение за раз, как в обычном цикле poll → обработать.
  if (c.processing) {
    c.processing.left--
    if (c.processing.left <= 0) finishProcessing(ctx, c, g)
  }
  if (!c.processing && c.buffer.length > 0) {
    const next = c.buffer.shift()!
    c.processing = { ...next, left: rng.duration([cfg.processTicks, cfg.processTicks + 1]) }
  }

  // Автокоммит по таймеру: фиксируется позиция — всё, что poll уже отдал.
  if (cfg.commitMode === 'auto' && w.tick - c.lastAutoCommit >= cfg.autoCommitInterval) {
    c.lastAutoCommit = w.tick
    const offsets: Record<number, number> = {}
    for (const p of c.assigned) {
      const pos = c.positions[p] ?? 0
      if (pos > (g.committed[p] ?? 0)) offsets[p] = pos
    }
    if (Object.keys(offsets).length > 0) send(ctx, { kind: 'commit', consumer: c.id, group: g.name, offsets, mode: 'auto' }, cfg.netLatency)
  }

  // Пачка разобрана — коммитим обработанное и зовём poll снова.
  if (c.processing || c.buffer.length > 0) return
  if (cfg.commitMode === 'after-process') {
    const offsets: Record<number, number> = {}
    for (const p of c.assigned) {
      const next = c.processedNext[p] ?? 0
      if (next > (g.committed[p] ?? 0)) offsets[p] = next
    }
    const pendingCommit = w.net.some((m) => m.kind === 'commit' && m.consumer === c.id)
    if (!pendingCommit && Object.keys(offsets).length > 0) {
      send(ctx, { kind: 'commit', consumer: c.id, group: g.name, offsets, mode: 'after-process' }, cfg.netLatency)
    }
  }

  for (const part of c.assigned) {
    if (c.fetching.includes(part)) {
      // Запрос ушёл брокеру, который умер или перестал быть лидером, — ответа не будет.
      const inflight = w.net.some((m) => m.kind === 'fetch' && m.consumer === c.id && m.partition === part) ||
        w.net.some((m) => m.kind === 'fetch-resp' && m.consumer === c.id && m.partition === part)
      if (inflight) continue
      c.fetching = c.fetching.filter((x) => x !== part)
    }
    const p = w.partitions[part]!
    if (p.leader === null) continue
    c.fetching.push(part)
    send(ctx, { kind: 'fetch', consumer: c.id, broker: p.leader, partition: part, offset: c.positions[part] ?? 0, generation: g.generation }, cfg.netLatency)
  }
}

function finishProcessing(ctx: Ctx, c: Consumer, g: Group): void {
  const w = ctx.world
  const job = c.processing!
  c.processing = null
  const r = getRec(w, job.rec)
  const times = (r.processed[g.name] ?? 0) + 1
  r.processed[g.name] = times
  c.processedNext[job.partition] = job.offset + 1
  c.processedCount++
  w.stats.processed++
  const latency = w.tick - r.createdTick
  if (times === 1) {
    w.stats.e2eLatency += latency
    w.stats.e2eCount++
  }
  trace(ctx, r, times === 1 ? 'process' : 'reprocess', {
    consumer: c.name,
    group: g.name,
    partition: job.partition,
    offset: job.offset,
    latency,
    times,
  })
  emit(ctx, 'consumer.process', { consumer: [c.id], partition: [job.partition], rec: [r.id] }, {
    consumer: c.name,
    group: g.name,
    rec: r.id,
    partition: job.partition,
    offset: job.offset,
    latency,
    times,
  })
  if (times > 1) {
    const dupInLog = job.offset !== r.offset
    g.stats.reprocessed++
    w.stats.reprocessed++
    emit(ctx, 'rec.reprocess', { consumer: [c.id], partition: [job.partition], rec: [r.id] }, {
      consumer: c.name,
      group: g.name,
      rec: r.id,
      partition: job.partition,
      offset: job.offset,
      firstOffset: r.offset,
      times,
      cause: dupInLog ? 'duplicate' : 'rebalance',
    })
  }
}

function handleFetch(ctx: Ctx, m: Extract<Msg, { kind: 'fetch' }>): void {
  const w = ctx.world
  const p = w.partitions[m.partition]!
  const reply = (entries: { partition: number; offset: number; rec: number }[], error: string | null) =>
    send(ctx, {
      kind: 'fetch-resp',
      consumer: m.consumer,
      broker: m.broker,
      partition: m.partition,
      offset: m.offset,
      entries,
      error,
      generation: m.generation,
    }, w.config.netLatency)

  if (p.leader !== m.broker) {
    reply([], 'NOT_LEADER_OR_FOLLOWER')
    return
  }
  // Потребитель видит только закоммиченное — то, что ниже high watermark.
  const log = p.logs[m.broker]!.log
  const upto = Math.min(p.hw, m.offset + w.config.maxPollRecords)
  const entries = []
  for (let o = m.offset; o < upto; o++) entries.push({ partition: p.id, offset: o, rec: log[o]!.rec })
  reply(entries, m.offset > log.length ? 'OFFSET_OUT_OF_RANGE' : null)
}

function handleFetchResp(ctx: Ctx, m: Extract<Msg, { kind: 'fetch-resp' }>): void {
  const w = ctx.world
  const c = w.consumers[m.consumer]!
  c.fetching = c.fetching.filter((x) => x !== m.partition)
  const g = w.groups.find((x) => x.name === c.group)!
  // Ответ на запрос прошлого поколения группы: партиция могла уже уйти другому.
  if (c.state !== 'active' || m.generation !== g.generation || !c.assigned.includes(m.partition)) return
  if (m.error === 'OFFSET_OUT_OF_RANGE') {
    // Лог стал короче, чем позиция: после нечистых выборов такое бывает. auto.offset.reset.
    c.positions[m.partition] = w.partitions[m.partition]!.hw
    return
  }
  if (m.error || m.entries.length === 0) return
  if ((c.positions[m.partition] ?? 0) !== m.offset) return

  for (const e of m.entries) {
    c.buffer.push({ ...e })
    trace(ctx, getRec(w, e.rec), 'fetch', { consumer: c.name, group: g.name, partition: e.partition, offset: e.offset, broker: m.broker })
  }
  c.positions[m.partition] = m.offset + m.entries.length
  const p = w.partitions[m.partition]!
  emit(ctx, 'consumer.fetch', { consumer: [c.id], partition: [m.partition], broker: [m.broker], rec: m.entries.map((e) => e.rec) }, {
    consumer: c.name,
    group: g.name,
    partition: m.partition,
    broker: m.broker,
    from: m.offset,
    count: m.entries.length,
    recs: m.entries.map((e) => e.rec),
    hw: p.hw,
    leo: p.leader === null ? 0 : leo(p, p.leader),
  })
}

function handleCommit(ctx: Ctx, m: Extract<Msg, { kind: 'commit' }>): void {
  const w = ctx.world
  const g = w.groups.find((x) => x.name === m.group)
  const c = w.consumers[m.consumer]
  if (!g || !c) return
  applyCommit(ctx, g, c, m.offsets, m.mode)
}

function applyCommit(ctx: Ctx, g: Group, c: Consumer, offsets: Record<number, number>, mode: string): void {
  const w = ctx.world
  const changed: Record<number, [number, number]> = {}
  const recs: number[] = []
  for (const [k, next] of Object.entries(offsets)) {
    const part = Number(k)
    const prev = g.committed[part] ?? 0
    if (next <= prev) continue
    g.committed[part] = next
    changed[part] = [prev, next]
    const log = leaderLog(w.partitions[part]!) ?? w.partitions[part]!.logs[w.partitions[part]!.replicas[0]!]!.log
    for (let o = prev; o < Math.min(next, log.length); o++) {
      const r = getRec(w, log[o]!.rec)
      recs.push(r.id)
      trace(ctx, r, 'offset', {
        group: g.name,
        consumer: c.name,
        partition: part,
        committed: next,
        mode,
        processed: (r.processed[g.name] ?? 0) > 0,
      })
    }
  }
  if (Object.keys(changed).length === 0) return
  const unprocessed = recs.filter((id) => (getRec(w, id).processed[g.name] ?? 0) === 0)
  emit(ctx, 'offset.commit', { consumer: [c.id], partition: Object.keys(changed).map(Number), rec: recs }, {
    group: g.name,
    consumer: c.name,
    offsets: changed,
    mode,
    count: recs.length,
    unprocessed,
  })
}

/* ─────────────────────────── финиш ─────────────────────────── */

function phaseFinish(ctx: Ctx): void {
  const w = ctx.world
  const allCreated = w.producers.every((p) => p.created >= ctx.scenario.producers[p.spec]!.messages)
  if (!allCreated) return
  const producersIdle = w.producers.every((p) => p.ready.length === 0 && p.inFlight.length === 0 && Object.keys(p.open).length === 0)
  if (!producersIdle) return
  if (w.net.some((m) => m.kind === 'produce' || m.kind === 'produce-resp' || m.kind === 'commit')) return
  if (w.brokers.some((b) => !b.alive && b.detectAt !== null && b.detectAt > w.tick)) return
  if (ctx.scenario.faults.some((f) => f.at > w.tick)) return

  for (const p of w.partitions) {
    if (p.leader === null) continue
    if (leo(p, p.leader) !== p.hw) return
  }
  for (const g of w.groups) {
    if (g.rebalanceUntil !== null) return
    const members = w.consumers.filter((c) => c.group === g.name)
    if (members.some((c) => c.state === 'waiting' && joinTick(ctx, c) > w.tick)) return
    if (members.some((c) => c.state === 'dead' && c.deadDetectAt !== null && c.deadDetectAt >= w.tick)) return
    const active = members.filter((c) => c.state === 'active')
    if (active.length === 0) continue
    if (active.some((c) => c.processing || c.buffer.length > 0)) return
    for (const p of w.partitions) {
      if (p.leader === null) continue
      if ((g.committed[p.id] ?? 0) < p.hw) return
    }
  }
  w.finished = true
  w.finishReason = 'all-done'
}

/** Лаг группы по партиции: сколько закоммиченных сообщений она ещё не подтвердила. */
export function lagOf(w: KafkaWorld, g: Group, partition: number): number {
  const p = w.partitions[partition]!
  return Math.max(0, p.hw - (g.committed[partition] ?? 0))
}
