/**
 * Какие упражнения exercism/problem-specifications превращаются в задачи темы
 * «Алгоритмы» и как их тест-кейсы ложатся на Go.
 *
 * В canonical-data.json кейсы описаны без языка: `property` — имя операции,
 * `input` — объект с аргументами, `expected` — ответ или `{ error }`. Здесь для
 * каждой операции сказано, какая функция на Go её реализует, в каком порядке
 * идут аргументы и каких они типов. Остальное генератор выводит сам.
 *
 * Добавить задачу: дописать сюда упражнение и запустить `pnpm tasks:exercism`.
 * Генератор положит заготовку (index.mdx с draft: true, starter.go,
 * solution.go) — их дальше пишут руками, — а check.go перегенерирует при каждом
 * запуске.
 */

/** Тип Go, в который превращается значение из JSON. */
export type GoType =
  | 'int'
  | 'string'
  | 'bool'
  | { slice: GoType }
  | { array: GoType; len: number }
  | { map: [GoType, GoType] }
  /** fields: ключ в JSON → [поле в Go, тип]. Сам тип объявляет starter.go. */
  | { struct: string; fields: Record<string, [string, GoType]> }

export interface Operation {
  /** `property` из canonical-data. */
  property: string
  /** Имя теста в check.go, без префикса Test. */
  test: string
  /** Функция на Go. У операции с `call` не используется. */
  func: string
  /** Аргументы в порядке вызова: ключ в `input` и тип в Go. */
  args: [key: string, type: GoType][]
  result: GoType
  /** Функция возвращает (результат, error), и кейсы с `{ error }` ждут err != nil. */
  error?: boolean
  /** Чем заменить `{ error }`, если ошибку функция не возвращает: -1 у бинарного поиска. */
  errorValue?: unknown
  /**
   * Свой вызов вместо `func(аргументы)`: `$string` подставляется как аргумент.
   * Такой операции нет в заготовке — она проверяет уже объявленные функции.
   */
  call?: string
}

export interface Exercise {
  /** Папка в problem-specifications/exercises. */
  exercise: string
  /** Папка задачи в src/content/tasks/algorithms. */
  slug: string
  operations: Operation[]
  /** Объявления типов, которые нужны сигнатурам, — для заготовки. */
  types?: string
  /** uuid кейсов, которые не берём, и почему. */
  skip?: Record<string, string>
}

/** Коммит problem-specifications, из которого собраны тесты. Обновлять осознанно: кейсы могут поменяться. */
export const SPECS_REF = '9943fd751b684ba1a6a8673c3700a6e79c10a808'

export const SPECS_REPO = 'https://github.com/exercism/problem-specifications.git'

const ints = { slice: 'int' } as const
const strs = { slice: 'string' } as const
const grid = { slice: ints } as const

export const EXERCISES: Exercise[] = [
  {
    exercise: 'isogram',
    slug: 'isogram',
    operations: [{ property: 'isIsogram', test: 'IsIsogram', func: 'IsIsogram', args: [['phrase', 'string']], result: 'bool' }],
  },
  {
    exercise: 'matching-brackets',
    slug: 'matching-brackets',
    operations: [{ property: 'isPaired', test: 'IsPaired', func: 'IsPaired', args: [['value', 'string']], result: 'bool' }],
  },
  {
    exercise: 'roman-numerals',
    slug: 'roman-numerals',
    operations: [{ property: 'roman', test: 'ToRoman', func: 'ToRoman', args: [['number', 'int']], result: 'string' }],
  },
  {
    exercise: 'binary-search',
    slug: 'binary-search',
    operations: [
      { property: 'find', test: 'Search', func: 'Search', args: [['array', ints], ['value', 'int']], result: 'int', errorValue: -1 },
    ],
  },
  {
    exercise: 'run-length-encoding',
    slug: 'run-length-encoding',
    operations: [
      { property: 'encode', test: 'Encode', func: 'Encode', args: [['string', 'string']], result: 'string' },
      { property: 'decode', test: 'Decode', func: 'Decode', args: [['string', 'string']], result: 'string' },
      {
        property: 'consistency',
        test: 'EncodeDecode',
        func: 'Decode',
        args: [['string', 'string']],
        result: 'string',
        call: 'Decode(Encode($string))',
      },
    ],
  },
  {
    exercise: 'word-count',
    slug: 'word-count',
    operations: [
      { property: 'countWords', test: 'WordCount', func: 'WordCount', args: [['sentence', 'string']], result: { map: ['string', 'int'] } },
    ],
  },
  {
    exercise: 'pascals-triangle',
    slug: 'pascals-triangle',
    operations: [{ property: 'rows', test: 'Pascal', func: 'Pascal', args: [['count', 'int']], result: grid }],
  },
  {
    exercise: 'sieve',
    slug: 'sieve',
    operations: [{ property: 'primes', test: 'Primes', func: 'Primes', args: [['limit', 'int']], result: ints }],
  },
  {
    exercise: 'anagram',
    slug: 'anagram',
    operations: [
      {
        property: 'findAnagrams',
        test: 'Anagrams',
        func: 'Anagrams',
        args: [['subject', 'string'], ['candidates', strs]],
        result: strs,
      },
    ],
  },
  {
    exercise: 'prime-factors',
    slug: 'prime-factors',
    operations: [{ property: 'factors', test: 'Factors', func: 'Factors', args: [['value', 'int']], result: ints }],
  },
  {
    exercise: 'spiral-matrix',
    slug: 'spiral-matrix',
    operations: [{ property: 'spiralMatrix', test: 'Spiral', func: 'Spiral', args: [['size', 'int']], result: grid }],
  },
  {
    exercise: 'game-of-life',
    slug: 'game-of-life',
    operations: [{ property: 'tick', test: 'Tick', func: 'Tick', args: [['matrix', grid]], result: grid }],
  },
  {
    exercise: 'flower-field',
    slug: 'flower-field',
    operations: [{ property: 'annotate', test: 'Annotate', func: 'Annotate', args: [['garden', strs]], result: strs }],
  },
  {
    exercise: 'all-your-base',
    slug: 'all-your-base',
    operations: [
      {
        property: 'rebase',
        test: 'Rebase',
        func: 'Rebase',
        args: [['inputBase', 'int'], ['digits', ints], ['outputBase', 'int']],
        result: ints,
        error: true,
      },
    ],
  },
  {
    exercise: 'sublist',
    slug: 'sublist',
    operations: [
      { property: 'sublist', test: 'Compare', func: 'Compare', args: [['listOne', ints], ['listTwo', ints]], result: 'string' },
    ],
  },
  {
    exercise: 'change',
    slug: 'coin-change',
    operations: [
      {
        property: 'findFewestCoins',
        test: 'FewestCoins',
        func: 'FewestCoins',
        args: [['coins', ints], ['target', 'int']],
        result: ints,
        error: true,
      },
    ],
  },
  {
    exercise: 'knapsack',
    slug: 'knapsack',
    types: 'type Item struct {\n\tWeight int\n\tValue  int\n}',
    operations: [
      {
        property: 'maximumValue',
        test: 'MaxValue',
        func: 'MaxValue',
        args: [
          ['maximumWeight', 'int'],
          ['items', { slice: { struct: 'Item', fields: { weight: ['Weight', 'int'], value: ['Value', 'int'] } } }],
        ],
        result: 'int',
      },
    ],
  },
  {
    exercise: 'dominoes',
    slug: 'dominoes',
    operations: [
      { property: 'canChain', test: 'CanChain', func: 'CanChain', args: [['dominoes', { slice: { array: 'int', len: 2 } }]], result: 'bool' },
    ],
  },
]
