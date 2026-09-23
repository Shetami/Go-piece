import type { Completion, Sample, ShardConfig, ShardEvent, ShardEventType, ShardLog, ShardScenario, ShardWorld, Stream } from './types.ts'
import { DEFAULT_SHARD_CONFIG } from './types.ts'

export interface Ctx {
  world: ShardWorld
  events: ShardEvent[]
  scenario: ShardScenario
  log: ShardLog
}

export function resolveConfig(partial: Partial<ShardConfig> | undefined): ShardConfig {
  return { ...DEFAULT_SHARD_CONFIG, ...(partial ?? {}) }
}

export const newShard = (id: number, workers: number) => ({
  id,
  name: `s${id + 1}`,
  workers: Array.from({ length: workers }, () => null),
  queue: [],
  served: 0,
  owns: 0,
})

export function createWorld(scenario: ShardScenario): ShardWorld {
  const config = resolveConfig(scenario.config)
  const shards = Array.from({ length: config.shards }, (_, i) => newShard(i, config.workers))
  const owner = Array.from({ length: config.keys }, (_, k) => ownerOf(config, k, config.shards))
  for (const o of owner) shards[o]!.owns++
  return {
    tick: 0,
    config,
    rng: { arrive: seedOf(config.seed, 1), key: seedOf(config.seed, 2), work: seedOf(config.seed, 3) },
    shards,
    owner,
    moving: {},
    groups: [],
    nextId: 1,
    nextGroup: 1,
    x: multiplierAt(scenario, 0),
    qMilestone: Array.from({ length: config.shards }, () => FIRST_MILESTONE),
    skewed: false,
    stats: { requests: 0, subRequests: 0, done: 0, scatter: 0, migratingHits: 0, moved: 0, maxQueue: 0 },
    finished: false,
  }
}

export const FIRST_MILESTONE = 8

export function emit(ctx: Ctx, type: ShardEventType, actors: ShardEvent['actors'] = {}, payload: Record<string, unknown> = {}): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export const emptySample = (x: number, n: number): Sample => ({ arrived: 0, done: 0, load: Array.from({ length: n }, () => 0), migrating: 0, x })

/* ───────────────────────────── случайность ───────────────────────────── */

const seedOf = (seed: number, stream: number) => (Math.imul((seed | 0) + stream * 0x632be5ab, 0x9e3779b1) ^ 0x5bd1e995) >>> 0

export function random(w: ShardWorld, stream: Stream): number {
  w.rng[stream] = (w.rng[stream] + 0x6d2b79f5) >>> 0
  let t = w.rng[stream]
  t = Math.imul(t ^ (t >>> 15), t | 1)
  t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296
}

export function poisson(w: ShardWorld, lambda: number): number {
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

export function geometric(w: ShardWorld, mean: number): number {
  if (mean <= 1) return 1
  const u = 1 - random(w, 'work')
  return 1 + Math.floor(Math.log(u) / Math.log(1 - 1 / mean))
}

/* ───────────────────────────── раскладка ───────────────────────────── */

/** Хеш ключа — детерминированный и «перемешивающий», как настоящий хеш от id. */
export function hash(x: number, salt = 0): number {
  let h = Math.imul(x ^ 0x9e3779b9, 0x85ebca6b) ^ Math.imul(salt + 1, 0xc2b2ae35)
  h = Math.imul(h ^ (h >>> 13), 0x27d4eb2f)
  return ((h ^ (h >>> 16)) >>> 0) / 4294967296
}

/**
 * Кольцо согласованного хеширования: у каждого шарда vnodes точек на окружности,
 * ключ достаётся ближайшей точке по часовой стрелке. Кольцо считается заново
 * при каждом вопросе, зато в мире не нужно хранить его состояние.
 */
function ringOwner(key: number, shards: number, vnodes: number): number {
  const k = hash(key, 7)
  let best = -1
  let bestPos = Number.POSITIVE_INFINITY
  let first = -1
  let firstPos = Number.POSITIVE_INFINITY
  for (let s = 0; s < shards; s++) {
    for (let v = 0; v < vnodes; v++) {
      const pos = hash(s * 1000 + v, 13)
      if (pos < firstPos) {
        firstPos = pos
        first = s
      }
      if (pos >= k && pos < bestPos) {
        bestPos = pos
        best = s
      }
    }
  }
  return best === -1 ? first : best
}

/** Какому шарду принадлежит ключ при таком числе шардов. */
export function ownerOf(c: ShardConfig, key: number, shards: number): number {
  switch (c.scheme) {
    case 'mod':
      // Хеш от ключа по модулю числа шардов: при смене числа шардов переезжает почти всё.
      return Math.floor(hash(key) * 1e9) % shards
    case 'ring':
      return ringOwner(key, shards, Math.max(1, c.vnodes))
    case 'range':
      // Диапазоны по возрастанию ключа: соседние ключи лежат рядом.
      return Math.min(shards - 1, Math.floor((key / c.keys) * shards))
  }
}

/** Доля ключей, которая сменит владельца при переходе с from на to шардов. */
export function movedShare(c: ShardConfig, from: number, to: number): number {
  let moved = 0
  for (let k = 0; k < c.keys; k++) if (ownerOf(c, k, from) !== ownerOf(c, k, to)) moved++
  return moved / c.keys
}

const cdfCache = new Map<string, number[]>()
export function zipfCdf(n: number, s: number): number[] {
  const key = `${n}:${s}`
  const hit = cdfCache.get(key)
  if (hit) return hit
  const w = Array.from({ length: n }, (_, i) => 1 / (i + 1) ** s)
  const total = w.reduce((a, b) => a + b, 0)
  let acc = 0
  const cdf = w.map((x) => (acc += x / total))
  cdfCache.set(key, cdf)
  return cdf
}

export function sampleKey(w: ShardWorld): number {
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

/* ───────────────────────────── производные ───────────────────────────── */

export function multiplierAt(scenario: Pick<ShardScenario, 'phases'>, tick: number): number {
  let x = 1
  for (const p of [...scenario.phases].sort((a, b) => a.at - b.at)) if (p.at <= tick) x = p.x
  return x
}

export const busyOf = (s: { workers: (unknown | null)[] }) => s.workers.filter((x) => x !== null).length

export const loadOf = (s: { workers: (unknown | null)[]; queue: unknown[] }) => busyOf(s) + s.queue.length

/** Во сколько раз самый загруженный шард загружен сильнее среднего. */
export function imbalance(w: ShardWorld): number {
  const loads = w.shards.map((s) => s.served)
  const total = loads.reduce((a, b) => a + b, 0)
  if (total === 0) return 1
  return (Math.max(...loads) * w.shards.length) / total
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
