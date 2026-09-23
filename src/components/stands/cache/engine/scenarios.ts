import type { CacheScenario } from './types.ts'

/**
 * Пресеты стенда. У каждого одна мысль, которую он доказывает, — поле claim.
 * База — 6 воркеров, запрос в среднем 4 тика: она успевает полтора запроса за тик.
 * Приходит два запроса за тик — без кэша базе не справиться.
 */

export const hotKeys: CacheScenario = {
  id: 'hot-keys',
  title: 'Горячие ключи',
  claim:
    'Двести ключей, популярность неравномерная, как почти всегда в жизни. Кэш на двадцать ключей — десятая часть данных — забирает больше половины чтений, и база, которая одна бы не справилась, работает спокойно.',
  config: { cacheSize: 20, seed: 2 },
  phases: [],
  faults: [],
  watchFor: ['db.queue'],
  stopAfter: 200,
}

export const stampede: CacheScenario = {
  id: 'stampede',
  title: 'Лавина на горячем ключе',
  claim:
    'Самый горячий ключ получает 40% запросов, и его запрос в базу тяжёлый. Каждый раз, когда запись истекает, все, кто спросил ключ, пока новое значение едет из базы, тоже идут в базу — за одним и тем же.',
  config: { cacheSize: 40, hotShare: 30, ttl: 30, hotTime: 12 },
  phases: [],
  faults: [],
  watchFor: ['cache.expired', 'db.stampede'],
  stopAfter: 160,
}

export const coldStart: CacheScenario = {
  id: 'cold-start',
  title: 'Холодный кэш',
  claim:
    'Кэш перезапустился и опустел. Все чтения разом пошли в базу, рассчитанную на треть потока. Очередь в базе растёт, кэш наполняется медленно — потому что медленно отвечает база.',
  config: { keys: 1000, cacheSize: 200, rate: 3 },
  phases: [],
  faults: [
    { kind: 'warm', at: 1, count: 200 },
    { kind: 'flush', at: 80 },
  ],
  watchFor: ['cache.flush', 'db.queue', 'db.drained'],
  stopAfter: 280,
}

export const staleData: CacheScenario = {
  id: 'stale',
  title: 'Устаревшие данные',
  claim:
    'Каждый десятый запрос — запись. Без инвалидации кэш отдаёт старое значение, пока запись не истечёт. Удаление ключа при записи почти всё лечит, но не совсем: чтение, начатое до записи, может положить старое значение в кэш уже после удаления.',
  config: { cacheSize: 40, writeShare: 10, ttl: 60, invalidation: 'none' },
  phases: [],
  faults: [],
  watchFor: ['req.stale', 'cache.stale-set'],
  stopAfter: 200,
}

export const ttlSync: CacheScenario = {
  id: 'ttl-sync',
  title: 'Одновременное истечение',
  claim:
    'Кэш прогрели при старте: полсотни ключей положили в один тик с одинаковым TTL. Ровно через TTL они истекают разом, и база получает всплеск промахов. Случайная добавка к TTL размазывает истечение во времени.',
  config: { keys: 50, zipf: 0.6, cacheSize: 50, ttl: 50, rate: 6 },
  phases: [],
  faults: [{ kind: 'warm', at: 1, count: 50 }],
  watchFor: ['cache.warm', 'cache.expired', 'db.queue'],
  stopAfter: 160,
}

export const CACHE_SCENARIOS: CacheScenario[] = [hotKeys, stampede, coldStart, staleData, ttlSync]

export function cacheScenarioById(id: string): CacheScenario | undefined {
  return CACHE_SCENARIOS.find((s) => s.id === id)
}
