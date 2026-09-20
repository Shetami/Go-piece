import type { GcScenario } from './types.ts'

/**
 * Шесть пресетов. У каждого одна мысль, которую он доказывает, — поле claim.
 * Сценарий без claim заводить нельзя: получится красивая анимация, а не урок.
 */

export const firstCycle: GcScenario = {
  id: 'first-cycle',
  title: 'Один цикл целиком',
  claim: 'Сборка идёт одновременно с программой: мир останавливается только на двух коротких паузах по краям разметки.',
  config: { gomaxprocs: 4, gogc: 100, initialGoal: 36, heapCapacity: 200 },
  workloads: [
    {
      name: 'worker',
      count: 4,
      spawnAt: 0,
      stackSlots: 3,
      phases: [
        { kind: 'cpu', ticks: [2, 4] },
        { kind: 'alloc', blocks: 3 },
        { kind: 'cpu', ticks: [3, 5] },
        { kind: 'drop' },
      ],
      repeat: 'forever',
    },
    {
      name: 'cache',
      count: 1,
      spawnAt: 0,
      stackSlots: 2,
      phases: [
        { kind: 'cpu', ticks: [8, 12] },
        { kind: 'alloc', blocks: 2, retain: true },
      ],
      repeat: 10,
    },
  ],
  watchFor: ['gc.trigger', 'stw.enter', 'mark.drained', 'cycle.done'],
  stopAfter: 300,
}

export const writeBarrier: GcScenario = {
  id: 'write-barrier',
  title: 'Барьер записи',
  claim: 'Без барьера записи объект, спрятанный за уже просмотренным, остаётся белым — и подметальщик освобождает живую память.',
  config: { gomaxprocs: 4, gogc: 60, initialGoal: 30, heapCapacity: 160, scanRate: 1 },
  workloads: [
    {
      name: 'builder',
      count: 3,
      spawnAt: 0,
      stackSlots: 3,
      phases: [
        { kind: 'alloc', blocks: 2 },
        { kind: 'cpu', ticks: [2, 4] },
        { kind: 'alloc', blocks: 2 },
        { kind: 'cpu', ticks: [3, 6] },
        { kind: 'drop' },
      ],
      repeat: 'forever',
    },
    {
      name: 'cache',
      count: 1,
      spawnAt: 0,
      stackSlots: 2,
      phases: [
        { kind: 'cpu', ticks: [6, 9] },
        { kind: 'alloc', blocks: 2, retain: true },
      ],
      repeat: 6,
    },
    {
      name: 'shuffler',
      count: 1,
      spawnAt: 2,
      stackSlots: 2,
      phases: [
        { kind: 'cpu', ticks: 2 },
        { kind: 'move' },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['barrier.shade', 'barrier.missed', 'heap.lost', 'mark.drained'],
  stopAfter: 300,
}

export const assistPressure: GcScenario = {
  id: 'assist-pressure',
  title: 'Помощь в разметке',
  claim: 'Кто выделяет быстрее, чем сборщик размечает, тот размечает сам: аллокация оплачивается работой.',
  config: { gomaxprocs: 4, gogc: 100, initialGoal: 40, heapCapacity: 260, scanRate: 3 },
  workloads: [
    {
      name: 'greedy',
      count: 4,
      spawnAt: 0,
      stackSlots: 3,
      phases: [
        { kind: 'alloc', blocks: 3 },
        { kind: 'cpu', ticks: 1 },
      ],
      repeat: 'forever',
    },
    {
      name: 'cache',
      count: 2,
      spawnAt: 0,
      stackSlots: 2,
      phases: [
        { kind: 'cpu', ticks: [4, 7] },
        { kind: 'alloc', blocks: 3, retain: true },
      ],
      repeat: 8,
    },
  ],
  watchFor: ['assist.begin', 'gc.trigger', 'cycle.done'],
  stopAfter: 300,
}

export const gogcDial: GcScenario = {
  id: 'gogc-dial',
  title: 'Ручка GOGC',
  claim: 'GOGC не ускоряет сборку — он меняет размер кучи: вдвое больше мусора терпим, вдвое реже собираем.',
  config: { gomaxprocs: 4, gogc: 100, initialGoal: 30, heapCapacity: 300 },
  workloads: [
    {
      name: 'worker',
      count: 4,
      spawnAt: 0,
      stackSlots: 3,
      phases: [
        { kind: 'cpu', ticks: [1, 3] },
        { kind: 'alloc', blocks: 3 },
        { kind: 'cpu', ticks: [2, 4] },
        { kind: 'drop' },
      ],
      repeat: 'forever',
    },
    {
      name: 'cache',
      count: 1,
      spawnAt: 0,
      stackSlots: 2,
      phases: [
        { kind: 'cpu', ticks: [6, 10] },
        { kind: 'alloc', blocks: 2, retain: true },
      ],
      repeat: 10,
    },
  ],
  watchFor: ['gc.trigger', 'pacer.set', 'cycle.done'],
  stopAfter: 300,
}

export const memLimit: GcScenario = {
  id: 'mem-limit',
  title: 'GOMEMLIMIT и спираль',
  claim: 'Когда живое подбирается к лимиту памяти, сборщик перестаёт останавливаться: память удерживается ценой процессора.',
  config: { gomaxprocs: 4, gogc: 100, memLimit: 70, initialGoal: 30, heapCapacity: 200, scanRate: 2 },
  workloads: [
    {
      name: 'leak',
      count: 2,
      spawnAt: 0,
      stackSlots: 2,
      phases: [
        { kind: 'cpu', ticks: [2, 3] },
        { kind: 'alloc', blocks: 2, retain: true },
      ],
      repeat: 16,
    },
    {
      name: 'worker',
      count: 4,
      spawnAt: 0,
      stackSlots: 3,
      phases: [
        { kind: 'cpu', ticks: [2, 4] },
        { kind: 'alloc', blocks: 2 },
        { kind: 'drop' },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['limit.hit', 'assist.begin', 'gc.trigger', 'oom'],
  stopAfter: 320,
}

export const garbageVsLive: GcScenario = {
  id: 'garbage-vs-live',
  title: 'Мусор почти бесплатен',
  claim: 'Разметка стоит по живому, а не по мусору: гора недолговечных объектов не добавляет сборщику работы.',
  config: { gomaxprocs: 4, gogc: 100, initialGoal: 60, heapCapacity: 260, scanRate: 4 },
  workloads: [
    {
      name: 'churn',
      count: 4,
      spawnAt: 0,
      stackSlots: 2,
      phases: [
        { kind: 'alloc', blocks: 3 },
        { kind: 'cpu', ticks: 1 },
        { kind: 'drop' },
      ],
      repeat: 'forever',
    },
    {
      name: 'cache',
      count: 1,
      spawnAt: 0,
      stackSlots: 2,
      phases: [
        { kind: 'cpu', ticks: [5, 8] },
        { kind: 'alloc', blocks: 2, retain: true },
      ],
      repeat: 10,
    },
  ],
  watchFor: ['gc.trigger', 'mark.drained', 'cycle.done'],
  stopAfter: 300,
}

export const GC_SCENARIOS: GcScenario[] = [
  firstCycle,
  writeBarrier,
  assistPressure,
  gogcDial,
  memLimit,
  garbageVsLive,
]

export function gcScenarioById(id: string): GcScenario | undefined {
  return GC_SCENARIOS.find((s) => s.id === id)
}
