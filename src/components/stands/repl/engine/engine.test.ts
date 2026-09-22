import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { REPL_SCENARIOS, conflict, failover, monotonic, readYourWrites, slot, stream, syncHang } from './scenarios.ts'
import type { ReplConfig, ReplScenario } from './types.ts'
import { lastLsn } from './world.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: ReplScenario) => new Simulation(s).runToEnd(s.stopAfter)
const withConfig = (s: ReplScenario, config: Partial<ReplConfig>): ReplScenario => ({ ...s, config: { ...s.config, ...config } })

describe('общие свойства движка', () => {
  test('прогон воспроизводим', () => {
    for (const sc of REPL_SCENARIOS) assert.equal(JSON.stringify(run(sc).events), JSON.stringify(run(sc).events), sc.id)
  })

  test('ни один сценарий не падает и доходит до внятного конца', () => {
    for (const sc of REPL_SCENARIOS) {
      const sim = run(sc)
      assert.ok(sim.finished && sim.world.finishReason, sc.id)
    }
  })

  test('реплика никогда не опережает ведущего, а позиции идут по порядку', () => {
    for (const sc of REPL_SCENARIOS) {
      for (const f of run(sc).history) {
        const w = f.world
        for (const n of w.nodes) {
          if (n.id === w.primary || n.broken) continue
          assert.ok(n.replayLsn <= n.flushLsn && n.flushLsn <= n.writeLsn, `${sc.id}, тик ${w.tick}: у ${n.name} replay ≤ flush ≤ write нарушено`)
          assert.ok(n.writeLsn <= lastLsn(w), `${sc.id}, тик ${w.tick}: ${n.name} получила WAL, которого у ведущего нет`)
        }
      }
    }
  })

  test('проигранное на реплике совпадает с тем, что написано в WAL', () => {
    for (const sc of REPL_SCENARIOS) {
      const sim = run(sc)
      const w = sim.world
      for (const n of w.nodes) {
        if (n.broken || !n.up) continue
        for (const [key, got] of Object.entries(n.kv)) {
          const last = [...w.wal].reverse().find((r) => r.kind === 'commit' && r.key === key && r.lsn <= n.replayLsn)
          assert.equal(got.value, last?.value ?? 0, `${sc.id}: ${n.name}.${key}`)
        }
      }
    }
  })
})

describe('поток WAL', () => {
  test('медленная далёкая реплика отстаёт сильнее быстрой', () => {
    const sim = run(stream)
    const lagAt = (tick: number, id: number) => {
      const w = sim.at(tick)!.world
      return lastLsn(w) - w.nodes[id]!.replayLsn
    }
    assert.ok(lagAt(30, 2) > lagAt(30, 1) * 2, `r1 ${lagAt(30, 1)}, r2 ${lagAt(30, 2)}`)
    assert.ok(sim.world.stats.maxLag >= 10)
  })

  test('асинхронный ведущий не ждёт реплик: каждый коммит подтверждается за тик', () => {
    assert.equal(run(stream).world.stats.maxCommitWait, 1)
  })

  test('в конце обе реплики догоняют', () => {
    const w = run(stream).world
    for (const n of w.nodes.slice(1)) assert.equal(n.replayLsn, lastLsn(w))
  })
})

describe('чтение с реплик', () => {
  test('сразу после своего коммита реплика показывает старое', () => {
    const sim = run(readYourWrites)
    assert.equal(sim.world.stats.ownStale, sim.world.stats.reads)
  })

  test('remote_apply спасает только чтение с синхронной реплики', () => {
    const sim = run(withConfig(readYourWrites, { syncCommit: 'remote_apply', standbys: 'first-r1' }))
    const bad = sim.events.filter((e) => e.type === 'read.own-stale')
    assert.ok(bad.length > 0, 'на асинхронной r2 своё должно быть не видно')
    assert.ok(bad.every((e) => e.payload.node === 'r2'), 'на синхронной r1 своё видно всегда')
  })

  test('чтение с ведущего всегда видит своё', () => {
    const sc: ReplScenario = {
      ...readYourWrites,
      clients: readYourWrites.clients.map((c) => ({ ...c, ops: c.ops.map((o) => (o.kind === 'read' ? { ...o, from: 'primary' as const } : o)) })),
    }
    assert.equal(run(sc).world.stats.staleReads, 0)
  })

  test('чтения по двум репликам с разным отставанием идут назад во времени', () => {
    assert.ok(run(monotonic).world.stats.backwards >= 2)
  })

  test('если читать всегда с одной реплики, время назад не идёт', () => {
    const sc: ReplScenario = {
      ...monotonic,
      clients: monotonic.clients.map((c) => ({ ...c, ops: c.ops.map((o) => (o.kind === 'read' ? { ...o, from: 2 } : o)) })),
    }
    const sim = run(sc)
    assert.equal(sim.world.stats.backwards, 0)
    assert.ok(sim.world.stats.staleReads > 0, 'устаревать данные при этом не перестают')
  })
})

describe('переключение', () => {
  test('асинхронно: подтверждённые коммиты теряются', () => {
    const sim = run(failover)
    assert.equal(sim.countEvents('failover.promote'), 1)
    assert.ok(sim.world.stats.lost > 0)
    assert.equal(sim.world.stats.lost, sim.countEvents('commit.lost'))
  })

  test('с синхронной репликой не теряется ни один подтверждённый коммит', () => {
    for (const standbys of ['first-r1', 'any-1'] as const) {
      const sim = run(withConfig(failover, { standbys }))
      assert.equal(sim.world.stats.lost, 0, standbys)
    }
  })

  test('синхронность стоит времени: коммитов за то же время меньше', () => {
    assert.ok(run(withConfig(failover, { standbys: 'any-1' })).world.stats.acked * 2 < run(failover).world.stats.acked)
  })

  test('без автоматического переключения запись стоит до конца', () => {
    const sim = run(withConfig(failover, { failover: false }))
    assert.equal(sim.countEvents('failover.promote'), 0)
    assert.ok(sim.world.stats.downTicks > run(failover).world.stats.downTicks * 3)
  })

  test('новой ведущей становится самая свежая реплика, и клиенты пишут в неё', () => {
    const sim = run(failover)
    const e = sim.firstEvent('failover.promote')!
    const before = sim.at(e.tick - 1)!.world
    const best = Math.max(...before.nodes.slice(1).map((n) => n.flushLsn))
    assert.ok(Number(e.payload.end) >= best)
    assert.ok(sim.events.some((x) => x.type === 'commit.ack' && x.tick > e.tick))
  })
})

describe('синхронная реплика', () => {
  test('FIRST 1 (r1): r1 упала — коммиты висят, пока не вернётся', () => {
    const sim = run(syncHang)
    assert.ok(sim.countEvents('commit.hang') > 0)
    assert.ok(sim.world.stats.maxCommitWait >= 12)
  })

  test('ANY 1: коммиты подтверждает r2', () => {
    const sim = run(withConfig(syncHang, { standbys: 'any-1' }))
    assert.equal(sim.countEvents('commit.hang'), 0)
    assert.ok(sim.world.stats.maxCommitWait < 8)
  })
})

describe('слот репликации', () => {
  test('со слотом WAL растёт, пока реплики нет, и реплика догоняет', () => {
    const sim = run(slot)
    assert.ok(sim.world.stats.maxRetained > slot.config.walKeep! * 2)
    assert.ok(sim.countEvents('wal.retained') > 0)
    const r1 = sim.world.nodes[1]!
    assert.equal(r1.broken, false)
    assert.equal(r1.replayLsn, lastLsn(sim.world))
  })

  test('без слота WAL не растёт, но реплика ломается', () => {
    const sim = run(withConfig(slot, { slots: false }))
    assert.ok(sim.world.stats.maxRetained <= slot.config.walKeep!)
    assert.equal(sim.countEvents('replica.broken'), 1)
    assert.equal(sim.world.nodes[1]!.broken, true)
  })
})

describe('конфликты с восстановлением', () => {
  test('отчёт отменяется через max_standby_streaming_delay', () => {
    const sim = run(conflict)
    const wait = sim.firstEvent('conflict.wait')!
    const cancel = sim.firstEvent('conflict.cancel')!
    assert.equal(cancel.tick - wait.tick, conflict.config.maxStandbyDelay)
  })

  test('hot_standby_feedback спасает отчёт, но мусор копится на ведущем', () => {
    const plain = run(conflict)
    const fb = run(withConfig(conflict, { hotStandbyFeedback: true }))
    assert.equal(fb.world.stats.cancels, 0)
    assert.equal(fb.countEvents('query.done'), 1)
    assert.ok(fb.world.stats.maxDead > plain.world.stats.maxDead * 3)
  })

  test('без предела ожидания отчёт доживает, но реплика отстаёт', () => {
    const plain = run(conflict)
    const inf = run(withConfig(conflict, { maxStandbyDelay: -1 }))
    assert.equal(inf.world.stats.cancels, 0)
    assert.ok(inf.world.stats.maxLag > plain.world.stats.maxLag * 2)
  })
})
