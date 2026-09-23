import type { Completion, ResilConfig, ResilEvent, ResilEventType, ResilLog, ResilScenario, ResilWorld, Sample, Stream } from './types.ts'
import { DEFAULT_RESIL_CONFIG } from './types.ts'

export interface Ctx {
  world: ResilWorld
  events: ResilEvent[]
  scenario: ResilScenario
  log: ResilLog
}

export function resolveConfig(partial: Partial<ResilConfig> | undefined): ResilConfig {
  return { ...DEFAULT_RESIL_CONFIG, ...(partial ?? {}) }
}

export function createWorld(scenario: ResilScenario): ResilWorld {
  const config = resolveConfig(scenario.config)
  return {
    tick: 0,
    config,
    rng: { arrive: seedOf(config.seed, 1), kind: seedOf(config.seed, 2), work: seedOf(config.seed, 3) },
    dep: 'ok',
    depFactor: 1,
    workers: Array.from({ length: config.workers }, () => null),
    queue: [],
    requests: {},
    depQueue: [],
    depWorkers: Array.from({ length: config.depWorkers }, () => null),
    breaker: 'closed',
    breakerFails: 0,
    breakerUntil: 0,
    probing: false,
    tokens: config.burst,
    nextId: 1,
    x: multiplierAt(scenario, 0),
    qMilestone: FIRST_MILESTONE,
    saturated: false,
    stats: {
      arrived: 0,
      ok: 0,
      degraded: 0,
      errors: 0,
      rejected: 0,
      depCalls: 0,
      depTimeouts: 0,
      depErrors: 0,
      retries: 0,
      blocked: 0,
      busyTicks: 0,
      waitTicks: 0,
      maxQueue: 0,
      localSlow: 0,
    },
    finished: false,
  }
}

export const FIRST_MILESTONE = 8

export function emit(ctx: Ctx, type: ResilEventType, actors: ResilEvent['actors'] = {}, payload: Record<string, unknown> = {}): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export const emptySample = (x: number): Sample => ({ arrived: 0, ok: 0, degraded: 0, errors: 0, rejected: 0, busy: 0, waiting: 0, queue: 0, depQueue: 0, x })

/* ───────────────────────────── случайность ───────────────────────────── */

const seedOf = (seed: number, stream: number) => (Math.imul((seed | 0) + stream * 0x632be5ab, 0x9e3779b1) ^ 0x5bd1e995) >>> 0

export function random(w: ResilWorld, stream: Stream): number {
  w.rng[stream] = (w.rng[stream] + 0x6d2b79f5) >>> 0
  let t = w.rng[stream]
  t = Math.imul(t ^ (t >>> 15), t | 1)
  t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296
}

export function poisson(w: ResilWorld, lambda: number): number {
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

export function geometric(w: ResilWorld, mean: number): number {
  if (mean <= 1) return 1
  const u = 1 - random(w, 'work')
  return 1 + Math.floor(Math.log(u) / Math.log(1 - 1 / mean))
}

/* ───────────────────────────── производные ───────────────────────────── */

export function multiplierAt(scenario: Pick<ResilScenario, 'phases'>, tick: number): number {
  let x = 1
  for (const p of [...scenario.phases].sort((a, b) => a.at - b.at)) if (p.at <= tick) x = p.x
  return x
}

export const busyOf = (w: ResilWorld) => w.workers.filter((x) => x !== null).length

/** Сколько воркеров A сейчас просто ждут ответа B. */
export const waitingOf = (w: ResilWorld) =>
  w.workers.filter((id) => id !== null && w.requests[id]?.stage === 'dep').length

/** Ёмкость A, если бы ожидание B ничего не стоило. */
export const capacityOf = (c: ResilConfig) => c.workers / c.ownTime

/** Ёмкость B. */
export const depCapacityOf = (c: ResilConfig) => c.depWorkers / c.depTime

export function percentile(values: number[], p: number): number {
  if (values.length === 0) return 0
  const sorted = [...values].sort((a, b) => a - b)
  const rank = Math.ceil((p / 100) * sorted.length)
  return sorted[Math.min(sorted.length, Math.max(1, rank)) - 1]!
}

/** Задержки ответов, завершённых на тиках (from, to]. */
export function latencies(done: readonly Completion[], to: number, from = 0, pick?: (c: Completion) => boolean): number[] {
  const out: number[] = []
  for (const c of done) if (c.t > from && c.t <= to && c.kind !== 'rejected' && (!pick || pick(c))) out.push(c.lat)
  return out
}

/** Доля ответов, которые дошли до пользователя хоть в каком-то виде. */
export function successShare(w: ResilWorld): number {
  const s = w.stats
  const total = s.ok + s.degraded + s.errors + s.rejected
  return total === 0 ? 1 : (s.ok + s.degraded) / total
}
