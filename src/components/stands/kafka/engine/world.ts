import type {
  KafkaConfig,
  KafkaEvent,
  KafkaEventType,
  KafkaScenario,
  KafkaWorld,
  Msg,
  Partition,
  Rec,
  StepKind,
} from './types.ts'
import { DEFAULT_KAFKA_CONFIG } from './types.ts'

export interface Ctx {
  world: KafkaWorld
  events: KafkaEvent[]
  scenario: KafkaScenario
}

/** Конфиг сценария поверх умолчаний — с поправкой на невозможные сочетания. */
export function resolveConfig(partial: Partial<KafkaConfig> | undefined): KafkaConfig {
  const c = { ...DEFAULT_KAFKA_CONFIG, ...(partial ?? {}) }
  c.brokers = clamp(c.brokers, 1, 5)
  c.partitions = clamp(c.partitions, 1, 6)
  c.replicationFactor = clamp(c.replicationFactor, 1, c.brokers)
  c.minInsyncReplicas = clamp(c.minInsyncReplicas, 1, 5)
  c.batchSize = Math.max(1, c.batchSize)
  c.maxInFlight = Math.max(1, c.maxInFlight)
  c.maxPollRecords = Math.max(1, c.maxPollRecords)
  c.netLatency = Math.max(1, c.netLatency)
  c.processTicks = Math.max(1, c.processTicks)
  return c
}

function clamp(v: number, lo: number, hi: number): number {
  return Math.max(lo, Math.min(hi, Math.round(v)))
}

/**
 * Раскладка реплик как у Kafka без стоек: партиция p живёт на брокерах
 * p, p+1, … по кругу, и первый из них — предпочтительный лидер. Так лидеры
 * равномерно размазаны по кластеру.
 */
export function createWorld(scenario: KafkaScenario): KafkaWorld {
  const config = resolveConfig(scenario.config)
  const partitions: Partition[] = Array.from({ length: config.partitions }, (_, id) => {
    const replicas = Array.from({ length: config.replicationFactor }, (_, i) => (id + i) % config.brokers)
    const leader = replicas[0]!
    return {
      id,
      replicas,
      leader,
      epoch: 0,
      isr: [...replicas],
      logs: Object.fromEntries(replicas.map((b) => [b, { broker: b, log: [], hw: 0 }])),
      hw: 0,
      followerLeo: Object.fromEntries(replicas.map((b) => [b, 0])),
      caughtUpAt: Object.fromEntries(replicas.map((b) => [b, 0])),
      lastFetchLeo: Object.fromEntries(replicas.map((b) => [b, 0])),
      lastFetchAt: Object.fromEntries(replicas.map((b) => [b, 0])),
      fetching: Object.fromEntries(replicas.map((b) => [b, null])),
      electAt: null,
      pending: [],
      stats: { appended: 0, duplicates: 0, truncated: 0 },
    }
  })

  const groups = [...new Set(scenario.consumers.map((c) => c.group))].map((name) => ({
    name,
    generation: 0,
    committed: {},
    rebalanceUntil: null,
    rebalanceReason: null,
    stats: { rebalances: 0, reprocessed: 0, skipped: 0 },
  }))

  return {
    tick: 0,
    config,
    brokers: Array.from({ length: config.brokers }, (_, id) => ({ id, alive: true, slow: 1, detectAt: null })),
    partitions,
    producers: scenario.producers.map((p, id) => ({
      id,
      name: p.name,
      spec: id,
      created: 0,
      nextSendAt: Math.max(1, p.startAt ?? 1),
      open: {},
      ready: [],
      inFlight: [],
      metadata: Object.fromEntries(partitions.map((pt) => [pt.id, pt.leader])),
      sticky: id % config.partitions,
      dropNext: 0,
      stats: { requests: 0, acked: 0, retries: 0, failed: 0 },
    })),
    consumers: scenario.consumers.map((c, id) => ({
      id,
      name: c.name,
      group: c.group,
      spec: id,
      state: 'waiting',
      assigned: [],
      positions: {},
      processedNext: {},
      buffer: [],
      processing: null,
      fetching: [],
      lastAutoCommit: 0,
      deadDetectAt: null,
      processedCount: 0,
    })),
    groups,
    recs: [],
    batches: [],
    net: [],
    nextMsgId: 1,
    stats: {
      produced: 0,
      acked: 0,
      committed: 0,
      lost: 0,
      failed: 0,
      duplicates: 0,
      dedups: 0,
      requests: 0,
      recsSent: 0,
      retries: 0,
      ackLatency: 0,
      ackCount: 0,
      e2eLatency: 0,
      e2eCount: 0,
      processed: 0,
      reprocessed: 0,
      skipped: 0,
      rebalances: 0,
      elections: 0,
    },
    finished: false,
  }
}

export function emit(
  ctx: Ctx,
  type: KafkaEventType,
  actors: KafkaEvent['actors'] = {},
  payload: Record<string, unknown> = {},
): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

/** Записать шаг в путь сообщения. */
export function trace(ctx: Ctx, rec: Rec, step: StepKind, data: Record<string, unknown> = {}): void {
  rec.trace.push({ tick: ctx.world.tick, step, data })
}

export function getRec(w: KafkaWorld, id: number): Rec {
  const r = w.recs[id - 1]
  if (!r) throw new Error(`нет сообщения m${id}`)
  return r
}

type DistributiveOmit<T, K extends PropertyKey> = T extends unknown ? Omit<T, K> : never

/** Отправить по сети. Доставка — через `latency` тиков. */
export function send(ctx: Ctx, msg: DistributiveOmit<Msg, 'id' | 'deliverAt'>, latency: number): void {
  const w = ctx.world
  w.net.push({ ...msg, id: w.nextMsgId++, deliverAt: w.tick + Math.max(1, latency) } as Msg)
}

/** Лог лидера партиции или null, если лидера сейчас нет. */
export function leaderLog(p: Partition) {
  return p.leader === null ? null : (p.logs[p.leader]?.log ?? null)
}

/** Log end offset — оффсет, который получит следующая запись. */
export function leo(p: Partition, broker: number): number {
  return p.logs[broker]?.log.length ?? 0
}

/** Общий префикс двух логов: до какого оффсета записи совпадают. */
export function commonPrefix(a: { rec: number; epoch: number }[], b: { rec: number; epoch: number }[]): number {
  const n = Math.min(a.length, b.length)
  for (let i = 0; i < n; i++) {
    if (a[i]!.rec !== b[i]!.rec || a[i]!.epoch !== b[i]!.epoch) return i
  }
  return n
}

/**
 * Номер партиции для ключа — ровно как в Java-клиенте Kafka:
 * `toPositive(murmur2(keyBytes)) % numPartitions`.
 * Совместимость здесь важна: сервисы на разных языках должны класть один ключ
 * в одну партицию, иначе порядок по ключу рассыпается.
 */
export function partitionForKey(key: string, partitions: number): number {
  return (murmur2(new TextEncoder().encode(key)) & 0x7fffffff) % partitions
}

/** org.apache.kafka.common.utils.Utils.murmur2 — побитово. */
export function murmur2(data: Uint8Array): number {
  const length = data.length
  const m = 0x5bd1e995
  const r = 24
  let h = (0x9747b28c ^ length) | 0
  const length4 = length >> 2
  for (let i = 0; i < length4; i++) {
    const i4 = i * 4
    let k = (data[i4]! & 0xff) + ((data[i4 + 1]! & 0xff) << 8) + ((data[i4 + 2]! & 0xff) << 16) + ((data[i4 + 3]! & 0xff) << 24)
    k = Math.imul(k, m)
    k ^= k >>> r
    k = Math.imul(k, m)
    h = Math.imul(h, m)
    h ^= k
  }
  // Хвост из 1–3 байт. В Java это switch с проваливанием: case 3 → 2 → 1.
  const tail = length & ~3
  const rest = length % 4
  if (rest >= 3) h ^= (data[tail + 2]! & 0xff) << 16
  if (rest >= 2) h ^= (data[tail + 1]! & 0xff) << 8
  if (rest >= 1) {
    h ^= data[tail]! & 0xff
    h = Math.imul(h, m)
  }
  h ^= h >>> 13
  h = Math.imul(h, m)
  h ^= h >>> 15
  return h | 0
}
