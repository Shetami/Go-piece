import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { SHARD_SCENARIOS, addShard, hotKey, ranges, scatter } from './scenarios.ts'
import type { ShardConfig, ShardScenario } from './types.ts'
import { imbalance, latencies, movedShare, ownerOf, percentile, resolveConfig } from './world.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: ShardScenario) => new Simulation(s).runToEnd()
const withConfig = (s: ShardScenario, config: Partial<ShardConfig>): ShardScenario => ({ ...s, config: { ...s.config, ...config } })
const pct = (sim: Simulation, p: number, kind?: 'point' | 'scatter') => percentile(latencies(sim.log.done, sim.tick, 0, kind), p)

describe('общие свойства движка', () => {
  test('прогон воспроизводим', () => {
    for (const sc of SHARD_SCENARIOS) assert.equal(JSON.stringify(run(sc).events), JSON.stringify(run(sc).events), sc.id)
  })

  test('у каждого ключа ровно один владелец, и счётчики шардов сходятся', () => {
    for (const sc of [...SHARD_SCENARIOS, withConfig(addShard, { scheme: 'ring' })]) {
      for (const f of run(sc).history) {
        const w = f.world
        assert.equal(w.owner.length, w.config.keys, sc.id)
        for (const o of w.owner) assert.ok(o >= 0 && o < w.shards.length, `${sc.id}, тик ${w.tick}`)
        const owns = w.shards.reduce((a, s) => a + s.owns, 0)
        assert.equal(owns, w.config.keys, `${sc.id}, тик ${w.tick}`)
      }
    }
  })

  test('веерный запрос уходит во все шарды и считается один раз', () => {
    const sim = run(scatter)
    const s = sim.world.stats
    assert.ok(s.scatter > 10)
    assert.ok(s.subRequests > s.requests, 'веерные запросы умножают нагрузку на шарды')
    assert.equal(sim.log.done.filter((c) => c.kind === 'scatter').length, s.scatter)
  })
})

describe('раскладка ключей', () => {
  test('кольцо переносит около доли нового шарда, остаток от деления — почти всё', () => {
    const c = resolveConfig(addShard.config)
    const mod = movedShare({ ...c, scheme: 'mod' }, 3, 4)
    const ring = movedShare({ ...c, scheme: 'ring' }, 3, 4)
    assert.ok(mod > 0.6, `mod ${mod}`)
    assert.ok(ring > 0.15 && ring < 0.35, `ring ${ring}`)
  })

  test('чем больше виртуальных узлов, тем ровнее раскладка', () => {
    const c = resolveConfig({ ...addShard.config, scheme: 'ring' })
    const spread = (vnodes: number) => {
      const counts = Array.from({ length: 4 }, () => 0)
      for (let k = 0; k < c.keys; k++) counts[ownerOf({ ...c, vnodes }, k, 4)]!++
      return Math.max(...counts) / (c.keys / 4)
    }
    assert.ok(spread(64) < spread(4), `64: ${spread(64)}, 4: ${spread(4)}`)
  })
})

describe('добавляем шард', () => {
  test('переезд идёт долго, и всё это время часть запросов дороже', () => {
    const sim = run(addShard)
    const start = sim.firstEvent('reshard.start')!
    const done = sim.firstEvent('reshard.done')!
    assert.ok(Number(start.payload.share) > 0.6, `переехало ${start.payload.share}`)
    assert.ok(done.tick - start.tick > 30, `переезд занял ${done.tick - start.tick}`)
    assert.ok(sim.world.stats.migratingHits > 5)
  })

  test('с кольцом переезжает меньше ключей и переезд заканчивается быстрее', () => {
    const ring = run(withConfig(addShard, { scheme: 'ring' }))
    const mod = run(addShard)
    assert.ok(ring.world.stats.moved * 2 < mod.world.stats.moved, `${ring.world.stats.moved} против ${mod.world.stats.moved}`)
    const took = (sim: Simulation) => sim.firstEvent('reshard.done')!.tick - sim.firstEvent('reshard.start')!.tick
    assert.ok(took(ring) < took(mod))
    assert.ok(ring.world.stats.migratingHits < mod.world.stats.migratingHits)
  })

  test('после переезда ключи распределены по четырём шардам', () => {
    const w = run(addShard).world
    assert.equal(w.shards.length, 4)
    for (const s of w.shards) assert.ok(s.owns > 0, `${s.name} без ключей`)
  })
})

describe('горячий ключ', () => {
  test('один ключ перегружает свой шард, пока соседние скучают', () => {
    const sim = run(hotKey)
    assert.ok(sim.firstEvent('shard.skew'))
    assert.ok(imbalance(sim.world) > 1.5, `перекос ${imbalance(sim.world)}`)
    assert.ok(sim.world.stats.maxQueue > 20)
  })

  test('без горячего ключа те же шарды загружены ровно', () => {
    const sim = run(withConfig(hotKey, { hotShare: 0 }))
    assert.ok(imbalance(sim.world) < 1.2, `перекос ${imbalance(sim.world)}`)
    assert.ok(pct(sim, 99) * 2 < pct(run(hotKey), 99))
  })

  test('смена схемы раскладки горячему ключу не помогает', () => {
    for (const scheme of ['mod', 'range'] as const) {
      const sim = run(withConfig(hotKey, { scheme }))
      assert.ok(imbalance(sim.world) > 1.4, `${scheme}: перекос ${imbalance(sim.world)}`)
    }
  })
})

describe('диапазоны', () => {
  test('горячая часть ключей собирается на одном шарде', () => {
    const sim = run(ranges)
    assert.ok(imbalance(sim.world) > 2, `перекос ${imbalance(sim.world)}`)
    assert.ok(sim.world.shards[0]!.served > sim.world.shards[3]!.served * 5)
  })

  test('хеш вместо диапазонов выравнивает нагрузку', () => {
    const ring = run(withConfig(ranges, { scheme: 'ring' }))
    assert.ok(imbalance(ring.world) < 1.5)
    assert.ok(pct(ring, 99) * 2 < pct(run(ranges), 99), `${pct(ring, 99)} против ${pct(run(ranges), 99)}`)
  })
})

describe('запрос без ключа шардирования', () => {
  test('веерный запрос ждёт самый медленный шард и потому медленнее точечного', () => {
    const sim = run(scatter)
    assert.ok(pct(sim, 50, 'scatter') > pct(sim, 50, 'point') * 2, `${pct(sim, 50, 'scatter')} против ${pct(sim, 50, 'point')}`)
  })

  test('при той же ёмкости больше шардов — хуже веерным запросам', () => {
    const two = run(withConfig(scatter, { shards: 2, workers: 4 }))
    const eight = run(withConfig(scatter, { shards: 8, workers: 1 }))
    assert.ok(eight.world.stats.subRequests > two.world.stats.subRequests * 1.5)
    assert.ok(pct(eight, 50, 'scatter') > pct(two, 50, 'scatter') * 2)
  })

  test('без веерных запросов те же шарды справляются спокойно', () => {
    const none = run(withConfig(scatter, { scatterShare: 0 }))
    assert.ok(none.world.stats.maxQueue * 2 < run(scatter).world.stats.maxQueue)
  })
})
