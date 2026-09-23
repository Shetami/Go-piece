import type { CacheConfig, CacheEvent, CacheEventType, CacheLog, CacheScenario, CacheWorld, Completion, Sample, Stream } from './types.ts'
import { DEFAULT_CACHE_CONFIG } from './types.ts'

export interface Ctx {
  world: CacheWorld
  events: CacheEvent[]
  scenario: CacheScenario
  log: CacheLog
}

export function resolveConfig(partial: Partial<CacheConfig> | undefined): CacheConfig {
  return { ...DEFAULT_CACHE_CONFIG, ...(partial ?? {}) }
}

export function createWorld(scenario: CacheScenario): CacheWorld {
  const config = resolveConfig(scenario.config)
  return {
    tick: 0,
    config,
    rng: { arrive: seedOf(config.seed, 1), key: seedOf(config.seed, 2), work: seedOf(config.seed, 3), ttl: seedOf(config.seed, 4) },
    cache: {},
    db: Array.from({ length: config.keys }, () => 0),
    queue: [],
    workers: Array.from({ length: config.dbWorkers }, () => null),
    inflight: {},
    nextId: 1,
    x: multiplierAt(scenario, 0),
    qMilestone: FIRST_MILESTONE,
    stampedeKeys: [],
    stats: { reads: 0, writes: 0, hits: 0, misses: 0, joins: 0, swr: 0, stale: 0, dbReads: 0, dbWrites: 0, maxDbQueue: 0, maxSameKey: 0 },
    finished: false,
  }
}

export const FIRST_MILESTONE = 8
/** Столько одновременных запросов в базу за один ключ — уже лавина. */
export const STAMPEDE = 4

export function emit(ctx: Ctx, type: CacheEventType, actors: CacheEvent['actors'] = {}, payload: Record<string, unknown> = {}): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export const emptySample = (x: number): Sample => ({ arrived: 0, hits: 0, misses: 0, stale: 0, dbStarted: 0, dbQueue: 0, dbBusy: 0, cached: 0, x })

/* ───────────────────────────── случайность ───────────────────────────── */

const seedOf = (seed: number, stream: number) => (Math.imul((seed | 0) + stream * 0x632be5ab, 0x9e3779b1) ^ 0x5bd1e995) >>> 0

export function random(w: CacheWorld, stream: Stream): number {
  w.rng[stream] = (w.rng[stream] + 0x6d2b79f5) >>> 0
  let t = w.rng[stream]
  t = Math.imul(t ^ (t >>> 15), t | 1)
  t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296
}

export function poisson(w: CacheWorld, lambda: number): number {
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

export function geometric(w: CacheWorld, mean: number): number {
  if (mean <= 1) return 1
  const u = 1 - random(w, 'work')
  return 1 + Math.floor(Math.log(u) / Math.log(1 - 1 / mean))
}

/** Функция распределения Ципфа: доля запросов к ключам 0..i. Не хранится в мире — считается один раз. */
const cdfCache = new Map<string, number[]>()
export function zipfCdf(n: number, s: number): number[] {
  const k = `${n}:${s}`
  const hit = cdfCache.get(k)
  if (hit) return hit
  const w = Array.from({ length: n }, (_, i) => 1 / (i + 1) ** s)
  const total = w.reduce((a, b) => a + b, 0)
  let acc = 0
  const cdf = w.map((x) => (acc += x / total))
  cdfCache.set(k, cdf)
  return cdf
}

/** Какой ключ спросили: самый горячий с вероятностью hotShare, иначе — по Ципфу. */
export function sampleKey(w: CacheWorld): number {
  const c = w.config
  if (random(w, 'key') < c.hotShare / 100) return 0
  const u = random(w, 'key')
  const cdf = zipfCdf(c.keys, c.zipf)
  let lo = 0
  let hi = cdf.length - 1
  while (lo < hi) {
    const mid = (lo + hi) >> 1
    if (cdf[mid]! < u) lo = mid + 1
    else hi = mid
  }
  return lo
}

/** Какая доля чтений может попасть в кэш размера size при идеальном LRU и вечных записях. */
export function idealHitRatio(c: CacheConfig): number {
  if (c.cacheSize <= 0) return 0
  const cdf = zipfCdf(c.keys, c.zipf)
  const h = c.hotShare / 100
  return h + (1 - h) * (cdf[Math.min(c.keys, c.cacheSize) - 1] ?? 1)
}

/* ───────────────────────────── производные ───────────────────────────── */

export function multiplierAt(scenario: Pick<CacheScenario, 'phases'>, tick: number): number {
  let x = 1
  for (const p of [...scenario.phases].sort((a, b) => a.at - b.at)) if (p.at <= tick) x = p.x
  return x
}

export const dbCapacity = (c: CacheConfig) => c.dbWorkers / c.dbTime

export const cachedCount = (w: CacheWorld) => Object.keys(w.cache).length

export function percentile(values: number[], p: number): number {
  if (values.length === 0) return 0
  const sorted = [...values].sort((a, b) => a - b)
  const rank = Math.ceil((p / 100) * sorted.length)
  return sorted[Math.min(sorted.length, Math.max(1, rank)) - 1]!
}

/** Задержки чтений, завершённых на тиках (from, to]. */
export function readLatencies(done: readonly Completion[], to: number, from = 0): number[] {
  const out: number[] = []
  for (const c of done) if (c.t > from && c.t <= to && c.kind !== 'write') out.push(c.lat)
  return out
}
