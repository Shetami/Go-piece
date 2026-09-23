import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { RESIL_SCENARIOS, breaker, bulkhead, cascade, degrade, rateLimit } from './scenarios.ts'
import type { ResilConfig, ResilScenario } from './types.ts'
import { latencies, percentile, successShare } from './world.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: ResilScenario) => new Simulation(s).runToEnd()
const withConfig = (s: ResilScenario, config: Partial<ResilConfig>): ResilScenario => ({ ...s, config: { ...s.config, ...config } })
const pct = (sim: Simulation, p: number, pick?: (c: { needsDep: boolean }) => boolean) => percentile(latencies(sim.log.done, sim.tick, 0, pick), p)
const local = (c: { needsDep: boolean }) => !c.needsDep

describe('общие свойства движка', () => {
  test('прогон воспроизводим', () => {
    for (const sc of RESIL_SCENARIOS) assert.equal(JSON.stringify(run(sc).events), JSON.stringify(run(sc).events), sc.id)
  })

  test('каждый запрос учтён ровно один раз', () => {
    for (const sc of [...RESIL_SCENARIOS, withConfig(cascade, { timeout: 8, breaker: true, fallback: true })]) {
      const sim = run(sc)
      const s = sim.world.stats
      const inside = Object.keys(sim.world.requests).length
      assert.equal(s.arrived, s.ok + s.degraded + s.errors + s.rejected + inside, sc.id)
      assert.equal(sim.log.done.length, s.ok + s.degraded + s.errors + s.rejected, sc.id)
    }
  })

  test('воркер A занят ровно тем запросом, который на нём числится', () => {
    for (const sc of RESIL_SCENARIOS) {
      for (const f of run(sc).history) {
        f.world.workers.forEach((id, i) => {
          if (id === null) return
          const r = f.world.requests[id]
          assert.ok(r, `${sc.id}: воркер ${i} держит несуществующий запрос`)
          assert.equal(r.worker, i, `${sc.id}, тик ${f.world.tick}`)
        })
      }
    }
  })
})

describe('каскад', () => {
  test('медленная зависимость занимает воркеры и топит даже независимые запросы', () => {
    const sim = run(cascade)
    assert.ok(sim.firstEvent('pool.saturated'))
    assert.ok(sim.world.stats.maxQueue > 50, `очередь ${sim.world.stats.maxQueue}`)
    assert.ok(pct(sim, 99, local) > 50, `p99 независимых ${pct(sim, 99, local)}`)
    assert.ok(sim.world.stats.waitTicks > sim.world.stats.busyTicks * 0.6, 'большую часть времени воркеры просто ждут')
  })

  test('таймаут возвращает сервису управление, а предохранитель — ещё и ёмкость', () => {
    const none = run(cascade)
    const to = run(withConfig(cascade, { timeout: 8 }))
    const cb = run(withConfig(cascade, { timeout: 8, breaker: true }))
    assert.ok(to.world.stats.maxQueue * 2 < none.world.stats.maxQueue, `${to.world.stats.maxQueue} против ${none.world.stats.maxQueue}`)
    assert.ok(pct(cb, 99, local) < pct(to, 99, local), `с предохранителем ${pct(cb, 99, local)}, без ${pct(to, 99, local)}`)
    assert.ok(cb.world.stats.depCalls < to.world.stats.depCalls / 1.5, 'в мёртвую зависимость почти не ходим')
  })

  test('переборка защищает независимые запросы, не трогая таймауты', () => {
    const sim = run(withConfig(cascade, { bulkhead: 2 }))
    assert.ok(sim.world.stats.blocked > 20)
    assert.ok(pct(sim, 99, local) < 15, `p99 независимых ${pct(sim, 99, local)}`)
    assert.ok(sim.world.stats.maxQueue < 10)
  })
})

describe('предохранитель', () => {
  test('без него каждый запрос честно ждёт таймаут', () => {
    const sim = run(breaker)
    assert.ok(sim.world.stats.depTimeouts > 40, `таймаутов ${sim.world.stats.depTimeouts}`)
    assert.ok(pct(sim, 50) > 20, `медиана ${pct(sim, 50)}`)
  })

  test('с ним сервис перестаёт ходить в мёртвую зависимость и проверяет её пробами', () => {
    const sim = run(withConfig(breaker, { breaker: true }))
    const open = sim.firstEvent('breaker.open')
    assert.ok(open && open.tick > breaker.faults[0]!.at)
    assert.ok(sim.countEvents('breaker.half') >= 2, 'пробы должны повторяться')
    assert.ok(sim.world.stats.depTimeouts * 3 < run(breaker).world.stats.depTimeouts)
    assert.ok(pct(sim, 50) <= 3, `медиана ${pct(sim, 50)}`)
  })

  test('когда зависимость оживает, предохранитель замыкается сам', () => {
    const sim = run(withConfig(breaker, { breaker: true }))
    const close = [...sim.events].reverse().find((e) => e.type === 'breaker.close')
    const recover = sim.firstEvent('dep.recover')!
    assert.ok(close && close.tick >= recover.tick, `замкнулся на тике ${close?.tick}, зависимость ожила на ${recover.tick}`)
    assert.equal(sim.world.breaker, 'closed')
  })
})

describe('переборки', () => {
  test('без переборки ожидание зависимости занимает весь пул', () => {
    const sim = run(bulkhead)
    assert.ok(sim.world.stats.waitTicks > sim.world.stats.busyTicks * 0.6)
    assert.ok(pct(sim, 99, local) > 50)
  })

  test('с переборкой независимые запросы отвечают как обычно', () => {
    const sim = run(withConfig(bulkhead, { bulkhead: 2 }))
    assert.ok(pct(sim, 50, local) <= 3, `медиана независимых ${pct(sim, 50, local)}`)
    assert.ok(pct(sim, 99, local) < 15)
    assert.equal(sim.world.stats.errors, 0, 'с заглушкой отказов нет')
    assert.ok(sim.world.stats.degraded > 20)
  })
})

describe('ограничитель частоты', () => {
  test('без него всплеск нагрузки растит очередь и задержку', () => {
    const sim = run(rateLimit)
    assert.ok(sim.world.stats.maxQueue > 80, `очередь ${sim.world.stats.maxQueue}`)
    assert.ok(pct(sim, 50) > 20, `медиана ${pct(sim, 50)}`)
  })

  test('с ним лишние получают отказ, а принятые обслуживаются быстро', () => {
    const sim = run(withConfig(rateLimit, { rateLimit: 1.2 }))
    assert.ok(sim.world.stats.rejected > 100)
    assert.ok(sim.world.stats.maxQueue < 10, `очередь ${sim.world.stats.maxQueue}`)
    assert.ok(pct(sim, 99) < 20, `p99 ${pct(sim, 99)}`)
    assert.ok(sim.world.stats.ok > run(rateLimit).world.stats.ok * 0.7, 'полезных ответов не сильно меньше')
  })
})

describe('деградация', () => {
  test('без заглушки половина запросов превращается в ошибки', () => {
    const sim = run(degrade)
    assert.ok(sim.world.stats.errors > 50)
    assert.ok(successShare(sim.world) < 0.75)
  })

  test('с заглушкой отвечают все, просто часть — урезанно', () => {
    const sim = run(withConfig(degrade, { fallback: true }))
    assert.equal(sim.world.stats.errors, 0)
    assert.equal(successShare(sim.world), 1)
    assert.ok(sim.world.stats.degraded > 50)
    assert.ok(pct(sim, 99) < 30)
  })
})
