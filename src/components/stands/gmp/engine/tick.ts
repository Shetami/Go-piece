import type { G, M, P, Phase, WaitReason } from './types.ts'
import type { Ctx } from './world.ts'
import { emit, getG, getM, getP, goready, runnextStealable, runqput, wakeWorthwhile } from './world.ts'
import type { Rng } from './rng.ts'

/**
 * Один шаг всей системы.
 *
 * Порядок фаз фиксирован и важен: от него зависит, какие события вообще
 * возможны. Менять его — значит менять выводы, которые стенд доказывает.
 */
export function tick(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  if (w.finished) return
  w.tick++

  phaseSpawn(ctx, rng)
  phaseSysmon(ctx)
  phaseWaiters(ctx)
  phaseExecute(ctx, rng)
  phaseFindRunnable(ctx, rng)
  phaseAccount(ctx)
}

function phaseSpawn(ctx: Ctx, _rng: Rng): void {
  const w = ctx.world
  while (w.spawnPtr < ctx.plan.length) {
    const item = ctx.plan[w.spawnPtr]
    if (!item || item.tick >= w.tick) break
    w.spawnPtr++
    createG(ctx, item.workload, item.onP)
  }
}

function createG(ctx: Ctx, workload: number, onP: number): void {
  const w = ctx.world
  const wl = ctx.scenario.workloads[workload]
  if (!wl) throw new Error(`нет нагрузки #${workload}`)
  const id = w.nextGid++
  const g: G = {
    id,
    workload,
    name: wl.name,
    state: 'runnable',
    phaseIdx: -1,
    phaseLeft: 0,
    repeatsLeft: wl.repeat === 'forever' ? Number.POSITIVE_INFINITY : (wl.repeat ?? 1) - 1,
    quantumUsed: 0,
    createdTick: w.tick,
    runTicks: 0,
    waitTicks: 0,
    runqTicks: 0,
  }
  w.gs.push(g)
  runqput(ctx, onP, id, true)
  emit(ctx, 'g.created', { g: [id], p: [onP] }, {
    name: wl.name,
    to: w.config.runnext ? 'runnext' : 'runq',
  })
}

function phaseSysmon(ctx: Ctx): void {
  const w = ctx.world

  for (const m of w.ms) {
    if (m.state !== 'syscall' || m.p === null) continue
    if (m.syscallTicks < w.config.retakeThreshold) continue
    const p = getP(w, m.p)
    p.m = null
    p.state = 'idle'
    p.handoffFrom = m.id
    m.p = null
    m.state = 'blocked'
    emit(ctx, 'sysmon.retake', { m: [m.id], p: [p.id], g: m.g ? [m.g] : [] }, {
      syscallTicks: m.syscallTicks,
      threshold: w.config.retakeThreshold,
    })
  }

  if (!w.config.asyncPreemption) return

  for (const p of w.ps) {
    if (p.state !== 'running' || p.m === null) continue
    const m = getM(w, p.m)
    if (m.state !== 'running' || m.g === null) continue
    const g = getG(w, m.g)
    if (g.quantumUsed < w.config.quantum) continue

    emit(ctx, 'sysmon.preempt', { g: [g.id], m: [m.id], p: [p.id] }, {
      ranTicks: g.quantumUsed,
      quantum: w.config.quantum,
    })
    g.state = 'runnable'
    g.quantumUsed = 0
    w.globrunq.push(g.id)
    m.g = null
    m.state = 'spinning'
    emit(ctx, 'g.preempted', { g: [g.id], p: [p.id] }, { reason: 'quantum', to: 'globrunq' })
  }
}

function phaseWaiters(ctx: Ctx): void {
  const w = ctx.world

  const net = w.netpoll.filter((e) => e.readyAt <= w.tick)
  if (net.length > 0) {
    w.netpoll = w.netpoll.filter((e) => e.readyAt > w.tick)
    emit(ctx, 'net.ready', { g: net.map((e) => e.g) }, { count: net.length })
    for (const e of net) goready(ctx, e.g, null, 'netpoller')
  }

  const fired = w.timers.filter((e) => e.readyAt <= w.tick)
  if (fired.length > 0) {
    w.timers = w.timers.filter((e) => e.readyAt > w.tick)
    for (const e of fired) goready(ctx, e.g, null, 'timer')
  }
}

function phaseExecute(ctx: Ctx, rng: Rng): void {
  const w = ctx.world

  const running = w.ms.filter((m) => m.state === 'running' && m.g !== null)
  for (const m of running) {
    if (m.state !== 'running' || m.g === null) continue
    const g = getG(w, m.g)

    if (g.phaseIdx === -1) {
      if (!advance(ctx, m, g, rng)) continue
    }
    if (g.state !== 'running' || m.g === null) continue

    g.runTicks++
    g.quantumUsed++
    if (Number.isFinite(g.phaseLeft)) g.phaseLeft--
    if (g.phaseLeft <= 0) advance(ctx, m, g, rng)
  }

  const inSyscall = w.ms.filter((m) => (m.state === 'syscall' || m.state === 'blocked') && m.g !== null)
  for (const m of inSyscall) {
    if (m.g === null) continue
    const g = getG(w, m.g)
    m.syscallTicks++
    if (Number.isFinite(g.phaseLeft)) g.phaseLeft--
    if (g.phaseLeft <= 0) syscallExit(ctx, m, g, rng)
  }
}

/** Завершить текущую фазу и войти в следующую. false — горутина больше не на процессоре. */
function advance(ctx: Ctx, m: M, g: G, rng: Rng): boolean {
  const wl = ctx.scenario.workloads[g.workload]
  if (!wl) throw new Error(`нет нагрузки #${g.workload}`)

  const cur = g.phaseIdx >= 0 ? wl.phases[g.phaseIdx] : undefined
  if (cur && cur.kind === 'mutex') releaseMutex(ctx, cur.lock, g.id)

  let next = g.phaseIdx + 1
  if (next >= wl.phases.length) {
    if (g.repeatsLeft > 0) {
      g.repeatsLeft--
      next = 0
    } else {
      die(ctx, m, g)
      return false
    }
  }
  g.phaseIdx = next
  const phase = wl.phases[next]
  if (!phase) throw new Error(`нет фазы #${next}`)
  return enterPhase(ctx, m, g, phase, rng)
}

function enterPhase(ctx: Ctx, m: M, g: G, ph: Phase, rng: Rng): boolean {
  const w = ctx.world

  switch (ph.kind) {
    case 'cpu': {
      g.phaseLeft = rng.duration(ph.ticks)
      return true
    }
    case 'syscall': {
      g.phaseLeft = rng.duration(ph.ticks)
      g.state = 'syscall'
      m.state = 'syscall'
      m.syscallTicks = 0
      if (m.p !== null) getP(w, m.p).state = 'syscall'
      emit(ctx, 'm.syscallEnter', { g: [g.id], m: [m.id], p: m.p === null ? [] : [m.p] }, {
        expectedTicks: g.phaseLeft,
      })
      return false
    }
    case 'net': {
      block(g, 'net')
      w.netpoll.push({ g: g.id, readyAt: w.tick + rng.duration(ph.ticks) })
      releaseM(m)
      emit(ctx, 'g.blocked', { g: [g.id], m: [m.id] }, { on: 'net', costsThread: false })
      return false
    }
    case 'sleep': {
      block(g, 'sleep')
      w.timers.push({ g: g.id, readyAt: w.tick + rng.duration(ph.ticks) })
      releaseM(m)
      emit(ctx, 'g.blocked', { g: [g.id], m: [m.id] }, { on: 'sleep' })
      return false
    }
    case 'mutex': {
      const mu = (w.mutexes[ph.lock] ??= { holder: null, waitq: [] })
      if (mu.holder === null || mu.holder === g.id) {
        mu.holder = g.id
        g.phaseLeft = rng.duration(ph.ticks)
        return true
      }
      block(g, 'mutex')
      mu.waitq.push(g.id)
      releaseM(m)
      emit(ctx, 'g.blocked', { g: [g.id], m: [m.id] }, { on: 'mutex', lock: ph.lock })
      return false
    }
    case 'chanSend':
    case 'chanRecv': {
      const ch = (w.chans[ph.chan] ??= { sendq: [], recvq: [] })
      const partnerQ = ph.kind === 'chanSend' ? ch.recvq : ch.sendq
      const ownQ = ph.kind === 'chanSend' ? ch.sendq : ch.recvq
      const partner = partnerQ.shift()
      if (partner !== undefined) {
        goready(ctx, partner, m.p, ph.kind)
        g.phaseLeft = 1
        return true
      }
      block(g, ph.kind === 'chanSend' ? 'chan-send' : 'chan-recv')
      ownQ.push(g.id)
      releaseM(m)
      emit(ctx, 'g.blocked', { g: [g.id], m: [m.id] }, {
        on: 'chan',
        chan: ph.chan,
        dir: ph.kind === 'chanSend' ? 'send' : 'recv',
      })
      return false
    }
  }
}

function syscallExit(ctx: Ctx, m: M, g: G, rng: Rng): void {
  const w = ctx.world
  const keptP = m.state === 'syscall'
  emit(ctx, 'm.syscallExit', { g: [g.id], m: [m.id], p: m.p === null ? [] : [m.p] }, {
    keptP,
    syscallTicks: m.syscallTicks,
  })

  if (keptP && m.p !== null) {
    getP(w, m.p).state = 'running'
    m.state = 'running'
    g.state = 'running'
    g.quantumUsed = 0
    advance(ctx, m, g, rng)
    return
  }

  g.state = 'runnable'
  m.g = null
  const idle = w.ps.find((p) => p.state === 'idle')
  if (idle) {
    idle.m = m.id
    idle.state = 'running'
    idle.handoffFrom = null
    m.p = idle.id
    m.state = 'running'
    m.g = g.id
    g.state = 'running'
    g.quantumUsed = 0
    advance(ctx, m, g, rng)
    return
  }
  w.globrunq.push(g.id)
  m.state = 'idle'
  emit(ctx, 'm.parked', { m: [m.id] }, { reason: 'no-p-after-syscall' })
}

function block(g: G, reason: WaitReason): void {
  g.state = 'waiting'
  g.waitReason = reason
}

function releaseM(m: M): void {
  m.g = null
  m.state = 'spinning'
}

function releaseMutex(ctx: Ctx, lock: string, gid: number): void {
  const w = ctx.world
  const mu = w.mutexes[lock]
  if (!mu || mu.holder !== gid) return
  mu.holder = null
  const nextG = mu.waitq.shift()
  if (nextG !== undefined) {
    mu.holder = nextG
    goready(ctx, nextG, null, 'mutex-unlock')
  }
}

function die(ctx: Ctx, m: M, g: G): void {
  const w = ctx.world
  g.state = 'dead'
  g.finishedTick = w.tick
  releaseM(m)
  emit(ctx, 'g.finished', { g: [g.id], m: [m.id] }, {
    lifetime: w.tick - g.createdTick,
    runTicks: g.runTicks,
    waitTicks: g.waitTicks,
    runqTicks: g.runqTicks,
  })
}

function phaseFindRunnable(ctx: Ctx, rng: Rng): void {
  const w = ctx.world

  const spinning = w.ms.filter((m) => m.state === 'spinning').length
  if (spinning === 0 && wakeWorthwhile(w)) {
    const p =
      w.ps.find((x) => x.state === 'idle' && (x.runq.length > 0 || x.runnext !== null)) ??
      w.ps.find((x) => x.state === 'idle')
    if (p) {
      const m = takeOrSpawnM(ctx)
      if (m) {
        p.m = m.id
        p.state = 'running'
        m.p = p.id
        m.state = 'spinning'
        if (p.handoffFrom !== null) {
          emit(ctx, 'p.handoff', { p: [p.id], m: [p.handoffFrom, m.id] }, {
            from: p.handoffFrom,
            to: m.id,
          })
          p.handoffFrom = null
        }
        emit(ctx, 'm.spinning', { m: [m.id], p: [p.id] }, {})
      }
    }
  }

  for (const m of [...w.ms]) {
    if (m.state !== 'spinning' || m.p === null) continue
    const p = getP(w, m.p)
    p.schedtick++

    let gid: number | null = null
    let source = ''

    const every = w.config.globalCheckEvery
    if (every && p.schedtick % every === 0) {
      const first = w.globrunq.shift()
      if (first !== undefined) {
        gid = first
        source = 'global-61'
        emit(ctx, 'p.globalCheck', { p: [p.id], g: [first] }, { found: true, schedtick: p.schedtick })
      } else {
        emit(ctx, 'p.globalCheck', { p: [p.id] }, { found: false, schedtick: p.schedtick })
      }
    }
    if (gid === null && p.runnext !== null) {
      gid = p.runnext
      p.runnext = null
      source = 'runnext'
    }
    if (gid === null) {
      const local = p.runq.shift()
      if (local !== undefined) {
        gid = local
        source = 'local'
      }
    }
    if (gid === null) {
      const glob = w.globrunq.shift()
      if (glob !== undefined) {
        gid = glob
        source = 'global'
      }
    }
    if (gid === null && w.config.workStealing) {
      const stolen = steal(ctx, p, rng)
      if (stolen !== null) {
        gid = stolen
        source = 'stolen'
      }
    }

    if (gid !== null) {
      const g = getG(w, gid)
      m.g = gid
      m.state = 'running'
      g.state = 'running'
      g.quantumUsed = 0
      emit(ctx, 'g.scheduled', { g: [gid], m: [m.id], p: [p.id] }, { source })
    } else {
      m.state = 'idle'
      m.p = null
      p.m = null
      p.state = 'idle'
      p.idleTicks++
      emit(ctx, 'm.parked', { m: [m.id] }, { reason: 'no-work' })
      emit(ctx, 'p.idle', { p: [p.id] }, { idleTicks: p.idleTicks })
    }
  }
}

function takeOrSpawnM(ctx: Ctx): M | undefined {
  const w = ctx.world
  const idle = w.ms.find((m) => m.state === 'idle')
  if (idle) return idle
  if (w.ms.length >= w.config.maxThreads) return undefined
  const m: M = {
    id: w.nextMid++,
    state: 'idle',
    p: null,
    g: null,
    syscallTicks: 0,
    spawnedTick: w.tick,
  }
  w.ms.push(m)
  emit(ctx, 'm.spawned', { m: [m.id] }, { totalThreads: w.ms.length, gomaxprocs: w.config.gomaxprocs })
  return m
}

/** До четырёх попыток украсть половину очереди у случайного другого P. */
function steal(ctx: Ctx, p: P, rng: Rng): number | null {
  const w = ctx.world
  const others = w.ps.filter((x) => x.id !== p.id)
  if (others.length === 0) return null

  for (let attempt = 1; attempt <= 4; attempt++) {
    const victim = rng.pick(others)
    if (!victim) break
    const len = victim.runq.length
    const n = len - Math.floor(len / 2)
    if (n === 0) {
      if (attempt === 4 && victim.runnext !== null && runnextStealable(w, victim)) {
        const gid = victim.runnext
        victim.runnext = null
        emit(ctx, 'p.stole', { p: [p.id, victim.id], g: [gid] }, {
          from: victim.id,
          count: 1,
          attempt,
          runnext: true,
        })
        return gid
      }
      continue
    }
    const batch = victim.runq.splice(0, n)
    const first = batch.shift()
    if (first === undefined) continue
    p.runq.push(...batch)
    emit(ctx, 'p.stole', { p: [p.id, victim.id] }, { from: victim.id, count: n, attempt })
    return first
  }
  emit(ctx, 'p.stealFailed', { p: [p.id] }, { attempts: 4 })
  return null
}

function phaseAccount(ctx: Ctx): void {
  const w = ctx.world
  for (const g of w.gs) {
    if (g.state === 'waiting') g.waitTicks++
    else if (g.state === 'runnable') g.runqTicks++
  }

  const alive = w.gs.filter((g) => g.state !== 'dead')
  const pending = w.spawnPtr < ctx.plan.length

  if (alive.length === 0 && !pending) {
    w.finished = true
    w.finishReason = 'all-done'
    return
  }
  const allWaiting = alive.length > 0 && alive.every((g) => g.state === 'waiting')
  if (allWaiting && w.netpoll.length === 0 && w.timers.length === 0 && !pending) {
    emit(ctx, 'deadlock', { g: alive.map((g) => g.id) }, {
      blocked: alive.length,
      reasons: alive.map((g) => g.waitReason),
    })
    w.finished = true
    w.finishReason = 'deadlock'
  }
}
