import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { CACHE_SCENARIOS, coldStart, hotKeys, staleData, stampede, ttlSync } from './scenarios.ts'
import type { CacheConfig, CacheScenario } from './types.ts'
import { cachedCount, idealHitRatio, percentile, readLatencies } from './world.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: CacheScenario) => new Simulation(s).runToEnd()
const withConfig = (s: CacheScenario, config: Partial<CacheConfig>): CacheScenario => ({ ...s, config: { ...s.config, ...config } })
const hitRatio = (sim: Simulation) => sim.world.stats.hits / Math.max(1, sim.world.stats.reads)
const pct = (sim: Simulation, p: number, from = 0, to = Number.POSITIVE_INFINITY) => percentile(readLatencies(sim.log.done, Math.min(to, sim.tick), from), p)
const p99 = (sim: Simulation, from = 0, to = Number.POSITIVE_INFINITY) => pct(sim, 99, from, to)
/** Сколько запросов в базу ушло за один ключ. */
const dbReadsFor = (sim: Simulation, key: number) => sim.events.filter((e) => e.type === 'db.done' && e.payload.key === key).length

describe('общие свойства движка', () => {
  test('прогон воспроизводим', () => {
    for (const sc of CACHE_SCENARIOS) assert.equal(JSON.stringify(run(sc).events), JSON.stringify(run(sc).events), sc.id)
  })

  test('кэш не больше своего размера, а в нём только существующие ключи', () => {
    for (const sc of CACHE_SCENARIOS) {
      for (const f of run(sc).history) {
        assert.ok(cachedCount(f.world) <= f.world.config.cacheSize, `${sc.id}, тик ${f.world.tick}`)
        for (const k of Object.keys(f.world.cache)) assert.ok(Number(k) < f.world.config.keys, sc.id)
      }
    }
  })

  test('каждое чтение либо попало, либо промахнулось, либо получило просроченное', () => {
    for (const sc of [...CACHE_SCENARIOS, withConfig(stampede, { staleWhileRevalidate: true }), withConfig(stampede, { coalesce: true })]) {
      const s = run(sc).world.stats
      assert.equal(s.reads, s.hits + s.misses + s.swr, sc.id)
    }
  })

  test('кэш в модели никогда не опережает базу', () => {
    for (const sc of [staleData, withConfig(staleData, { invalidation: 'update' })]) {
      for (const f of run(sc).history) {
        for (const [k, e] of Object.entries(f.world.cache)) assert.ok(e.version <= f.world.db[Number(k)]!, `${sc.id}, ключ ${k}`)
      }
    }
  })
})

describe('горячие ключи', () => {
  test('кэш на десятую часть ключей забирает больше половины чтений', () => {
    const sim = run(hotKeys)
    assert.ok(hitRatio(sim) > 0.4, `попаданий ${hitRatio(sim)}`)
    assert.ok(idealHitRatio(sim.world.config) > 0.6)
    assert.ok(sim.world.stats.dbReads < sim.world.stats.reads * 0.6)
  })

  test('без кэша та же нагрузка топит базу', () => {
    const none = run(withConfig(hotKeys, { cacheSize: 0 }))
    const sim = run(hotKeys)
    assert.ok(none.world.stats.maxDbQueue > sim.world.stats.maxDbQueue * 10, `${none.world.stats.maxDbQueue} против ${sim.world.stats.maxDbQueue}`)
    assert.ok(p99(none) > p99(sim) * 3)
  })

  test('на равномерной популярности тот же кэш почти бесполезен', () => {
    const flat = run(withConfig(hotKeys, { zipf: 0 }))
    assert.ok(hitRatio(flat) < 0.15, `попаданий ${hitRatio(flat)}`)
  })

  test('вдвое больший кэш добавляет заметно меньше, чем первый', () => {
    const small = hitRatio(run(hotKeys))
    const big = hitRatio(run(withConfig(hotKeys, { cacheSize: 40 })))
    assert.ok(big > small && big - small < small, `${small} → ${big}`)
  })
})

describe('лавина на горячем ключе', () => {
  test('за одним истёкшим ключом в базу уходит пачка одинаковых запросов', () => {
    const sim = run(stampede)
    assert.ok(sim.world.stats.maxSameKey >= 4, `${sim.world.stats.maxSameKey}`)
    assert.ok(sim.firstEvent('db.stampede'))
    assert.ok(dbReadsFor(sim, 0) > 8, `запросов за горячий ключ ${dbReadsFor(sim, 0)}`)
  })

  test('объединение запросов оставляет ровно один запрос в базу за ключ', () => {
    const sim = run(withConfig(stampede, { coalesce: true }))
    assert.equal(sim.world.stats.maxSameKey, 1)
    assert.equal(sim.countEvents('db.stampede'), 0)
    assert.ok(sim.world.stats.joins > 20)
    assert.ok(dbReadsFor(sim, 0) < dbReadsFor(run(stampede), 0))
  })

  test('отдача просроченного значения убирает ожидание у пользователей', () => {
    const swr = run(withConfig(stampede, { staleWhileRevalidate: true }))
    assert.ok(swr.world.stats.swr > 30)
    assert.ok(p99(swr) < p99(run(stampede)), `${p99(swr)} против ${p99(run(stampede))}`)
  })
})

describe('холодный кэш', () => {
  const flushAt = coldStart.faults[1]!.at

  test('после сброса кэша нагрузка на базу подскакивает, а попадания восстанавливаются медленно', () => {
    const sim = run(coldStart)
    const hits = (a: number, b: number) => {
      const win = sim.log.series.slice(a + 1, b + 1)
      const reads = win.reduce((x, s) => x + s.hits + s.misses, 0)
      return win.reduce((x, s) => x + s.hits, 0) / Math.max(1, reads)
    }
    assert.ok(hits(20, flushAt) > 0.65, `до сброса попаданий ${hits(20, flushAt)}`)
    assert.ok(hits(flushAt, flushAt + 30) < 0.35, `сразу после ${hits(flushAt, flushAt + 30)}`)
    assert.ok(hits(flushAt + 80, sim.tick) < hits(20, flushAt), 'и через 80 тиков кэш ещё не тот, что был')
    assert.ok(pct(sim, 90, flushAt, flushAt + 30) > pct(sim, 90, 20, flushAt) * 2, 'задержка чтений выросла')
    assert.ok(sim.world.stats.maxDbQueue > 25, `очередь ${sim.world.stats.maxDbQueue}`)
    const drained = sim.events.find((e) => e.type === 'db.drained' && e.tick > flushAt)
    assert.ok(drained && drained.tick > flushAt + 60, `очередь разошлась на тике ${drained?.tick}`)
  })

  test('объединение одинаковых запросов заметно смягчает сброс', () => {
    const plain = run(coldStart)
    const coalesced = run(withConfig(coldStart, { coalesce: true }))
    assert.ok(coalesced.world.stats.maxDbQueue < plain.world.stats.maxDbQueue, `${coalesced.world.stats.maxDbQueue} против ${plain.world.stats.maxDbQueue}`)
    assert.ok(coalesced.world.stats.dbReads < plain.world.stats.dbReads)
  })
})

describe('устаревшие данные', () => {
  test('без инвалидации кэш отдаёт старое, пока запись не истечёт', () => {
    const sim = run(staleData)
    assert.ok(sim.world.stats.stale > 20, `устаревших чтений ${sim.world.stats.stale}`)
  })

  test('удаление ключа при записи почти всё лечит, но оставляет гонку', () => {
    const sim = run(withConfig(staleData, { invalidation: 'delete' }))
    assert.ok(sim.world.stats.stale * 10 < run(staleData).world.stats.stale)
    assert.ok(sim.firstEvent('cache.stale-set'), 'гонка cache-aside должна случиться хотя бы раз')
  })

  test('запись нового значения прямо в кэш убирает и устаревшие чтения, и лишние промахи', () => {
    const sim = run(withConfig(staleData, { invalidation: 'update' }))
    assert.equal(sim.world.stats.stale, 0)
    assert.ok(sim.world.stats.dbReads < run(withConfig(staleData, { invalidation: 'delete' })).world.stats.dbReads)
  })
})

describe('одновременное истечение', () => {
  test('прогретые в один тик ключи истекают разом и заваливают базу', () => {
    const sim = run(ttlSync)
    const ttl = ttlSync.config.ttl!
    const spike = sim.events.find((e) => e.type === 'db.queue' && e.tick > ttl)
    assert.ok(spike && spike.tick < ttl + 15, `всплеск на тике ${spike?.tick}`)
    assert.ok(sim.world.stats.maxDbQueue > 60, `очередь ${sim.world.stats.maxDbQueue}`)
  })

  test('разброс TTL, объединение запросов и отдача просроченного чинят это по-разному', () => {
    const plain = run(ttlSync)
    const jitter = run(withConfig(ttlSync, { ttlJitter: 50 }))
    const coalesce = run(withConfig(ttlSync, { coalesce: true }))
    const swr = run(withConfig(ttlSync, { staleWhileRevalidate: true }))
    assert.ok(jitter.world.stats.maxDbQueue * 3 < plain.world.stats.maxDbQueue, `${jitter.world.stats.maxDbQueue} против ${plain.world.stats.maxDbQueue}`)
    assert.ok(coalesce.world.stats.maxDbQueue < jitter.world.stats.maxDbQueue)
    assert.equal(p99(swr), 1, 'просроченное значение отдаётся мгновенно')
    assert.ok(swr.world.stats.misses === 0)
  })
})
