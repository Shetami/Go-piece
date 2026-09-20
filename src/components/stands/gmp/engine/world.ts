import type {
  Config,
  EventType,
  G,
  M,
  P,
  Scenario,
  SimEvent,
  World,
} from './types.ts'
import { DEFAULT_CONFIG } from './types.ts'

/** Одна запланированная горутина: кто, когда и на каком P рождается. */
export interface SpawnPlanItem {
  tick: number
  workload: number
  onP: number
}

export interface Ctx {
  world: World
  events: SimEvent[]
  plan: SpawnPlanItem[]
  scenario: Scenario
}

export function resolveConfig(partial: Partial<Config> | undefined): Config {
  return { ...DEFAULT_CONFIG, ...(partial ?? {}) }
}

export function buildSpawnPlan(scenario: Scenario, gomaxprocs: number): SpawnPlanItem[] {
  const plan: SpawnPlanItem[] = []
  scenario.workloads.forEach((w, wi) => {
    for (let i = 0; i < w.count; i++) {
      const tick = w.spawnAt === 'staggered' ? i : w.spawnAt
      const onP = w.spawnOn ?? i % gomaxprocs
      plan.push({ tick, workload: wi, onP: Math.min(onP, gomaxprocs - 1) })
    }
  })
  plan.sort((a, b) => a.tick - b.tick || a.workload - b.workload || a.onP - b.onP)
  return plan
}

export function createWorld(scenario: Scenario): World {
  const config = resolveConfig(scenario.config)
  const ps: P[] = []
  for (let i = 0; i < config.gomaxprocs; i++) {
    ps.push({
      id: i,
      state: 'idle',
      m: null,
      runq: [],
      runnext: null,
      schedtick: 0,
      idleTicks: 0,
      handoffFrom: null,
    })
  }
  return {
    tick: 0,
    config,
    gs: [],
    ms: [],
    ps,
    globrunq: [],
    netpoll: [],
    timers: [],
    chans: {},
    mutexes: {},
    nextGid: 1,
    nextMid: 1,
    spawnPtr: 0,
    finished: false,
  }
}

export function getG(w: World, id: number): G {
  const g = w.gs[id - 1]
  if (!g) throw new Error(`нет горутины G${id}`)
  return g
}

export function getM(w: World, id: number): M {
  const m = w.ms[id - 1]
  if (!m) throw new Error(`нет потока M${id}`)
  return m
}

export function getP(w: World, id: number): P {
  const p = w.ps[id]
  if (!p) throw new Error(`нет процессора P${id}`)
  return p
}

export function emit(
  ctx: Ctx,
  type: EventType,
  actors: SimEvent['actors'],
  payload: Record<string, unknown> = {},
): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

/**
 * Положить горутину в очередь P.
 *
 * `next = true` — горутина только что разбужена и попадает в слот runnext,
 * минуя очередь. Именно этот слот делает пинг-понг по каналу быстрым.
 * При переполнении локальной очереди половина уезжает в глобальную —
 * так делает runqputslow в настоящем рантайме.
 */
export function runqput(ctx: Ctx, pid: number, gid: number, next: boolean): void {
  const w = ctx.world
  const p = getP(w, pid)

  if (next && w.config.runnext) {
    const kicked = p.runnext
    p.runnext = gid
    if (kicked !== null) pushLocal(ctx, p, kicked)
    return
  }
  pushLocal(ctx, p, gid)
}

function pushLocal(ctx: Ctx, p: P, gid: number): void {
  const w = ctx.world
  if (p.runq.length >= w.config.runqCapacity) {
    const half = Math.floor(w.config.runqCapacity / 2)
    const batch = p.runq.splice(0, half)
    batch.push(gid)
    w.globrunq.push(...batch)
    emit(ctx, 'p.runqOverflow', { p: [p.id], g: [gid] }, {
      fromRunq: half,
      total: batch.length,
      globrunqSize: w.globrunq.length,
    })
    return
  }
  p.runq.push(gid)
}

/** Разбудить горутину: в runnext текущего P, иначе в глобальную очередь. */
export function goready(ctx: Ctx, gid: number, byPid: number | null, by: string): void {
  const w = ctx.world
  const g = getG(w, gid)
  g.state = 'runnable'
  g.waitReason = undefined

  let target = byPid
  if (target === null) {
    const idle = w.ps.find((p) => p.state === 'idle')
    target = idle ? idle.id : null
  }

  if (target === null) {
    w.globrunq.push(gid)
    emit(ctx, 'g.ready', { g: [gid] }, { by, to: 'globrunq' })
    return
  }
  runqput(ctx, target, gid, true)
  emit(ctx, 'g.ready', { g: [gid], p: [target] }, {
    by,
    to: w.config.runnext ? 'runnext' : 'runq',
  })
}

/** Есть ли вообще работа — для проверки на взаимную блокировку. */
export function workAvailable(w: World): boolean {
  if (w.globrunq.length > 0) return true
  return w.ps.some((p) => p.runq.length > 0 || p.runnext !== null)
}

/**
 * Стоит ли будить поток ради свободного P.
 *
 * Аналог wakep: рантайм не поднимает поток на каждую готовую горутину.
 * Смысл есть, только если работа лежит в глобальной очереди, у процессора
 * без потока, или у кого-то в очереди набралось хотя бы два элемента —
 * иначе украсть будет нечего и поток тут же уснёт обратно.
 */
export function wakeWorthwhile(w: World): boolean {
  if (w.globrunq.length > 0) return true
  for (const p of w.ps) {
    const load = p.runq.length + (p.runnext === null ? 0 : 1)
    if (load === 0) continue
    if (p.state === 'idle') return true
    if (w.config.workStealing && (load >= 2 || runnextStealable(w, p))) return true
  }
  return false
}

/** Сколько тиков P не забирает горутину из runnext, прежде чем её разрешено украсть. */
export const RUNNEXT_GRACE = 3

/**
 * Можно ли украсть горутину из runnext этого P.
 *
 * Настоящий runqgrab сначала даёт жертве шанс взять её самой (usleep(3)) и крадёт,
 * только если та так и не освободилась. В модели «не освободилась» — текущая
 * горутина жертвы работает уже RUNNEXT_GRACE тиков и больше.
 */
export function runnextStealable(w: World, p: P): boolean {
  if (p.runnext === null || p.runq.length > 0) return false
  if (p.m === null) return true
  const m = w.ms[p.m - 1]
  if (!m || m.g === null) return false
  const cur = w.gs[m.g - 1]
  return cur !== undefined && cur.quantumUsed >= RUNNEXT_GRACE
}
