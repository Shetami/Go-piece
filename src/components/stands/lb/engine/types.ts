/**
 * Типы модели балансировщика перед несколькими репликами сервиса.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Каждая реплика — тот же сервис, что в
 * стенде «Нагрузка»: несколько воркеров и очередь. Перед ними балансировщик:
 * он выбирает реплику по алгоритму, видит только число своих незавершённых
 * запросов к каждой, проверяет здоровье реплик и может исключать их из ротации.
 * Реплики ломаются по расписанию: тормозят, отвечают ошибками или зависают.
 * Автомасштабирование добавляет реплики, которым нужно время на запуск.
 *
 * Время — в тиках, сеть мгновенная, балансировщик один. Полный список
 * упрощений — в конце лекции.
 */

export type Algo = 'round-robin' | 'random' | 'least-conn' | 'p2c'

export const ALGOS: Algo[] = ['round-robin', 'random', 'least-conn', 'p2c']

export interface LbConfig {
  /** Сколько реплик в начале. */
  replicas: number
  /** Воркеров в каждой реплике. */
  workers: number
  /** Среднее время обработки на здоровой реплике, тиков. */
  serviceTime: number
  /** Сколько запросов приходит в среднем за тик — на весь сервис. */
  rate: number
  algo: Algo
  /** Таймаут клиента, тиков; 0 — ждать сколько угодно. */
  timeout: number
  /** Повторы на другой реплике после ошибки или таймаута. */
  retries: number
  /** Как часто балансировщик проверяет здоровье реплик, тиков; 0 — не проверяет. */
  healthEvery: number
  /** После скольких проваленных проверок подряд реплика исключается. */
  healthFails: number
  /** Пассивное исключение: реплика с несколькими ошибками подряд выводится из ротации на время. */
  outlier: boolean
  /** Автомасштабирование по загрузке. */
  autoscale: boolean
  /** Целевая загрузка реплик для автомасштабирования, процентов. */
  scaleTarget: number
  /** Сколько тиков запускается новая реплика. */
  bootTime: number
  maxReplicas: number
  seed: number
}

export const DEFAULT_LB_CONFIG: LbConfig = {
  replicas: 4,
  workers: 2,
  serviceTime: 4,
  rate: 1.2,
  algo: 'round-robin',
  timeout: 0,
  retries: 0,
  healthEvery: 5,
  healthFails: 2,
  outlier: false,
  autoscale: false,
  scaleTarget: 70,
  bootTime: 25,
  maxReplicas: 8,
  seed: 1,
}

/** Сколько ошибок подряд исключают реплику при пассивной проверке. */
export const OUTLIER_ERRORS = 3
/** На сколько тиков. */
export const OUTLIER_EJECT = 20
/** Как часто автомасштабирование принимает решение, тиков. */
export const SCALE_EVERY = 10

export interface Phase {
  at: number
  x: number
}

export type Fault =
  /** Реплика работает в factor раз медленнее: шумный сосед, деградация диска. */
  | { kind: 'slow'; replica: number; at: number; until?: number; factor: number }
  /** Процесс жив и проходит /healthz, но на запросы сразу отвечает 500: сломан пул соединений с базой. */
  | { kind: 'errors'; replica: number; at: number; until?: number }
  /** Процесс упал: соединения отклоняются, проверка здоровья проваливается. */
  | { kind: 'crash'; replica: number; at: number; until?: number }
  /** Реплика принимает запросы и не отвечает: дедлок, бесконечная пауза. */
  | { kind: 'hang'; replica: number; at: number; until?: number }

export type ReplicaState = 'up' | 'slow' | 'errors' | 'crash' | 'hang' | 'booting'

export interface Request {
  id: number
  origin: number
  attempt: number
  bornAt: number
  arriveAt: number
  deadline: number | null
  service: number
  startAt: number | null
  finishAt: number | null
  /** Клиент ушёл по таймауту. */
  abandoned: boolean
}

export interface Replica {
  id: number
  name: string
  state: ReplicaState
  /** Во сколько раз медленнее обычного. */
  factor: number
  workers: (Request | null)[]
  queue: Request[]
  bootUntil: number
  /** Балансировщик считает реплику живой и шлёт на неё запросы. */
  inRotation: boolean
  /** Исключена пассивно до этого тика; 0 — нет. */
  ejectedUntil: number
  /** Незавершённые запросы балансировщика к реплике — всё, что он о ней знает. */
  outstanding: number
  errorsInRow: number
  failedChecks: number
  passedChecks: number
  served: number
  errors: number
  /** Запросов отправлено на реплику за прогон. */
  routed: number
}

export interface PendingRetry {
  origin: number
  attempt: number
  bornAt: number
  at: number
  /** Реплика, на которой провалилась прошлая попытка, — её обходим. */
  avoid: number
}

export interface Completion {
  t: number
  lat: number
  replica: number
  outcome: 'ok' | 'error' | 'wasted'
}

export interface Sample {
  arrived: number
  ok: number
  errors: number
  timeouts: number
  x: number
  /** Незавершённые запросы по репликам: очередь плюс в работе. */
  load: number[]
  /** Реплик в ротации. */
  inRotation: number
  /** Реплик всего, включая запускающиеся. */
  replicas: number
}

export type LbEventType =
  | 'req.route'
  | 'req.done'
  | 'req.error'
  | 'req.timeout'
  | 'req.retry'
  | 'req.fail'
  | 'req.nobackend'
  | 'load.change'
  | 'replica.fault'
  | 'replica.recover'
  | 'health.fail'
  | 'health.eject'
  | 'health.restore'
  | 'outlier.eject'
  | 'outlier.return'
  | 'queue.grow'
  | 'scale.decide'
  | 'scale.ready'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<LbEventType, Importance> = {
  'req.route': 'low',
  'req.done': 'low',
  'req.error': 'normal',
  'req.timeout': 'normal',
  'req.retry': 'normal',
  'req.fail': 'normal',
  'req.nobackend': 'normal',
  'load.change': 'key',
  'replica.fault': 'key',
  'replica.recover': 'key',
  'health.fail': 'normal',
  'health.eject': 'key',
  'health.restore': 'key',
  'outlier.eject': 'key',
  'outlier.return': 'normal',
  'queue.grow': 'key',
  'scale.decide': 'key',
  'scale.ready': 'key',
}

export interface LbEvent {
  tick: number
  type: LbEventType
  actors: { replica?: number[]; req?: number[] }
  payload: Record<string, unknown>
}

export interface LbScenario {
  id: string
  title: string
  claim: string
  config: Partial<LbConfig>
  phases: Phase[]
  faults: Fault[]
  watchFor: LbEventType[]
  stopAfter: number
}

export interface LbStats {
  logical: number
  attempts: number
  ok: number
  errors: number
  timeouts: number
  retries: number
  wasted: number
  /** Пользователь не получил ответа. */
  failed: number
  noBackend: number
  /** Самая длинная очередь на одной реплике. */
  maxQueue: number
}

export type Stream = 'arrive' | 'work' | 'route'

export interface LbWorld {
  tick: number
  config: LbConfig
  rng: Record<Stream, number>
  replicas: Replica[]
  retryQ: PendingRetry[]
  nextId: number
  nextOrigin: number
  /** Указатель круговой балансировки. */
  rr: number
  x: number
  /** Следующая длина очереди реплики, о которой стоит сообщить, — по репликам. */
  qMilestone: number[]
  /** Сумма занятых воркеров в ротации и их число с прошлого решения автомасштабирования. */
  busyAcc: number
  capAcc: number
  stats: LbStats
  finished: boolean
  finishReason?: 'stop-after'
}

export interface Frame {
  world: LbWorld
  events: LbEvent[]
}

/** Журнал прогона: только дописывается и в снимки не копируется. */
export interface LbLog {
  series: Sample[]
  done: Completion[]
}
