import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { KAFKA_SCENARIOS, acks, batching, group, isr, journey, keys, retry } from './scenarios.ts'
import { murmur2, partitionForKey } from './world.ts'
import type { KafkaConfig, KafkaScenario, StepKind } from './types.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: KafkaScenario, seed = 1, ticks?: number) => new Simulation(s, seed).runToEnd(ticks ?? s.stopAfter)

const withConfig = (s: KafkaScenario, config: Partial<KafkaConfig>): KafkaScenario => ({
  ...s,
  config: { ...s.config, ...config },
})

const firstTick = (steps: { tick: number; step: StepKind }[], step: StepKind) => steps.find((x) => x.step === step)?.tick

describe('общие свойства движка', () => {
  test('один seed даёт посимвольно одинаковую ленту событий', () => {
    for (const sc of KAFKA_SCENARIOS) {
      const a = run(sc, 7)
      const b = run(sc, 7)
      assert.equal(JSON.stringify(a.events), JSON.stringify(b.events), `${sc.id}: два прогона с одним seed разошлись`)
    }
  })

  test('каждый пресет доходит до конца сам, а не по лимиту тиков', () => {
    for (const sc of KAFKA_SCENARIOS) {
      const sim = run(sc)
      assert.equal(sim.world.finishReason, 'all-done', `${sc.id}: прогон не завершился за ${sc.stopAfter} тиков`)
    }
  })

  test('HW никогда не обгоняет конец лога лидера', () => {
    for (const sc of KAFKA_SCENARIOS) {
      for (const snap of run(sc).history) {
        for (const p of snap.world.partitions) {
          if (p.leader === null) continue
          assert.ok(p.hw <= p.logs[p.leader]!.log.length, `${sc.id} t=${snap.world.tick}: HW p${p.id} за концом лога`)
        }
      }
    }
  })

  test('лог реплики из ISR — всегда префикс лога лидера', () => {
    for (const sc of KAFKA_SCENARIOS) {
      for (const snap of run(sc).history) {
        for (const p of snap.world.partitions) {
          if (p.leader === null) continue
          const lead = p.logs[p.leader]!.log
          for (const b of p.isr) {
            const log = p.logs[b]!.log
            log.forEach((e, i) => {
              assert.equal(e.rec, lead[i]?.rec, `${sc.id} t=${snap.world.tick}: B${b} разошёлся с лидером p${p.id} на оффсете ${i}`)
            })
          }
        }
      }
    }
  })

  test('путь каждого сообщения идёт вперёд по времени', () => {
    for (const sc of KAFKA_SCENARIOS) {
      for (const r of run(sc).world.recs) {
        for (let i = 1; i < r.trace.length; i++) {
          assert.ok(r.trace[i]!.tick >= r.trace[i - 1]!.tick, `${sc.id}: m${r.id} шагнул назад во времени`)
        }
      }
    }
  })

  test('партиция по ключу — как в Java-клиенте Kafka', () => {
    // Векторы из UtilsTest.testMurmur2 в исходниках Kafka.
    const enc = (s: string) => new TextEncoder().encode(s)
    assert.equal(murmur2(enc('21')), -973932308)
    assert.equal(murmur2(enc('foobar')), -790332482)
    assert.equal(murmur2(enc('a-little-bit-long-string')), -985981536)
    assert.equal(murmur2(enc('a-little-bit-longer-string')), -1486304829)
    assert.equal(murmur2(enc('lkjh234lh9fiuh90y23oiuhsafujhadof229phr9h19h89h8')), -58897971)
    assert.equal(murmur2(enc('abc')), 479470107)
    assert.equal(partitionForKey('alice', 3), 0)
  })
})

describe('путь одного сообщения', () => {
  test('acks=all: подтверждение приходит только после коммита на ISR', () => {
    const sim = run(journey)
    for (const r of sim.world.recs) {
      const t = (s: StepKind) => firstTick(r.trace, s)
      assert.ok(r.state === 'acked', `m${r.id} не подтверждено`)
      assert.ok(t('append')! < t('commit')!, `m${r.id}: закоммичено раньше записи`)
      assert.ok(t('commit')! <= t('ack')!, `m${r.id}: ack раньше коммита`)
      assert.ok(t('commit')! <= t('fetch')!, `m${r.id}: потребитель прочитал незакоммиченное`)
      assert.ok(t('process')! <= t('offset')!, `m${r.id}: оффсет закоммичен раньше обработки`)
    }
  })

  test('каждое сообщение обработано группой ровно один раз', () => {
    const sim = run(journey)
    for (const r of sim.world.recs) assert.equal(r.processed.billing, 1, `m${r.id}`)
  })

  test('acks=1 подтверждает до репликации', () => {
    const sim = run(withConfig(journey, { acks: 1 }))
    const early = sim.world.recs.filter((r) => firstTick(r.trace, 'ack')! < firstTick(r.trace, 'commit')!)
    assert.ok(early.length > 0, 'при acks=1 ни одно подтверждение не обогнало коммит')
  })
})

describe('пакеты и linger', () => {
  test('linger=0 без нагрузки: по запросу на сообщение', () => {
    const sim = run(batching)
    assert.equal(sim.world.stats.requests, 36)
  })

  test('linger=5: запросов в разы меньше, задержка выше', () => {
    const base = run(batching)
    const lingered = run(withConfig(batching, { lingerTicks: 5 }))
    assert.ok(lingered.world.stats.requests * 3 <= base.world.stats.requests, `запросов ${lingered.world.stats.requests}`)
    assert.ok(lingered.avgAck > base.avgAck, 'linger не добавил задержки')
  })
})

describe('ключи', () => {
  test('один ключ — одна партиция, и группа обрабатывает его по порядку', () => {
    const sim = run(keys)
    const byKey = new Map<string, number[]>()
    for (const r of sim.world.recs) byKey.set(r.key!, [...(byKey.get(r.key!) ?? []), r.id])
    for (const [k, ids] of byKey) {
      const parts = new Set(ids.map((id) => sim.world.recs[id - 1]!.partition))
      assert.equal(parts.size, 1, `ключ ${k} попал в несколько партиций`)
      const order = ids.map((id) => firstTick(sim.world.recs[id - 1]!.trace, 'process')!)
      assert.deepEqual(order, [...order].sort((a, b) => a - b), `ключ ${k} обработан не по порядку`)
    }
  })

  test('горячий ключ забирает больше половины записей в одну партицию', () => {
    const sim = run(keys)
    const counts = sim.world.partitions.map((p) => p.logs[p.leader!]!.log.length)
    assert.ok(Math.max(...counts) * 2 > sim.world.recs.length, `распределение ${counts.join('/')}`)
  })

  test('потребителей больше, чем партиций, — лишние простаивают', () => {
    const sim = run({ ...keys, consumers: [...keys.consumers, { name: 'c5', group: 'feed' }] })
    const c5 = sim.world.consumers.find((c) => c.name === 'c5')!
    assert.deepEqual(c5.assigned, [])
    assert.equal(c5.processedCount, 0)
  })
})

describe('acks и падение лидера', () => {
  test('acks=1: подтверждённые сообщения теряются', () => {
    const sim = run(acks)
    assert.ok(sim.world.stats.lost > 0)
    assert.ok(sim.countEvents('log.truncate') > 0, 'старый лидер не обрезал лог после возвращения')
  })

  test('acks=all: то же падение, ни одной потери', () => {
    const sim = run(withConfig(acks, { acks: 'all' }))
    assert.equal(sim.world.stats.lost, 0)
    assert.equal(sim.world.stats.acked, 30)
  })
})

describe('ISR и min.insync.replicas', () => {
  test('медленный фолловер вылетает из ISR по лагу раньше, чем падает брокер', () => {
    const sim = run(isr)
    const shrink = sim.events.find((e) => e.type === 'isr.shrink' && e.payload.reason === 'lag')
    const down = sim.events.find((e) => e.type === 'broker.down')
    assert.ok(shrink && down && shrink.tick < down.tick)
  })

  test('ISR меньше min.insync — запись отклоняется, часть неудач всё же лежит в логе', () => {
    const sim = run(isr)
    assert.ok(sim.events.some((e) => e.type === 'produce.error' && e.payload.error === 'NOT_ENOUGH_REPLICAS'))
    const failed = sim.world.recs.filter((r) => r.state === 'failed')
    assert.ok(failed.length > 0)
    assert.ok(failed.some((r) => r.committedTick !== null), 'ни одна «неудача» не оказалась в логе')
  })

  test('min.insync.replicas=1: записи не останавливаются', () => {
    const sim = run(withConfig(isr, { minInsyncReplicas: 1 }))
    assert.equal(sim.world.stats.failed, 0)
  })
})

describe('группа потребителей', () => {
  test('партиция в группе всегда у одного потребителя', () => {
    for (const snap of run(group).history) {
      for (const g of snap.world.groups) {
        const owners = new Map<number, string>()
        for (const c of snap.world.consumers) {
          if (c.group !== g.name || c.state !== 'active') continue
          for (const p of c.assigned) {
            assert.ok(!owners.has(p), `t=${snap.world.tick}: p${p} у ${owners.get(p)} и ${c.name}`)
            owners.set(p, c.name)
          }
        }
      }
    }
  })

  test('упавший уносит незакоммиченное: его обработают ещё раз', () => {
    const sim = run(group)
    assert.ok(sim.world.groups.find((g) => g.name === 'billing')!.stats.reprocessed > 0)
    assert.equal(sim.world.stats.skipped, 0)
  })

  test('другая группа читает тот же топик независимо и целиком', () => {
    const sim = run(group)
    for (const r of sim.world.recs) assert.equal(r.processed.audit, 1, `m${r.id}`)
  })

  test('автокоммит: необработанное пропускается навсегда', () => {
    const sim = run(withConfig(group, { commitMode: 'auto' }))
    assert.ok(sim.world.stats.skipped > 0)
    const never = sim.world.recs.filter((r) => (r.processed.billing ?? 0) === 0)
    assert.ok(never.length > 0)
  })
})

describe('потерянный ответ', () => {
  test('без идемпотентности повтор оставляет дубль, и его обрабатывают дважды', () => {
    const sim = run(retry)
    assert.ok(sim.world.stats.duplicates > 0)
    assert.ok(sim.world.stats.reprocessed > 0)
  })

  test('с идемпотентностью брокер узнаёт повтор и не пишет его', () => {
    const sim = run(withConfig(retry, { idempotence: true }))
    assert.equal(sim.world.stats.duplicates, 0)
    assert.ok(sim.world.stats.dedups > 0)
    assert.equal(sim.world.stats.reprocessed, 0)
  })
})
