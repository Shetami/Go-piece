import type { TxnScenario } from './types.ts'

/**
 * Пресеты стенда. У каждого одна мысль, которую он доказывает, — поле claim.
 * Сценарий без claim заводить нельзя: получится красивая анимация, а не урок.
 *
 * Первые шесть — к лекции о транзакциях: аномалии и то, как от них защищают
 * уровни изоляции. Остальные — к лекции о PostgreSQL: версии строк, очистка,
 * очередь к строке, SSI и долговечность коммита.
 */

export const dirty: TxnScenario = {
  id: 'dirty',
  title: 'Грязное чтение',
  claim:
    'На READ UNCOMMITTED транзакция видит чужие изменения до коммита. Если их потом откатят, она успеет принять решение по данным, которых никогда не было.',
  config: { isolation: 'read-uncommitted' },
  rows: [{ key: 'acc', value: 100 }],
  sessions: [
    {
      name: 'T1',
      at: 1,
      ops: [
        { kind: 'update', key: 'acc', set: { kind: 'delta', d: -100 } },
        { kind: 'rollback', after: 3 },
      ],
    },
    {
      name: 'T2',
      at: 2,
      ops: [{ kind: 'read', key: 'acc' }, { kind: 'commit' }],
    },
  ],
  watchFor: ['anomaly.dirty', 'txn.rollback'],
  stopAfter: 20,
}

export const nonrepeatable: TxnScenario = {
  id: 'nonrepeatable',
  title: 'Неповторяемое чтение',
  claim:
    'На READ COMMITTED каждый оператор берёт свой снимок. Два одинаковых SELECT в одной транзакции могут вернуть разное — если между ними кто-то закоммитил изменение.',
  config: { isolation: 'read-committed' },
  rows: [{ key: 'acc', value: 100 }],
  sessions: [
    {
      name: 'T1',
      at: 1,
      ops: [
        { kind: 'read', key: 'acc' },
        { kind: 'read', key: 'acc', after: 5 },
        { kind: 'commit' },
      ],
    },
    {
      name: 'T2',
      at: 3,
      ops: [{ kind: 'update', key: 'acc', set: { kind: 'delta', d: -30 } }, { kind: 'commit' }],
    },
  ],
  watchFor: ['snap.take', 'txn.commit', 'anomaly.nonrepeatable'],
  stopAfter: 20,
}

export const phantom: TxnScenario = {
  id: 'phantom',
  title: 'Фантом',
  claim:
    'Чтение по условию на READ COMMITTED может во второй раз найти строки, которых не было в первый: их вставили и закоммитили между запросами. В PostgreSQL от этого защищает уже REPEATABLE READ.',
  config: { isolation: 'read-committed' },
  rows: [
    { key: 'o1', value: 120 },
    { key: 'o2', value: 80 },
    { key: 'o3', value: 150 },
  ],
  sessions: [
    {
      name: 'T1',
      at: 1,
      ops: [
        { kind: 'scan', min: 100 },
        { kind: 'scan', min: 100, after: 5 },
        { kind: 'commit' },
      ],
    },
    {
      name: 'T2',
      at: 2,
      ops: [{ kind: 'insert', key: 'o4', value: 300 }, { kind: 'commit' }],
    },
  ],
  watchFor: ['row.insert', 'anomaly.phantom'],
  stopAfter: 20,
}

export const lostUpdate: TxnScenario = {
  id: 'lost-update',
  title: 'Потерянное обновление',
  claim:
    'Две транзакции читают баланс, прибавляют каждая своё и записывают. На READ COMMITTED вторая запись затирает первую — и никакой ошибки при этом нет.',
  config: { isolation: 'read-committed', rmw: 'app' },
  rows: [{ key: 'acc', value: 100 }],
  sessions: [
    {
      name: 'T1',
      at: 1,
      ops: [
        { kind: 'read', key: 'acc', rmw: true },
        { kind: 'update', key: 'acc', set: { kind: 'read', d: 50 }, after: 3 },
        { kind: 'commit' },
      ],
    },
    {
      name: 'T2',
      at: 2,
      ops: [
        { kind: 'read', key: 'acc', rmw: true },
        { kind: 'update', key: 'acc', set: { kind: 'read', d: 30 }, after: 3 },
        { kind: 'commit' },
      ],
    },
  ],
  invariant: { kind: 'conserve', keys: ['acc'], label: 'каждое подтверждённое пополнение на счету' },
  watchFor: ['lock.wait', 'row.recheck', 'anomaly.lost', 'txn.abort', 'invariant.broken'],
  stopAfter: 30,
}

export const writeSkew: TxnScenario = {
  id: 'write-skew',
  title: 'Перекос записи',
  claim:
    'Две транзакции проверяют одно условие и меняют РАЗНЫЕ строки. Конфликта записи нет, REPEATABLE READ пропускает обе — и условие, которое проверяла каждая, нарушено.',
  config: { isolation: 'repeatable-read' },
  rows: [
    { key: 'alice', value: 1 },
    { key: 'bob', value: 1 },
  ],
  sessions: [
    {
      name: 'T1',
      at: 1,
      ops: [
        { kind: 'read', key: 'alice' },
        { kind: 'read', key: 'bob' },
        { kind: 'update', key: 'alice', set: { kind: 'const', value: 0 }, when: { sumOf: ['alice', 'bob'], gte: 2 } },
        { kind: 'commit' },
      ],
    },
    {
      name: 'T2',
      at: 1,
      ops: [
        { kind: 'read', key: 'alice' },
        { kind: 'read', key: 'bob' },
        { kind: 'update', key: 'bob', set: { kind: 'const', value: 0 }, when: { sumOf: ['alice', 'bob'], gte: 2 } },
        { kind: 'commit' },
      ],
    },
  ],
  invariant: { kind: 'min-sum', keys: ['alice', 'bob'], min: 1, label: 'дежурит хотя бы один врач' },
  watchFor: ['row.update', 'ssi.conflict', 'txn.abort', 'invariant.broken'],
  stopAfter: 20,
}

export const deadlock: TxnScenario = {
  id: 'deadlock',
  title: 'Взаимная блокировка',
  claim:
    'Два перевода между одними счетами, но в разном порядке: каждый держит одну строку и ждёт другую. Ждать можно вечно — поэтому база ищет цикл и убивает одну из транзакций.',
  config: { isolation: 'read-committed', deadlockTimeout: 3 },
  rows: [
    { key: 'a', value: 100 },
    { key: 'b', value: 100 },
  ],
  sessions: [
    {
      name: 'T1',
      at: 1,
      ops: [
        { kind: 'update', key: 'a', set: { kind: 'delta', d: -10 } },
        { kind: 'update', key: 'b', set: { kind: 'delta', d: 10 }, after: 2 },
        { kind: 'commit' },
      ],
    },
    {
      name: 'T2',
      at: 2,
      ops: [
        { kind: 'update', key: 'b', set: { kind: 'delta', d: -20 } },
        { kind: 'update', key: 'a', set: { kind: 'delta', d: 20 }, after: 2 },
        { kind: 'commit' },
      ],
    },
  ],
  invariant: { kind: 'conserve', keys: ['a', 'b'], label: 'все подтверждённые переводы на месте' },
  watchFor: ['lock.wait', 'deadlock.found', 'txn.abort'],
  stopAfter: 30,
}

export const versions: TxnScenario = {
  id: 'versions',
  title: 'Версии строки',
  claim:
    'UPDATE не перезаписывает строку, а добавляет новую версию. Старые остаются в таблице, пока их может видеть хоть один снимок, — и транзакция со старым снимком продолжает читать прошлое.',
  config: { isolation: 'read-committed', autovacuum: 6 },
  rows: [{ key: 'acc', value: 100 }],
  sessions: [
    {
      name: 'R',
      isolation: 'repeatable-read',
      at: 1,
      ops: [
        { kind: 'read', key: 'acc' },
        { kind: 'read', key: 'acc', after: 8 },
        { kind: 'commit' },
      ],
    },
    {
      name: 'T1',
      at: 2,
      ops: [{ kind: 'update', key: 'acc', set: { kind: 'delta', d: 10 } }, { kind: 'commit' }],
    },
    {
      name: 'T2',
      at: 5,
      ops: [{ kind: 'update', key: 'acc', set: { kind: 'delta', d: 10 } }, { kind: 'commit' }],
    },
  ],
  watchFor: ['row.update', 'vacuum.blocked', 'vacuum.run'],
  minTicks: 12,
  stopAfter: 30,
}

export const rowLock: TxnScenario = {
  id: 'row-lock',
  title: 'Очередь к строке',
  claim:
    'UPDATE … SET v = v + 10 не теряет обновлений даже на READ COMMITTED: он ждёт, пока строку отпустят, и перечитывает свежую версию. Цена — очередь: каждый следующий ждёт всех предыдущих.',
  config: { isolation: 'read-committed' },
  rows: [{ key: 'acc', value: 100 }],
  sessions: [
    {
      name: 'T1',
      at: 1,
      ops: [
        { kind: 'update', key: 'acc', set: { kind: 'delta', d: 10 } },
        { kind: 'commit', after: 3 },
      ],
    },
    {
      name: 'T2',
      at: 2,
      ops: [{ kind: 'update', key: 'acc', set: { kind: 'delta', d: 10 } }, { kind: 'commit', after: 2 }],
    },
    {
      name: 'T3',
      at: 3,
      ops: [{ kind: 'update', key: 'acc', set: { kind: 'delta', d: 10 } }, { kind: 'commit', after: 2 }],
    },
  ],
  invariant: { kind: 'conserve', keys: ['acc'], label: 'каждое подтверждённое пополнение на счету' },
  watchFor: ['lock.wait', 'row.recheck', 'txn.commit'],
  stopAfter: 30,
}

export const ssi: TxnScenario = {
  id: 'ssi',
  title: 'SERIALIZABLE и повтор',
  claim:
    'SERIALIZABLE не блокирует чтение — он следит, кто чего не увидел. Когда из таких зависимостей складывается цикл, одна транзакция получает ошибку, и приложение обязано её повторить.',
  config: { isolation: 'serializable', retry: true },
  rows: writeSkew.rows,
  sessions: writeSkew.sessions,
  invariant: writeSkew.invariant,
  watchFor: ['ssi.conflict', 'txn.abort', 'txn.retry', 'stmt.skip'],
  stopAfter: 30,
}

export const bloat: TxnScenario = {
  id: 'bloat',
  title: 'Забытая транзакция',
  claim:
    'Одна открытая транзакция со старым снимком останавливает очистку всей таблицы: автовакуум приходит, но убрать не может ничего, и таблица растёт, хотя живая строка в ней одна.',
  config: { isolation: 'read-committed', autovacuum: 8 },
  rows: [{ key: 'hits', value: 0 }],
  sessions: [
    {
      name: 'W',
      at: 1,
      ops: [{ kind: 'update', key: 'hits', set: { kind: 'delta', d: 1 } }, { kind: 'commit' }],
      repeat: 16,
      every: 1,
    },
    {
      name: 'R',
      isolation: 'repeatable-read',
      at: 4,
      ops: [
        { kind: 'read', key: 'hits' },
        { kind: 'commit', after: 32 },
      ],
    },
  ],
  watchFor: ['vacuum.blocked', 'txn.commit'],
  minTicks: 48,
  stopAfter: 60,
}

export const durability: TxnScenario = {
  id: 'durability',
  title: 'Коммит и падение',
  claim:
    'Коммит долговечен, когда его запись в WAL на диске. С synchronous_commit=on клиент ждёт fsync, зато падение не теряет ничего подтверждённого; с off — подтверждение приходит сразу, а последние коммиты может унести падение.',
  config: { syncCommit: true, fsyncTicks: 3, walWriterDelay: 5, crashAt: 17, replayRate: 4, autovacuum: 5 },
  rows: [
    { key: 'k1', value: 0 },
    { key: 'k2', value: 0 },
    { key: 'k3', value: 0 },
  ],
  sessions: [1, 2, 3].map((i) => ({
    name: `C${i}`,
    at: i,
    ops: [
      { kind: 'update' as const, key: `k${i}`, set: { kind: 'delta' as const, d: 10 } },
      { kind: 'commit' as const },
    ],
    repeat: 8,
    every: 1,
  })),
  invariant: { kind: 'conserve', keys: ['k1', 'k2', 'k3'], label: 'все подтверждённые коммиты пережили падение' },
  watchFor: ['wal.flush', 'db.crash', 'commit.lost', 'db.recovered'],
  stopAfter: 60,
}

export const TXN_SCENARIOS: TxnScenario[] = [
  dirty,
  nonrepeatable,
  phantom,
  lostUpdate,
  writeSkew,
  deadlock,
  versions,
  rowLock,
  ssi,
  bloat,
  durability,
]

export function txnScenarioById(id: string): TxnScenario | undefined {
  return TXN_SCENARIOS.find((s) => s.id === id)
}
