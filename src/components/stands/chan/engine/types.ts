/**
 * Типы модели каналов Go.
 *
 * Как и в соседних стендах, это МОДЕЛЬ. Устройство `hchan` — кольцевой буфер,
 * очереди ожидания, порядок проверок в chansend/chanrecv — взято из рантайма
 * почти без изменений. А вот время измеряется в тиках, значение в канале — это
 * просто id горутины-отправителя, и планировщик здесь предельно упрощён:
 * никаких M, кражи работы и вытеснения. Полный список упрощений — в конце лекции.
 */

/** Длительность в тиках: точное число или диапазон [min, max]. */
export type Duration = number | [number, number]

/** Описание канала в сценарии. `cap: 0` — небуферизованный, точка встречи. */
export interface ChanSpec {
  name: string
  cap: number
  /**
   * Канал, который объявили, но не создали (`var ch chan int`).
   * Любая операция на нём блокируется навсегда.
   */
  nil?: boolean
}

/** Один case оператора select. */
export interface SelectCase {
  chan: string
  op: 'send' | 'recv'
}

/** Что делает горутина. */
export type ChanPhase =
  /** Считает, каналов не трогает. */
  | { kind: 'cpu'; ticks: Duration }
  /** Отправляет `count` значений в канал. */
  | { kind: 'send'; chan: string; count?: number }
  /** Принимает `count` значений. */
  | { kind: 'recv'; chan: string; count?: number }
  /** Выбор из нескольких операций. `default: true` — не блокироваться, если никто не готов. */
  | { kind: 'select'; cases: SelectCase[]; default?: boolean; count?: number }
  /** Закрывает канал. */
  | { kind: 'close'; chan: string }

export interface ChanWorkload {
  name: string
  count: number
  /** На каком тике рождаются: число — все сразу, 'staggered' — по одной в тик. */
  spawnAt: number | 'staggered'
  phases: ChanPhase[]
  repeat?: number | 'forever'
}

export interface ChanConfig {
  gomaxprocs: number
  /**
   * Передача из рук в руки: если получатель уже ждёт, значение уходит ему
   * напрямую, минуя буфер. Выключите — и увидите цену лишнего копирования.
   */
  directHandoff: boolean
  /** Разбуженная горутина попадает в runnext и исполняется следующей, а не в хвост очереди. */
  runnext: boolean
  /** Обнаруживать полную взаимную блокировку (checkdead). */
  deadlockDetect: boolean
  /**
   * Сколько тиков горутина считается «залипшей», чтобы стенд назвал её утечкой.
   * К рантайму отношения не имеет — это мерка самого стенда.
   */
  leakAfter: number
}

export const DEFAULT_CHAN_CONFIG: ChanConfig = {
  gomaxprocs: 4,
  directHandoff: true,
  runnext: true,
  deadlockDetect: true,
  leakAfter: 40,
}

/** Горутина в очереди ожидания канала — в рантайме это структура `sudog`. */
export interface Waiter {
  g: number
  /** Горутина пришла из select и стоит сразу в нескольких очередях. */
  fromSelect: boolean
  /** С какого тика ждёт. */
  since: number
}

export interface Chan {
  id: number
  name: string
  cap: number
  /**
   * Кольцевой буфер. В ячейке — id горутины, чьё значение там лежит:
   * так на схеме видно, кто кого ждёт.
   */
  buf: (number | null)[]
  /** Сколько значений в буфере, куда пишет отправитель, откуда читает получатель. */
  qcount: number
  sendx: number
  recvx: number
  closed: boolean
  isNil: boolean
  sendq: Waiter[]
  recvq: Waiter[]
  /** Тик, на котором брали блокировку канала, — для подсветки. */
  lockedAt: number
  stats: {
    sent: number
    received: number
    direct: number
    buffered: number
    blockedSends: number
    blockedRecvs: number
    maxQcount: number
  }
}

export type ChanGState = 'runnable' | 'running' | 'waiting' | 'done'

/** Чего ждёт горутина. Для select в `chans` лежат сразу все каналы его case'ов. */
export interface WaitInfo {
  kind: 'send' | 'recv' | 'select'
  chans: number[]
  since: number
}

export interface Goroutine {
  id: number
  name: string
  workload: number
  state: ChanGState
  wait: WaitInfo | null
  phaseIdx: number
  /** Сколько тиков осталось в фазе cpu. */
  phaseLeft: number
  /** Сколько операций осталось в текущей фазе send/recv/select. */
  opsLeft: number
  repeatsLeft: number
  createdTick: number
  runTicks: number
  waitTicks: number
  /** Сколько раз паркована и сколько раз разбужена. */
  blocks: number
  wakeups: number
  sent: number
  received: number
  finishedTick?: number
}

export type ChanEventType =
  | 'g.start'
  | 'g.done'
  | 'g.ready'
  | 'send.direct'
  | 'send.buffer'
  | 'send.block'
  | 'buf.full'
  | 'recv.direct'
  | 'recv.buffer'
  | 'recv.block'
  | 'recv.wake'
  | 'recv.closed'
  | 'chan.close'
  | 'send.closed'
  | 'chan.nil'
  | 'select.ready'
  | 'select.default'
  | 'select.block'
  | 'deadlock'
  | 'leak'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<ChanEventType, Importance> = {
  'g.start': 'low',
  'g.done': 'low',
  'g.ready': 'normal',
  'send.direct': 'key',
  'send.buffer': 'low',
  'send.block': 'key',
  'buf.full': 'normal',
  'recv.direct': 'key',
  'recv.buffer': 'low',
  'recv.block': 'key',
  'recv.wake': 'key',
  'recv.closed': 'normal',
  'chan.close': 'key',
  'send.closed': 'key',
  'chan.nil': 'key',
  'select.ready': 'normal',
  'select.default': 'normal',
  'select.block': 'key',
  deadlock: 'key',
  leak: 'key',
}

export interface ChanEvent {
  tick: number
  type: ChanEventType
  /** Участники — для подсветки на схеме. */
  actors: { g?: number[]; chan?: number[]; p?: number[] }
  payload: Record<string, unknown>
}

export interface ChanScenario {
  id: string
  title: string
  /** Одна мысль, которую сценарий доказывает. Без неё сценарий не заводится. */
  claim: string
  config: Partial<ChanConfig>
  chans: ChanSpec[]
  workloads: ChanWorkload[]
  /** Типы событий, на которых стенд автоматически ставит паузу. */
  watchFor: ChanEventType[]
  stopAfter?: number
  seed?: number
}

export interface ChanStats {
  /** Все успешные передачи значений. */
  transfers: number
  /** Из них — из рук в руки, минуя буфер. */
  direct: number
  /** И через буфер. */
  buffered: number
  parks: number
  wakeups: number
  /** Суммарные тики ожидания — из них считается среднее. */
  waitTicks: number
  selects: number
  selectDefaults: number
  selectBlocks: number
  closes: number
  /** Тики, в которые хотя бы один процессор простаивал без работы. */
  idleSlots: number
  busySlots: number
}

/** Что делал процессор в этом тике — проекция для схемы. */
export interface Slot {
  p: number
  g: number | null
  op: string | null
  chan: number | null
}

export interface ChanWorld {
  tick: number
  config: ChanConfig
  chans: Chan[]
  gs: Goroutine[]
  /** Готовые к исполнению, в порядке очереди. Разбуженные попадают в начало — это runnext. */
  runq: number[]
  slots: Slot[]
  stats: ChanStats
  nextGid: number
  spawnPtr: number
  finished: boolean
  finishReason?: 'all-done' | 'deadlock' | 'panic' | 'stop-after'
  /** Текст паники, если прогон закончился ею. */
  panic?: string
}

export interface Snapshot {
  world: ChanWorld
  /** События, порождённые ИМЕННО этим тиком. */
  events: ChanEvent[]
}
