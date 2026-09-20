/**
 * Типы модели сборщика мусора Go.
 *
 * Как и стенд планировщика, это МОДЕЛЬ, а не эмулятор. Куча здесь — сетка
 * одинаковых блоков, а не спаны с классами размеров; время — тики, а не
 * наносекунды. Всё, что модель упрощает сознательно, перечислено в разборах
 * событий (поле `model`) и в разделе лекции «Что модель упрощает».
 */

/** Трёхцветная разметка: белый — не найден, серый — найден, но не просмотрен, чёрный — просмотрен. */
export type Color = 'white' | 'grey' | 'black'

/**
 * Фазы цикла. В рантайме их так и зовут: sweep termination, mark, mark termination, sweep.
 * Две из них — паузы, когда весь мир остановлен.
 */
export type GcPhase = 'off' | 'stw-start' | 'mark' | 'stw-end' | 'sweep'

export type Duration = number | [number, number]

/** Что делает горутина-мутатор. Мутатор — это любой код программы: он портит разметку, пока GC её строит. */
export type MutPhase =
  /** Считает, ничего не выделяя. */
  | { kind: 'cpu'; ticks: Duration }
  /** Выделяет структуру из `blocks` блоков: голова в слот стека, остальные — цепочкой за ней. */
  | { kind: 'alloc'; blocks: number; retain?: boolean }
  /** Обнуляет слот стека: то, что в нём лежало, становится мусором. */
  | { kind: 'drop' }
  /** Перекладывает указатель из непросмотренного объекта в уже чёрный — сцена, ради которой нужен барьер записи. */
  | { kind: 'move' }

export interface MutWorkload {
  /** Метка для схемы и разборов: «worker», «cache», «shuffler». */
  name: string
  count: number
  spawnAt: number | 'staggered'
  phases: MutPhase[]
  repeat?: number | 'forever'
  /** Сколько указателей горутина держит на стеке. Это её корни. */
  stackSlots?: number
}

export interface GcConfig {
  gomaxprocs: number
  /** GOGC: цель = живое × (1 + GOGC/100). false — GC выключен (GOGC=off). */
  gogc: number | false
  /** GOMEMLIMIT в блоках. false — не задан. */
  memLimit: number | false
  /** Физический размер кучи в блоках: больше в модель просто не влезает. */
  heapCapacity: number
  /** Доля процессоров под фоновую разметку. В Go — 0.25. */
  gcCpuShare: number
  /** Сколько блоков просматривает один маркер за тик. */
  scanRate: number
  /** Сколько блоков освобождает фоновый подметальщик за тик. */
  sweepRate: number
  /** Длительность каждой из двух пауз, в тиках. */
  stwTicks: number
  /** Барьер записи во время разметки. Выключение показывает, зачем он нужен. */
  writeBarrier: boolean
  /** Помощь в разметке: аллокация оплачивает работу сборщика. */
  markAssist: boolean
  /** Простаивающий P подрабатывает маркером. */
  idleWorkers: boolean
  /** Объекты, выделенные во время разметки, сразу чёрные. */
  allocBlack: boolean
  /** Цель по куче до первого цикла. В Go — 4 МБ. */
  initialGoal: number
}

export const DEFAULT_GC_CONFIG: GcConfig = {
  gomaxprocs: 4,
  gogc: 100,
  memLimit: false,
  heapCapacity: 220,
  gcCpuShare: 0.25,
  scanRate: 3,
  sweepRate: 6,
  stwTicks: 2,
  writeBarrier: true,
  markAssist: true,
  idleWorkers: true,
  allocBlack: true,
  initialGoal: 32,
}

/** Блок кучи. Он же ячейка на схеме, он же объект: в модели объект всегда ровно один блок. */
export interface Cell {
  id: number
  used: boolean
  color: Color
  /** Исходящие указатели: id блоков или null. */
  slots: (number | null)[]
  /** Индекс нагрузки, чья горутина его выделила, — для цвета на схеме. */
  owner: number
  bornTick: number
  bornCycle: number
  /** Выделен чёрным во время разметки. */
  bornBlack: boolean
  /** Подметальщик освободил его, хотя объект был достижим. Такого не должно случаться никогда. */
  lost: boolean
}

export type MutState = 'runnable' | 'running' | 'assist' | 'stopped' | 'done'

export interface Mutator {
  id: number
  name: string
  workload: number
  /** Корни этой горутины: указатели на её стеке. */
  stack: (number | null)[]
  /** Стек просмотрен в текущем цикле — после этого он «чёрный» и барьер Дейкстры для него не нужен. */
  stackScanned: boolean
  state: MutState
  phaseIdx: number
  phaseLeft: number
  repeatsLeft: number
  /** Долг помощи в блоках сканирования: набрали аллокацией — отдай разметкой. */
  assistDebt: number
  /** Ходил ли в помощь в этом цикле — чтобы не сыпать событиями. */
  assistedThisCycle: boolean
  createdTick: number
  /** Тиков собственной работы программы. */
  cpuTicks: number
  /** Тиков, отданных разметке вместо своей работы. */
  assistTicks: number
  /** Тиков, простоянных в паузах. */
  stwTicks: number
  /** Тиков, когда мутатор был готов работать, но процессор достался кому-то другому. */
  waitTicks: number
  allocated: number
}

/** Чем был занят процессор в этом тике — проекция для схемы. */
export type SlotKind = 'dedicated' | 'fractional' | 'idle-worker' | 'mutator' | 'assist' | 'sweeper' | 'stw' | 'free'

export interface SlotView {
  p: number
  kind: SlotKind
  mut?: number
  /** Сколько блоков просмотрено или освобождено этим слотом. */
  work?: number
}

export type GcEventType =
  | 'gc.trigger'
  | 'gc.forced'
  | 'stw.enter'
  | 'stw.exit'
  | 'mark.roots'
  | 'mark.stack'
  | 'mark.scan'
  | 'mark.drained'
  | 'assist.begin'
  | 'assist.work'
  | 'barrier.shade'
  | 'barrier.missed'
  | 'alloc'
  | 'alloc.black'
  | 'heap.lost'
  | 'sweep.begin'
  | 'sweep.freed'
  | 'sweep.done'
  | 'cycle.done'
  | 'pacer.set'
  | 'limit.hit'
  | 'oom'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<GcEventType, Importance> = {
  'gc.trigger': 'key',
  'gc.forced': 'key',
  'stw.enter': 'key',
  'stw.exit': 'normal',
  'mark.roots': 'normal',
  'mark.stack': 'normal',
  'mark.scan': 'low',
  'mark.drained': 'key',
  'assist.begin': 'key',
  'assist.work': 'low',
  'barrier.shade': 'normal',
  'barrier.missed': 'key',
  alloc: 'low',
  'alloc.black': 'low',
  'heap.lost': 'key',
  'sweep.begin': 'normal',
  'sweep.freed': 'low',
  'sweep.done': 'normal',
  'cycle.done': 'key',
  'pacer.set': 'normal',
  'limit.hit': 'key',
  oom: 'key',
}

export interface GcEvent {
  tick: number
  type: GcEventType
  /** Участники — для подсветки на схеме. */
  actors: { cells?: number[]; mut?: number[] }
  payload: Record<string, unknown>
}

export interface GcScenario {
  id: string
  title: string
  /** Одна мысль, которую сценарий доказывает. Без неё сценарий не заводится. */
  claim: string
  config: Partial<GcConfig>
  workloads: MutWorkload[]
  /** Типы событий, на которых стенд ставит автопаузу. */
  watchFor: GcEventType[]
  stopAfter?: number
  seed?: number
}

export interface GcStats {
  /** Завершённых циклов. */
  cycles: number
  /** Суммарная длительность пауз и самая долгая из них. */
  stwTicks: number
  maxPause: number
  /** Слото-тики: чем занимались процессоры за весь прогон. */
  gcSlotTicks: number
  mutatorSlotTicks: number
  assistTicks: number
  allocated: number
  freed: number
  /** Освобождённых достижимых объектов — при исправном барьере всегда 0. */
  lost: number
  /** Самая большая куча за прогон и самое большое превышение цели. */
  peakHeap: number
  maxOvershoot: number
}

export interface GcWorld {
  tick: number
  config: GcConfig
  phase: GcPhase
  /** Тиков до конца текущей паузы. */
  stwLeft: number
  /** Номер текущего (или последнего) цикла, с единицы. */
  cycle: number
  /** Освобождено, отдано помощи и потрачено на GC в текущем цикле — для итога цикла. */
  cycleFreed: number
  cycleAssistTicks: number
  cycleGcTicks: number
  /** Тик, на котором начался текущий цикл, — для длительности в 'cycle.done'. */
  cycleStart: number
  cells: Cell[]
  muts: Mutator[]
  /** Глобальные переменные программы: вторая половина корней. */
  globals: (number | null)[]
  /** Очередь серых объектов — фронт разметки. */
  greyq: number[]
  /** Блоки, которые подметальщик ещё не освободил. */
  sweepq: number[]
  /** Живое по итогам прошлого цикла: от него пейсер считает цель. */
  heapMarked: number
  goal: number
  trigger: number
  /** Сколько работы разметки предстоит в этом цикле по оценке пейсера и сколько уже сделано. */
  scanWorkTarget: number
  scanWorkDone: number
  /** Сколько сканирования обязан оплатить мутатор за один выделенный блок. */
  assistRatio: number
  /** Скорость аллокации, блоков за тик (скользящее среднее) — вход пейсера. */
  allocRate: number
  /** Барьер записи включается на время разметки. */
  barrierOn: boolean
  slots: SlotView[]
  stats: GcStats
  nextMutId: number
  spawnPtr: number
  finished: boolean
  finishReason?: 'all-done' | 'oom' | 'stop-after'
}

export interface Snapshot {
  world: GcWorld
  /** События, порождённые именно этим тиком. */
  events: GcEvent[]
}
