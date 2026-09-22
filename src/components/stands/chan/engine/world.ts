import type {
  Chan,
  ChanConfig,
  ChanEvent,
  ChanEventType,
  ChanScenario,
  ChanWorld,
  Goroutine,
  Waiter,
} from './types.ts'
import { DEFAULT_CHAN_CONFIG } from './types.ts'

/** Одна запланированная горутина: кто и когда рождается. */
export interface SpawnPlanItem {
  tick: number
  workload: number
}

export interface Ctx {
  world: ChanWorld
  events: ChanEvent[]
  plan: SpawnPlanItem[]
  scenario: ChanScenario
}

export function resolveConfig(partial: Partial<ChanConfig> | undefined): ChanConfig {
  return { ...DEFAULT_CHAN_CONFIG, ...(partial ?? {}) }
}

export function buildSpawnPlan(scenario: ChanScenario): SpawnPlanItem[] {
  const plan: SpawnPlanItem[] = []
  scenario.workloads.forEach((w, wi) => {
    for (let i = 0; i < w.count; i++) {
      plan.push({ tick: w.spawnAt === 'staggered' ? i : w.spawnAt, workload: wi })
    }
  })
  plan.sort((a, b) => a.tick - b.tick || a.workload - b.workload)
  return plan
}

export function createWorld(scenario: ChanScenario): ChanWorld {
  const config = resolveConfig(scenario.config)
  const chans: Chan[] = scenario.chans.map((spec, id) => ({
    id,
    name: spec.name,
    cap: spec.nil ? 0 : spec.cap,
    buf: Array.from({ length: spec.nil ? 0 : spec.cap }, () => null),
    qcount: 0,
    sendx: 0,
    recvx: 0,
    closed: false,
    isNil: spec.nil === true,
    sendq: [],
    recvq: [],
    lockedAt: -1,
    stats: {
      sent: 0,
      received: 0,
      direct: 0,
      buffered: 0,
      blockedSends: 0,
      blockedRecvs: 0,
      maxQcount: 0,
    },
  }))

  return {
    tick: 0,
    config,
    chans,
    gs: [],
    runq: [],
    // Пустые слоты с самого начала: на нулевом тике схема должна выглядеть так же,
    // как на любом другом, — просто все процессоры простаивают.
    slots: Array.from({ length: config.gomaxprocs }, (_, p) => ({ p, g: null, op: null, chan: null })),
    stats: {
      transfers: 0,
      direct: 0,
      buffered: 0,
      parks: 0,
      wakeups: 0,
      waitTicks: 0,
      selects: 0,
      selectDefaults: 0,
      selectBlocks: 0,
      closes: 0,
      idleSlots: 0,
      busySlots: 0,
    },
    nextGid: 1,
    spawnPtr: 0,
    finished: false,
  }
}

export function emit(
  ctx: Ctx,
  type: ChanEventType,
  actors: ChanEvent['actors'] = {},
  payload: Record<string, unknown> = {},
): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export function getG(w: ChanWorld, id: number): Goroutine {
  const g = w.gs[id - 1]
  if (!g) throw new Error(`нет горутины G${id}`)
  return g
}

export function getChan(w: ChanWorld, id: number): Chan {
  const c = w.chans[id]
  if (!c) throw new Error(`нет канала #${id}`)
  return c
}

/** Канал по имени из сценария. Опечатка в имени роняет прогон, а не молча меняет смысл. */
export function chanByName(w: ChanWorld, name: string): Chan {
  const c = w.chans.find((x) => x.name === name)
  if (!c) throw new Error(`в сценарии нет канала «${name}»`)
  return c
}

/* ───────────────────────── очередь готовых ───────────────────────── */

/** Поставить горутину в конец общей очереди готовых. */
export function enqueue(w: ChanWorld, id: number): void {
  if (!w.runq.includes(id)) w.runq.push(id)
}

/**
 * Парковка: горутина уходит в ожидание и встаёт в очереди всех каналов,
 * которые она ждёт. Для select таких каналов несколько — по одному sudog на case.
 */
export function park(
  ctx: Ctx,
  g: Goroutine,
  kind: 'send' | 'recv' | 'select',
  entries: { chan: Chan; op: 'send' | 'recv' }[],
  /** Каналы для подписи на схеме. По умолчанию — те же, в чьих очередях горутина встала.
   *  Отличаются они только у nil-канала: очереди у него нет, а ждут его вечно. */
  waitChans?: number[],
): void {
  const w = ctx.world
  g.state = 'waiting'
  g.wait = { kind, chans: waitChans ?? entries.map((e) => e.chan.id), since: w.tick }
  g.blocks++
  w.stats.parks++
  w.runq = w.runq.filter((x) => x !== g.id)

  const waiter: Waiter = { g: g.id, fromSelect: kind === 'select', since: w.tick }
  for (const e of entries) {
    if (e.op === 'send') e.chan.sendq.push({ ...waiter })
    else e.chan.recvq.push({ ...waiter })
  }
}

/**
 * Пробуждение: горутину вынимают из очередей ВСЕХ каналов, где она стояла,
 * и ставят в очередь готовых. С runnext — в начало: разбуженная исполнится следующей.
 */
export function wake(ctx: Ctx, id: number, reason: string, chan: Chan | null): void {
  const w = ctx.world
  const g = getG(w, id)
  const waited = g.wait ? w.tick - g.wait.since : 0
  const viaSelect = g.wait?.kind === 'select'

  for (const c of w.chans) {
    c.sendq = c.sendq.filter((x) => x.g !== id)
    c.recvq = c.recvq.filter((x) => x.g !== id)
  }
  g.state = 'runnable'
  g.wait = null
  g.wakeups++
  w.stats.wakeups++

  if (w.config.runnext) w.runq.unshift(id)
  else w.runq.push(id)

  emit(ctx, 'g.ready', { g: [id], chan: chan ? [chan.id] : [] }, {
    reason,
    waited,
    chan: chan?.name ?? null,
    viaSelect,
    runnext: w.config.runnext,
  })
}

/** Первый ждущий в очереди. Очереди канала строго FIFO — это гарантия рантайма. */
export function firstWaiter(q: Waiter[]): Waiter | undefined {
  return q[0]
}

export function removeWaiter(q: Waiter[], id: number): void {
  const i = q.findIndex((x) => x.g === id)
  if (i >= 0) q.splice(i, 1)
}

/** Живые горутины, которые сейчас способны исполняться. */
export function runnableCount(w: ChanWorld): number {
  return w.gs.filter((g) => g.state === 'runnable' || g.state === 'running').length
}
