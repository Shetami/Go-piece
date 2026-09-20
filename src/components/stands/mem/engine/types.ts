/**
 * Типы модели памяти Go.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Классы размеров и путь аллокации взяты
 * из рантайма почти без изменений, а вот время измеряется в тиках, спан мелкого
 * класса всегда одна страница, и мусор убирается не сборщиком, а по заранее
 * заданному времени жизни. Полный список упрощений — в конце лекции.
 */

/** Страница рантайма: 8 КБ. Куча раздаётся страницами, а не байтами. */
export const PAGE = 8192

/**
 * Классы размеров из runtime/sizeclasses.go.
 *
 * Аллокатор не умеет выдавать «33 байта» — он выдаёт ближайший класс сверху.
 * Разница между запрошенным размером и классом и есть внутренняя фрагментация.
 */
export const SIZE_CLASSES = [
  8, 16, 24, 32, 48, 64, 80, 96, 112, 128, 144, 160, 176, 192, 208, 224, 240, 256,
  288, 320, 352, 384, 416, 448, 480, 512, 576, 640, 704, 768, 896, 1024, 1152, 1280,
  1408, 1536, 1792, 2048, 2304, 2688, 3072, 3200, 3456, 4096, 4864, 5376, 6144, 6528,
  6784, 6912, 8192, 9472, 9728, 10240, 10880, 12288, 13568, 14336, 16384, 18432,
  19072, 20480, 21760, 24576, 27264, 28672, 32768,
] as const

/** Объекты больше последнего класса идут мимо кэшей — прямо в кучу, страницами. */
export const MAX_SMALL = SIZE_CLASSES[SIZE_CLASSES.length - 1]!
/** Объекты меньше этого размера и без указателей внутри складываются в общий блок. */
export const TINY_SIZE = 16

export type Duration = number | [number, number]

/** Что делает горутина. */
export type MemPhase =
  /** Считает, ничего не выделяя. */
  | { kind: 'cpu'; ticks: Duration }
  /**
   * Просит `count` объектов размера `size`.
   * `escapes: false` — компилятор доказал, что указатель не уезжает наружу,
   * и объект остаётся на стеке. `lifetime` — сколько тиков объект живёт в куче.
   */
  | { kind: 'alloc'; size: number | [number, number]; count?: number; escapes?: boolean; lifetime?: Duration; pointers?: boolean }
  /** Рекурсия: `depth` кадров по `frame` байт — так растёт стек. */
  | { kind: 'recurse'; depth: number; frame: number }

export interface MemWorkload {
  name: string
  count: number
  spawnAt: number | 'staggered'
  phases: MemPhase[]
  repeat?: number | 'forever'
}

export interface MemConfig {
  gomaxprocs: number
  /** Сколько страниц в арене — физический потолок модели. */
  heapPages: number
  /** Сколько страниц занимает спан мелкого класса. В рантайме зависит от класса. */
  spanPages: number
  /** Стартовый размер стека горутины. В Go — 2 КБ. */
  stackStart: number
  /** Сколько байт стека копируется за тик при росте. */
  stackCopyRate: number
  /** Маленькие объекты без указателей складываются в один блок на 16 байт. */
  tinyAllocator: boolean
  /** Анализ побега. Выключить — и всё уедет в кучу, как в языке без него. */
  escapeAnalysis: boolean
  /** Учитывать ли борьбу за общую блокировку mcentral. */
  centralLock: boolean
  /** Через сколько тиков простоя свободные страницы возвращаются ОС. */
  scavengeAfter: number
  /** Раз в сколько тиков рантайм проверяет, не пора ли ужать стеки (в Go — на сборке мусора). */
  shrinkEvery: number
}

export const DEFAULT_MEM_CONFIG: MemConfig = {
  gomaxprocs: 4,
  heapPages: 96,
  spanPages: 1,
  stackStart: 2048,
  stackCopyRate: 4096,
  tinyAllocator: true,
  escapeAnalysis: true,
  centralLock: true,
  scavengeAfter: 40,
  shrinkEvery: 60,
}

/** Чем занята страница кучи. */
export type PageKind = 'free' | 'span' | 'large' | 'stack' | 'returned'

export interface Page {
  id: number
  kind: PageKind
  /** id спана, большого объекта или стека, которому принадлежит страница. */
  owner: number | null
  /** С какого тика страница свободна — для возврата памяти операционной системе. */
  freeSince: number | null
}

/** Спан: несколько страниц, нарезанных на объекты одного класса. */
export interface Span {
  id: number
  sizeClass: number
  /** Размер объекта этого класса в байтах. */
  objSize: number
  pages: number[]
  /** Сколько объектов всего помещается и сколько занято. */
  slots: number
  used: number
  /** Где лежит спан: в кэше процессора, в центральном списке или уже пуст. */
  owner: { kind: 'mcache'; p: number } | { kind: 'mcentral' } | { kind: 'free' }
  createdTick: number
}

export interface HeapObject {
  id: number
  size: number
  /** Размер класса, в который объект округлили. */
  slotSize: number
  sizeClass: number
  span: number | null
  /** Большой объект живёт не в спане, а на своих страницах. */
  pages: number[]
  owner: number
  bornTick: number
  diesAt: number
  tiny: boolean
}

/**
 * Кусок под маленькие стеки: одна страница, нарезанная на стеки одного размера.
 * Четыре горутины со стеком по 2 КБ делят одну страницу — в рантайме этим
 * занимается stackpool.
 */
export interface StackChunk {
  id: number
  size: number
  page: number
  slots: number
  free: number[]
}

export type GState = 'runnable' | 'running' | 'growing' | 'done'

export interface Goroutine {
  id: number
  name: string
  workload: number
  state: GState
  /** Размер стека и сколько из него занято кадрами. */
  stackSize: number
  stackUsed: number
  /** Большой стек лежит на своих страницах, маленький — в слоте общего куска. */
  stackPages: number[]
  stackSlot: { chunk: number; index: number } | null
  /** Сколько тиков осталось копировать стек. */
  copyLeft: number
  /** Глубина, на которую горутина уже погрузилась в текущей рекурсии. */
  depth: number
  phaseIdx: number
  phaseLeft: number
  repeatsLeft: number
  createdTick: number
  cpuTicks: number
  /** Тики, потерянные на росте стека и на ожидании блокировки mcentral. */
  stallTicks: number
  allocBytes: number
  /** Сколько объектов осталось выделить в текущей фазе. */
  allocLeft: number
  stackGrows: number
}

/** Кэш процессора: по одному текущему спану на класс размеров плюс блок для мелочи. */
export interface Mcache {
  p: number
  /** Класс → id спана, из которого сейчас берутся объекты. */
  spans: Record<number, number>
  /** Блок для мелких объектов: сколько байт из 16 уже занято и сколько объектов в нём. */
  tinyOffset: number
  tinyObjects: number
  tinySpanObj: number | null
}

export interface Mcentral {
  sizeClass: number
  /** Спаны со свободными слотами и спаны, отданные кэшам целиком. */
  partial: number[]
  full: number[]
  /** Тик, на котором блокировку взяли в последний раз, — для учёта борьбы за неё. */
  lockedAt: number
  lockedBy: number | null
}

export type MemEventType =
  | 'g.start'
  | 'alloc.stack'
  | 'alloc.tiny'
  | 'alloc.fast'
  | 'alloc.large'
  | 'class.waste'
  | 'cache.refill'
  | 'central.contended'
  | 'central.empty'
  | 'heap.grow'
  | 'os.map'
  | 'frag.external'
  | 'obj.free'
  | 'span.free'
  | 'scavenge'
  | 'stack.grow'
  | 'stack.shrink'
  | 'oom'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<MemEventType, Importance> = {
  'g.start': 'low',
  'alloc.stack': 'low',
  'alloc.tiny': 'normal',
  'alloc.fast': 'low',
  'alloc.large': 'key',
  'class.waste': 'key',
  'cache.refill': 'normal',
  'central.contended': 'key',
  'central.empty': 'normal',
  'heap.grow': 'normal',
  'os.map': 'key',
  'frag.external': 'key',
  'obj.free': 'low',
  'span.free': 'normal',
  scavenge: 'key',
  'stack.grow': 'key',
  'stack.shrink': 'normal',
  oom: 'key',
}

export interface MemEvent {
  tick: number
  type: MemEventType
  actors: { g?: number[]; p?: number[]; pages?: number[]; span?: number[] }
  payload: Record<string, unknown>
}

export interface MemScenario {
  id: string
  title: string
  /** Одна мысль, которую сценарий доказывает. Без неё сценарий не заводится. */
  claim: string
  config: Partial<MemConfig>
  workloads: MemWorkload[]
  watchFor: MemEventType[]
  stopAfter?: number
  seed?: number
}

export interface MemStats {
  /** Аллокации по путям: стек, общий блок для мелочи, кэш процессора, центральный список, куча. */
  onStack: number
  tiny: number
  /** Сколько блоков по 16 байт пришлось открыть под мелочь. */
  tinyBlocks: number
  fast: number
  refills: number
  large: number
  /** Запрошено и фактически выдано байт — по всем аллокациям. */
  askedBytes: number
  slotBytes: number
  /**
   * То же, но только по объектам, которые округлялись до класса размеров.
   * Мелочь, сложенную в общий блок, считать здесь нельзя: она не теряет память,
   * а экономит, и смешивать эти два счёта — значит прятать потери округления.
   */
  roundAsked: number
  roundSlot: number
  /** Сколько раз пришлось ждать общую блокировку и сколько тиков потеряно. */
  contended: number
  contendedTicks: number
  /** Страницы: сколько взято у ОС и сколько ей возвращено. */
  mapped: number
  returned: number
  peakPages: number
  stackGrows: number
  stackCopyBytes: number
  stackShrinks: number
  spansCreated: number
  spansFreed: number
}

export interface MemWorld {
  tick: number
  config: MemConfig
  pages: Page[]
  spans: Span[]
  stackChunks: StackChunk[]
  objects: HeapObject[]
  gs: Goroutine[]
  mcaches: Mcache[]
  mcentrals: Mcentral[]
  /** Что происходило на каждом процессоре в этом тике — проекция для схемы. */
  slots: { p: number; g: number | null; path: string | null }[]
  stats: MemStats
  nextGid: number
  nextSpanId: number
  nextChunkId: number
  nextObjId: number
  spawnPtr: number
  finished: boolean
  finishReason?: 'all-done' | 'oom' | 'stop-after'
}

export interface Snapshot {
  world: MemWorld
  events: MemEvent[]
}

/** Класс размеров, в который попадёт объект: первый, который не меньше запрошенного. */
export function classFor(size: number): { index: number; size: number } | null {
  for (let i = 0; i < SIZE_CLASSES.length; i++) {
    const c = SIZE_CLASSES[i]!
    if (size <= c) return { index: i, size: c }
  }
  return null
}
