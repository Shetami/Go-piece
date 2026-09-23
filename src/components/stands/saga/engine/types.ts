/**
 * Типы модели согласованности между сервисами.
 *
 * Как и остальные стенды, это МОДЕЛЬ. Сервис заказов принимает заказ, пишет
 * его в свою базу и должен рассказать о нём остальным через брокер. Дальше
 * событие обрабатывают платежи, а за ними — доставка. Каждый шаг может не
 * удаться, сервис может упасть между записью в базу и отправкой события,
 * брокер может быть недоступен, а событие — прийти дважды.
 *
 * Сравниваются два способа публикации: «две записи подряд» (сначала база,
 * потом брокер) и transactional outbox, где событие пишется в ту же
 * транзакцию, а отправляет его отдельный процесс. Плюс идемпотентность
 * потребителя и сага с компенсацией.
 *
 * Время — в тиках, сеть мгновенная. Полный список упрощений — в конце лекции.
 */

export type Publish = 'dual-write' | 'outbox'

export interface SagaConfig {
  /** Заказов за тик. */
  rate: number
  publish: Publish
  /** Как часто отправщик outbox забирает накопленное, тиков. */
  relayEvery: number
  /** Задержка доставки события, тиков. */
  brokerDelay: number
  /** Потребитель проверяет, не обрабатывал ли он это событие раньше. */
  idempotent: boolean
  /** Доля неудачных списаний, процентов: их потребитель повторяет. */
  payFail: number
  /** Доля неудачных отправок, процентов. */
  shipFail: number
  /** Через сколько тиков потребитель повторяет неудачную обработку. */
  retryAfter: number
  /** Сколько раз потребитель повторяет, прежде чем сдаться. */
  retries: number
  /** Сага: при неудачной доставке деньги возвращаются компенсацией. */
  compensate: boolean
  /** Сколько событий одновременно обрабатывает каждый потребитель. */
  workers: number
  /** Среднее время обработки события, тиков. */
  workTime: number
  seed: number
}

export const DEFAULT_SAGA_CONFIG: SagaConfig = {
  rate: 0.5,
  publish: 'dual-write',
  relayEvery: 3,
  brokerDelay: 2,
  idempotent: false,
  payFail: 0,
  shipFail: 0,
  retryAfter: 4,
  retries: 3,
  compensate: false,
  workers: 3,
  workTime: 2,
  seed: 1,
}

export type Fault =
  /** Сервис заказов падает: всё, что не успело уйти в брокер, пропадает. */
  | { kind: 'crash'; at: number }
  /** Брокер недоступен: опубликовать событие нельзя. */
  | { kind: 'broker-down'; at: number; until: number }
  /** Доставка дублирует события: потребитель получает их дважды. */
  | { kind: 'duplicates'; at: number; until: number }

/** Состояние заказа в базе сервиса заказов. */
export type OrderState = 'created' | 'paid' | 'shipped' | 'cancelled'

export interface Order {
  id: number
  createdAt: number
  state: OrderState
  /** Событие о заказе опубликовано в брокер. */
  published: boolean
  /** Сколько раз с заказа списали деньги: больше одного — это дубль. */
  charges: number
  /** Деньги возвращены компенсацией. */
  refunded: boolean
  /** Когда заказ дошёл до конечного состояния. */
  settledAt: number | null
}

/** Строка outbox: событие, записанное в ту же транзакцию, что и заказ. */
export interface OutboxRow {
  id: number
  order: number
  kind: EventKind
  writtenAt: number
  sentAt: number | null
}

export type EventKind = 'order.created' | 'order.paid' | 'payment.refund'

export interface Message {
  id: number
  kind: EventKind
  order: number
  /** Ключ идемпотентности: у повторной доставки он тот же. */
  key: string
  publishedAt: number
  arriveAt: number
  attempt: number
  duplicate: boolean
}

export interface Work {
  msg: Message
  startAt: number
  finishAt: number
  consumer: 'pay' | 'ship'
}

export interface Completion {
  t: number
  /** Сколько заказ шёл от создания до конечного состояния. */
  lat: number
  kind: 'shipped' | 'cancelled' | 'stuck'
}

export interface Sample {
  created: number
  shipped: number
  cancelled: number
  /** Заказов, застрявших между состояниями прямо сейчас. */
  inFlight: number
  /** Событий в брокере и в обработке. */
  inBroker: number
  outbox: number
  x: number
}

export type SagaEventType =
  | 'order.created'
  | 'outbox.write'
  | 'event.publish'
  | 'event.lost'
  | 'relay.send'
  | 'event.deliver'
  | 'event.duplicate'
  | 'pay.ok'
  | 'pay.double'
  | 'pay.fail'
  | 'pay.retry'
  | 'pay.giveup'
  | 'ship.ok'
  | 'ship.fail'
  | 'saga.compensate'
  | 'saga.stuck'
  | 'svc.crash'
  | 'broker.down'
  | 'broker.up'

export type Importance = 'key' | 'normal' | 'low'

export const EVENT_IMPORTANCE: Record<SagaEventType, Importance> = {
  'order.created': 'low',
  'outbox.write': 'low',
  'event.publish': 'low',
  'event.lost': 'key',
  'relay.send': 'low',
  'event.deliver': 'low',
  'event.duplicate': 'normal',
  'pay.ok': 'low',
  'pay.double': 'key',
  'pay.fail': 'normal',
  'pay.retry': 'normal',
  'pay.giveup': 'key',
  'ship.ok': 'low',
  'ship.fail': 'normal',
  'saga.compensate': 'key',
  'saga.stuck': 'key',
  'svc.crash': 'key',
  'broker.down': 'key',
  'broker.up': 'key',
}

export interface SagaEvent {
  tick: number
  type: SagaEventType
  actors: { order?: number[]; msg?: number[] }
  payload: Record<string, unknown>
}

export interface SagaScenario {
  id: string
  title: string
  claim: string
  config: Partial<SagaConfig>
  faults: Fault[]
  watchFor: SagaEventType[]
  stopAfter: number
}

export interface SagaStats {
  orders: number
  paid: number
  shipped: number
  cancelled: number
  /** События, пропавшие вместе с упавшим сервисом или недоступным брокером. */
  lost: number
  duplicates: number
  /** Списания сверх одного на заказ — настоящие потерянные деньги. */
  doubleCharges: number
  payFails: number
  retries: number
  giveups: number
  shipFails: number
  compensations: number
  /** Заказы, застрявшие в промежуточном состоянии навсегда. */
  stuck: number
  /** Сумма тиков, которые заказы провели в промежуточном состоянии. */
  inconsistentTicks: number
}

export type Stream = 'arrive' | 'fail' | 'work'

export interface SagaWorld {
  tick: number
  config: SagaConfig
  rng: Record<Stream, number>
  orders: Order[]
  outbox: OutboxRow[]
  /** Сообщения в пути до потребителя. */
  broker: Message[]
  /** Обработка у потребителей. */
  work: Work[]
  /** Ключи идемпотентности, уже обработанные потребителями. */
  seen: string[]
  /** Отложенные повторы обработки. */
  retryQ: { msg: Message; at: number }[]
  brokerUp: boolean
  duplicating: boolean
  nextId: number
  nextMsg: number
  x: number
  stats: SagaStats
  finished: boolean
  finishReason?: 'stop-after'
}

export interface Frame {
  world: SagaWorld
  events: SagaEvent[]
}

/** Журнал прогона: только дописывается и в снимки не копируется. */
export interface SagaLog {
  series: Sample[]
  done: Completion[]
}
