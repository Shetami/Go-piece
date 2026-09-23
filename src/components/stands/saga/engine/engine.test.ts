import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { SAGA_SCENARIOS, dualWrite, duplicates, outbox, sagaFlow } from './scenarios.ts'
import type { SagaConfig, SagaScenario } from './types.ts'
import { consistency, outcome } from './world.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: SagaScenario) => new Simulation(s).runToEnd()
const withConfig = (s: SagaScenario, config: Partial<SagaConfig>): SagaScenario => ({ ...s, config: { ...s.config, ...config } })

describe('общие свойства движка', () => {
  test('прогон воспроизводим', () => {
    for (const sc of SAGA_SCENARIOS) assert.equal(JSON.stringify(run(sc).events), JSON.stringify(run(sc).events), sc.id)
  })

  test('состояния заказов меняются только вперёд и остаются согласованными', () => {
    for (const sc of SAGA_SCENARIOS) {
      const sim = run(sc)
      for (const o of sim.world.orders) {
        if (o.state === 'shipped' || o.state === 'paid') assert.ok(o.charges >= 1, `${sc.id}: заказ ${o.id} без списания`)
        if (o.state === 'created') assert.equal(o.charges, 0, `${sc.id}: заказ ${o.id} оплачен, но числится созданным`)
        if (o.refunded) assert.equal(o.state, 'cancelled', `${sc.id}: возврат без отмены`)
      }
      const s = sim.world.stats
      assert.equal(s.orders, sim.world.orders.length, sc.id)
      assert.equal(s.shipped + s.cancelled, outcome(sim.world).settled, sc.id)
    }
  })

  test('ключ идемпотентности у дубля тот же, что у оригинала', () => {
    const sim = run(duplicates)
    const dup = sim.events.filter((e) => e.type === 'event.duplicate')
    assert.ok(dup.length > 20)
    for (const f of sim.history) {
      for (const m of f.world.broker) assert.equal(m.key, `${m.kind}:${m.order}`)
    }
  })
})

describe('две записи подряд', () => {
  test('падение сервиса и недоступный брокер теряют события навсегда', () => {
    const sim = run(dualWrite)
    assert.ok(sim.world.stats.lost > 5, `потеряно ${sim.world.stats.lost}`)
    assert.ok(sim.firstEvent('event.lost'))
    assert.equal(sim.world.stats.stuck, sim.world.stats.lost, 'каждое потерянное событие — застрявший заказ')
    const stuck = sim.world.orders.filter((o) => o.state === 'created' && o.published)
    assert.ok(stuck.length > 5, 'заказ есть, события о нём нет')
  })

  test('outbox с теми же сбоями не теряет ничего', () => {
    const sim = run(withConfig(dualWrite, { publish: 'outbox' }))
    assert.equal(sim.world.stats.lost, 0)
    assert.equal(sim.world.stats.stuck, 0)
    assert.ok(consistency(sim.world) > consistency(run(dualWrite).world))
  })
})

describe('transactional outbox', () => {
  test('пока брокер лежит, строки копятся в базе, а потом уезжают', () => {
    const sim = run(outbox)
    const down = sim.firstEvent('broker.down')!
    const up = sim.firstEvent('broker.up')!
    const atDown = sim.log.series[Math.min(up.tick - 1, sim.tick)]!
    assert.ok(atDown.outbox > 3, `в outbox накопилось ${atDown.outbox}`)
    const after = sim.events.find((e) => e.type === 'relay.send' && e.tick >= up.tick)
    assert.ok(after && Number(after.payload.rows) > 3, 'после возвращения брокера отправщик догоняет')
    assert.ok(down.tick < up.tick)
    assert.equal(sim.world.stats.lost, 0)
  })

  test('цена — задержка: событие ждёт отправщика', () => {
    const fast = run(withConfig(outbox, { relayEvery: 1 }))
    const slow = run(withConfig(outbox, { relayEvery: 8 }))
    const avg = (sim: Simulation) => sim.log.done.reduce((a, c) => a + c.lat, 0) / Math.max(1, sim.log.done.length)
    assert.ok(avg(slow) > avg(fast), `${avg(slow)} против ${avg(fast)}`)
  })
})

describe('дубли', () => {
  test('без идемпотентности дубль доставки означает двойное списание', () => {
    const sim = run(duplicates)
    assert.ok(sim.world.stats.doubleCharges > 20, `двойных списаний ${sim.world.stats.doubleCharges}`)
    assert.ok(sim.firstEvent('pay.double'))
    assert.ok(sim.world.orders.some((o) => o.charges > 1))
  })

  test('с ключом идемпотентности повтор безопасен', () => {
    const sim = run(withConfig(duplicates, { idempotent: true }))
    assert.ok(sim.world.stats.duplicates > 20, 'дубли всё так же приходят')
    assert.equal(sim.world.stats.doubleCharges, 0)
    for (const o of sim.world.orders) assert.ok(o.charges <= 1, `заказ ${o.id}: списаний ${o.charges}`)
  })
})

describe('сага', () => {
  test('без компенсации неудачный шаг оставляет заказ оплаченным навсегда', () => {
    const sim = run(sagaFlow)
    assert.ok(sim.world.stats.giveups > 5)
    assert.ok(sim.world.stats.stuck > 5, `застряло ${sim.world.stats.stuck}`)
    assert.ok(sim.firstEvent('saga.stuck'))
    assert.ok(sim.world.orders.some((o) => o.state === 'paid' && o.settledAt === null))
  })

  test('компенсация возвращает деньги и доводит заказ до конца', () => {
    const sim = run(withConfig(sagaFlow, { compensate: true }))
    assert.ok(sim.world.stats.compensations > 5)
    assert.equal(sim.world.stats.stuck, 0)
    assert.ok(sim.world.stats.cancelled > 5)
    assert.ok(consistency(sim.world) > consistency(run(sagaFlow).world) + 0.1)
    for (const o of sim.world.orders) if (o.refunded) assert.equal(o.state, 'cancelled')
  })

  test('больше повторов — меньше компенсаций, но дольше неопределённость', () => {
    const few = run(withConfig(sagaFlow, { compensate: true }))
    const many = run(withConfig(sagaFlow, { compensate: true, retries: 3 }))
    assert.ok(many.world.stats.compensations < few.world.stats.compensations)
    assert.ok(many.world.stats.retries > few.world.stats.retries)
  })
})
