import type { Scenario } from './types.ts'

/**
 * Шесть пресетов. У каждого одна мысль, которую он доказывает, — поле claim.
 * Сценарий без claim заводить нельзя: получится красивая анимация, а не урок.
 */

export const calmStart: Scenario = {
  id: 'calm-start',
  title: 'Спокойный старт',
  claim: 'Базовый цикл планировщика: очередь → исполнение → вытеснение по кванту.',
  config: { gomaxprocs: 4, quantum: 12 },
  workloads: [
    { name: 'worker', count: 8, spawnAt: 0, phases: [{ kind: 'cpu', ticks: [8, 14] }], repeat: 3 },
  ],
  watchFor: ['g.scheduled', 'sysmon.preempt'],
  stopAfter: 300,
}

export const skewAndStealing: Scenario = {
  id: 'skew-and-stealing',
  title: 'Перекос и кража',
  claim: 'Без work stealing три процессора простаивают рядом с одним перегруженным.',
  config: { gomaxprocs: 4, quantum: 100 },
  workloads: [
    {
      name: 'worker',
      count: 24,
      spawnAt: 0,
      spawnOn: 0,
      phases: [{ kind: 'cpu', ticks: [6, 12] }],
      repeat: 2,
    },
  ],
  watchFor: ['p.stole', 'p.stealFailed', 'p.idle'],
  stopAfter: 400,
}

export const blockingSyscalls: Scenario = {
  id: 'blocking-syscalls',
  title: 'Блокирующие вызовы',
  claim: 'Долгие системные вызовы отбирают P и заставляют рантайм плодить потоки далеко за GOMAXPROCS.',
  config: { gomaxprocs: 4, retakeThreshold: 5 },
  workloads: [
    {
      name: 'io',
      count: 20,
      spawnAt: 0,
      phases: [
        { kind: 'cpu', ticks: [2, 4] },
        { kind: 'syscall', ticks: [20, 30] },
      ],
      repeat: 3,
    },
  ],
  watchFor: ['sysmon.retake', 'p.handoff', 'm.spawned'],
  stopAfter: 500,
}

export const netVsFiles: Scenario = {
  id: 'net-vs-files',
  title: 'Сеть против файлов',
  claim: 'Сетевое ожидание не стоит ни одного потока, файловое — стоит каждое.',
  config: { gomaxprocs: 4, retakeThreshold: 5 },
  workloads: [
    {
      name: 'net',
      count: 10,
      spawnAt: 0,
      phases: [
        { kind: 'cpu', ticks: [2, 3] },
        { kind: 'net', ticks: [20, 30] },
      ],
      repeat: 3,
    },
    {
      name: 'file',
      count: 10,
      spawnAt: 0,
      phases: [
        { kind: 'cpu', ticks: [2, 3] },
        { kind: 'syscall', ticks: [20, 30] },
      ],
      repeat: 3,
    },
  ],
  watchFor: ['net.ready', 'm.spawned', 'sysmon.retake'],
  stopAfter: 500,
}

export const greedyLoop: Scenario = {
  id: 'greedy-loop',
  title: 'Жадный цикл',
  claim: 'До Go 1.14 одна горутина без точек безопасности вешала весь процессор; асинхронное вытеснение это чинит.',
  config: { gomaxprocs: 1, quantum: 15 },
  workloads: [
    { name: 'hog', count: 1, spawnAt: 0, spawnOn: 0, phases: [{ kind: 'cpu', ticks: 'forever' }] },
    {
      name: 'worker',
      count: 5,
      spawnAt: 1,
      spawnOn: 0,
      phases: [{ kind: 'cpu', ticks: [4, 8] }],
      repeat: 3,
    },
  ],
  watchFor: ['sysmon.preempt', 'g.scheduled'],
  stopAfter: 300,
}

export const channelPingPong: Scenario = {
  id: 'channel-ping-pong',
  title: 'Пинг-понг по каналу',
  claim: 'Слот runnext пропускает разбуженного партнёра вперёд очереди — без него каждый обмен идёт в хвост.',
  config: { gomaxprocs: 1, quantum: 30 },
  workloads: [
    {
      name: 'pinger',
      count: 1,
      spawnAt: 0,
      spawnOn: 0,
      phases: [
        { kind: 'chanSend', chan: 'ball' },
        { kind: 'chanRecv', chan: 'ball' },
      ],
      repeat: 'forever',
    },
    {
      name: 'ponger',
      count: 1,
      spawnAt: 0,
      spawnOn: 0,
      phases: [
        { kind: 'chanRecv', chan: 'ball' },
        { kind: 'chanSend', chan: 'ball' },
      ],
      repeat: 'forever',
    },
    {
      name: 'background',
      count: 6,
      spawnAt: 0,
      spawnOn: 0,
      phases: [{ kind: 'cpu', ticks: [5, 9] }],
      repeat: 'forever',
    },
  ],
  watchFor: ['g.ready', 'g.blocked', 'g.scheduled'],
  stopAfter: 300,
}

export const SCENARIOS: Scenario[] = [
  calmStart,
  skewAndStealing,
  blockingSyscalls,
  netVsFiles,
  greedyLoop,
  channelPingPong,
]

export function scenarioById(id: string): Scenario | undefined {
  return SCENARIOS.find((s) => s.id === id)
}
