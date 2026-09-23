import type { ShardScenario } from './types.ts'

/**
 * Пресеты стенда. У каждого одна мысль, которую он доказывает, — поле claim.
 * Шард — 2 воркера, запрос в среднем 4 тика: ползапроса за тик. Три шарда
 * вместе успевают полтора, приходит 1,2 — запас есть, пока раскладка ровная.
 */

export const addShard: ShardScenario = {
  id: 'add-shard',
  title: 'Добавляем шард',
  claim:
    'Ключи разложены по остатку от деления хеша на число шардов. Добавляем четвёртый шард — и владельца меняют три четверти ключей. Пока они переезжают, запросы к ним стоят вдвое дороже.',
  config: { scheme: 'mod', shards: 3, keys: 120, zipf: 0.3, rate: 1.2, migrateRate: 2 },
  phases: [],
  faults: [{ kind: 'add', at: 60 }],
  watchFor: ['reshard.start', 'req.migrating', 'reshard.done'],
  stopAfter: 200,
}

export const hotKey: ShardScenario = {
  id: 'hot-key',
  title: 'Горячий ключ',
  claim:
    'Один ключ собирает 40% запросов. Он целиком лежит на одном шарде, и никакая раскладка этого не исправит: его шард перегружен, соседние скучают. Шардирование делит данные, а не нагрузку на один ключ.',
  config: { scheme: 'ring', shards: 3, keys: 120, rate: 1.2, hotShare: 40 },
  phases: [],
  faults: [],
  watchFor: ['shard.skew', 'queue.grow'],
  stopAfter: 160,
}

export const ranges: ShardScenario = {
  id: 'ranges',
  title: 'Диапазоны',
  claim:
    'Ключи разложены диапазонами: соседние лежат вместе. Это удобно для выборок по диапазону, но популярность обычно неравномерна — и горячая часть ключей оказывается на одном шарде.',
  config: { scheme: 'range', shards: 4, keys: 120, rate: 1.4, zipf: 1 },
  phases: [],
  faults: [],
  watchFor: ['shard.skew', 'queue.grow'],
  stopAfter: 160,
}

export const scatter: ShardScenario = {
  id: 'scatter',
  title: 'Запрос без ключа шардирования',
  claim:
    'Каждый пятый запрос не знает ключа шардирования и уходит во все шарды сразу. Он ждёт самый медленный ответ и умножает нагрузку на число шардов: чем больше шардов, тем хуже такому запросу.',
  config: { scheme: 'ring', shards: 4, keys: 120, rate: 1.2, scatterShare: 20 },
  phases: [],
  faults: [],
  watchFor: ['scatter.done', 'queue.grow'],
  stopAfter: 160,
}

export const SHARD_SCENARIOS: ShardScenario[] = [addShard, hotKey, ranges, scatter]

export function shardScenarioById(id: string): ShardScenario | undefined {
  return SHARD_SCENARIOS.find((s) => s.id === id)
}
