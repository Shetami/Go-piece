/**
 * Типы модели потоковой репликации PostgreSQL.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Устройство взято из PostgreSQL почти без
 * изменений: ведущий сервер пишет WAL, walsender отправляет его репликам,
 * реплика записывает, сбрасывает на диск и проигрывает его, а в ответ шлёт
 * позиции write/flush/replay — те самые, что видны в pg_stat_replication.
 * Коммит ждёт реплик по synchronous_commit и synchronous_standby_names,
 * очистка на ведущем может конфликтовать с запросами на реплике, слот
 * репликации держит WAL, пока реплика его не заберёт.
 *
 * А вот время — в тиках, запись WAL — это целая транзакция из одного UPDATE,
 * сеть доставляет всё за фиксированную задержку, а переключение на реплику
 * делает идеальный внешний менеджер вроде Patroni. Полный список упрощений —
 * в конце лекции о репликации.
 */

/** synchronous_commit. */
export type SyncLevel = 'off' | 'local' | 'remote_write' | 'on' | 'remote_apply'

export const SYNC_LEVELS: SyncLevel[] = ['off', 'local', 'remote_write', 'on', 'remote_apply']

/** synchronous_standby_names: никого, FIRST 1 (r1) или ANY 1 (все реплики). */
export type Standbys = 'none' | 'first-r1' | 'any-1'

export interface ReplConfig {
  replicas: number
  syncCommit: SyncLevel
  standbys: Standbys
  /** Задержка сети в одну сторону, если у реплики не задана своя. */
  netLatency: number
  /** Сколько записей WAL реплика проигрывает за тик. */
  replayRate: number
  /** Сколько записей ведущий досылает за тик реплике, которая догоняет. */
  catchupRate: number
  /** Автоматическое переключение на реплику при падении ведущего. */
  failover: boolean
  /** Через сколько тиков менеджер кластера замечает, что ведущий умер. */
  detectAfter: number
  /** Слоты репликации: ведущий хранит WAL, пока его не заберут все реплики. */
  slots: boolean
  /** wal_keep_size без слотов — сколько последних записей WAL хранить. */
  walKeep: number
  hotStandbyFeedback: boolean
  /** max_standby_streaming_delay; -1 — ждать запрос сколько угодно. */
  maxStandbyDelay: number
  /** Интервал автовакуума на ведущем; 0 — выключен. */
  vacuumEvery: number
}

export const DEFAULT_REPL_CONFIG: ReplConfig = {
  replicas: 2,
  syncCommit: 'on',
  standbys: 'none',
  netLatency: 1,
  replayRate: 4,
  catchupRate: 6,
  failover: true,
  detectAfter: 4,
  slots: true,
  walKeep: 16,
  hotStandbyFeedback: false,
  maxStandbyDelay: 4,
  vacuumEvery: 0,
}

export interface ReplicaSpec {
  name: string
  /** Своя задержка сети — реплика в другом дата-центре. */
  latency?: number
  replayRate?: number
}

/** Откуда читает клиент. */
export type ReadFrom = 'primary' | 'any' | number

export type ClientOp =
  /** UPDATE t SET v = v + 1 WHERE k = key; COMMIT — на ведущем. */
  | { kind: 'write'; key: string; after?: number }
  /** SELECT v FROM t WHERE k = key — там, куда скажут. */
  | { kind: 'read'; key: string; from: ReadFrom; after?: number }
  /** Долгий отчёт на реплике: снимок берётся в начале и держится ticks тиков. */
  | { kind: 'query'; key: string; from: number; ticks: number; after?: number }

export interface ClientSpec {
  name: string
  at: number
  ops: ClientOp[]
  repeat?: number | 'forever'
  /** Пауза между кругами. */
  every?: number
}

export type Fault =
  | { kind: 'primary-down'; at: number }
  /** Реплика недоступна: упала или пропала сеть до неё. */
  | { kind: 'replica-down'; replica: number; at: number; until?: number }
  /** Реплика проигрывает WAL медленнее — тяжёлый диск или долгие запросы. */
  | { kind: 'replica-slow'; replica: number; at: number; until: number; rate: number }

export type NodeRole = 'primary' | 'replica'

/** Запись WAL. Для простоты — целая транзакция: одно изменение и её коммит. */
export interface WalRecord {
  lsn: number
  kind: 'commit' | 'cleanup'
  key?: string
  value?: number
  /** Кто писал — клиент. */
  client?: number
  tick: number
  /** cleanup: удалены версии, заменённые не позже этого LSN. */
  maxSuperseded?: number
  removed?: number
}

export interface Query {
  client: number
  key: string
  /** LSN, до которого реплика проиграла WAL к началу запроса, — его снимок. */
  snapshot: number
  startTick: number
  endTick: number
}

export interface NodeState {
  id: number
  name: string
  role: NodeRole
  up: boolean
  /** Отстал так, что нужный ему WAL уже удалён: нужна новая копия. */
  broken: boolean
  latency: number
  replayRate: number
  /** Позиции: получено и записано, сброшено на диск, проиграно. */
  writeLsn: number
  flushLsn: number
  replayLsn: number
  /** Закоммиченные данные, видимые на этом узле. */
  kv: Record<string, { value: number; lsn: number }>
  queries: Query[]
  /** Проигрывание остановлено: конфликт с запросом, с этого тика. */
  conflictSince: number | null
  slowUntil: number
  slowRate: number
  /** Тик, когда узел стал ведущим. */
  promotedAt: number | null
}

/** Что ведущий знает о реплике из её ответов — строка pg_stat_replication. */
export interface ReplicaView {
  /** Докуда walsender уже отправил WAL этой реплике. */
  sent: number
  write: number
  flush: number
  replay: number
  /** xmin из hot_standby_feedback: LSN самого старого снимка на реплике. */
  feedback: number | null
  lastReply: number
}

/** Сообщение в сети. */
export type Message =
  | { kind: 'wal'; from: number; to: number; lsn: number; arriveAt: number }
  | { kind: 'reply'; from: number; to: number; write: number; flush: number; replay: number; feedback: number | null; arriveAt: number }

export type ClientState = 'idle' | 'running' | 'commit-wait' | 'reading' | 'done' | 'error-wait'

export interface Client {
  id: number
  name: string
  state: ClientState
  opIdx: number
  nextAt: number
  iter: number
  repeatsLeft: number
  /** Коммит, которого ждёт клиент. */
  waitLsn: number | null
  waitSince: number | null
  /** Какой ведущий принял этот коммит — если он умрёт, исход неизвестен. */
  waitNode: number | null
  /** Свои подтверждённые записи по ключу — для проверки «читаю своё». */
  ownAcked: Record<string, number>
  /** Самый свежий LSN, увиденный по ключу, — для монотонного чтения. */
  seen: Record<string, number>
  lastRead: { key: string; value: number | null; node: number; tick: number } | null
  /** Для круговой балансировки чтений по репликам. */
  rr: number
  query: Query | null
  acked: number
  errors: number
}

export type ReplEventType =
  | 'wal.write'
  | 'replica.receive'
  | 'replica.replay'
  | 'commit.wait'
  | 'commit.ack'
  | 'commit.hang'
  | 'read.ok'
  | 'read.stale'
  | 'read.own-stale'
  | 'read.backwards'
  | 'read.error'
  | 'write.error'
  | 'primary.down'
  | 'replica.down'
  | 'replica.up'
  | 'replica.slow'
  | 'failover.promote'
  | 'commit.lost'
  | 'commit.unknown'
  | 'wal.retained'
  | 'wal.removed'
  | 'replica.broken'
  | 'vacuum.run'
  | 'feedback.hold'
  | 'conflict.wait'
  | 'conflict.cancel'
  | 'query.start'
  | 'query.done'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<ReplEventType, Importance> = {
  'wal.write': 'low',
  'replica.receive': 'low',
  'replica.replay': 'low',
  'commit.wait': 'normal',
  'commit.ack': 'normal',
  'commit.hang': 'key',
  'read.ok': 'low',
  'read.stale': 'normal',
  'read.own-stale': 'key',
  'read.backwards': 'key',
  'read.error': 'normal',
  'write.error': 'key',
  'primary.down': 'key',
  'replica.down': 'key',
  'replica.up': 'key',
  'replica.slow': 'normal',
  'failover.promote': 'key',
  'commit.lost': 'key',
  'commit.unknown': 'key',
  'wal.retained': 'normal',
  'wal.removed': 'normal',
  'replica.broken': 'key',
  'vacuum.run': 'normal',
  'feedback.hold': 'normal',
  'conflict.wait': 'key',
  'conflict.cancel': 'key',
  'query.start': 'normal',
  'query.done': 'normal',
}

export interface ReplEvent {
  tick: number
  type: ReplEventType
  actors: { node?: number[]; client?: number[]; lsn?: number[] }
  payload: Record<string, unknown>
}

export interface ReplScenario {
  id: string
  title: string
  claim: string
  config: Partial<ReplConfig>
  replicas: ReplicaSpec[]
  keys: string[]
  clients: ClientSpec[]
  faults: Fault[]
  watchFor: ReplEventType[]
  minTicks?: number
  stopAfter?: number
}

export interface ReplStats {
  commits: number
  acked: number
  lost: number
  unknown: number
  writeErrors: number
  reads: number
  staleReads: number
  ownStale: number
  backwards: number
  readErrors: number
  /** Сколько тиков в сумме коммиты ждали подтверждения. */
  commitWaitTicks: number
  maxCommitWait: number
  cancels: number
  maxLag: number
  maxRetained: number
  /** Мёртвые версии на ведущем — их держит hot_standby_feedback или ещё не дошёл вакуум. */
  maxDead: number
  /** Тики, когда запись была невозможна: ведущего нет. */
  downTicks: number
}

/** Что клиент делал на этом тике — строка расписания. */
export type Mark =
  | { kind: 'write' }
  | { kind: 'wait' }
  | { kind: 'ack' }
  | { kind: 'read'; node: number; bad: boolean }
  | { kind: 'query'; node: number }
  | { kind: 'error' }

export interface ReplWorld {
  tick: number
  config: ReplConfig
  nodes: NodeState[]
  primary: number
  wal: WalRecord[]
  nextLsn: number
  /** Самая старая запись WAL, которая ещё хранится на ведущем. */
  walOldest: number
  /** Изменения, заменённые новыми и ещё не убранные вакуумом: LSN замены. */
  dead: number[]
  view: Record<number, ReplicaView>
  net: Message[]
  clients: Client[]
  /** Тик, когда умер ведущий, — пока нового не выбрали. */
  primaryDownSince: number | null
  /** Подтверждённые клиентам коммиты — LSN и кто. */
  ackedLsns: { lsn: number; client: number; key: string; lost: boolean }[]
  marks: (Mark | null)[]
  stats: ReplStats
  finished: boolean
  finishReason?: 'all-done' | 'stop-after'
}

export interface Frame {
  world: ReplWorld
  events: ReplEvent[]
}
