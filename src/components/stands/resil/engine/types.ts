/**
 * Типы модели сервиса с зависимостью.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Сервис A принимает запросы и часть из
 * них не может обслужить сам — ему нужен сервис B. Пока A ждёт ответа B, его
 * воркер занят: именно так чужая медлительность превращается в свою.
 *
 * Дальше в модель добавлены обычные средства защиты: таймаут вызова, повторы,
 * предохранитель (circuit breaker), переборки (отдельный лимит воркеров на
 * вызовы B), ограничитель частоты на входе и ответ-заглушка вместо ошибки.
 *
 * Время — в тиках, сеть мгновенная. Полный список упрощений — в конце лекции.
 */

export type BreakerState = 'closed' | 'open' | 'half-open'

export interface ResilConfig {
  /** Воркеров в сервисе A. */
  workers: number
  /** Сколько тиков A работает сам над запросом. */
  ownTime: number
  /** Доля запросов, которым нужен сервис B, процентов. */
  depShare: number
  rate: number
  /** Воркеров в сервисе B. */
  depWorkers: number
  /** Среднее время ответа B, тиков. */
  depTime: number
  /** Таймаут вызова B; 0 — ждать сколько угодно. */
  timeout: number
  /** Повторы вызова B после ошибки или таймаута. */
  retries: number
  breaker: boolean
  /** Сколько неудач подряд размыкают предохранитель. */
  breakerFails: number
  /** На сколько тиков он размыкается. */
  breakerOpen: number
  /** Переборки: сколько воркеров A могут одновременно ждать B; 0 — все. */
  bulkhead: number
  /** Ограничитель частоты на входе, запросов за тик; 0 — выключен. */
  rateLimit: number
  /** Запас ограничителя: сколько запросов можно принять сверх нормы разом. */
  burst: number
  /** Отвечать заглушкой вместо ошибки, когда B недоступен. */
  fallback: boolean
  seed: number
}

export const DEFAULT_RESIL_CONFIG: ResilConfig = {
  workers: 6,
  ownTime: 2,
  depShare: 50,
  rate: 1.2,
  depWorkers: 4,
  depTime: 4,
  timeout: 0,
  retries: 0,
  breaker: false,
  breakerFails: 5,
  breakerOpen: 20,
  bulkhead: 0,
  rateLimit: 0,
  burst: 5,
  fallback: false,
  seed: 1,
}

export interface Phase {
  at: number
  x: number
}

export type Fault =
  /** B отвечает в factor раз медленнее. */
  | { kind: 'slow'; at: number; until?: number; factor: number }
  /** B отвечает ошибками. */
  | { kind: 'errors'; at: number; until?: number }
  /** B не отвечает вовсе: только таймаут. */
  | { kind: 'hang'; at: number; until?: number }

export type DepState = 'ok' | 'slow' | 'errors' | 'hang'

export type Stage = 'queue' | 'own' | 'dep' | 'done'

export interface Request {
  id: number
  /** Нужен ли этому запросу сервис B. */
  needsDep: boolean
  bornAt: number
  stage: Stage
  /** Когда закончится текущий шаг. */
  until: number
  /** Занятый воркер A. */
  worker: number | null
  attempt: number
  /** Тик, когда начался текущий вызов B. */
  calledAt: number | null
  /** Место в очереди B. */
  inDep: boolean
}

export interface DepCall {
  id: number
  req: number
  arriveAt: number
  startAt: number | null
  finishAt: number | null
}

export interface Completion {
  t: number
  lat: number
  kind: 'ok' | 'degraded' | 'error' | 'rejected'
  needsDep: boolean
}

export interface Sample {
  arrived: number
  ok: number
  degraded: number
  errors: number
  rejected: number
  /** Воркеров A занято всего и из них ждут B. */
  busy: number
  waiting: number
  queue: number
  depQueue: number
  x: number
}

export type ResilEventType =
  | 'req.accept'
  | 'req.reject'
  | 'req.done'
  | 'dep.call'
  | 'dep.ok'
  | 'dep.timeout'
  | 'dep.error'
  | 'dep.retry'
  | 'req.fallback'
  | 'req.fail'
  | 'breaker.open'
  | 'breaker.half'
  | 'breaker.close'
  | 'pool.saturated'
  | 'bulkhead.block'
  | 'queue.grow'
  | 'dep.fault'
  | 'dep.recover'
  | 'load.change'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<ResilEventType, Importance> = {
  'req.accept': 'low',
  'req.reject': 'normal',
  'req.done': 'low',
  'dep.call': 'low',
  'dep.ok': 'low',
  'dep.timeout': 'normal',
  'dep.error': 'normal',
  'dep.retry': 'normal',
  'req.fallback': 'normal',
  'req.fail': 'normal',
  'breaker.open': 'key',
  'breaker.half': 'normal',
  'breaker.close': 'key',
  'pool.saturated': 'key',
  'bulkhead.block': 'normal',
  'queue.grow': 'key',
  'dep.fault': 'key',
  'dep.recover': 'key',
  'load.change': 'key',
}

export interface ResilEvent {
  tick: number
  type: ResilEventType
  actors: { req?: number[]; worker?: number[] }
  payload: Record<string, unknown>
}

export interface ResilScenario {
  id: string
  title: string
  claim: string
  config: Partial<ResilConfig>
  phases: Phase[]
  faults: Fault[]
  watchFor: ResilEventType[]
  stopAfter: number
}

export interface ResilStats {
  arrived: number
  ok: number
  /** Ответы-заглушки: пользователь получил урезанный ответ. */
  degraded: number
  errors: number
  rejected: number
  depCalls: number
  depTimeouts: number
  depErrors: number
  retries: number
  blocked: number
  /** Сумма занятых воркеров A и из них ждущих B — по тикам. */
  busyTicks: number
  waitTicks: number
  maxQueue: number
  /** Запросы, которые не нуждались в B, но пострадали: ждали в очереди дольше двух своих времён. */
  localSlow: number
}

export type Stream = 'arrive' | 'kind' | 'work'

export interface ResilWorld {
  tick: number
  config: ResilConfig
  rng: Record<Stream, number>
  dep: DepState
  depFactor: number
  /** Воркеры A: id запроса или null. */
  workers: (number | null)[]
  queue: number[]
  requests: Record<number, Request>
  depQueue: DepCall[]
  depWorkers: (DepCall | null)[]
  /** Предохранитель. */
  breaker: BreakerState
  breakerFails: number
  breakerUntil: number
  /** В полуоткрытом состоянии пропущен пробный вызов и ответа ещё нет. */
  probing: boolean
  /** Токены ограничителя частоты. */
  tokens: number
  nextId: number
  x: number
  qMilestone: number
  saturated: boolean
  stats: ResilStats
  finished: boolean
  finishReason?: 'stop-after'
}

export interface Frame {
  world: ResilWorld
  events: ResilEvent[]
}

/** Журнал прогона: только дописывается и в снимки не копируется. */
export interface ResilLog {
  series: Sample[]
  done: Completion[]
}
