import type { Chan, ChanGState, ChanPhase, Goroutine, Slot } from './types.ts'
import type { Ctx } from './world.ts'
import {
  chanByName,
  emit,
  enqueue,
  firstWaiter,
  getG,
  park,
  wake,
} from './world.ts'
import type { Rng } from './rng.ts'

/**
 * Один шаг всей системы.
 *
 * Порядок фиксирован: сначала рождаются горутины, потом процессоры разбирают
 * очередь готовых, и только в конце подводится итог тика. Разбуженная горутина
 * попадает в начало очереди — это runnext, и её видно уже в этом же тике.
 */
export function tick(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  if (w.finished) return
  w.tick++
  w.slots = []

  phaseSpawn(ctx)
  phaseRun(ctx, rng)
  phaseAccount(ctx)
}

/* ───────────────────────────── горутины ───────────────────────────── */

function phaseSpawn(ctx: Ctx): void {
  const w = ctx.world
  while (w.spawnPtr < ctx.plan.length) {
    const item = ctx.plan[w.spawnPtr]
    if (!item || item.tick >= w.tick) break
    w.spawnPtr++
    createG(ctx, item.workload)
  }
}

function createG(ctx: Ctx, workload: number): void {
  const w = ctx.world
  const wl = ctx.scenario.workloads[workload]
  if (!wl) throw new Error(`нет нагрузки #${workload}`)
  const g: Goroutine = {
    id: w.nextGid++,
    name: wl.name,
    workload,
    state: 'runnable',
    wait: null,
    phaseIdx: -1,
    phaseLeft: 0,
    opsLeft: 0,
    repeatsLeft: wl.repeat === 'forever' ? Number.POSITIVE_INFINITY : (wl.repeat ?? 1) - 1,
    createdTick: w.tick,
    runTicks: 0,
    waitTicks: 0,
    blocks: 0,
    wakeups: 0,
    sent: 0,
    received: 0,
  }
  w.gs.push(g)
  enqueue(w, g.id)
  emit(ctx, 'g.start', { g: [g.id] }, { name: wl.name, phases: wl.phases.length })
}

/**
 * Перейти к следующей фазе. false — горутина закончилась.
 * Отдельная функция потому, что фазу завершает не только сама горутина:
 * парковку снимает тот, кто передал ей значение.
 */
function advance(ctx: Ctx, g: Goroutine, rng: Rng): boolean {
  const w = ctx.world
  const wl = ctx.scenario.workloads[g.workload]
  if (!wl) throw new Error(`нет нагрузки #${g.workload}`)

  let next = g.phaseIdx + 1
  if (next >= wl.phases.length) {
    if (g.repeatsLeft > 0) {
      g.repeatsLeft--
      next = 0
    } else {
      g.state = 'done'
      g.finishedTick = w.tick
      w.runq = w.runq.filter((x) => x !== g.id)
      emit(ctx, 'g.done', { g: [g.id] }, {
        name: g.name,
        sent: g.sent,
        received: g.received,
        blocks: g.blocks,
        waited: g.waitTicks,
        lived: w.tick - g.createdTick,
      })
      return false
    }
  }
  g.phaseIdx = next
  const ph = wl.phases[next]!
  if (ph.kind === 'cpu') g.phaseLeft = rng.duration(ph.ticks)
  else if (ph.kind === 'close') g.opsLeft = 1
  else g.opsLeft = ph.count ?? 1
  return true
}

/**
 * Операция с каналом завершена — у той горутины, которая её выполняла,
 * и у той, которую при этом разбудили: проснувшийся отправитель не повторяет
 * отправку, его значение уже забрали.
 */
function completeOp(ctx: Ctx, g: Goroutine, rng: Rng): void {
  g.opsLeft--
  if (g.opsLeft > 0) return
  advance(ctx, g, rng)
}

/* ──────────────────────── операции с каналами ──────────────────────── */

/** Паника рантайма: прогон на этом кончается, как кончилась бы программа. */
function panic(ctx: Ctx, text: string): void {
  const w = ctx.world
  w.finished = true
  w.finishReason = 'panic'
  w.panic = text
}

/** Вечная парковка: у nil-канала нет очередей, разбудить из неё некому. */
function parkForever(ctx: Ctx, g: Goroutine, kind: 'send' | 'recv', c: Chan): void {
  park(ctx, g, kind, [], [c.id])
  emit(ctx, 'chan.nil', { g: [g.id], chan: [c.id] }, { chan: c.name, op: kind })
}

function doSend(ctx: Ctx, g: Goroutine, c: Chan, rng: Rng): Slot['op'] {
  const w = ctx.world
  if (c.isNil) {
    parkForever(ctx, g, 'send', c)
    return 'nil'
  }
  c.lockedAt = w.tick

  if (c.closed) {
    emit(ctx, 'send.closed', { g: [g.id], chan: [c.id] }, { chan: c.name, asleep: false })
    panic(ctx, 'send on closed channel')
    return 'panic'
  }

  // 1. Получатель уже ждёт — значение уходит прямо ему, минуя буфер.
  const recvW = firstWaiter(c.recvq)
  if (recvW && (c.cap === 0 || w.config.directHandoff)) {
    const r = getG(w, recvW.g)
    g.sent++
    r.received++
    c.stats.sent++
    c.stats.received++
    c.stats.direct++
    w.stats.transfers++
    w.stats.direct++
    emit(ctx, 'send.direct', { g: [g.id, recvW.g], chan: [c.id] }, {
      chan: c.name,
      to: recvW.g,
      waited: w.tick - recvW.since,
      viaSelect: recvW.fromSelect,
      cap: c.cap,
    })
    wake(ctx, recvW.g, 'значение передано напрямую', c)
    completeOp(ctx, r, rng)
    completeOp(ctx, g, rng)
    return 'send.direct'
  }

  // 2. В буфере есть место.
  if (c.qcount < c.cap) {
    c.buf[c.sendx] = g.id
    c.sendx = (c.sendx + 1) % c.cap
    c.qcount++
    g.sent++
    c.stats.sent++
    c.stats.buffered++
    c.stats.maxQcount = Math.max(c.stats.maxQcount, c.qcount)
    w.stats.transfers++
    w.stats.buffered++
    emit(ctx, 'send.buffer', { g: [g.id], chan: [c.id] }, {
      chan: c.name,
      qcount: c.qcount,
      cap: c.cap,
      free: c.cap - c.qcount,
    })
    if (c.qcount === c.cap) {
      emit(ctx, 'buf.full', { g: [g.id], chan: [c.id] }, { chan: c.name, cap: c.cap, waiting: c.sendq.length })
    }
    // Получатель ждал, но передачу из рук в руки выключили — будим его,
    // и значение он заберёт из буфера уже своим ходом. Лишний переезд налицо.
    if (recvW) wake(ctx, recvW.g, 'в буфере появилось значение', c)
    completeOp(ctx, g, rng)
    return 'send.buffer'
  }

  // 3. Места нет и получателя нет — паркуемся в sendq.
  park(ctx, g, 'send', [{ chan: c, op: 'send' }])
  c.stats.blockedSends++
  emit(ctx, 'send.block', { g: [g.id], chan: [c.id] }, {
    chan: c.name,
    cap: c.cap,
    qcount: c.qcount,
    why: c.cap === 0 ? 'получателя нет' : 'буфер полон',
    queue: c.sendq.length,
  })
  return 'send.block'
}

function doRecv(ctx: Ctx, g: Goroutine, c: Chan, rng: Rng): Slot['op'] {
  const w = ctx.world
  if (c.isNil) {
    parkForever(ctx, g, 'recv', c)
    return 'nil'
  }
  c.lockedAt = w.tick

  // 1. В буфере что-то есть — берём оттуда. Даже если канал уже закрыт:
  //    закрытие не выбрасывает то, что успели положить.
  if (c.qcount > 0) {
    const from = c.buf[c.recvx]
    c.buf[c.recvx] = null
    c.recvx = (c.recvx + 1) % c.cap
    c.qcount--
    g.received++
    c.stats.received++
    emit(ctx, 'recv.buffer', { g: [g.id], chan: [c.id] }, {
      chan: c.name,
      from,
      qcount: c.qcount,
      cap: c.cap,
    })

    // Буфер был полон, и кто-то стоял в очереди на отправку: освободившуюся
    // ячейку тут же занимает его значение — в хвост, а не в голову.
    const sendW = firstWaiter(c.sendq)
    if (sendW) {
      const s = getG(w, sendW.g)
      c.buf[c.sendx] = sendW.g
      c.sendx = (c.sendx + 1) % c.cap
      c.qcount++
      s.sent++
      c.stats.sent++
      c.stats.buffered++
      w.stats.transfers++
      w.stats.buffered++
      emit(ctx, 'recv.wake', { g: [g.id, sendW.g], chan: [c.id] }, {
        chan: c.name,
        sender: sendW.g,
        waited: w.tick - sendW.since,
        qcount: c.qcount,
        cap: c.cap,
      })
      wake(ctx, sendW.g, 'в буфере освободилась ячейка', c)
      completeOp(ctx, s, rng)
    }
    completeOp(ctx, g, rng)
    return 'recv.buffer'
  }

  // 2. Буфера нет или он пуст, но отправитель уже стоит и держит значение.
  const sendW = firstWaiter(c.sendq)
  if (sendW) {
    const s = getG(w, sendW.g)
    g.received++
    s.sent++
    c.stats.sent++
    c.stats.received++
    c.stats.direct++
    w.stats.transfers++
    w.stats.direct++
    emit(ctx, 'recv.direct', { g: [g.id, sendW.g], chan: [c.id] }, {
      chan: c.name,
      from: sendW.g,
      waited: w.tick - sendW.since,
      viaSelect: sendW.fromSelect,
      cap: c.cap,
    })
    wake(ctx, sendW.g, 'значение забрали из рук', c)
    completeOp(ctx, s, rng)
    completeOp(ctx, g, rng)
    return 'recv.direct'
  }

  // 3. Канал закрыт и пуст — нулевое значение, без всякой блокировки.
  if (c.closed) {
    emit(ctx, 'recv.closed', { g: [g.id], chan: [c.id] }, { chan: c.name })
    completeOp(ctx, g, rng)
    return 'recv.closed'
  }

  // 4. Ждать нечего — паркуемся в recvq.
  park(ctx, g, 'recv', [{ chan: c, op: 'recv' }])
  c.stats.blockedRecvs++
  emit(ctx, 'recv.block', { g: [g.id], chan: [c.id] }, {
    chan: c.name,
    cap: c.cap,
    queue: c.recvq.length,
  })
  return 'recv.block'
}

function doClose(ctx: Ctx, g: Goroutine, c: Chan, rng: Rng): Slot['op'] {
  const w = ctx.world
  if (c.isNil) {
    panic(ctx, 'close of nil channel')
    emit(ctx, 'chan.nil', { g: [g.id], chan: [c.id] }, { chan: c.name, op: 'close' })
    return 'panic'
  }
  if (c.closed) {
    panic(ctx, 'close of closed channel')
    emit(ctx, 'chan.close', { g: [g.id], chan: [c.id] }, { chan: c.name, again: true })
    return 'panic'
  }

  c.closed = true
  c.lockedAt = w.tick
  w.stats.closes++
  const receivers = [...c.recvq]
  const senders = [...c.sendq]
  emit(ctx, 'chan.close', { g: [g.id], chan: [c.id] }, {
    chan: c.name,
    receivers: receivers.length,
    senders: senders.length,
    buffered: c.qcount,
    again: false,
  })

  // Все, кто ждал приёма, просыпаются разом. Значение им достанется нулевое,
  // и они это увидят, когда снова доберутся до процессора.
  for (const wtr of receivers) wake(ctx, wtr.g, 'канал закрыт', c)

  // А вот отправитель, уснувший в sendq, просыпается паникой.
  const first = senders[0]
  if (first) {
    emit(ctx, 'send.closed', { g: [first.g], chan: [c.id] }, { chan: c.name, asleep: true })
    panic(ctx, 'send on closed channel')
    return 'close'
  }

  completeOp(ctx, g, rng)
  return 'close'
}

/** Готова ли операция прямо сейчас — та самая проверка, которую selectgo делает по всем case. */
function ready(c: Chan, op: 'send' | 'recv'): boolean {
  if (c.isNil) return false
  if (op === 'send') return c.closed || c.recvq.length > 0 || c.qcount < c.cap
  return c.closed || c.qcount > 0 || c.sendq.length > 0
}

function doSelect(
  ctx: Ctx,
  g: Goroutine,
  ph: Extract<ChanPhase, { kind: 'select' }>,
  rng: Rng,
): Slot['op'] {
  const w = ctx.world
  const cases = ph.cases.map((cs) => ({ c: chanByName(w, cs.chan), op: cs.op }))
  const avail = cases.filter((x) => ready(x.c, x.op))

  // 1. Готов хотя бы один case — выбираем случайный. Порядок в исходнике ничего не решает.
  if (avail.length > 0) {
    const pick = rng.pick(avail)
    w.stats.selects++
    emit(ctx, 'select.ready', { g: [g.id], chan: [pick.c.id] }, {
      chan: pick.c.name,
      op: pick.op,
      ready: avail.length,
      cases: cases.length,
    })
    return pick.op === 'send' ? doSend(ctx, g, pick.c, rng) : doRecv(ctx, g, pick.c, rng)
  }

  // 2. Никто не готов, но есть default — уходим, не блокируясь.
  if (ph.default) {
    w.stats.selectDefaults++
    emit(ctx, 'select.default', { g: [g.id], chan: cases.map((x) => x.c.id) }, {
      cases: cases.length,
      chans: cases.map((x) => x.c.name).join(', '),
    })
    completeOp(ctx, g, rng)
    return 'select.default'
  }

  // 3. Встаём в очередь сразу ко всем каналам. Проснёмся от первого, кто отзовётся.
  const live = cases.filter((x) => !x.c.isNil)
  park(ctx, g, 'select', live.map((x) => ({ chan: x.c, op: x.op })), cases.map((x) => x.c.id))
  w.stats.selectBlocks++
  emit(ctx, 'select.block', { g: [g.id], chan: cases.map((x) => x.c.id) }, {
    cases: cases.length,
    queued: live.length,
    nils: cases.length - live.length,
    chans: cases.map((x) => `${x.op === 'send' ? '→' : '←'}${x.c.name}`).join(', '),
  })
  return 'select.block'
}

/* ──────────────────────────── исполнение ──────────────────────────── */

function step(ctx: Ctx, g: Goroutine, rng: Rng): { op: Slot['op']; chan: number | null } {
  const w = ctx.world
  g.state = 'running'
  if (g.phaseIdx === -1 && !advance(ctx, g, rng)) return { op: 'done', chan: null }

  const wl = ctx.scenario.workloads[g.workload]
  const ph = wl?.phases[g.phaseIdx]
  if (!ph) return { op: 'done', chan: null }

  switch (ph.kind) {
    case 'cpu': {
      g.runTicks++
      g.phaseLeft--
      if (g.phaseLeft <= 0) advance(ctx, g, rng)
      return { op: 'cpu', chan: null }
    }
    case 'send': {
      const c = chanByName(w, ph.chan)
      return { op: doSend(ctx, g, c, rng), chan: c.id }
    }
    case 'recv': {
      const c = chanByName(w, ph.chan)
      return { op: doRecv(ctx, g, c, rng), chan: c.id }
    }
    case 'close': {
      const c = chanByName(w, ph.chan)
      return { op: doClose(ctx, g, c, rng), chan: c.id }
    }
    case 'select':
      return { op: doSelect(ctx, g, ph, rng), chan: null }
  }
}

function phaseRun(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  /** Те, кто отработал этот тик, возвращаются в очередь ПОСЛЕ обхода —
   *  иначе одна горутина заняла бы сразу несколько процессоров. */
  const ran: number[] = []
  /** А эти были разбужены прямо в этом тике и ждут следующего. */
  const deferred: number[] = []
  const used = new Set<number>()

  let p = 0
  let guard = 0
  while (p < w.config.gomaxprocs && guard++ < 512) {
    const id = w.runq.shift()
    if (id === undefined) {
      w.slots.push({ p, g: null, op: null, chan: null })
      w.stats.idleSlots++
      p++
      continue
    }
    const g = getG(w, id)
    // В очереди мог остаться след от горутины, которую уже разбудили или завершили.
    if (g.state !== 'runnable') continue
    // Разбуженная в этом же тике дождётся следующего: один тик — один шаг горутины.
    if (used.has(id)) {
      deferred.push(id)
      continue
    }

    const { op, chan } = step(ctx, g, rng)
    used.add(id)
    w.slots.push({ p, g: g.id, op, chan })
    w.stats.busySlots++
    // step мог припарковать или завершить горутину — TypeScript об этом не знает.
    if ((g.state as ChanGState) === 'running') {
      g.state = 'runnable'
      ran.push(g.id)
    }
    p++
    if (w.finished) break
  }
  while (p < w.config.gomaxprocs) {
    w.slots.push({ p, g: null, op: null, chan: null })
    p++
  }
  // Сначала отложенные — их уже разбудили, — потом отработавшие.
  for (const id of deferred) enqueue(w, id)
  for (const id of ran) enqueue(w, id)
}

function phaseAccount(ctx: Ctx): void {
  const w = ctx.world
  for (const g of w.gs) {
    if (g.state === 'waiting') {
      g.waitTicks++
      w.stats.waitTicks++
    }
  }
  for (const c of w.chans) c.stats.maxQcount = Math.max(c.stats.maxQcount, c.qcount)
  if (w.finished) return

  const alive = w.gs.filter((g) => g.state !== 'done')
  const spawnDone = w.spawnPtr >= ctx.plan.length

  if (alive.length === 0 && spawnDone) {
    w.finished = true
    w.finishReason = 'all-done'
    return
  }

  // Взаимная блокировка: живые горутины есть, но все они спят, и разбудить
  // их некому — никто больше не родится и никто не работает.
  if (w.config.deadlockDetect && spawnDone && alive.length > 0) {
    const awake = alive.filter((g) => g.state !== 'waiting')
    if (awake.length === 0) {
      emit(ctx, 'deadlock', { g: alive.map((g) => g.id) }, {
        waiting: alive.length,
        chans: [...new Set(alive.flatMap((g) => g.wait?.chans ?? []))]
          .map((id) => w.chans[id]?.name)
          .filter(Boolean)
          .join(', '),
      })
      w.finished = true
      w.finishReason = 'deadlock'
      return
    }
  }

  // Прогон дошёл до лимита. Кто к этому моменту спит давно — тот уже не проснётся.
  const limit = ctx.scenario.stopAfter
  if (limit && w.tick >= limit) {
    const stuck = alive.filter(
      (g) => g.state === 'waiting' && g.wait !== null && w.tick - g.wait.since >= w.config.leakAfter,
    )
    if (stuck.length > 0) {
      emit(ctx, 'leak', { g: stuck.map((g) => g.id) }, {
        count: stuck.length,
        longest: Math.max(...stuck.map((g) => w.tick - (g.wait?.since ?? w.tick))),
        chans: [...new Set(stuck.flatMap((g) => g.wait?.chans ?? []))]
          .map((id) => w.chans[id]?.name)
          .filter(Boolean)
          .join(', '),
        total: alive.length,
      })
    }
    w.finished = true
    w.finishReason = 'stop-after'
  }
}
