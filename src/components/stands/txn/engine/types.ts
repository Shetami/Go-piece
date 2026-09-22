/**
 * Типы модели транзакций.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Правила взяты из PostgreSQL почти без
 * изменений: UPDATE не перезаписывает строку, а создаёт новую версию с xmin/xmax,
 * видимость решает снимок, строку блокирует xmax пишущей транзакции, на READ
 * COMMITTED заблокированное UPDATE перечитывает свежую версию, на REPEATABLE READ
 * получает ошибку сериализации, SERIALIZABLE ловит опасные rw-зависимости (SSI),
 * VACUUM убирает только то, что не видит ни один снимок, а коммит долговечен,
 * когда его запись в WAL сброшена на диск.
 *
 * А вот время измеряется в тиках, таблица одна, индексов нет, одна операция
 * длится ровно один тик, а у SSI нет ложных срабатываний от укрупнения блокировок.
 * Полный список упрощений — в конце лекции о PostgreSQL.
 */

/** Уровни изоляции. READ UNCOMMITTED здесь учебный: в PostgreSQL он работает как READ COMMITTED. */
export type Isolation = 'read-uncommitted' | 'read-committed' | 'repeatable-read' | 'serializable'

export const ISOLATIONS: Isolation[] = ['read-uncommitted', 'read-committed', 'repeatable-read', 'serializable']

/** Как приложение делает «прочитать — изменить — записать». */
export type RmwStyle =
  /** SELECT, потом UPDATE со значением, посчитанным в приложении. */
  | 'app'
  /** SELECT … FOR UPDATE, потом UPDATE: строка заблокирована с момента чтения. */
  | 'for-update'
  /** Одним запросом: UPDATE … SET v = v + d. Считает сама база, по текущей версии. */
  | 'atomic'

/** Новое значение в UPDATE. */
export type Expr =
  /** SET v = 42 */
  | { kind: 'const'; value: number }
  /** SET v = v + d — считается по той версии строки, которую UPDATE в итоге меняет. */
  | { kind: 'delta'; d: number }
  /** SET v = <прочитанное раньше> + d — значение посчитало приложение. */
  | { kind: 'read'; d: number }

/** Проверка в приложении перед записью: «если дежурных хотя бы двое — снимаю себя». */
export interface Cond {
  /** Сумма значений, которые транзакция прочитала по этим ключам. */
  sumOf: string[]
  gte: number
}

interface OpBase {
  /** Сколько тиков пройдёт после предыдущей операции. По умолчанию 1 — следующий тик. */
  after?: number
}

export type Op =
  /** SELECT v FROM t WHERE k = key. `rmw` — это чтение перед записью, его стиль задаёт настройка. */
  | (OpBase & { kind: 'read'; key: string; forUpdate?: boolean; rmw?: boolean })
  /** SELECT count(*), sum(v) FROM t WHERE v >= min — чтение по условию, а не по ключу. */
  | (OpBase & { kind: 'scan'; min?: number })
  | (OpBase & { kind: 'update'; key: string; set: Expr; when?: Cond })
  | (OpBase & { kind: 'insert'; key: string; value: number })
  | (OpBase & { kind: 'delete'; key: string })
  | (OpBase & { kind: 'commit' })
  | (OpBase & { kind: 'rollback' })

export type OpKind = Op['kind']

/** Сценарий одной «сессии» — клиента, который выполняет транзакцию. */
export interface TxnSpec {
  name: string
  /** null — уровень из общей настройки. */
  isolation?: Isolation | null
  /** На каком тике выполняется первая операция. */
  at: number
  ops: Op[]
  /** Сколько раз повторить транзакцию целиком (каждый раз — новая транзакция). */
  repeat?: number
  /** Через сколько тиков после конца начинать следующий повтор. */
  every?: number
}

export interface RowSpec {
  key: string
  value: number
}

/** Что должно оставаться верным. Проверяется по закоммиченному состоянию. */
export type Invariant =
  /**
   * Каждое подтверждённое клиенту изменение на месте: значение ключа равно
   * исходному плюс все изменения транзакций, которые получили «COMMIT».
   */
  | { kind: 'conserve'; keys: string[]; label: string }
  /** Сумма по ключам не ниже порога: «дежурит хотя бы один врач». */
  | { kind: 'min-sum'; keys: string[]; min: number; label: string }

export interface TxnConfig {
  /** default_transaction_isolation. */
  isolation: Isolation
  rmw: RmwStyle
  /** Клиент повторяет транзакцию, упавшую с ошибкой сериализации или дедлоком. */
  retry: boolean
  /** deadlock_timeout: через сколько тиков ожидания блокировки искать цикл. */
  deadlockTimeout: number
  /** Интервал автовакуума в тиках; 0 — выключен. */
  autovacuum: number
  /** synchronous_commit: ждать ли, пока запись о коммите окажется на диске. */
  syncCommit: boolean
  /** Сколько тиков длится сброс WAL на диск (fsync). */
  fsyncTicks: number
  /** wal_writer_delay: как часто фоновый процесс сбрасывает WAL сам. */
  walWriterDelay: number
  /** Интервал контрольной точки; 0 — только в начале. */
  checkpointEvery: number
  /** На каком тике сервер падает; 0 — не падает. */
  crashAt: number
  /** Сколько записей WAL восстановление проигрывает за тик. */
  replayRate: number
  /** Сколько слотов в странице кучи. */
  pageSlots: number
}

export const DEFAULT_TXN_CONFIG: TxnConfig = {
  isolation: 'read-committed',
  rmw: 'app',
  retry: false,
  deadlockTimeout: 3,
  autovacuum: 0,
  syncCommit: true,
  fsyncTicks: 1,
  walWriterDelay: 4,
  checkpointEvery: 0,
  crashAt: 0,
  replayRate: 4,
  pageSlots: 4,
}

/** Статус транзакции в pg_xact (бывший CLOG). */
export type XactStatus = 'in-progress' | 'committed' | 'aborted'

/** Снимок: какие транзакции считать завершёнными. */
export interface Snapshot {
  /** Все xid меньше этого завершены. */
  xmin: number
  /** Все xid начиная с этого ещё не начались. */
  xmax: number
  /** Активные в момент снимка — их изменения не видны, даже когда они закоммитятся. */
  xip: number[]
  tick: number
}

/** Версия строки в куче. */
export interface Tuple {
  id: number
  key: string
  value: number
  /** Кто создал версию. */
  xmin: number
  /** Кто удалил или заменил её. null — версия последняя. */
  xmax: number | null
  /** xmax поставлен блокировкой FOR UPDATE, а не изменением. */
  lockOnly: boolean
  /** Следующая версия той же строки (t_ctid). */
  next: number | null
  page: number
  slot: number
  /** Номер записи WAL, которая создала версию, и которая поставила xmax. */
  lsnIn: number
  lsnOut: number | null
}

export type TxnState =
  /** Ещё не началась. */
  | 'idle'
  | 'active'
  /** Ждёт блокировку строки. */
  | 'waiting'
  /** COMMIT отправлен, ждёт сброса WAL. */
  | 'committing'
  | 'committed'
  | 'aborted'
  /** Сценарий сессии исчерпан. */
  | 'done'

export interface WaitInfo {
  /** Кого ждём — метка попытки (сессия:прогон). */
  on: string
  /** Его xid, если он уже есть. */
  xid: number | null
  key: string
  /** С какого тика стоим в очереди к строке — по нему решается, кто следующий. */
  since: number
  /** С какого тика ждём нынешнего держателя — от него отсчитывается deadlock_timeout. */
  timerFrom: number
  /** Строка свободна, но к ней уже стоит очередь, и мы в её хвосте. */
  behind: boolean
  /** Поиск дедлока уже был — PostgreSQL делает его один раз за ожидание. */
  checked: boolean
}

/** Что транзакция изменила — для проверки инварианта после её подтверждения. */
export interface Effect {
  key: string
  kind: 'set' | 'add' | 'insert' | 'delete'
  value: number
}

export interface Txn {
  /** Номер сессии в сценарии. */
  spec: number
  name: string
  /** Сквозной номер транзакции в сессии — для меток SSI. */
  run: number
  /** Какой по счёту круг сессии с repeat: W#1, W#2… */
  iter: number
  /** У сессии больше одного круга — тогда номер круга виден в подписи. */
  looped: boolean
  /** Сколько раз эту транзакцию уже повторяли после ошибки: T1′, T1″… */
  retries: number
  isolation: Isolation
  state: TxnState
  /** Номер транзакции. PostgreSQL выдаёт его только при первой записи. */
  xid: number | null
  opIdx: number
  /** На каком тике выполнится следующая операция. */
  nextAt: number
  /** Снимок транзакции (REPEATABLE READ, SERIALIZABLE) или последнего оператора (READ COMMITTED). */
  snapshot: Snapshot | null
  /** Держит ли снимок горизонт очистки прямо сейчас. */
  holdsSnapshot: boolean
  /** Что приложение прочитало: последнее значение по ключу. null — строки не было. */
  reads: Record<string, number | null>
  /** Все прочитанные значения по порядку — для поиска неповторяемого чтения. */
  seen: Record<string, (number | null)[]>
  scans: { count: number; sum: number; min: number }[]
  wait: WaitInfo | null
  /** SSI: кого мы не увидели (out) и кто не увидел нас (in). Номера сессий-попыток. */
  rwOut: string[]
  rwIn: string[]
  /** Ключи под SIREAD-блокировкой; '*' — вся таблица (последовательное чтение). */
  sireads: string[]
  effects: Effect[]
  /** Ключи, которые транзакция изменила, — для SSI и для «своих» записей. */
  wrote: string[]
  startTick: number
  endTick: number | null
  commitLsn: number | null
  /** Клиент получил «COMMIT» — неважно, пережил ли коммит падение. */
  acked: boolean
  ackTick: number | null
  /** Почему транзакция прервана: serialization, deadlock, unique, crash. */
  error: string | null
  /** Коммит подтверждён клиенту, но его запись в WAL не пережила падения. */
  lost: boolean
  repeatsLeft: number
  /** Когда сессия начнёт следующую транзакцию — повтор после ошибки или по расписанию. */
  restartAt: number | null
  /** Следующая транзакция — это повтор упавшей, а не очередной круг. */
  retryNext: boolean
}

export type TxnEventType =
  | 'txn.begin'
  | 'txn.xid'
  | 'snap.take'
  | 'row.read'
  | 'row.scan'
  | 'row.update'
  | 'row.insert'
  | 'row.delete'
  | 'row.lock'
  | 'row.recheck'
  | 'stmt.skip'
  | 'lock.wait'
  | 'lock.granted'
  | 'commit.wait'
  | 'txn.commit'
  | 'txn.rollback'
  | 'txn.abort'
  | 'txn.retry'
  | 'ssi.conflict'
  | 'deadlock.found'
  | 'anomaly.dirty'
  | 'anomaly.nonrepeatable'
  | 'anomaly.phantom'
  | 'anomaly.lost'
  | 'invariant.broken'
  | 'wal.flush'
  | 'checkpoint'
  | 'vacuum.run'
  | 'vacuum.blocked'
  | 'db.crash'
  | 'db.recovered'
  | 'commit.lost'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<TxnEventType, Importance> = {
  'txn.begin': 'low',
  'txn.xid': 'low',
  'snap.take': 'low',
  'row.read': 'normal',
  'row.scan': 'normal',
  'row.update': 'key',
  'row.insert': 'normal',
  'row.delete': 'normal',
  'row.lock': 'normal',
  'row.recheck': 'key',
  'stmt.skip': 'normal',
  'lock.wait': 'key',
  'lock.granted': 'normal',
  'commit.wait': 'normal',
  'txn.commit': 'key',
  'txn.rollback': 'normal',
  'txn.abort': 'key',
  'txn.retry': 'normal',
  'ssi.conflict': 'normal',
  'deadlock.found': 'key',
  'anomaly.dirty': 'key',
  'anomaly.nonrepeatable': 'key',
  'anomaly.phantom': 'key',
  'anomaly.lost': 'key',
  'invariant.broken': 'key',
  'wal.flush': 'low',
  checkpoint: 'normal',
  'vacuum.run': 'normal',
  'vacuum.blocked': 'key',
  'db.crash': 'key',
  'db.recovered': 'key',
  'commit.lost': 'key',
}

export interface TxnEvent {
  tick: number
  type: TxnEventType
  /** Участники — для подсветки: сессии и версии строк. */
  actors: { txn?: number[]; tuple?: number[]; key?: string[] }
  payload: Record<string, unknown>
}

/** Запись WAL. */
export interface WalRecord {
  lsn: number
  xid: number | null
  kind: 'insert' | 'update' | 'delete' | 'lock' | 'commit' | 'abort' | 'checkpoint'
  key?: string
  tick: number
}

export interface TxnScenario {
  id: string
  title: string
  /** Одна мысль, которую сценарий доказывает. Без неё сценарий не заводится. */
  claim: string
  config: Partial<TxnConfig>
  rows: RowSpec[]
  sessions: TxnSpec[]
  invariant?: Invariant
  /** Типы событий, на которых стенд автоматически ставит паузу. */
  watchFor: TxnEventType[]
  /** Не заканчивать прогон раньше этого тика — чтобы успел, например, пройти автовакуум. */
  minTicks?: number
  stopAfter?: number
}

export interface TxnStats {
  commits: number
  aborts: number
  rollbacks: number
  serializationFailures: number
  deadlocks: number
  retries: number
  /** Сколько раз и сколько тиков в сумме ждали блокировку. */
  waits: number
  waitTicks: number
  /** Сколько тиков коммиты ждали сброса WAL. */
  commitWaitTicks: number
  flushes: number
  anomalies: number
  vacuumRuns: number
  vacuumRemoved: number
  lostCommits: number
  /** Самая большая доля мёртвых версий в куче за прогон. */
  maxDead: number
}

/** Что сессия делала в этом тике — строка в расписании прогона. */
export type Mark =
  | { kind: 'op'; op: OpKind; key?: string; bad: boolean }
  | { kind: 'wait' }
  | { kind: 'commit-wait' }
  | { kind: 'abort' }
  | { kind: 'down' }

export interface TxnWorld {
  tick: number
  config: TxnConfig
  tuples: Tuple[]
  nextTupleId: number
  txns: Txn[]
  /** Сессии, которые уже отработали: их последние транзакции для истории. */
  history: Txn[]
  nextXid: number
  xact: Record<number, XactStatus>
  wal: WalRecord[]
  nextLsn: number
  /** Всё до этого LSN включительно уже на диске. */
  flushedLsn: number
  /** Сброс в процессе: докуда и когда закончится. */
  flushing: { upTo: number; doneAt: number } | null
  lastFlushTick: number
  checkpointLsn: number
  /** Сервер лежит или восстанавливается до этого тика. */
  downUntil: number
  crashed: boolean
  /** Ожидаемые значения по подтверждённым изменениям — для инварианта conserve. */
  expected: Record<string, number | null>
  invariantBroken: boolean
  /** Расписание: для каждой сессии — что было на этом тике. */
  marks: (Mark | null)[]
  stats: TxnStats
  finished: boolean
  finishReason?: 'all-done' | 'stop-after'
}

/** Кадр истории: мир после тика и события этого тика. */
export interface Frame {
  world: TxnWorld
  /** События, порождённые ИМЕННО этим тиком. */
  events: TxnEvent[]
}
