/**
 * Типы модели кэша перед базой данных (cache-aside).
 *
 * Как и остальные стенды, это МОДЕЛЬ. Приложение сначала смотрит в кэш; при
 * промахе идёт в базу и кладёт ответ в кэш. База — тот же сервис с воркерами
 * и очередью, что в стенде «Нагрузка». Кэш — LRU ограниченного размера,
 * записи живут TTL тиков. Ключи популярны неравномерно (распределение Ципфа),
 * часть запросов — записи, которые меняют значение в базе и инвалидируют кэш.
 *
 * Время — в тиках, кэш отвечает за тик, сеть мгновенная. Полный список
 * упрощений — в конце лекции.
 */

export type Invalidation = 'none' | 'delete' | 'update'

export interface CacheConfig {
  /** Сколько разных ключей. */
  keys: number
  /** Показатель распределения Ципфа: 0 — все ключи одинаково популярны. */
  zipf: number
  /** Доля запросов к самому горячему ключу сверх распределения, процентов. */
  hotShare: number
  /** Запросов за тик. */
  rate: number
  /** Доля записей, процентов. */
  writeShare: number
  /** Сколько ключей помещается в кэш; 0 — кэша нет. */
  cacheSize: number
  /** Сколько тиков живёт запись; 0 — вечно. */
  ttl: number
  /** Случайная добавка к TTL, процентов. */
  ttlJitter: number
  /** Объединять одинаковые запросы в базу: пока один идёт, остальные ждут его ответа. */
  coalesce: boolean
  /** Отдавать просроченное значение сразу и обновлять его в фоне. */
  staleWhileRevalidate: boolean
  /** Что делать с кэшем при записи: ничего, удалить ключ или положить новое значение. */
  invalidation: Invalidation
  dbWorkers: number
  /** Среднее время запроса в базу, тиков. */
  dbTime: number
  /** Среднее время запроса за самый горячий ключ — обычно это тяжёлый агрегат; 0 — как у остальных. */
  hotTime: number
  seed: number
}

export const DEFAULT_CACHE_CONFIG: CacheConfig = {
  keys: 200,
  zipf: 1,
  hotShare: 0,
  rate: 2,
  writeShare: 0,
  cacheSize: 20,
  ttl: 0,
  ttlJitter: 0,
  coalesce: false,
  staleWhileRevalidate: false,
  invalidation: 'delete',
  dbWorkers: 6,
  dbTime: 4,
  hotTime: 0,
  seed: 1,
}

export interface Phase {
  at: number
  x: number
}

export type Fault =
  /** Кэш перезапустился и опустел. */
  | { kind: 'flush'; at: number }
  /** Прогрев: заранее положить в кэш самые популярные ключи — все в один тик. */
  | { kind: 'warm'; at: number; count: number }

export interface Entry {
  version: number
  /** Когда запись просрочится; Infinity — никогда. */
  expiresAt: number
  lastUsed: number
  setAt: number
}

/** Кто ждёт ответа из базы. */
export interface Waiter {
  id: number
  bornAt: number
}

export interface Query {
  id: number
  kind: 'read' | 'write' | 'refresh'
  key: number
  waiters: Waiter[]
  arriveAt: number
  startAt: number | null
  finishAt: number | null
  /** Версия, которую чтение увидело в базе в момент начала. */
  version: number | null
}

export interface Completion {
  t: number
  lat: number
  kind: 'hit' | 'miss' | 'write' | 'swr'
}

export interface Sample {
  arrived: number
  hits: number
  misses: number
  stale: number
  /** Запросов в базу, начатых на этом тике. */
  dbStarted: number
  dbQueue: number
  dbBusy: number
  cached: number
  x: number
}

export type CacheEventType =
  | 'req.hit'
  | 'req.miss'
  | 'req.join'
  | 'req.stale'
  | 'req.swr'
  | 'db.done'
  | 'write.done'
  | 'cache.expired'
  | 'cache.flush'
  | 'cache.warm'
  | 'cache.stale-set'
  | 'db.stampede'
  | 'db.queue'
  | 'db.drained'
  | 'load.change'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<CacheEventType, Importance> = {
  'req.hit': 'low',
  'req.miss': 'low',
  'req.join': 'normal',
  'req.stale': 'normal',
  'req.swr': 'normal',
  'db.done': 'low',
  'write.done': 'low',
  'cache.expired': 'normal',
  'cache.flush': 'key',
  'cache.warm': 'key',
  'cache.stale-set': 'key',
  'db.stampede': 'key',
  'db.queue': 'key',
  'db.drained': 'key',
  'load.change': 'key',
}

export interface CacheEvent {
  tick: number
  type: CacheEventType
  actors: { key?: number[]; query?: number[] }
  payload: Record<string, unknown>
}

export interface CacheScenario {
  id: string
  title: string
  claim: string
  config: Partial<CacheConfig>
  phases: Phase[]
  faults: Fault[]
  watchFor: CacheEventType[]
  stopAfter: number
}

export interface CacheStats {
  reads: number
  writes: number
  hits: number
  misses: number
  joins: number
  swr: number
  /** Попадания, отдавшие значение старее, чем в базе. */
  stale: number
  dbReads: number
  dbWrites: number
  maxDbQueue: number
  /** Самый большой всплеск запросов в базу за один ключ. */
  maxSameKey: number
}

export type Stream = 'arrive' | 'key' | 'work' | 'ttl'

export interface CacheWorld {
  tick: number
  config: CacheConfig
  rng: Record<Stream, number>
  /** Кэш: ключ → запись. */
  cache: Record<number, Entry>
  /** Версии значений в базе. */
  db: number[]
  queue: Query[]
  workers: (Query | null)[]
  /** Сколько запросов в базу за ключ сейчас в пути. */
  inflight: Record<number, number>
  nextId: number
  x: number
  qMilestone: number
  stampedeKeys: number[]
  stats: CacheStats
  finished: boolean
  finishReason?: 'stop-after'
}

export interface Frame {
  world: CacheWorld
  events: CacheEvent[]
}

/** Журнал прогона: только дописывается и в снимки не копируется. */
export interface CacheLog {
  series: Sample[]
  done: Completion[]
}
