/**
 * Типы модели шардированного хранилища.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Данные разрезаны по ключу на шарды;
 * каждый шард — сервис с воркерами и очередью, как в стенде «Нагрузка».
 * Способ раскладки — остаток от деления, кольцо согласованного хеширования
 * или диапазоны. Шарды можно добавлять на ходу: ключи, сменившие владельца,
 * переезжают, и до переезда запрос к ним стоит дороже. Часть запросов идёт
 * без ключа шардирования — такие спрашивают все шарды сразу.
 *
 * Время — в тиках, сеть мгновенная, данные не теряются. Полный список
 * упрощений — в конце лекции.
 */

export type Scheme = 'mod' | 'ring' | 'range'

export const SCHEMES: Scheme[] = ['mod', 'ring', 'range']

export interface ShardConfig {
  /** Сколько шардов в начале. */
  shards: number
  scheme: Scheme
  /** Виртуальных узлов на шард для кольца: чем больше, тем ровнее раскладка. */
  vnodes: number
  keys: number
  /** Неравномерность популярности ключей (распределение Ципфа). */
  zipf: number
  /** Доля запросов к самому горячему ключу сверх распределения, процентов. */
  hotShare: number
  /** Доля запросов без ключа шардирования: их приходится задавать всем шардам. */
  scatterShare: number
  rate: number
  workers: number
  serviceTime: number
  /** Сколько ключей переезжает за тик при решардировании. */
  migrateRate: number
  seed: number
}

export const DEFAULT_SHARD_CONFIG: ShardConfig = {
  shards: 3,
  scheme: 'mod',
  vnodes: 64,
  keys: 240,
  zipf: 0.8,
  hotShare: 0,
  scatterShare: 0,
  rate: 2,
  workers: 2,
  serviceTime: 4,
  migrateRate: 1,
  seed: 1,
}

export interface Phase {
  at: number
  x: number
}

export type Fault =
  /** Добавить шард на ходу. */
  | { kind: 'add'; at: number }
  /** Убрать последний шард. */
  | { kind: 'remove'; at: number }

export interface Request {
  id: number
  key: number
  /** Часть веерного запроса: id логического запроса. */
  group: number | null
  bornAt: number
  arriveAt: number
  startAt: number | null
  finishAt: number | null
  /** Ключ ещё не переехал: шард сначала сходит за ним в старый шард. */
  migrating: boolean
}

export interface Shard {
  id: number
  name: string
  workers: (Request | null)[]
  queue: Request[]
  /** Обработано запросов за прогон. */
  served: number
  /** Сколько ключей принадлежит шарду сейчас. */
  owns: number
}

/** Незакрытый веерный запрос: ждёт ответа от всех шардов. */
export interface Group {
  id: number
  bornAt: number
  left: number
  shards: number
}

export interface Completion {
  t: number
  lat: number
  shard: number
  kind: 'point' | 'scatter'
  migrating: boolean
}

export interface Sample {
  arrived: number
  done: number
  /** Запросов в очереди и в работе по шардам. */
  load: number[]
  /** Ключей осталось перевезти. */
  migrating: number
  x: number
}

export type ShardEventType =
  | 'req.route'
  | 'req.done'
  | 'scatter.start'
  | 'scatter.done'
  | 'req.migrating'
  | 'reshard.start'
  | 'reshard.move'
  | 'reshard.done'
  | 'shard.skew'
  | 'queue.grow'
  | 'load.change'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<ShardEventType, Importance> = {
  'req.route': 'low',
  'req.done': 'low',
  'scatter.start': 'low',
  'scatter.done': 'normal',
  'req.migrating': 'normal',
  'reshard.start': 'key',
  'reshard.move': 'low',
  'reshard.done': 'key',
  'shard.skew': 'key',
  'queue.grow': 'key',
  'load.change': 'key',
}

export interface ShardEvent {
  tick: number
  type: ShardEventType
  actors: { shard?: number[]; key?: number[] }
  payload: Record<string, unknown>
}

export interface ShardScenario {
  id: string
  title: string
  claim: string
  config: Partial<ShardConfig>
  phases: Phase[]
  faults: Fault[]
  watchFor: ShardEventType[]
  stopAfter: number
}

export interface ShardStats {
  requests: number
  /** Запросов, отправленных в шарды: веерный считается за столько, во сколько шардов ушёл. */
  subRequests: number
  done: number
  scatter: number
  /** Запросов к ключам, которые ещё не переехали. */
  migratingHits: number
  /** Ключей переехало за прогон. */
  moved: number
  maxQueue: number
}

export type Stream = 'arrive' | 'key' | 'work'

export interface ShardWorld {
  tick: number
  config: ShardConfig
  rng: Record<Stream, number>
  shards: Shard[]
  /** Владелец каждого ключа сейчас. */
  owner: number[]
  /** Ключи, которые ещё не переехали: ключ → старый владелец. */
  moving: Record<number, number>
  groups: Group[]
  nextId: number
  nextGroup: number
  x: number
  qMilestone: number[]
  /** Был ли уже отмечен перекос. */
  skewed: boolean
  stats: ShardStats
  finished: boolean
  finishReason?: 'stop-after'
}

export interface Frame {
  world: ShardWorld
  events: ShardEvent[]
}

/** Журнал прогона: только дописывается и в снимки не копируется. */
export interface ShardLog {
  series: Sample[]
  done: Completion[]
}
