import type { SagaScenario } from './types.ts'

/**
 * Пресеты стенда. У каждого одна мысль, которую он доказывает, — поле claim.
 * Заказы приходят примерно раз в два тика, событие идёт до потребителя 2 тика,
 * обработка — около 3 тиков.
 */

export const dualWrite: SagaScenario = {
  id: 'dual-write',
  title: 'Две записи',
  claim:
    'Сервис пишет заказ в базу, а потом отправляет событие в брокер. Между этими двумя действиями нет ничего общего: брокер недоступен или сервис перезапустился — и заказ есть, а события о нём нет уже никогда.',
  config: { publish: 'dual-write' },
  faults: [
    { kind: 'broker-down', at: 40, until: 60 },
    { kind: 'crash', at: 80 },
  ],
  watchFor: ['broker.down', 'event.lost', 'svc.crash', 'saga.stuck'],
  stopAfter: 140,
}

export const outbox: SagaScenario = {
  id: 'outbox',
  title: 'Transactional outbox',
  claim:
    'Событие пишется в ту же транзакцию, что и заказ, а отправляет его отдельный процесс. Недоступность брокера и падение сервиса теперь ничего не теряют: строки лежат в базе и уедут позже.',
  config: { publish: 'outbox', relayEvery: 3 },
  faults: [
    { kind: 'broker-down', at: 40, until: 60 },
    { kind: 'crash', at: 80 },
  ],
  watchFor: ['broker.down', 'svc.crash', 'broker.up', 'relay.send'],
  stopAfter: 140,
}

export const duplicates: SagaScenario = {
  id: 'duplicates',
  title: 'Дубли',
  claim:
    'Доставка «хотя бы один раз» иногда приносит событие дважды. Потребитель, который об этом не думает, списывает деньги дважды. Ключ идемпотентности делает повтор безопасным.',
  config: { publish: 'outbox', idempotent: false },
  faults: [{ kind: 'duplicates', at: 20, until: 120 }],
  watchFor: ['event.duplicate', 'pay.double'],
  stopAfter: 140,
}

export const sagaFlow: SagaScenario = {
  id: 'saga',
  title: 'Сага',
  claim:
    'Деньги списаны, а доставка не удалась — и откатить общую транзакцию нельзя, её нет. Компенсация возвращает деньги отдельным шагом: заказ приходит к понятному концу, хоть и к неуспешному.',
  config: { publish: 'outbox', idempotent: true, shipFail: 50, retries: 1, retryAfter: 4, compensate: false },
  faults: [],
  watchFor: ['ship.fail', 'pay.giveup', 'saga.stuck', 'saga.compensate'],
  stopAfter: 160,
}

export const SAGA_SCENARIOS: SagaScenario[] = [dualWrite, outbox, duplicates, sagaFlow]

export function sagaScenarioById(id: string): SagaScenario | undefined {
  return SAGA_SCENARIOS.find((s) => s.id === id)
}
