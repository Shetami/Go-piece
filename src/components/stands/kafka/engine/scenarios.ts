import type { KafkaScenario } from './types.ts'

/**
 * Семь пресетов. У каждого одна мысль, которую он доказывает, — поле claim.
 * Сценарий без claim заводить нельзя: получится красивая анимация, а не урок.
 */

export const journey: KafkaScenario = {
  id: 'journey',
  title: 'Путь одного сообщения',
  claim:
    'Между send() и обработкой сообщение проходит дюжину остановок, и продюсер получает «ок», только когда запись лежит на всех репликах из ISR.',
  config: { partitions: 3, acks: 'all', lingerTicks: 5, batchSize: 4 },
  producers: [{ name: 'orders', messages: 12, every: [2, 4], keys: ['alice', 'bob', 'carol', 'dave'] }],
  consumers: [
    { name: 'c1', group: 'billing' },
    { name: 'c2', group: 'billing' },
  ],
  faults: [],
  watchFor: ['batch.seal', 'produce.append', 'hw.advance', 'produce.ack', 'consumer.fetch', 'offset.commit'],
  focus: 1,
  stopAfter: 120,
}

export const batching: KafkaScenario = {
  id: 'batching',
  title: 'Пакеты и linger',
  claim:
    'Продюсер отправляет не сообщения, а пакеты на партицию. linger.ms покупает в разы меньше запросов ценой нескольких тиков задержки.',
  config: { partitions: 3, acks: 'all', lingerTicks: 0, batchSize: 4 },
  producers: [{ name: 'events', messages: 36, every: 1 }],
  consumers: [{ name: 'c1', group: 'stats' }],
  faults: [],
  watchFor: ['batch.seal', 'produce.request', 'produce.ack'],
  focus: 1,
  stopAfter: 120,
}

export const keys: KafkaScenario = {
  id: 'keys',
  title: 'Ключи и горячая партиция',
  claim:
    'Ключ определяет партицию, а партиция — одного потребителя в группе. Горячий ключ перегружает одну партицию, и лишние потребители этому не помогут.',
  config: { partitions: 4, acks: 'all', processTicks: 3, maxPollRecords: 4 },
  producers: [
    {
      name: 'clicks',
      messages: 40,
      every: 1,
      keys: ['alice', 'bob', 'carol', 'erin', 'frank', 'oscar'],
      weights: [14, 2, 2, 2, 2, 2],
    },
  ],
  consumers: [
    { name: 'c1', group: 'feed' },
    { name: 'c2', group: 'feed' },
    { name: 'c3', group: 'feed' },
    { name: 'c4', group: 'feed' },
  ],
  faults: [],
  watchFor: ['group.rebalance', 'consumer.fetch'],
  focus: 1,
  stopAfter: 200,
}

export const acks: KafkaScenario = {
  id: 'acks',
  title: 'acks=1 и падение лидера',
  claim:
    'acks=1 подтверждает запись, которая есть только на лидере. Если лидер падает раньше, чем фолловеры её скопировали, подтверждённое сообщение исчезает, и продюсер об этом не узнает.',
  config: { partitions: 1, acks: 1, lingerTicks: 0, batchSize: 1, replicaLagMax: 20 },
  producers: [{ name: 'payments', messages: 30, every: 1 }],
  consumers: [{ name: 'c1', group: 'ledger' }],
  faults: [
    // Реплики в другой зоне: копия доезжает до них в три раза дольше.
    { at: 1, kind: 'broker.slow', broker: 1, factor: 3 },
    { at: 1, kind: 'broker.slow', broker: 2, factor: 3 },
    { at: 14, kind: 'broker.down', broker: 0 },
    { at: 50, kind: 'broker.up', broker: 0 },
  ],
  watchFor: ['broker.down', 'leader.elect', 'rec.lost', 'log.truncate'],
  focus: 12,
  stopAfter: 140,
}

export const isr: KafkaScenario = {
  id: 'isr',
  title: 'ISR и min.insync.replicas',
  claim:
    'acks=all ждёт не все реплики, а реплики из ISR. Медленный фолловер тормозит каждую запись, пока его не вычеркнут из ISR, а min.insync.replicas решает, когда кластер перестаёт принимать записи вовсе.',
  config: { partitions: 1, acks: 'all', minInsyncReplicas: 2, lingerTicks: 2, batchSize: 2, replicaLagMax: 24, deliveryTimeout: 40 },
  producers: [{ name: 'orders', messages: 60, every: [1, 2] }],
  consumers: [{ name: 'c1', group: 'billing' }],
  faults: [
    { at: 6, kind: 'broker.slow', broker: 2, factor: 25 },
    { at: 45, kind: 'broker.down', broker: 1 },
    { at: 90, kind: 'broker.up', broker: 1 },
    { at: 92, kind: 'broker.slow', broker: 2, factor: 1 },
  ],
  watchFor: ['isr.shrink', 'produce.error', 'produce.failed', 'isr.expand'],
  focus: 29,
  stopAfter: 260,
}

export const group: KafkaScenario = {
  id: 'group',
  title: 'Группа и ребалансировка',
  claim:
    'Партиция в группе принадлежит ровно одному потребителю. Когда он падает, его партиции переходят к другим с последнего закоммиченного оффсета — и всё, что он успел обработать, но не закоммитил, обработают ещё раз.',
  config: { partitions: 4, acks: 'all', processTicks: 2, maxPollRecords: 4, sessionTimeout: 10 },
  producers: [{ name: 'orders', messages: 60, every: 1 }],
  consumers: [
    { name: 'c1', group: 'billing' },
    { name: 'c2', group: 'billing' },
    { name: 'c3', group: 'billing', joinAt: 24 },
    { name: 'audit', group: 'audit' },
  ],
  faults: [{ at: 44, kind: 'consumer.crash', consumer: 'c1' }],
  watchFor: ['group.rebalance', 'consumer.crash', 'rec.reprocess', 'rec.skipped'],
  focus: 1,
  stopAfter: 220,
}

export const retry: KafkaScenario = {
  id: 'retry',
  title: 'Потерянный ответ',
  claim:
    'Потерянный ответ неотличим от потерянной записи — продюсер повторяет запрос. Без идемпотентности в логе появляется дубль, с ней брокер узнаёт повтор по номеру последовательности.',
  config: { partitions: 1, acks: 'all', idempotence: false, lingerTicks: 2, batchSize: 3, requestTimeout: 10 },
  producers: [{ name: 'payments', messages: 15, every: [1, 2] }],
  consumers: [{ name: 'c1', group: 'ledger' }],
  faults: [{ at: 6, kind: 'drop.response', producer: 'payments' }],
  watchFor: ['resp.lost', 'produce.error', 'produce.duplicate', 'produce.dedup', 'rec.reprocess'],
  focus: 1,
  stopAfter: 120,
}

export const KAFKA_SCENARIOS: KafkaScenario[] = [journey, batching, keys, acks, isr, group, retry]

export function kafkaScenarioById(id: string): KafkaScenario | undefined {
  return KAFKA_SCENARIOS.find((s) => s.id === id)
}
