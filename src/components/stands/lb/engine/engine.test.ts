import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { LB_SCENARIOS, autoscale, blackHole, hang, slowReplica, twoChoices } from './scenarios.ts'
import type { LbConfig, LbScenario } from './types.ts'
import { okLatencies, percentile } from './world.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: LbScenario) => new Simulation(s).runToEnd()
const withConfig = (s: LbScenario, config: Partial<LbConfig>): LbScenario => ({ ...s, config: { ...s.config, ...config } })
const p99 = (sim: Simulation) => percentile(okLatencies(sim.log.done, sim.tick), 99)

describe('общие свойства движка', () => {
  test('прогон воспроизводим', () => {
    for (const sc of LB_SCENARIOS) assert.equal(JSON.stringify(run(sc).events), JSON.stringify(run(sc).events), sc.id)
  })

  test('незавершённые запросы балансировщика сходятся с тем, что лежит в репликах', () => {
    for (const sc of [...LB_SCENARIOS, withConfig(hang, { algo: 'least-conn' })]) {
      for (const f of run(sc).history) {
        for (const r of f.world.replicas) {
          const live = [...r.queue, ...r.workers].filter((q) => q && !q.abandoned).length
          assert.equal(r.outstanding, live, `${sc.id}, тик ${f.world.tick}, ${r.name}`)
        }
      }
    }
  })

  test('каждая попытка кончается ответом, ошибкой, таймаутом или ещё в пути', () => {
    for (const sc of LB_SCENARIOS) {
      const w = run(sc).world
      const s = w.stats
      const inside = w.replicas.reduce((a, r) => a + r.outstanding, 0)
      assert.equal(s.attempts, s.ok + s.errors + s.timeouts + s.noBackend + inside, sc.id)
    }
  })

  test('смена алгоритма не меняет, когда приходят запросы', () => {
    const a = run(slowReplica).log.series.map((s) => s.arrived)
    const b = run(withConfig(slowReplica, { algo: 'least-conn' })).log.series.map((s) => s.arrived)
    assert.deepEqual(a, b)
  })
})

describe('медленная реплика', () => {
  test('по кругу она получает свою четверть и тонет, наименьшее число соединений её разгружает', () => {
    const rr = run(slowReplica)
    const lc = run(withConfig(slowReplica, { algo: 'least-conn' }))
    const slow = (sim: Simulation) => sim.world.replicas[3]!
    assert.ok(rr.world.stats.maxQueue >= 20, `очередь ${rr.world.stats.maxQueue}`)
    assert.ok(lc.world.stats.maxQueue <= 3)
    assert.ok(slow(lc).routed * 1.5 < slow(rr).routed, `lc ${slow(lc).routed}, rr ${slow(rr).routed}`)
    assert.ok(p99(rr) > p99(lc) * 2, `p99 rr ${p99(rr)}, lc ${p99(lc)}`)
  })

  test('проверка здоровья медленную реплику не исключает', () => {
    assert.equal(run(slowReplica).countEvents('health.eject'), 0)
  })
})

describe('два случайных', () => {
  test('p2c почти как полное знание и заметно лучше случайного выбора', () => {
    const rnd = run(twoChoices)
    const p2c = run(withConfig(twoChoices, { algo: 'p2c' }))
    const lc = run(withConfig(twoChoices, { algo: 'least-conn' }))
    assert.ok(rnd.world.stats.maxQueue > p2c.world.stats.maxQueue * 2)
    assert.ok(p99(rnd) > p99(p2c))
    assert.ok(Math.abs(p99(p2c) - p99(lc)) <= 2, `p2c ${p99(p2c)}, lc ${p99(lc)}`)
  })
})

describe('чёрная дыра', () => {
  const after = (sim: Simulation) => sim.log.done.filter((c) => c.t > 20)
  const errShare = (sim: Simulation) => {
    const d = after(sim)
    return d.filter((c) => c.outcome === 'error').length / d.length
  }

  test('наименьшее число соединений отправляет в сломанную реплику больше половины трафика', () => {
    const lc = run(blackHole)
    const rr = run(withConfig(blackHole, { algo: 'round-robin' }))
    assert.ok(errShare(lc) > 0.5, `lc ${errShare(lc)}`)
    assert.ok(errShare(rr) < 0.35, `rr ${errShare(rr)}`)
    assert.equal(lc.countEvents('health.eject'), 0, '/healthz она проходит')
  })

  test('пассивное исключение по ошибкам срезает их в разы', () => {
    const lc = run(blackHole)
    const out = run(withConfig(blackHole, { outlier: true }))
    assert.ok(out.firstEvent('outlier.eject'))
    assert.ok(out.world.stats.errors * 3 < lc.world.stats.errors, `${out.world.stats.errors} против ${lc.world.stats.errors}`)
  })
})

describe('зависшая реплика', () => {
  test('таймауты на ней идут, пока проверки её не исключат, и прекращаются после', () => {
    const sim = run(hang)
    const eject = sim.firstEvent('health.eject')
    assert.ok(eject && eject.tick > 30)
    const timeouts = sim.events.filter((e) => e.type === 'req.timeout' && e.payload.replica === 'r3')
    assert.ok(timeouts.length >= 5)
    assert.ok(timeouts.every((e) => e.tick <= eject.tick + hang.config.timeout!))
  })

  test('повтор спасает запрос, но не задержку', () => {
    const withRetry = run(hang)
    const noRetry = run(withConfig(hang, { retries: 0 }))
    assert.ok(withRetry.world.stats.failed * 5 < noRetry.world.stats.failed, `${withRetry.world.stats.failed} против ${noRetry.world.stats.failed}`)
    assert.ok(p99(withRetry) > p99(noRetry))
  })

  test('наименьшее число соединений само обходит зависшую реплику', () => {
    const lc = run(withConfig(hang, { algo: 'least-conn' }))
    assert.ok(lc.world.stats.timeouts * 2 <= run(hang).world.stats.timeouts)
  })

  test('без проверок здоровья таймауты не прекращаются', () => {
    const sim = run(withConfig(hang, { healthEvery: 0 }))
    assert.ok(sim.world.stats.timeouts > run(hang).world.stats.timeouts)
    assert.ok(sim.events.some((e) => e.type === 'req.timeout' && e.tick > 100))
  })
})

describe('автомасштабирование', () => {
  test('реплики приходят через десяток тиков решения и ещё bootTime запуска', () => {
    const sim = run(autoscale)
    const decide = sim.firstEvent('scale.decide')
    const ready = sim.firstEvent('scale.ready')
    assert.ok(decide && decide.tick > autoscale.phases[0]!.at)
    assert.ok(ready && ready.tick === decide.tick + autoscale.config.bootTime!)
  })

  test('помогает, а быстрый запуск и запас по целевой загрузке — ещё больше', () => {
    const none = run(withConfig(autoscale, { autoscale: false }))
    const slow = run(autoscale)
    const fast = run(withConfig(autoscale, { bootTime: 5 }))
    const headroom = run(withConfig(autoscale, { scaleTarget: 50 }))
    assert.ok(p99(none) > p99(slow) && p99(slow) > p99(fast), `${p99(none)} > ${p99(slow)} > ${p99(fast)}`)
    assert.ok(headroom.world.stats.maxQueue < slow.world.stats.maxQueue)
  })
})
