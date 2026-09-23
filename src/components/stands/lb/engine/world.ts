import type { Completion, LbConfig, LbEvent, LbEventType, LbLog, LbScenario, LbWorld, Replica, Sample, Stream } from './types.ts'
import { DEFAULT_LB_CONFIG } from './types.ts'

export interface Ctx {
  world: LbWorld
  events: LbEvent[]
  scenario: LbScenario
  log: LbLog
}

export function resolveConfig(partial: Partial<LbConfig> | undefined): LbConfig {
  return { ...DEFAULT_LB_CONFIG, ...(partial ?? {}) }
}

export function newReplica(id: number, workers: number, booting: boolean, bootUntil = 0): Replica {
  return {
    id,
    name: `r${id + 1}`,
    state: booting ? 'booting' : 'up',
    factor: 1,
    workers: Array.from({ length: workers }, () => null),
    queue: [],
    bootUntil,
    inRotation: !booting,
    ejectedUntil: 0,
    outstanding: 0,
    errorsInRow: 0,
    failedChecks: 0,
    passedChecks: 0,
    served: 0,
    errors: 0,
    routed: 0,
  }
}

export function createWorld(scenario: LbScenario): LbWorld {
  const config = resolveConfig(scenario.config)
  const n = Math.max(1, Math.min(config.replicas, config.maxReplicas))
  return {
    tick: 0,
    config,
    rng: { arrive: seedOf(config.seed, 1), work: seedOf(config.seed, 2), route: seedOf(config.seed, 3) },
    replicas: Array.from({ length: n }, (_, i) => newReplica(i, config.workers, false)),
    retryQ: [],
    nextId: 1,
    nextOrigin: 1,
    rr: 0,
    x: multiplierAt(scenario, 0),
    qMilestone: Array.from({ length: n }, () => FIRST_MILESTONE),
    busyAcc: 0,
    capAcc: 0,
    stats: { logical: 0, attempts: 0, ok: 0, errors: 0, timeouts: 0, retries: 0, wasted: 0, failed: 0, noBackend: 0, maxQueue: 0 },
    finished: false,
  }
}

export const FIRST_MILESTONE = 8

export function emit(ctx: Ctx, type: LbEventType, actors: LbEvent['actors'] = {}, payload: Record<string, unknown> = {}): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export const emptySample = (x: number, n: number): Sample => ({
  arrived: 0,
  ok: 0,
  errors: 0,
  timeouts: 0,
  x,
  load: Array.from({ length: n }, () => 0),
  inRotation: n,
  replicas: n,
})

/* ───────────────────────────── случайность ───────────────────────────── */

const seedOf = (seed: number, stream: number) => (Math.imul((seed | 0) + stream * 0x632be5ab, 0x9e3779b1) ^ 0x5bd1e995) >>> 0

/** mulberry32 с состоянием в мире: снимок мира продолжает тот же прогон. */
export function random(w: LbWorld, stream: Stream): number {
  w.rng[stream] = (w.rng[stream] + 0x6d2b79f5) >>> 0
  let t = w.rng[stream]
  t = Math.imul(t ^ (t >>> 15), t | 1)
  t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296
}

export function poisson(w: LbWorld, lambda: number): number {
  if (lambda <= 0) return 0
  const limit = Math.exp(-lambda)
  let k = 0
  let p = random(w, 'arrive')
  while (p > limit) {
    k++
    p *= random(w, 'arrive')
  }
  return k
}

/** Геометрическое время обработки со средним mean — как в стенде «Нагрузка». */
export function geometric(w: LbWorld, mean: number): number {
  if (mean <= 1) return 1
  const u = 1 - random(w, 'work')
  return 1 + Math.floor(Math.log(u) / Math.log(1 - 1 / mean))
}

/* ───────────────────────────── производные ───────────────────────────── */

export function multiplierAt(scenario: Pick<LbScenario, 'phases'>, tick: number): number {
  let x = 1
  for (const p of [...scenario.phases].sort((a, b) => a.at - b.at)) if (p.at <= tick) x = p.x
  return x
}

/** Ёмкость одной здоровой реплики, запросов за тик. */
export const replicaCapacity = (c: LbConfig) => c.workers / c.serviceTime

export const busyOf = (r: Replica) => r.workers.filter((x) => x !== null).length

/** Реплика принимает запросы от балансировщика прямо сейчас. */
export const routable = (w: LbWorld, r: Replica) => r.inRotation && r.ejectedUntil <= w.tick && r.state !== 'booting'

export function percentile(values: number[], p: number): number {
  if (values.length === 0) return 0
  const sorted = [...values].sort((a, b) => a - b)
  const rank = Math.ceil((p / 100) * sorted.length)
  return sorted[Math.min(sorted.length, Math.max(1, rank)) - 1]!
}

export function okLatencies(done: readonly Completion[], to: number, from = 0): number[] {
  const out: number[] = []
  for (const c of done) if (c.t > from && c.t <= to && c.outcome === 'ok') out.push(c.lat)
  return out
}
