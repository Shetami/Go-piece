/**
 * Типы модели Kafka.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Устройство взято из Kafka почти без
 * изменений: лог партиции с оффсетами, лидер и фолловеры, ISR, high watermark,
 * аккумулятор продюсера с пакетами, acks, идемпотентность по номеру
 * последовательности, группы потребителей с коммитом оффсетов.
 *
 * А вот время измеряется в тиках, сеть доставляет всё за фиксированную
 * задержку, запросы идут по одному пакету, а контроллер и координатор группы
 * никогда не падают. Полный список упрощений — в конце лекции.
 */

/** Длительность в тиках: точное число или диапазон [min, max]. */
export type Duration = number | [number, number]

/** Сколько подтверждений ждёт продюсер: ни одного, только лидера, все реплики из ISR. */
export type Acks = 0 | 1 | 'all'

/**
 * Когда потребитель фиксирует оффсет.
 * auto — по таймеру коммитится всё, что poll уже отдал приложению.
 * after-process — после того как пачка обработана целиком.
 */
export type CommitMode = 'auto' | 'after-process'

export interface KafkaConfig {
  /** Кластер. */
  brokers: number
  partitions: number
  replicationFactor: number
  /** min.insync.replicas: при acks=all меньше реплик в ISR — запись отклоняется. */
  minInsyncReplicas: number
  /** unclean.leader.election.enable: можно ли выбрать лидером реплику не из ISR. */
  uncleanElection: boolean
  /** replica.lag.time.max.ms: сколько тиков фолловер может отставать, не вылетая из ISR. */
  replicaLagMax: number
  /** Сколько тиков контроллер замечает смерть брокера. */
  electionDelay: number

  /** Продюсер. */
  acks: Acks
  /** linger.ms: сколько пакет может ждать попутчиков. */
  lingerTicks: number
  /** batch.size — в модели в записях, а не в байтах. */
  batchSize: number
  /** enable.idempotence. */
  idempotence: boolean
  /** max.in.flight.requests.per.connection. */
  maxInFlight: number
  /** request.timeout.ms. */
  requestTimeout: number
  /** delivery.timeout.ms: после этого пакет считается неудачным. */
  deliveryTimeout: number

  /** Сеть: задержка в одну сторону. */
  netLatency: number

  /** Потребители. */
  commitMode: CommitMode
  /** auto.commit.interval.ms. */
  autoCommitInterval: number
  /** Сколько тиков приложение обрабатывает одно сообщение. */
  processTicks: number
  /** max.poll.records. */
  maxPollRecords: number
  /** session.timeout.ms: сколько группа ждёт упавшего, прежде чем отдать его партиции. */
  sessionTimeout: number
  /** Сколько тиков группа стоит на ребалансировке. */
  rebalanceTicks: number
}

export const DEFAULT_KAFKA_CONFIG: KafkaConfig = {
  brokers: 3,
  partitions: 3,
  replicationFactor: 3,
  minInsyncReplicas: 2,
  uncleanElection: false,
  replicaLagMax: 12,
  electionDelay: 4,

  acks: 'all',
  lingerTicks: 5,
  batchSize: 4,
  idempotence: true,
  maxInFlight: 5,
  requestTimeout: 15,
  deliveryTimeout: 60,

  netLatency: 1,

  commitMode: 'after-process',
  autoCommitInterval: 10,
  processTicks: 2,
  maxPollRecords: 4,
  sessionTimeout: 10,
  rebalanceTicks: 3,
}

/** Продюсер в сценарии. */
export interface ProducerSpec {
  name: string
  /** Сколько сообщений отправит за прогон. */
  messages: number
  /** Пауза между вызовами send(). */
  every: Duration
  startAt?: number
  /** Ключи сообщений. Пусто — сообщения без ключа. */
  keys?: string[]
  /** Веса ключей: так делается «горячий» ключ. */
  weights?: number[]
}

/** Потребитель в сценарии. Все с одинаковым group — одна группа. */
export interface ConsumerSpec {
  name: string
  group: string
  /** Тик, на котором потребитель подключается. По умолчанию — сразу. */
  joinAt?: number
}

/** Что сломать и когда. Брокеры — по номеру, клиенты — по имени. */
export type Fault =
  | { at: number; kind: 'broker.down'; broker: number }
  | { at: number; kind: 'broker.up'; broker: number }
  /** Фолловер на этом брокере начинает реплицировать в `factor` раз медленнее. */
  | { at: number; kind: 'broker.slow'; broker: number; factor: number }
  | { at: number; kind: 'consumer.crash'; consumer: string }
  /** Сеть теряет следующий ответ на запрос записи этому продюсеру. */
  | { at: number; kind: 'drop.response'; producer: string }

/* ─────────────────────────────── мир ─────────────────────────────── */

/** Запись в логе реплики. Оффсет — это индекс в массиве. */
export interface LogEntry {
  rec: number
  /** Эпоха лидера, при которой запись попала в лог. */
  epoch: number
}

export interface Replica {
  broker: number
  log: LogEntry[]
  /** High watermark, каким его знает эта реплика. */
  hw: number
}

/** Запрос на запись, ждущий репликации (acks=all) — в Kafka это purgatory. */
export interface PendingAck {
  producer: number
  batch: number
  req: number
  lastOffset: number
  baseOffset: number
  dedup: boolean
}

export interface Partition {
  id: number
  /** Брокеры с репликами. Первый — предпочтительный лидер. */
  replicas: number[]
  leader: number | null
  epoch: number
  isr: number[]
  /** Реплики по номеру брокера. */
  logs: Record<number, Replica>
  /** HW лидера — граница того, что видят потребители. */
  hw: number
  /** Что лидер знает о фолловерах: докуда у них есть лог и когда они последний раз догоняли. */
  followerLeo: Record<number, number>
  caughtUpAt: Record<number, number>
  /** LEO лидера и тик на момент прошлого запроса фолловера — по ним Kafka решает, догнал ли он. */
  lastFetchLeo: Record<number, number>
  lastFetchAt: Record<number, number>
  /** Куда ушёл запрос репликации от фолловера (номер лидера) — или его нет. */
  fetching: Record<number, number | null>
  /** На каком тике контроллер выберет нового лидера. */
  electAt: number | null
  pending: PendingAck[]
  stats: { appended: number; duplicates: number; truncated: number }
}

export interface Broker {
  id: number
  alive: boolean
  /** Во сколько раз медленнее реплицирует. 1 — нормально. */
  slow: number
  /** Тик, на котором контроллер заметит смерть брокера. */
  detectAt: number | null
}

export type BatchState = 'open' | 'ready' | 'in-flight' | 'done' | 'failed'

export interface Batch {
  id: number
  producer: number
  partition: number
  recs: number[]
  createdTick: number
  sealedTick: number | null
  sealReason: 'size' | 'linger' | null
  attempts: number
  retryAt: number
  state: BatchState
}

export interface InFlight {
  batch: number
  broker: number
  req: number
  sentTick: number
}

export interface Producer {
  id: number
  name: string
  spec: number
  created: number
  nextSendAt: number
  /** Открытый пакет на каждую партицию. */
  open: Record<number, number>
  /** Закрытые пакеты в порядке отправки. */
  ready: number[]
  inFlight: InFlight[]
  /** Кто лидер партиции — как это знает продюсер. Может устареть. */
  metadata: Record<number, number | null>
  /** Партиция, куда липнут сообщения без ключа. */
  sticky: number
  /** Сколько следующих ответов потеряет сеть. */
  dropNext: number
  stats: { requests: number; acked: number; retries: number; failed: number }
}

export type RecState = 'batched' | 'in-flight' | 'acked' | 'failed' | 'lost'

/** Шаги пути сообщения — из них собирается разбор на стенде. */
export type StepKind =
  | 'send'
  | 'partition'
  | 'batch'
  | 'seal'
  | 'request'
  | 'append'
  | 'replicate'
  | 'commit'
  | 'ack'
  | 'fetch'
  | 'process'
  | 'offset'
  | 'retry'
  | 'resp-lost'
  | 'dedup'
  | 'duplicate'
  | 'truncated'
  | 'lost'
  | 'failed'
  | 'reprocess'
  | 'skipped'

export interface TraceStep {
  tick: number
  step: StepKind
  data: Record<string, unknown>
}

/** Сообщение. В модели значение не хранится — только ключ и путь. */
export interface Rec {
  id: number
  producer: number
  key: string | null
  partition: number
  batch: number
  createdTick: number
  state: RecState
  /** Хоть раз попадало в лог лидера. */
  appended: boolean
  /** Оффсет первой копии. */
  offset: number | null
  committedTick: number | null
  ackedTick: number | null
  /** Сколько раз обработано — по группам. */
  processed: Record<string, number>
  /** Группы, которые это сообщение пропустили навсегда. */
  skipped: string[]
  trace: TraceStep[]
}

export type ConsumerState = 'waiting' | 'active' | 'dead'

export interface Fetched {
  partition: number
  offset: number
  rec: number
}

export interface Consumer {
  id: number
  name: string
  group: string
  spec: number
  state: ConsumerState
  assigned: number[]
  /** Следующий оффсет, который попросим у брокера. */
  positions: Record<number, number>
  /** Следующий после последнего обработанного. */
  processedNext: Record<number, number>
  /** Отданное poll'ом, но ещё не обработанное. */
  buffer: Fetched[]
  processing: (Fetched & { left: number }) | null
  /** Партиции с запросом fetch в пути. */
  fetching: number[]
  lastAutoCommit: number
  /** Когда группа заметит, что потребитель умер. */
  deadDetectAt: number | null
  processedCount: number
}

export interface Group {
  name: string
  generation: number
  /** Закоммиченные оффсеты — в Kafka это топик __consumer_offsets. */
  committed: Record<number, number>
  rebalanceUntil: number | null
  rebalanceReason: string | null
  stats: { rebalances: number; reprocessed: number; skipped: number }
}

export type Msg =
  | { id: number; deliverAt: number; kind: 'produce'; producer: number; broker: number; batch: number; partition: number }
  | {
      id: number
      deliverAt: number
      kind: 'produce-resp'
      producer: number
      broker: number
      batch: number
      partition: number
      req: number
      error: string | null
      baseOffset: number | null
    }
  | { id: number; deliverAt: number; kind: 'repl-fetch'; from: number; broker: number; partition: number; fetchOffset: number; epoch: number }
  | {
      id: number
      deliverAt: number
      kind: 'repl-resp'
      from: number
      broker: number
      partition: number
      fetchOffset: number
      entries: LogEntry[]
      hw: number
      epoch: number
    }
  | { id: number; deliverAt: number; kind: 'fetch'; consumer: number; broker: number; partition: number; offset: number; generation: number }
  | {
      id: number
      deliverAt: number
      kind: 'fetch-resp'
      consumer: number
      broker: number
      partition: number
      offset: number
      entries: Fetched[]
      error: string | null
      generation: number
    }
  | { id: number; deliverAt: number; kind: 'commit'; consumer: number; group: string; offsets: Record<number, number>; mode: CommitMode }

export type KafkaEventType =
  | 'rec.send'
  | 'batch.seal'
  | 'produce.request'
  | 'produce.append'
  | 'produce.ack'
  | 'produce.error'
  | 'produce.failed'
  | 'produce.dedup'
  | 'produce.duplicate'
  | 'resp.lost'
  | 'repl.fetch'
  | 'hw.advance'
  | 'isr.shrink'
  | 'isr.expand'
  | 'broker.down'
  | 'broker.up'
  | 'broker.slow'
  | 'leader.elect'
  | 'log.truncate'
  | 'rec.lost'
  | 'consumer.join'
  | 'consumer.crash'
  | 'group.rebalance'
  | 'consumer.fetch'
  | 'consumer.process'
  | 'offset.commit'
  | 'rec.reprocess'
  | 'rec.skipped'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<KafkaEventType, Importance> = {
  'rec.send': 'low',
  'batch.seal': 'normal',
  'produce.request': 'normal',
  'produce.append': 'normal',
  'produce.ack': 'normal',
  'produce.error': 'key',
  'produce.failed': 'key',
  'produce.dedup': 'key',
  'produce.duplicate': 'key',
  'resp.lost': 'key',
  'repl.fetch': 'low',
  'hw.advance': 'normal',
  'isr.shrink': 'key',
  'isr.expand': 'key',
  'broker.down': 'key',
  'broker.up': 'key',
  'broker.slow': 'normal',
  'leader.elect': 'key',
  'log.truncate': 'key',
  'rec.lost': 'key',
  'consumer.join': 'normal',
  'consumer.crash': 'key',
  'group.rebalance': 'key',
  'consumer.fetch': 'low',
  'consumer.process': 'low',
  'offset.commit': 'normal',
  'rec.reprocess': 'key',
  'rec.skipped': 'key',
}

export interface KafkaEvent {
  tick: number
  type: KafkaEventType
  /** Участники — для подсветки на схеме. */
  actors: {
    rec?: number[]
    broker?: number[]
    partition?: number[]
    producer?: number[]
    consumer?: number[]
  }
  payload: Record<string, unknown>
}

export interface KafkaScenario {
  id: string
  title: string
  /** Одна мысль, которую сценарий доказывает. Без неё сценарий не заводится. */
  claim: string
  config: Partial<KafkaConfig>
  producers: ProducerSpec[]
  consumers: ConsumerSpec[]
  faults: Fault[]
  /** Типы событий, на которых стенд автоматически ставит паузу. */
  watchFor: KafkaEventType[]
  /** Сообщение, чей путь стенд показывает сразу. */
  focus?: number
  stopAfter?: number
  seed?: number
}

export interface KafkaStats {
  produced: number
  acked: number
  committed: number
  lost: number
  failed: number
  /** Лишние копии в логе. */
  duplicates: number
  dedups: number
  requests: number
  /** Записей, ушедших в запросах, — для среднего размера пакета. */
  recsSent: number
  retries: number
  ackLatency: number
  ackCount: number
  /** От send() до первой обработки. */
  e2eLatency: number
  e2eCount: number
  processed: number
  reprocessed: number
  skipped: number
  rebalances: number
  elections: number
}

export interface KafkaWorld {
  tick: number
  config: KafkaConfig
  brokers: Broker[]
  partitions: Partition[]
  producers: Producer[]
  consumers: Consumer[]
  groups: Group[]
  recs: Rec[]
  batches: Batch[]
  net: Msg[]
  nextMsgId: number
  stats: KafkaStats
  finished: boolean
  finishReason?: 'all-done' | 'stop-after'
}

export interface Snapshot {
  world: KafkaWorld
  /** События, порождённые ИМЕННО этим тиком. */
  events: KafkaEvent[]
}
