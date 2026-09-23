import type { Completion, SagaConfig, SagaEvent, SagaEventType, SagaLog, SagaScenario, SagaWorld, Sample, Stream } from './types.ts'
import { DEFAULT_SAGA_CONFIG } from './types.ts'

export interface Ctx {
  world: SagaWorld
  events: SagaEvent[]
  scenario: SagaScenario
  log: SagaLog
}

export function resolveConfig(partial: Partial<SagaConfig> | undefined): SagaConfig {
  return { ...DEFAULT_SAGA_CONFIG, ...(partial ?? {}) }
}

export function createWorld(scenario: SagaScenario): SagaWorld {
  const config = resolveConfig(scenario.config)
  return {
    tick: 0,
    config,
    rng: { arrive: seedOf(config.seed, 1), fail: seedOf(config.seed, 2), work: seedOf(config.seed, 3) },
    orders: [],
    outbox: [],
    broker: [],
    work: [],
    seen: [],
    retryQ: [],
    brokerUp: true,
    duplicating: false,
    nextId: 1,
    nextMsg: 1,
    x: 1,
    stats: {
      orders: 0,
      paid: 0,
      shipped: 0,
      cancelled: 0,
      lost: 0,
      duplicates: 0,
      doubleCharges: 0,
      payFails: 0,
      retries: 0,
      giveups: 0,
      shipFails: 0,
      compensations: 0,
      stuck: 0,
      inconsistentTicks: 0,
    },
    finished: false,
  }
}

export function emit(ctx: Ctx, type: SagaEventType, actors: SagaEvent['actors'] = {}, payload: Record<string, unknown> = {}): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export const emptySample = (x: number): Sample => ({ created: 0, shipped: 0, cancelled: 0, inFlight: 0, inBroker: 0, outbox: 0, x })

/* ───────────────────────────── случайность ───────────────────────────── */

const seedOf = (seed: number, stream: number) => (Math.imul((seed | 0) + stream * 0x632be5ab, 0x9e3779b1) ^ 0x5bd1e995) >>> 0

export function random(w: SagaWorld, stream: Stream): number {
  w.rng[stream] = (w.rng[stream] + 0x6d2b79f5) >>> 0
  let t = w.rng[stream]
  t = Math.imul(t ^ (t >>> 15), t | 1)
  t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296
}

export function poisson(w: SagaWorld, lambda: number): number {
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

export function geometric(w: SagaWorld, mean: number): number {
  if (mean <= 1) return 1
  const u = 1 - random(w, 'work')
  return 1 + Math.floor(Math.log(u) / Math.log(1 - 1 / mean))
}

/* ───────────────────────────── производные ───────────────────────────── */

/** Заказ ещё не дошёл до конечного состояния. */
export const inFlight = (w: SagaWorld) => w.orders.filter((o) => o.state === 'created' || o.state === 'paid').length

/** Заказы, у которых списаны деньги, но доставки нет: временное расхождение. */
export const paidNotShipped = (w: SagaWorld) => w.orders.filter((o) => o.state === 'paid').length

/** Сколько заказов дошло до конца и сколько зависло навсегда. */
export function outcome(w: SagaWorld): { settled: number; stuck: number } {
  const settled = w.orders.filter((o) => o.state === 'shipped' || o.state === 'cancelled').length
  return { settled, stuck: w.orders.length - settled }
}

/** Доля заказов, которые дошли до понятного конца. */
export function consistency(w: SagaWorld): number {
  if (w.orders.length === 0) return 1
  return outcome(w).settled / w.orders.length
}

export function percentile(values: number[], p: number): number {
  if (values.length === 0) return 0
  const sorted = [...values].sort((a, b) => a - b)
  const rank = Math.ceil((p / 100) * sorted.length)
  return sorted[Math.min(sorted.length, Math.max(1, rank)) - 1]!
}

export function latencies(done: readonly Completion[], to: number, from = 0, kind?: Completion['kind']): number[] {
  const out: number[] = []
  for (const c of done) if (c.t > from && c.t <= to && (!kind || c.kind === kind)) out.push(c.lat)
  return out
}
