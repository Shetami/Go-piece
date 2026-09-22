import type { ReplScenario } from './types.ts'

/**
 * Пресеты стенда. У каждого одна мысль, которую он доказывает, — поле claim.
 */

export const stream: ReplScenario = {
  id: 'stream',
  title: 'Поток WAL',
  claim:
    'Реплика — это восстановление после сбоя, которое никогда не кончается: она получает WAL ведущего, сбрасывает его на диск и проигрывает. Отставание складывается из сети, диска и скорости проигрывания.',
  config: { syncCommit: 'on', standbys: 'none' },
  replicas: [
    { name: 'r1', latency: 1 },
    { name: 'r2', latency: 3, replayRate: 1 },
  ],
  keys: ['a', 'b', 'c'],
  clients: [
    { name: 'W1', at: 1, ops: [{ kind: 'write', key: 'a' }], repeat: 14, every: 1 },
    { name: 'W2', at: 2, ops: [{ kind: 'write', key: 'b' }], repeat: 14, every: 1 },
    { name: 'W3', at: 3, ops: [{ kind: 'write', key: 'c' }], repeat: 14, every: 1 },
  ],
  faults: [],
  watchFor: ['replica.receive', 'replica.replay'],
  stopAfter: 60,
}

export const readYourWrites: ReplScenario = {
  id: 'read-your-writes',
  title: 'Прочитать своё',
  claim:
    'Приложение сохраняет профиль на ведущем и сразу читает его с реплики — а там ещё старая версия. Коммит на ведущем ничего не обещает про реплики, если его не заставить ждать.',
  config: { syncCommit: 'on', standbys: 'none' },
  replicas: [
    { name: 'r1', latency: 1 },
    { name: 'r2', latency: 4 },
  ],
  keys: ['profile'],
  clients: [
    {
      name: 'U',
      at: 1,
      ops: [
        { kind: 'write', key: 'profile' },
        { kind: 'read', key: 'profile', from: 'any' },
      ],
      repeat: 6,
      every: 3,
    },
  ],
  faults: [],
  watchFor: ['read.own-stale', 'commit.ack'],
  stopAfter: 50,
}

export const monotonic: ReplScenario = {
  id: 'monotonic',
  title: 'Время назад',
  claim:
    'Чтения раскиданы по двум репликам с разным отставанием. Клиент видит счётчик 7, а следующим запросом — 4: данные «откатились», хотя никто ничего не откатывал.',
  config: { syncCommit: 'on', standbys: 'none' },
  replicas: [
    { name: 'r1', latency: 1 },
    { name: 'r2', latency: 3, replayRate: 1 },
  ],
  keys: ['likes'],
  clients: [
    { name: 'W1', at: 1, ops: [{ kind: 'write', key: 'likes' }], repeat: 12, every: 1 },
    { name: 'W2', at: 2, ops: [{ kind: 'write', key: 'likes' }], repeat: 12, every: 1 },
    { name: 'W3', at: 3, ops: [{ kind: 'write', key: 'likes' }], repeat: 12, every: 1 },
    { name: 'R', at: 8, ops: [{ kind: 'read', key: 'likes', from: 'any' }], repeat: 8, every: 1 },
  ],
  faults: [],
  watchFor: ['read.backwards'],
  stopAfter: 60,
}

export const failover: ReplScenario = {
  id: 'failover',
  title: 'Переключение и потери',
  claim:
    'С асинхронной репликацией ведущий подтверждает коммит, не дожидаясь реплик. Ведущий падает, реплика становится ведущей — и подтверждённые коммиты, которые до неё не дошли, исчезают.',
  config: { syncCommit: 'on', standbys: 'none', failover: true, detectAfter: 4 },
  replicas: [
    { name: 'r1', latency: 2 },
    { name: 'r2', latency: 2 },
  ],
  keys: ['orders', 'payments'],
  clients: [
    { name: 'W1', at: 1, ops: [{ kind: 'write', key: 'orders' }], repeat: 'forever', every: 1 },
    { name: 'W2', at: 2, ops: [{ kind: 'write', key: 'payments' }], repeat: 'forever', every: 1 },
  ],
  faults: [{ kind: 'primary-down', at: 14 }],
  watchFor: ['primary.down', 'failover.promote', 'commit.lost', 'commit.unknown'],
  stopAfter: 30,
}

export const syncHang: ReplScenario = {
  id: 'sync-hang',
  title: 'Синхронная реплика упала',
  claim:
    'С FIRST 1 (r1) каждый коммит ждёт одну конкретную реплику. Она падает — и запись на ведущем встаёт целиком, хотя сам ведущий жив. ANY 1 из двух переживает потерю одной.',
  config: { syncCommit: 'on', standbys: 'first-r1' },
  replicas: [
    { name: 'r1', latency: 1 },
    { name: 'r2', latency: 2 },
  ],
  keys: ['a', 'b'],
  clients: [
    { name: 'W1', at: 1, ops: [{ kind: 'write', key: 'a' }], repeat: 6, every: 1 },
    { name: 'W2', at: 2, ops: [{ kind: 'write', key: 'b' }], repeat: 6, every: 1 },
  ],
  faults: [{ kind: 'replica-down', replica: 1, at: 8, until: 24 }],
  watchFor: ['replica.down', 'commit.hang', 'replica.up'],
  stopAfter: 60,
}

export const slot: ReplScenario = {
  id: 'slot',
  title: 'Слот репликации',
  claim:
    'Реплика пропала на время. Со слотом ведущий хранит для неё весь WAL — и он растёт, пока она не вернётся. Без слота ведущий удаляет старый WAL, и вернувшаяся реплика уже не может догнать.',
  config: { syncCommit: 'on', standbys: 'none', slots: true, walKeep: 12, replicas: 1 },
  replicas: [{ name: 'r1', latency: 1 }],
  keys: ['a', 'b'],
  clients: [
    { name: 'W1', at: 1, ops: [{ kind: 'write', key: 'a' }], repeat: 20, every: 1 },
    { name: 'W2', at: 2, ops: [{ kind: 'write', key: 'b' }], repeat: 20, every: 1 },
  ],
  faults: [{ kind: 'replica-down', replica: 1, at: 6, until: 34 }],
  watchFor: ['replica.down', 'wal.retained', 'wal.removed', 'replica.up', 'replica.broken'],
  minTicks: 40,
  stopAfter: 70,
}

export const conflict: ReplScenario = {
  id: 'conflict',
  title: 'Отчёт на реплике',
  claim:
    'Вакуум на ведущем удаляет версии, которые ещё читает отчёт на реплике. Реплика ждёт max_standby_streaming_delay и отменяет запрос. hot_standby_feedback спасает запрос — ценой мусора на ведущем.',
  config: { syncCommit: 'on', standbys: 'none', vacuumEvery: 4, maxStandbyDelay: 4, replicas: 1 },
  replicas: [{ name: 'r1', latency: 1 }],
  keys: ['stats'],
  clients: [
    { name: 'W', at: 1, ops: [{ kind: 'write', key: 'stats' }], repeat: 22, every: 1 },
    { name: 'Q', at: 1, ops: [{ kind: 'query', key: 'stats', from: 1, ticks: 18 }] },
  ],
  faults: [],
  watchFor: ['conflict.wait', 'conflict.cancel', 'feedback.hold', 'query.done'],
  stopAfter: 70,
}

export const REPL_SCENARIOS: ReplScenario[] = [stream, readYourWrites, monotonic, failover, syncHang, slot, conflict]

export function replScenarioById(id: string): ReplScenario | undefined {
  return REPL_SCENARIOS.find((s) => s.id === id)
}
