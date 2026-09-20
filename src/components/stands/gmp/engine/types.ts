/**
 * Типы модели планировщика Go.
 *
 * Всё, что здесь описано, — это МОДЕЛЬ, а не эмулятор рантайма.
 * Список сознательных упрощений см. в README.
 */

export type GState = 'runnable' | 'running' | 'waiting' | 'syscall' | 'dead'

/** Причина, по которой горутина находится в waiting — нужна для подсветки и для detect deadlock. */
export type WaitReason = 'chan-send' | 'chan-recv' | 'mutex' | 'sleep' | 'net'

export type MState = 'idle' | 'spinning' | 'running' | 'syscall' | 'blocked'

export type PState = 'idle' | 'running' | 'syscall'

/** Длительность в тиках: точное число, диапазон [min, max] или «бесконечно». */
export type Duration = number | [number, number] | 'forever'

export type Phase =
  | { kind: 'cpu'; ticks: Duration }
  | { kind: 'syscall'; ticks: Duration }
  | { kind: 'net'; ticks: Duration }
  | { kind: 'sleep'; ticks: Duration }
  | { kind: 'mutex'; lock: string; ticks: Duration }
  | { kind: 'chanSend'; chan: string }
  | { kind: 'chanRecv'; chan: string }

export interface Workload {
  /** Метка для UI и отладки: «worker», «pinger», «io». */
  name: string
  count: number
  /** На каком тике рождаются: число — все сразу, 'staggered' — по одной в тик. */
  spawnAt: number | 'staggered'
  /** На каком P создаются. По умолчанию распределяются по кругу. */
  spawnOn?: number
  phases: Phase[]
  /** Сколько раз повторить список фаз. */
  repeat?: number | 'forever'
}

export interface Config {
  gomaxprocs: number
  /** Квант исполнения в тиках до принудительного вытеснения. */
  quantum: number
  /** Сколько тиков syscall'а терпит sysmon, прежде чем отобрать P. */
  retakeThreshold: number
  /** Вместимость локальной очереди P. В настоящем рантайме — 256. */
  runqCapacity: number
  workStealing: boolean
  /** Правило «каждый 61-й schedtick заглядывать в глобальную очередь». */
  globalCheckEvery: number | false
  runnext: boolean
  /** Асинхронное вытеснение сигналом (Go 1.14+). Выключение возвращает мир до 1.14. */
  asyncPreemption: boolean
  /** Максимум M. В рантайме по умолчанию 10000. */
  maxThreads: number
}

export const DEFAULT_CONFIG: Config = {
  gomaxprocs: 4,
  quantum: 20,
  retakeThreshold: 5,
  runqCapacity: 256,
  workStealing: true,
  globalCheckEvery: 61,
  runnext: true,
  asyncPreemption: true,
  maxThreads: 10_000,
}

export interface Scenario {
  id: string
  title: string
  /** Одна мысль, которую этот сценарий доказывает. Без неё сценарий не заводится. */
  claim: string
  config: Partial<Config>
  workloads: Workload[]
  /** Типы событий, на которых стенд автоматически ставит паузу. */
  watchFor: EventType[]
  stopAfter?: number
  seed?: number
}

export type EventType =
  | 'g.created'
  | 'g.scheduled'
  | 'g.preempted'
  | 'g.blocked'
  | 'g.ready'
  | 'g.finished'
  | 'p.stole'
  | 'p.stealFailed'
  | 'p.runqOverflow'
  | 'p.globalCheck'
  | 'p.handoff'
  | 'p.idle'
  | 'm.spawned'
  | 'm.spinning'
  | 'm.parked'
  | 'm.syscallEnter'
  | 'm.syscallExit'
  | 'sysmon.retake'
  | 'sysmon.preempt'
  | 'net.ready'
  | 'deadlock'

/** Насколько событие важно: определяет размер точки на таймлайне. */
export type Importance = 'key' | 'normal' | 'low'

export interface SimEvent {
  tick: number
  type: EventType
  /** Участники — для подсветки на схеме. */
  actors: { g?: number[]; m?: number[]; p?: number[] }
  payload: Record<string, unknown>
}

export const EVENT_IMPORTANCE: Record<EventType, Importance> = {
  'g.created': 'normal',
  'g.scheduled': 'normal',
  'g.preempted': 'normal',
  'g.blocked': 'normal',
  'g.ready': 'normal',
  'g.finished': 'normal',
  'p.stole': 'key',
  'p.stealFailed': 'normal',
  'p.runqOverflow': 'key',
  'p.globalCheck': 'normal',
  'p.handoff': 'key',
  'p.idle': 'low',
  'm.spawned': 'key',
  'm.spinning': 'normal',
  'm.parked': 'low',
  'm.syscallEnter': 'normal',
  'm.syscallExit': 'normal',
  'sysmon.retake': 'key',
  'sysmon.preempt': 'key',
  'net.ready': 'normal',
  deadlock: 'key',
}

export interface G {
  id: number
  /** Индекс workload'а, из которого она родилась. */
  workload: number
  name: string
  state: GState
  waitReason?: WaitReason
  phaseIdx: number
  /** Сколько тиков осталось в текущей фазе. Infinity для 'forever'. */
  phaseLeft: number
  repeatsLeft: number
  /** Тиков подряд на процессоре — сбрасывается при каждой потере P. */
  quantumUsed: number
  createdTick: number
  runTicks: number
  waitTicks: number
  runqTicks: number
  finishedTick?: number
}

export interface M {
  id: number
  state: MState
  p: number | null
  g: number | null
  /** Тиков в текущем системном вызове. */
  syscallTicks: number
  spawnedTick: number
}

export interface P {
  id: number
  state: PState
  m: number | null
  runq: number[]
  runnext: number | null
  schedtick: number
  idleTicks: number
  /** Если P только что отобран у потока — id этого потока, для события p.handoff. */
  handoffFrom: number | null
}

export interface WaitingEntry {
  g: number
  readyAt: number
}

export interface ChanState {
  /** Горутины, ждущие отправки. */
  sendq: number[]
  /** Горутины, ждущие приёма. */
  recvq: number[]
}

export interface World {
  tick: number
  config: Config
  gs: G[]
  ms: M[]
  ps: P[]
  globrunq: number[]
  /** Горутины в netpoller'е: сеть не занимает поток. */
  netpoll: WaitingEntry[]
  /** Таймеры: time.Sleep. */
  timers: WaitingEntry[]
  chans: Record<string, ChanState>
  mutexes: Record<string, { holder: number | null; waitq: number[] }>
  /** Счётчик для выдачи id новым горутинам. */
  nextGid: number
  nextMid: number
  /** Указатель в плане рождения горутин: сколько уже создано. */
  spawnPtr: number
  finished: boolean
  finishReason?: 'all-done' | 'deadlock' | 'stop-after'
}

export interface Snapshot {
  world: World
  /** События, порождённые ИМЕННО этим тиком. */
  events: SimEvent[]
}
