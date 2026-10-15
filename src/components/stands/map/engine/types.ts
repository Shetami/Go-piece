/**
 * Типы модели карты Go.
 *
 * Как и в соседних стендах, это МОДЕЛЬ. Раскладка групп и контрольных байтов,
 * последовательность проб, надгробия, прополка, рост и деление таблиц, каталог
 * и итератор взяты из internal/runtime/maps почти без изменений; старая карта
 * на корзинах — из runtime/map.go Go 1.23. Упрощено вот что:
 *
 * - хэш 32-битный, а не 64-битный: H2 — те же младшие 7 бит, H1 — остальные 25;
 * - потолок таблицы по умолчанию 32 слота вместо 1024 — иначе деление и каталог
 *   пришлось бы ждать тысячу вставок;
 * - ключи — строки, значения — числа; время — один тик на одну операцию с картой.
 */

/** Реализация: Swiss table (Go 1.24+) или карта на корзинах (до Go 1.24). */
export type Impl = 'swiss' | 'legacy'

export interface MapConfig {
  impl: Impl
  /**
   * Потолок ёмкости одной таблицы Swiss table. В рантайме — 1024 слота
   * (maxTableCapacity); в стенде меньше, чтобы деление было видно сразу.
   */
  maxTableSlots: number
  /** Подсказка размера: make(map[K]V, hint). 0 — просто make(map[K]V). */
  hint: number
  /** `var m map[string]int` — карту объявили, но не создали. */
  nilMap: boolean
}

export const DEFAULT_MAP_CONFIG: MapConfig = {
  impl: 'swiss',
  maxTableSlots: 32,
  hint: 0,
  nilMap: false,
}

/** Слотов в группе Swiss table и в корзине старой карты — 8 в обеих реализациях. */
export const GROUP_SLOTS = 8

/* ───────────────────────── Swiss table ───────────────────────── */

/** Контрольный байт пустого слота. */
export const CTRL_EMPTY = 0x80
/** Контрольный байт надгробия. */
export const CTRL_DELETED = 0xfe

export interface Slot {
  /** 0x80 — empty, 0xfe — deleted, 0..127 — занят, в байте лежит H2 ключа. */
  ctrl: number
  key: string | null
  val: number | null
  /** Хэш ключа — хранится ради отрисовки; рантайм его не хранит, а пересчитывает. */
  hash: number
}

export interface Group {
  slots: Slot[]
}

export interface Table {
  id: number
  /** Слотов всего: степень двойки, кратная 8. */
  capacity: number
  used: number
  /** Сколько ещё можно вставить до роста. Надгробия его не возвращают. */
  growthLeft: number
  /** Сколько старших бит хэша на самом деле определяют ключи этой таблицы. */
  localDepth: number
  /** Первая строка каталога, указывающая на таблицу; -1 — таблицу заменили. */
  index: number
  groups: Group[]
  /** Таблицу заменили ростом или делением; живёт, пока её держит итератор. */
  replaced: boolean
}

export interface SwissMap {
  used: number
  seed: number
  /**
   * Маленькая карта: одна группа без таблицы и каталога (dirLen == 0).
   * null — группа ещё не выделена: make без подсказки или с подсказкой ≤ 8.
   */
  small: Group | null
  /** Каталог: id таблиц, длина 2^globalDepth. Пустой — карта маленькая. */
  dir: number[]
  globalDepth: number
  tables: Record<number, Table>
  writing: number
  tombstonePossible: boolean
  clearSeq: number
}

/* ───────────────────────── Старая карта (до Go 1.24) ───────────────────────── */

/** Особые значения tophash: всё, что меньше 5, — не хэш, а состояние слота. */
export const TOP_EMPTY_REST = 0
export const TOP_EMPTY_ONE = 1
export const TOP_EVAC_X = 2
export const TOP_EVAC_Y = 3
export const TOP_EVAC_EMPTY = 4
export const MIN_TOPHASH = 5

export interface LSlot {
  top: number
  key: string | null
  val: number | null
  hash: number
}

/** Корзина: первая страница — сама корзина, остальные — цепочка корзин переполнения. */
export interface Bucket {
  pages: LSlot[][]
  /** Старая корзина уже перенесена в новый массив. */
  evacuated: boolean
}

export interface LegacyMap {
  count: number
  /** Корзин 2^B. */
  B: number
  hash0: number
  /** id массива корзин в `arrays`; null — корзины ещё не выделены (make без подсказки). */
  buckets: number | null
  /** Массив, из которого идёт эвакуация; null — рост не идёт. */
  oldbuckets: number | null
  arrays: Record<number, Bucket[]>
  /** Все корзины старого массива с номером меньше этого уже эвакуированы. */
  nevacuate: number
  /** Сколько корзин переполнения выделено — грубо, как в рантайме. */
  noverflow: number
  sameSizeGrow: boolean
  writing: number
}

/* ───────────────────────── Итератор ───────────────────────── */

export interface Yield {
  key: string
  val: number
  /** Значение перечитано из новой таблицы: итератор шёл по старой. */
  reread: boolean
}

export interface SwissIter {
  impl: 'swiss'
  entryOffset: number
  dirOffset: number
  dirIdx: number
  entryIdx: number
  /** Таблица, по которой идёт обход; null — между таблицами. */
  tab: number | null
  globalDepth: number
  /** Обход начался на маленькой карте. */
  small: boolean
  /** Маленькая карта успела стать таблицей: итератор держит старую группу. */
  smallGroup: Group | null
}

export interface LegacyIter {
  impl: 'legacy'
  /** Массив корзин, с которого начался обход, и B на тот момент. */
  arr: number
  B: number
  startBucket: number
  offset: number
  /** Сколько корзин пройдено. */
  bucketIdx: number
  /** Корзина, которую обходим сейчас: её массив и номер, и фильтр для старой корзины. */
  src: { arr: number; bucket: number; check: number | null } | null
  page: number
  i: number
}

export type Iter = (SwissIter | LegacyIter) & {
  /** Сколько слотов проверено — обход стоит O(ёмкости), а не O(len). */
  visited: number
  yielded: Yield[]
  done: boolean
  startTick: number
}

/* ───────────────────────── Команды ───────────────────────── */

export type MapOp =
  /** Создание карты — неявная первая команда любого сценария. */
  | { kind: 'make' }
  | { kind: 'put'; key: string; val?: number }
  | { kind: 'get'; key: string }
  | { kind: 'del'; key: string }
  /**
   * get/delete по ключам последней fill или collide: n == null — по всем,
   * n > 0 — по первым n, n < 0 — по последним |n|.
   */
  | { kind: 'last'; op: 'get' | 'del'; n: number | null }
  | { kind: 'clear' }
  | { kind: 'range' }
  | { kind: 'next'; count: number | 'all' }
  /** count новых ключей prefix1, prefix2, … */
  | { kind: 'fill'; count: number; prefix: string }
  /** count ключей, попадающих в одну группу (одну корзину). target — её номер. */
  | { kind: 'collide'; count: number; target?: number }
  /** Вторая горутина пишет в карту одновременно с первой. */
  | { kind: 'race'; mode: 'write' | 'read' | 'iter' }

/** Команда с номером строки сценария — чтобы лента знала, откуда она. */
export type LocatedOp = MapOp & { line: number }

/* ───────────────────────── События ───────────────────────── */

export type MapEventType =
  | 'map.make'
  | 'put.new'
  | 'put.update'
  | 'get.hit'
  | 'get.miss'
  | 'del.hit'
  | 'del.miss'
  | 'h2.false'
  | 'clear'
  | 'seed.reset'
  | 'nil.read'
  | 'nil.write'
  | 'race'
  | 'iter.start'
  | 'iter.next'
  | 'iter.reread'
  | 'iter.skip'
  | 'iter.done'
  // Swiss table
  | 'small.alloc'
  | 'small.promote'
  | 'tomb.set'
  | 'tomb.reuse'
  | 'tomb.prune'
  | 'table.grow'
  | 'table.split'
  | 'dir.grow'
  // Старая карта
  | 'bucket.overflow'
  | 'grow.start'
  | 'evac'
  | 'grow.done'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<MapEventType, Importance> = {
  'map.make': 'normal',
  'put.new': 'low',
  'put.update': 'normal',
  'get.hit': 'normal',
  'get.miss': 'normal',
  'del.hit': 'normal',
  'del.miss': 'low',
  'h2.false': 'key',
  clear: 'key',
  'seed.reset': 'key',
  'nil.read': 'normal',
  'nil.write': 'key',
  race: 'key',
  'iter.start': 'key',
  'iter.next': 'low',
  'iter.reread': 'key',
  'iter.skip': 'key',
  'iter.done': 'normal',
  'small.alloc': 'normal',
  'small.promote': 'key',
  'tomb.set': 'key',
  'tomb.reuse': 'key',
  'tomb.prune': 'key',
  'table.grow': 'key',
  'table.split': 'key',
  'dir.grow': 'key',
  'bucket.overflow': 'key',
  'grow.start': 'key',
  evac: 'normal',
  'grow.done': 'key',
}

export interface MapEvent {
  tick: number
  type: MapEventType
  /** Строка сценария, из которой пришла команда. */
  line: number
  payload: Record<string, unknown>
}

/* ───────────────────────── След операции ───────────────────────── */

/** Одна проверенная группа (Swiss) или страница корзины (старая карта). */
export interface Probe {
  /** Swiss: таблица (null — маленькая карта) и группа. */
  table?: number | null
  group?: number
  /** Старая карта: массив, корзина, страница цепочки. */
  arr?: number
  bucket?: number
  page?: number
  /** Слоты, где совпали 7 бит H2 (или tophash). */
  match: number[]
  /** Слоты, где сравнили ключ целиком, и результат. */
  compared: { slot: number; eq: boolean }[]
  /** В группе есть пустой слот — на ней поиск заканчивается. */
  stop: boolean
}

export interface OpTrace {
  /** Команда так, как она выглядит в Go. */
  code: string
  key?: string
  hash?: number
  probes: Probe[]
  /** Куда в итоге пришла операция. */
  hit?: { table?: number | null; group?: number; arr?: number; bucket?: number; page?: number; slot: number }
  /** Короткий итог для схемы. */
  outcome: string
  /** Шаг обхода: hit — слот, на котором стоит итератор. */
  iter?: boolean
}

/* ───────────────────────── Мир ───────────────────────── */

export interface MapStats {
  ops: number
  /** Сколько групп (страниц корзин) проверили все поиски. */
  probes: number
  /** Сколько раз сравнили ключ целиком. */
  compares: number
  /** Из них — ложные совпадения по 7 (8) битам хэша. */
  falseMatches: number
  /** Самая длинная последовательность проб. */
  maxProbe: number
  /** Сколько записей переложили при росте, делении и эвакуации. */
  moved: number
  grows: number
  splits: number
  dirGrows: number
  prunes: number
  overflows: number
  evacuated: number
  iterVisited: number
  iterYields: number
}

export interface MapWorld {
  tick: number
  config: MapConfig
  /** Ровно одно из двух — по config.impl; у nil-карты оба null. */
  swiss: SwissMap | null
  legacy: LegacyMap | null
  iter: (Iter & { id: number }) | null
  /** id для таблиц и массивов корзин. */
  nextId: number
  /** Счётчики для fill и collide, чтобы ключи не повторялись. */
  fillNext: Record<string, number>
  collideNext: number
  /** Ключи последней fill/collide — для `del @last`. */
  lastKeys: string[]
  stats: MapStats
  finished: boolean
  finishReason?: 'all-done' | 'panic' | 'fatal' | 'stop-after'
  /** Текст паники или fatal error. */
  panic?: string
}

export interface MapScenario {
  id: string
  title: string
  /** Одна мысль, которую сценарий доказывает. Без неё сценарий не заводится. */
  claim: string
  config: Partial<MapConfig>
  /** Команды по одной на строку — см. engine/script.ts. */
  script: string
  /** Типы событий, на которых стенд автоматически ставит паузу. */
  watchFor: MapEventType[]
  seed?: number
}

export interface Snapshot {
  world: MapWorld
  /** События, порождённые ИМЕННО этим тиком. */
  events: MapEvent[]
  /** Что делала операция этого тика — для подсветки на схеме. */
  trace: OpTrace | null
}
