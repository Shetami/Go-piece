import type { Completion, LoadConfig, LoadEvent, LoadEventType, LoadLog, LoadScenario, LoadWorld, Sample, Stream } from './types.ts'
import { DEFAULT_LOAD_CONFIG } from './types.ts'

export interface Ctx {
  world: LoadWorld
  events: LoadEvent[]
  scenario: LoadScenario
  log: LoadLog
}

export function resolveConfig(partial: Partial<LoadConfig> | undefined): LoadConfig {
  return { ...DEFAULT_LOAD_CONFIG, ...(partial ?? {}) }
}

export const emptySample = (x: number): Sample => ({ arrived: 0, retried: 0, rejected: 0, queue: 0, busy: 0, ok: 0, wasted: 0, timeouts: 0, x })

export function createWorld(scenario: LoadScenario): LoadWorld {
  const config = resolveConfig(scenario.config)
  return {
    tick: 0,
    config,
    rng: { arrive: seedOf(config.seed, 1), work: seedOf(config.seed, 2), retry: seedOf(config.seed, 3) },
    queue: [],
    workers: Array.from({ length: config.workers }, () => null),
    retryQ: [],
    nextId: 1,
    nextOrigin: 1,
    x: multiplierAt(scenario, 0),
    carry: 0,
    recent: [],
    stats: {
      arrivals: 0,
      logical: 0,
      admitted: 0,
      ok: 0,
      wasted: 0,
      cancelled: 0,
      timeouts: 0,
      retries: 0,
      rejected: 0,
      failed: 0,
      busyTicks: 0,
      inSystemSum: 0,
      sojournSum: 0,
      departures: 0,
      maxQueue: 0,
    },
    qMilestone: FIRST_MILESTONE,
    qPeak: 0,
    qSince: null,
    collapsedAt: null,
    finished: false,
  }
}

/** С какой длины очереди о ней стоит сообщать. Дальше — каждое удвоение. */
export const FIRST_MILESTONE = 8

export function emit(ctx: Ctx, type: LoadEventType, actors: LoadEvent['actors'] = {}, payload: Record<string, unknown> = {}): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

/* ───────────────────────────── случайность ───────────────────────────── */

/** Зерно перемешивается, чтобы соседние зёрна и потоки давали непохожие последовательности. */
const seedOf = (seed: number, stream: number) => (Math.imul((seed | 0) + stream * 0x632be5ab, 0x9e3779b1) ^ 0x5bd1e995) >>> 0

/** mulberry32: состояние живёт в мире, поэтому снимок мира продолжает тот же прогон. */
export function random(w: LoadWorld, stream: Stream): number {
  w.rng[stream] = (w.rng[stream] + 0x6d2b79f5) >>> 0
  let t = w.rng[stream]
  t = Math.imul(t ^ (t >>> 15), t | 1)
  t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296
}

/** Сколько запросов пришло за тик при пуассоновском потоке со средним lambda. */
export function poisson(w: LoadWorld, lambda: number): number {
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

/**
 * Время обработки: геометрическое распределение на 1, 2, 3… со средним mean —
 * дискретный родственник экспоненциального. Большинство запросов короче
 * среднего, редкие — намного длиннее.
 */
export function geometric(w: LoadWorld, mean: number): number {
  if (mean <= 1) return 1
  const u = 1 - random(w, 'work')
  return 1 + Math.floor(Math.log(u) / Math.log(1 - 1 / mean))
}

/* ───────────────────────────── производные ───────────────────────────── */

export function multiplierAt(scenario: Pick<LoadScenario, 'phases'>, tick: number): number {
  let x = 1
  for (const p of [...scenario.phases].sort((a, b) => a.at - b.at)) if (p.at <= tick) x = p.x
  return x
}

/** Среднее время обработки с учётом медленных запросов. */
export function meanService(c: LoadConfig): number {
  const s = c.slowShare / 100
  return c.serviceTime * (1 - s + s * c.slowFactor)
}

/** Ёмкость: сколько запросов в среднем сервис может обработать за тик. */
export const capacityOf = (c: LoadConfig) => c.workers / meanService(c)

/** Ожидаемая загрузка ρ при потоке rate·x. */
export const offeredLoad = (c: LoadConfig, x = 1) => (c.rate * x) / capacityOf(c)

export const busyOf = (w: LoadWorld) => w.workers.filter((r) => r !== null).length

export const inSystemOf = (w: LoadWorld) => w.queue.length + busyOf(w)

/** Перцентиль по ближайшему рангу: p99 — значение, которое не превысили 99% замеров. */
export function percentile(values: number[], p: number): number {
  if (values.length === 0) return 0
  const sorted = [...values].sort((a, b) => a - b)
  const rank = Math.ceil((p / 100) * sorted.length)
  return sorted[Math.min(sorted.length, Math.max(1, rank)) - 1]!
}

export const mean = (values: number[]) => (values.length === 0 ? 0 : values.reduce((a, b) => a + b, 0) / values.length)

/** Задержки успешных ответов, полученных на тиках (from, to]. */
export function okLatencies(done: readonly Completion[], to: number, from = 0): number[] {
  const out: number[] = []
  for (const c of done) if (c.t > from && c.t <= to && c.outcome === 'ok') out.push(c.lat)
  return out
}

/**
 * Закон Литтла по прогону до тика t: L — сколько запросов в среднем было в сервисе,
 * λ — сколько в среднем входило за тик, W — сколько в среднем пробыл ушедший.
 */
export function little(w: LoadWorld): { L: number; lambda: number; W: number } {
  const t = Math.max(1, w.tick)
  return {
    L: w.stats.inSystemSum / t,
    lambda: w.stats.admitted / t,
    W: w.stats.departures === 0 ? 0 : w.stats.sojournSum / w.stats.departures,
  }
}
