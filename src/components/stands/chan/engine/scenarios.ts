import type { ChanScenario } from './types.ts'

/**
 * Семь пресетов. У каждого одна мысль, которую он доказывает, — поле claim.
 * Сценарий без claim заводить нельзя: получится красивая анимация, а не урок.
 */

export const rendezvous: ChanScenario = {
  id: 'rendezvous',
  title: 'Точка встречи',
  claim:
    'Небуферизованный канал ничего не хранит: тот, кто пришёл первым, паркуется и ждёт второго, а значение передаётся из рук в руки.',
  config: { gomaxprocs: 2 },
  chans: [{ name: 'ch', cap: 0 }],
  workloads: [
    {
      name: 'sender',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'send', chan: 'ch' },
        { kind: 'cpu', ticks: [2, 5] },
      ],
      repeat: 'forever',
    },
    {
      name: 'receiver',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'recv', chan: 'ch' },
        { kind: 'cpu', ticks: [3, 7] },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['send.block', 'recv.direct', 'recv.block', 'send.direct'],
  stopAfter: 120,
}

export const buffered: ChanScenario = {
  id: 'buffer',
  title: 'Буфер и обратное давление',
  claim:
    'Буфер не ускоряет медленного получателя — он лишь откладывает момент, когда отправитель встанет в очередь и начнёт ждать.',
  config: { gomaxprocs: 4 },
  chans: [{ name: 'jobs', cap: 4 }],
  workloads: [
    {
      name: 'producer',
      count: 2,
      spawnAt: 0,
      phases: [
        { kind: 'send', chan: 'jobs', count: 3 },
        { kind: 'cpu', ticks: 1 },
      ],
      repeat: 'forever',
    },
    {
      name: 'worker',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'recv', chan: 'jobs' },
        { kind: 'cpu', ticks: [4, 7] },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['send.buffer', 'buf.full', 'send.block', 'recv.wake'],
  stopAfter: 140,
}

export const pipeline: ChanScenario = {
  id: 'pipeline',
  title: 'Конвейер',
  claim:
    'Скорость конвейера задаёт самое медленное звено: буферы перед ним переполняются, а буферы после него стоят пустыми.',
  config: { gomaxprocs: 4 },
  chans: [
    { name: 'in', cap: 3 },
    { name: 'out', cap: 3 },
  ],
  workloads: [
    {
      name: 'source',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'send', chan: 'in' },
        { kind: 'cpu', ticks: 1 },
      ],
      repeat: 'forever',
    },
    {
      name: 'stage',
      count: 2,
      spawnAt: 0,
      phases: [
        { kind: 'recv', chan: 'in' },
        { kind: 'cpu', ticks: [3, 6] },
        { kind: 'send', chan: 'out' },
      ],
      repeat: 'forever',
    },
    {
      name: 'sink',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'recv', chan: 'out' },
        { kind: 'cpu', ticks: 1 },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['buf.full', 'send.block', 'recv.block', 'recv.wake'],
  stopAfter: 160,
}

export const selectScenario: ChanScenario = {
  id: 'select',
  title: 'select',
  claim:
    'select не опрашивает каналы в цикле: он встаёт в очередь сразу ко всем и спит, пока не отзовётся первый. А default превращает ожидание в мгновенную проверку.',
  config: { gomaxprocs: 4 },
  chans: [
    { name: 'jobs', cap: 0 },
    { name: 'tick', cap: 1 },
    { name: 'quit', cap: 0 },
  ],
  workloads: [
    {
      name: 'worker',
      count: 1,
      spawnAt: 0,
      phases: [
        {
          kind: 'select',
          cases: [
            { chan: 'jobs', op: 'recv' },
            { chan: 'tick', op: 'recv' },
            { chan: 'quit', op: 'recv' },
          ],
        },
        { kind: 'cpu', ticks: [1, 2] },
      ],
      repeat: 'forever',
    },
    {
      name: 'jobber',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'send', chan: 'jobs' },
        { kind: 'cpu', ticks: [4, 7] },
      ],
      repeat: 'forever',
    },
    {
      name: 'ticker',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'cpu', ticks: [7, 11] },
        { kind: 'send', chan: 'tick' },
      ],
      repeat: 'forever',
    },
    {
      name: 'poller',
      count: 1,
      spawnAt: 2,
      phases: [
        { kind: 'select', cases: [{ chan: 'jobs', op: 'recv' }], default: true },
        { kind: 'cpu', ticks: [2, 4] },
      ],
      repeat: 'forever',
    },
  ],
  watchFor: ['select.block', 'select.ready', 'select.default'],
  stopAfter: 140,
}

export const closing: ChanScenario = {
  id: 'close',
  title: 'Закрытие',
  claim:
    'Закрытие — это широковещательный сигнал: оно будит всех получателей разом. Но закрывать канал может только отправитель, и только один: отправка в закрытый канал — паника.',
  config: { gomaxprocs: 4 },
  chans: [{ name: 'work', cap: 2 }],
  workloads: [
    {
      name: 'worker',
      count: 3,
      spawnAt: 0,
      phases: [
        { kind: 'recv', chan: 'work' },
        { kind: 'cpu', ticks: [2, 4] },
      ],
      repeat: 4,
    },
    {
      name: 'boss',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'cpu', ticks: 2 },
        { kind: 'send', chan: 'work', count: 4 },
        { kind: 'cpu', ticks: 6 },
        { kind: 'close', chan: 'work' },
        { kind: 'cpu', ticks: 8 },
        // Тот самый забытый отправитель, который узнаёт о закрытии последним.
        { kind: 'send', chan: 'work' },
      ],
      repeat: 1,
    },
  ],
  watchFor: ['chan.close', 'recv.closed', 'send.closed'],
  stopAfter: 80,
}

export const deadlock: ChanScenario = {
  id: 'deadlock',
  title: 'Взаимная блокировка',
  claim:
    'Две горутины, которые сперва отправляют, а потом принимают, не встретятся никогда. А операция на nil-канале не блокируется «пока что» — она блокируется навсегда.',
  config: { gomaxprocs: 4 },
  chans: [
    { name: 'a', cap: 0 },
    { name: 'b', cap: 0 },
    { name: 'never', cap: 0, nil: true },
  ],
  workloads: [
    {
      name: 'alice',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'cpu', ticks: 1 },
        { kind: 'send', chan: 'a' },
        { kind: 'recv', chan: 'b' },
      ],
      repeat: 1,
    },
    {
      name: 'bob',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'cpu', ticks: 1 },
        { kind: 'send', chan: 'b' },
        { kind: 'recv', chan: 'a' },
      ],
      repeat: 1,
    },
    {
      name: 'ghost',
      count: 1,
      spawnAt: 0,
      phases: [{ kind: 'recv', chan: 'never' }],
      repeat: 1,
    },
  ],
  watchFor: ['chan.nil', 'send.block', 'deadlock'],
  stopAfter: 40,
}

export const leak: ChanScenario = {
  id: 'leak',
  title: 'Забытый получатель',
  claim:
    'Горутина, которой некому отдать значение, не падает и не пишет в лог — она просто остаётся в памяти навсегда. Рантайм такую утечку не видит.',
  config: { gomaxprocs: 4, leakAfter: 30 },
  chans: [{ name: 'results', cap: 0 }],
  workloads: [
    {
      name: 'worker',
      count: 4,
      spawnAt: 0,
      phases: [
        { kind: 'cpu', ticks: [2, 6] },
        { kind: 'send', chan: 'results' },
      ],
      repeat: 1,
    },
    {
      // Забрал первый ответ и ушёл считать. Остальные три так и стоят с результатом в руках.
      name: 'collector',
      count: 1,
      spawnAt: 0,
      phases: [
        { kind: 'recv', chan: 'results' },
        { kind: 'cpu', ticks: 200 },
      ],
      repeat: 1,
    },
  ],
  watchFor: ['send.block', 'recv.direct', 'leak'],
  stopAfter: 80,
}

export const CHAN_SCENARIOS: ChanScenario[] = [
  rendezvous,
  buffered,
  pipeline,
  selectScenario,
  closing,
  deadlock,
  leak,
]

export function chanScenarioById(id: string): ChanScenario | undefined {
  return CHAN_SCENARIOS.find((s) => s.id === id)
}
