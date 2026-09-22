/**
 * Типы модели сервиса под нагрузкой.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Сервис — пул одинаковых воркеров (горутин,
 * потоков, соединений с базой) и очередь перед ними. Запросы приходят потоком,
 * ждут свободного воркера, обрабатываются и уходят. Клиент ждёт ответа не дольше
 * таймаута, а потом бросает запрос и, возможно, повторяет его.
 *
 * Этого хватает, чтобы воспроизвести главное из теории очередей: закон Литтла,
 * взрыв задержки у высокой загрузки, хвост из медленных запросов, долгое
 * рассасывание очереди после всплеска и метастабильный отказ от повторов.
 * Время — в тиках, воркеры одинаковые, сеть мгновенная. Полный список
 * упрощений — в конце лекции.
 */

/** Как приходят запросы: ровно через равные промежутки или случайно, пуассоновским потоком. */
export type Arrivals = 'even' | 'random'

/** Время обработки: всегда одно и то же или случайное с тем же средним. */
export type ServiceDist = 'fixed' | 'random'

export interface LoadConfig {
  /** Сколько запросов сервис обрабатывает одновременно. */
  workers: number
  /** Среднее время обработки запроса, тиков. */
  serviceTime: number
  serviceDist: ServiceDist
  /** Доля медленных запросов, процентов. */
  slowShare: number
  /** Во сколько раз медленный запрос дольше обычного. */
  slowFactor: number
  /** Сколько запросов в среднем приходит за тик — до множителей фаз нагрузки. */
  rate: number
  arrivals: Arrivals
  /** Таймаут клиента, тиков; 0 — ждать сколько угодно. */
  timeout: number
  /** Сколько раз клиент повторяет запрос после таймаута или отказа. */
  retries: number
  /** Экспоненциальная пауза со случайным разбросом между повторами; иначе повтор сразу. */
  backoff: boolean
  /** Сервис бросает запрос, чей клиент уже ушёл: отмена через context. */
  cancelOnTimeout: boolean
  /** Сколько запросов может ждать в очереди; 0 — без предела. Лишним сразу отказ. */
  queueLimit: number
  /** Зерно генератора случайных чисел: один и тот же прогон при каждом запуске. */
  seed: number
}

export const DEFAULT_LOAD_CONFIG: LoadConfig = {
  workers: 4,
  serviceTime: 4,
  serviceDist: 'random',
  slowShare: 0,
  slowFactor: 12,
  rate: 0.6,
  arrivals: 'random',
  timeout: 0,
  retries: 0,
  backoff: false,
  cancelOnTimeout: false,
  queueLimit: 0,
  seed: 7,
}

/** Фаза нагрузки: с тика at поток умножается на x. */
export interface Phase {
  at: number
  x: number
}

/** Одна попытка запроса. Повтор — новая попытка того же логического запроса. */
export interface Request {
  id: number
  /** Номер логического запроса — того, чего ждёт пользователь. */
  origin: number
  /** 0 — первая попытка, дальше повторы. */
  attempt: number
  /** Когда пользователь начал ждать: приход первой попытки. */
  bornAt: number
  arriveAt: number
  /** Тик, после которого клиент перестаёт ждать; null — ждёт сколько угодно. */
  deadline: number | null
  /** Сколько тиков нужно на обработку. */
  service: number
  slow: boolean
  startAt: number | null
  finishAt: number | null
  /** Клиент ушёл, а запрос остался в сервисе. */
  abandoned: boolean
}

/** Повтор, который клиент отправит позже. */
export interface PendingRetry {
  origin: number
  attempt: number
  bornAt: number
  at: number
}

/** Запрос покинул сервис: ответил, отработал впустую или был отменён. */
export interface Completion {
  t: number
  id: number
  /** Сколько ждал пользователь — с первой попытки. */
  lat: number
  /** Сколько эта попытка пробыла в сервисе. */
  sojourn: number
  /** Сколько из них простояла в очереди. */
  wait: number
  outcome: 'ok' | 'wasted' | 'cancelled'
  slow: boolean
}

/** Что происходило на тике — точка графиков. */
export interface Sample {
  /** Пришло попыток, включая повторы. */
  arrived: number
  retried: number
  rejected: number
  queue: number
  busy: number
  ok: number
  wasted: number
  timeouts: number
  /** Множитель нагрузки на этом тике. */
  x: number
}

export type LoadEventType =
  | 'req.arrive'
  | 'req.start'
  | 'req.done'
  | 'req.slow'
  | 'req.retry'
  | 'req.timeout'
  | 'req.wasted'
  | 'req.reject'
  | 'req.fail'
  | 'load.change'
  | 'queue.grow'
  | 'queue.drained'
  | 'overload.collapse'
  | 'overload.recover'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<LoadEventType, Importance> = {
  'req.arrive': 'low',
  'req.start': 'low',
  'req.done': 'low',
  'req.slow': 'normal',
  'req.retry': 'normal',
  'req.timeout': 'normal',
  'req.wasted': 'normal',
  'req.reject': 'normal',
  'req.fail': 'normal',
  'load.change': 'key',
  'queue.grow': 'key',
  'queue.drained': 'key',
  'overload.collapse': 'key',
  'overload.recover': 'key',
}

export interface LoadEvent {
  tick: number
  type: LoadEventType
  /** Попытки запросов и воркеры, о которых событие. */
  actors: { req?: number[]; worker?: number[] }
  payload: Record<string, unknown>
}

export interface LoadScenario {
  id: string
  title: string
  claim: string
  config: Partial<LoadConfig>
  phases: Phase[]
  watchFor: LoadEventType[]
  stopAfter: number
}

export interface LoadStats {
  /** Попыток, включая повторы. */
  arrivals: number
  /** Логических запросов — того, что прислали пользователи. */
  logical: number
  /** Попыток, попавших в очередь, — без отказов. */
  admitted: number
  ok: number
  wasted: number
  cancelled: number
  timeouts: number
  retries: number
  rejected: number
  /** Пользователь так и не получил ответа: кончились повторы. */
  failed: number
  /** Сумма занятых воркеров по тикам — для загрузки. */
  busyTicks: number
  /** Сумма запросов в сервисе по тикам — для закона Литтла. */
  inSystemSum: number
  /** Сумма времени в сервисе по ушедшим попыткам. */
  sojournSum: number
  departures: number
  maxQueue: number
}

/** Приход запросов, время их обработки, паузы перед повторами. */
export type Stream = 'arrive' | 'work' | 'retry'

export interface LoadWorld {
  tick: number
  config: LoadConfig
  /**
   * Состояния генераторов случайных чисел. Потоки независимы: поменяйте долю
   * медленных запросов — и запросы придут в те же тики, что и раньше.
   */
  rng: Record<Stream, number>
  queue: Request[]
  workers: (Request | null)[]
  retryQ: PendingRetry[]
  nextId: number
  nextOrigin: number
  /** Текущий множитель нагрузки. */
  x: number
  /** Дробный остаток для равномерного потока. */
  carry: number
  /** Чем кончилась работа воркеров над последними запросами — чтобы заметить работу впустую. */
  recent: ('ok' | 'wasted')[]
  stats: LoadStats
  /** Следующая длина очереди, о которой стоит сообщить. */
  qMilestone: number
  /** Самая длинная очередь с тех пор, как она последний раз была пуста. */
  qPeak: number
  /** Тик, когда очередь начала расти. */
  qSince: number | null
  /** С какого тика сервис почти весь работает впустую; null — работает с пользой. */
  collapsedAt: number | null
  finished: boolean
  finishReason?: 'stop-after'
}

export interface Frame {
  world: LoadWorld
  events: LoadEvent[]
}

/**
 * Журнал прогона. Только дописывается, поэтому в снимки не копируется:
 * графики и перцентили на тике t читают из него всё, что было до t.
 */
export interface LoadLog {
  /** series[t] — тик t; series[0] пустой. */
  series: Sample[]
  done: Completion[]
}
