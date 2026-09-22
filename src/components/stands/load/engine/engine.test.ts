import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { LOAD_SCENARIOS, burst, knee, little as littleSc, retryStorm, shedding, tail } from './scenarios.ts'
import type { LoadConfig, LoadScenario } from './types.ts'
import { little, mean, okLatencies, percentile } from './world.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: LoadScenario) => new Simulation(s).runToEnd()
const withConfig = (s: LoadScenario, config: Partial<LoadConfig>): LoadScenario => ({ ...s, config: { ...s.config, ...config } })
const okWaits = (sim: Simulation, pick: (slow: boolean) => boolean = () => true) =>
  sim.log.done.filter((c) => c.outcome === 'ok' && pick(c.slow)).map((c) => c.wait)
const lat = (sim: Simulation) => okLatencies(sim.log.done, sim.tick)
/** Сколько ответов пользователи получили на тиках (from, to]. */
const okBetween = (sim: Simulation, from: number, to: number) => sim.log.series.slice(from + 1, to + 1).reduce((a, s) => a + s.ok, 0)

describe('общие свойства движка', () => {
  test('прогон воспроизводим', () => {
    for (const sc of LOAD_SCENARIOS) assert.equal(JSON.stringify(run(sc).events), JSON.stringify(run(sc).events), sc.id)
  })

  test('ни один сценарий не падает и доходит до конца', () => {
    for (const sc of LOAD_SCENARIOS) {
      const sim = run(sc)
      assert.ok(sim.finished && sim.world.finishReason, sc.id)
      assert.equal(sim.log.series.length, sim.history.length, `${sc.id}: журнал и история разошлись`)
    }
  })

  test('каждая попытка учтена ровно один раз', () => {
    for (const sc of [...LOAD_SCENARIOS, withConfig(retryStorm, { cancelOnTimeout: true }), withConfig(shedding, { queueLimit: 8, timeout: 12, retries: 2 })]) {
      const w = run(sc).world
      const s = w.stats
      const inside = w.queue.length + w.workers.filter((r) => r !== null).length
      assert.equal(s.arrivals, s.admitted + s.rejected, `${sc.id}: приход = принято + отказано`)
      assert.equal(s.admitted, s.ok + s.wasted + s.cancelled + inside, `${sc.id}: принятое либо ушло, либо ещё внутри`)
    }
  })

  test('воркеров не больше, чем задано, и очередь не длиннее предела', () => {
    for (const sc of [...LOAD_SCENARIOS, withConfig(shedding, { queueLimit: 8 })]) {
      for (const f of run(sc).history) {
        assert.ok(f.world.workers.length === f.world.config.workers, sc.id)
        if (f.world.config.queueLimit > 0) assert.ok(f.world.queue.length <= f.world.config.queueLimit, `${sc.id}, тик ${f.world.tick}`)
      }
    }
  })

  test('смена доли медленных запросов не меняет, когда приходят запросы', () => {
    const a = run(tail).log.series.map((s) => s.arrived)
    const b = run(withConfig(tail, { slowShare: 0 })).log.series.map((s) => s.arrived)
    assert.deepEqual(a, b)
  })
})

describe('закон Литтла', () => {
  test('L = λ·W с точностью до нескольких процентов', () => {
    for (const sc of [littleSc, knee, tail, burst]) {
      const { L, lambda, W } = little(run(sc).world)
      assert.ok(Math.abs(L - lambda * W) / L < 0.05, `${sc.id}: L ${L.toFixed(2)}, λW ${(lambda * W).toFixed(2)}`)
    }
  })
})

describe('колено', () => {
  test('от 50% к 90% ожидание в очереди растёт больше чем в десять раз', () => {
    const hi = mean(okWaits(run(knee)))
    const lo = mean(okWaits(run(withConfig(knee, { rate: 0.5 }))))
    assert.ok(hi > lo * 10, `90%: ${hi.toFixed(2)}, 50%: ${lo.toFixed(2)}`)
    assert.ok(run(knee).world.stats.maxQueue >= 10)
  })

  test('без случайности очереди нет даже на 90%', () => {
    const sim = run(withConfig(knee, { arrivals: 'even', serviceDist: 'fixed' }))
    assert.equal(sim.world.stats.maxQueue, 0)
    assert.equal(percentile(lat(sim), 99), 4)
  })

  test('пятый воркер заметно укорачивает хвост', () => {
    const four = percentile(lat(run(knee)), 99)
    const five = percentile(lat(run(withConfig(knee, { workers: 5 }))), 99)
    assert.ok(five < four, `4 воркера: p99 ${four}, 5 воркеров: ${five}`)
  })
})

describe('медленный хвост', () => {
  test('медиана на месте, p99 вырастает в разы', () => {
    const slow = lat(run(tail))
    const fast = lat(run(withConfig(tail, { slowShare: 0 })))
    assert.ok(percentile(slow, 50) - percentile(fast, 50) <= 1, `p50 ${percentile(slow, 50)} против ${percentile(fast, 50)}`)
    assert.ok(percentile(slow, 99) >= percentile(fast, 99) * 3, `p99 ${percentile(slow, 99)} против ${percentile(fast, 99)}`)
  })

  test('обычные запросы ждут в очереди дольше, когда рядом медленные', () => {
    const withSlow = mean(okWaits(run(tail), (slow) => !slow))
    const without = mean(okWaits(run(withConfig(tail, { slowShare: 0 }))))
    assert.ok(withSlow > without * 5, `${withSlow.toFixed(2)} против ${without.toFixed(2)}`)
  })
})

describe('всплеск', () => {
  test('очередь рассасывается дольше, чем длился всплеск', () => {
    const sim = run(burst)
    const end = burst.phases[1]!.at
    const drained = sim.events.find((e) => e.type === 'queue.drained' && e.tick > end)
    assert.ok(drained, 'очередь так и не опустела')
    assert.ok(drained.tick - end > end - burst.phases[0]!.at, `опустела на тике ${drained.tick}`)
  })

  test('хуже всего задержка уже после всплеска', () => {
    const sim = run(burst)
    const end = burst.phases[1]!.at
    const during = percentile(okLatencies(sim.log.done, end, burst.phases[0]!.at), 99)
    const after = percentile(okLatencies(sim.log.done, end + 40, end), 99)
    assert.ok(after > during, `во время ${during}, после ${after}`)
  })
})

describe('шторм повторов', () => {
  test('всплеск кончился, а сервис работает впустую до конца прогона', () => {
    const sim = run(retryStorm)
    const collapse = sim.firstEvent('overload.collapse')
    assert.ok(collapse && collapse.tick > retryStorm.phases[1]!.at, 'отказ должен наступить после всплеска')
    assert.equal(sim.countEvents('overload.recover'), 0)
    assert.equal(okBetween(sim, 150, sim.tick), 0, 'последние 50 тиков пользователи не получают ответов')
    assert.ok(sim.world.stats.busyTicks / (sim.tick * 4) > 0.9, 'и всё это время воркеры заняты')
    assert.ok(sim.world.stats.arrivals > sim.world.stats.logical * 2.5, 'повторы втрое умножают нагрузку')
  })

  test('без повторов тот же всплеск проходит сам', () => {
    const sim = run(withConfig(retryStorm, { retries: 0 }))
    assert.ok(sim.firstEvent('overload.recover'))
    assert.ok(okBetween(sim, 150, sim.tick) > 30)
  })

  test('отмена по context спасает, пауза между повторами — нет', () => {
    const cancel = run(withConfig(retryStorm, { cancelOnTimeout: true }))
    assert.equal(cancel.countEvents('overload.collapse'), 0)
    assert.equal(cancel.world.stats.wasted, 0)
    assert.ok(okBetween(cancel, 150, cancel.tick) > 30)

    const backoff = run(withConfig(retryStorm, { backoff: true }))
    assert.ok(backoff.firstEvent('overload.collapse'))
    assert.equal(backoff.countEvents('overload.recover'), 0)
  })

  test('ограниченная очередь тоже не даёт сервису захлебнуться', () => {
    const sim = run(withConfig(retryStorm, { queueLimit: 8 }))
    assert.equal(sim.countEvents('overload.collapse'), 0)
    assert.ok(sim.world.stats.rejected > 0)
  })
})

describe('сброс нагрузки', () => {
  test('без предела очередь и задержка растут весь прогон', () => {
    const sim = run(shedding)
    const early = percentile(okLatencies(sim.log.done, 50), 50)
    const late = percentile(okLatencies(sim.log.done, sim.tick, sim.tick - 50), 50)
    assert.ok(late > early * 2, `медиана в начале ${early}, в конце ${late}`)
    assert.ok(sim.world.stats.maxQueue > 40)
  })

  test('с пределом лишним отказ, а принятые обслуживаются быстро', () => {
    const open = run(shedding)
    const limited = run(withConfig(shedding, { queueLimit: 8 }))
    assert.ok(limited.world.stats.rejected > 0)
    assert.equal(limited.world.stats.ok, open.world.stats.ok, 'ответов столько же — ёмкость та же')
    assert.ok(percentile(lat(limited), 99) * 2 <= percentile(lat(open), 99), `p99 ${percentile(lat(limited), 99)} против ${percentile(lat(open), 99)}`)
  })
})
