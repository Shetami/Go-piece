import type { MemScenario } from './types.ts'

/**
 * Шесть пресетов. У каждого одна мысль, которую он доказывает, — поле claim.
 * Сценарий без claim заводить нельзя: получится красивая анимация, а не урок.
 */

export const fastPath: MemScenario = {
  id: 'fast-path',
  title: 'Быстрый путь',
  claim: 'Обычная аллокация не доходит до кучи: объект берётся из спана в кэше своего процессора, без единой блокировки.',
  config: { gomaxprocs: 4, heapPages: 96 },
  workloads: [
    {
      name: 'worker',
      count: 4,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size: 64, count: 6, lifetime: [10, 25] },
        { kind: 'cpu', ticks: [1, 3] },
      ],
      repeat: 'forever',
    },
    {
      name: 'tiny',
      count: 3,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size: [1, 8], count: 10, pointers: false, lifetime: [10, 20] },
        { kind: 'cpu', ticks: 1 },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['alloc.tiny', 'cache.refill', 'heap.grow', 'os.map'],
  stopAfter: 240,
}

export const sizeClasses: MemScenario = {
  id: 'size-classes',
  title: 'Классы размеров',
  claim: 'Аллокатор выдаёт не байты, а класс размера: один лишний байт в структуре может стоить трети выделенной памяти.',
  config: { gomaxprocs: 2, heapPages: 96 },
  workloads: [
    {
      name: 'tight',
      count: 2,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size: 32, count: 6, lifetime: [15, 30] },
        { kind: 'cpu', ticks: 2 },
      ],
      repeat: 'forever',
    },
    {
      name: 'plus-one',
      count: 2,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size: 33, count: 6, lifetime: [15, 30] },
        { kind: 'cpu', ticks: 2 },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['class.waste', 'cache.refill', 'heap.grow'],
  stopAfter: 240,
}

export const contention: MemScenario = {
  id: 'contention',
  title: 'Давка за спанами',
  claim: 'Быстрый путь кончается, когда кэш процессора пустеет: за спанами все идут в один общий список под блокировкой.',
  config: { gomaxprocs: 8, heapPages: 128, spanPages: 2 },
  workloads: [
    {
      name: 'chunky',
      count: 8,
      spawnAt: 0,
      phases: [{ kind: 'alloc', size: 2048, count: 4, lifetime: [20, 40] }],
      repeat: 'forever',
    },
  ],
  watchFor: ['central.contended', 'cache.refill', 'central.empty', 'heap.grow'],
  stopAfter: 240,
}

export const largeObjects: MemScenario = {
  id: 'large-objects',
  title: 'Большие объекты',
  claim: 'Объект больше 32 КБ идёт мимо всех кэшей прямо в кучу и требует непрерывных страниц — их может не найтись даже в пустой памяти.',
  config: { gomaxprocs: 4, heapPages: 128, scavengeAfter: 18 },
  workloads: [
    {
      name: 'buffers',
      count: 2,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size: [40960, 98304], count: 1, lifetime: [12, 24] },
        { kind: 'cpu', ticks: [3, 6] },
      ],
      repeat: 'forever',
    },
    {
      name: 'small',
      count: 4,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size: 128, count: 4, lifetime: [30, 60] },
        { kind: 'cpu', ticks: 2 },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['alloc.large', 'os.map', 'frag.external', 'scavenge'],
  stopAfter: 240,
}

export const escapeAnalysis: MemScenario = {
  id: 'escape',
  title: 'Побег в кучу',
  claim: 'Пока указатель не уезжает из функции, объект живёт на стеке и не стоит куче ничего — escape analysis решает это ещё при компиляции.',
  config: { gomaxprocs: 4, heapPages: 96 },
  workloads: [
    {
      name: 'local',
      count: 3,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size: 512, count: 5, escapes: false },
        { kind: 'cpu', ticks: 2 },
      ],
      repeat: 'forever',
    },
    {
      name: 'escaping',
      count: 3,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size: 512, count: 5, escapes: true, lifetime: [25, 50] },
        { kind: 'cpu', ticks: 2 },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['alloc.stack', 'cache.refill', 'heap.grow', 'stack.grow'],
  stopAfter: 240,
}

export const stackGrowth: MemScenario = {
  id: 'stack-growth',
  title: 'Рост стека',
  claim: 'Стек горутины начинается с 2 КБ и растёт копированием: старт дёшев, а за глубину приходится платить переездами.',
  config: { gomaxprocs: 4, heapPages: 96, stackCopyRate: 2048, shrinkEvery: 50 },
  workloads: [
    {
      name: 'deep',
      count: 2,
      spawnAt: 0,
      phases: [
        { kind: 'recurse', depth: 24, frame: 512 },
        { kind: 'cpu', ticks: [4, 8] },
      ],
      repeat: 'forever',
    },
    {
      name: 'flat',
      count: 4,
      spawnAt: 0,
      phases: [
        { kind: 'recurse', depth: 2, frame: 256 },
        { kind: 'cpu', ticks: 3 },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['stack.grow', 'stack.shrink', 'os.map'],
  stopAfter: 240,
}

export const MEM_SCENARIOS: MemScenario[] = [
  fastPath,
  sizeClasses,
  contention,
  largeObjects,
  escapeAnalysis,
  stackGrowth,
]

export function memScenarioById(id: string): MemScenario | undefined {
  return MEM_SCENARIOS.find((s) => s.id === id)
}
