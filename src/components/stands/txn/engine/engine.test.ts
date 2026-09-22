import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import {
  TXN_SCENARIOS,
  bloat,
  deadlock,
  dirty,
  durability,
  lostUpdate,
  nonrepeatable,
  phantom,
  rowLock,
  ssi,
  versions,
  writeSkew,
} from './scenarios.ts'
import { invariantState } from './tick.ts'
import type { TxnConfig, TxnScenario } from './types.ts'
import { committedValue, isDead, labelOf, takeSnapshot, viewOf, visibleTo } from './world.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: TxnScenario) => new Simulation(s).runToEnd(s.stopAfter)

const withConfig = (s: TxnScenario, config: Partial<TxnConfig>): TxnScenario => ({
  ...s,
  config: { ...s.config, ...config },
})

const invariantOk = (sim: Simulation) => invariantState(sim.world, sim.scenario.invariant)?.ok

/** Все прочитанные значения по ключу у сессии с этим именем — по всем её транзакциям. */
const readsOf = (sim: Simulation, name: string, key: string) =>
  sim.events
    .filter((e) => e.type === 'row.read' && String(e.payload.txn).startsWith(name) && e.payload.key === key)
    .map((e) => e.payload.value)

describe('общие свойства движка', () => {
  test('прогон воспроизводим: два запуска дают посимвольно одинаковую ленту', () => {
    for (const sc of TXN_SCENARIOS) {
      assert.equal(JSON.stringify(run(sc).events), JSON.stringify(run(sc).events), `${sc.id}: прогоны разошлись`)
    }
  })

  test('ни один сценарий не падает и доходит до внятного конца', () => {
    for (const sc of TXN_SCENARIOS) {
      const sim = run(sc)
      assert.ok(sim.finished, `${sc.id}: прогон не завершился`)
      assert.ok(sim.world.finishReason, `${sc.id}: не сказано, почему прогон кончился`)
    }
  })

  test('любому снимку в любой момент видно не больше одной версии каждой строки', () => {
    for (const sc of TXN_SCENARIOS) {
      for (const f of run(sc).history) {
        const w = f.world
        const views = [
          { snap: takeSnapshot(w, null), own: null as number | null },
          ...w.txns.flatMap((t) => {
            const snap = viewOf(w, t)
            return snap ? [{ snap, own: t.xid }] : []
          }),
        ]
        for (const { snap, own } of views) {
          const keys = w.tuples.filter((t) => visibleTo(w, t, snap, own)).map((t) => t.key)
          assert.equal(new Set(keys).size, keys.length, `${sc.id}, тик ${w.tick}: две видимые версии одной строки`)
        }
      }
    }
  })

  test('каждая транзакция кончается ровно одним исходом, а ожидание — пробуждением или ошибкой', () => {
    for (const sc of TXN_SCENARIOS) {
      const sim = run(sc)
      const s = sim.world.stats
      const begins = sim.countEvents('txn.begin')
      const ends = s.commits + s.rollbacks + s.aborts
      const open = sim.world.txns.filter((t) => t.state === 'active' || t.state === 'waiting' || t.state === 'committing').length
      assert.equal(begins, ends + open, `${sc.id}: начато ${begins}, закончено ${ends}, открыто ${open}`)
    }
  })
})

describe('грязное чтение', () => {
  test('READ UNCOMMITTED видит незакоммиченное, которое потом откатят', () => {
    const sim = run(dirty)
    assert.equal(sim.countEvents('anomaly.dirty'), 1)
    assert.deepEqual(readsOf(sim, 'T2', 'acc'), [0])
    assert.equal(committedValue(sim.world, 'acc'), 100, 'откат не вернул значение')
  })

  test('READ COMMITTED чужого незакоммиченного не видит', () => {
    const sim = run(withConfig(dirty, { isolation: 'read-committed' }))
    assert.equal(sim.countEvents('anomaly.dirty'), 0)
    assert.deepEqual(readsOf(sim, 'T2', 'acc'), [100])
  })
})

describe('неповторяемое чтение и фантом', () => {
  test('READ COMMITTED: второе чтение видит чужой коммит', () => {
    const sim = run(nonrepeatable)
    assert.deepEqual(readsOf(sim, 'T1', 'acc'), [100, 70])
    assert.equal(sim.countEvents('anomaly.nonrepeatable'), 1)
  })

  test('REPEATABLE READ: один снимок — одно значение', () => {
    const sim = run(withConfig(nonrepeatable, { isolation: 'repeatable-read' }))
    assert.deepEqual(readsOf(sim, 'T1', 'acc'), [100, 100])
    assert.equal(sim.world.stats.anomalies, 0)
  })

  test('фантом есть на READ COMMITTED и исчезает на REPEATABLE READ', () => {
    const counts = (s: Simulation) => s.events.filter((e) => e.type === 'row.scan').map((e) => e.payload.count)
    assert.deepEqual(counts(run(phantom)), [2, 3])
    assert.equal(run(phantom).countEvents('anomaly.phantom'), 1)
    const rr = run(withConfig(phantom, { isolation: 'repeatable-read' }))
    assert.deepEqual(counts(rr), [2, 2])
  })
})

describe('потерянное обновление', () => {
  test('READ COMMITTED и расчёт в приложении: вторая запись затирает первую без ошибки', () => {
    const sim = run(lostUpdate)
    assert.equal(sim.world.stats.aborts, 0, 'никто не получил ошибку — в этом и беда')
    assert.equal(sim.countEvents('anomaly.lost'), 1)
    assert.equal(committedValue(sim.world, 'acc'), 130)
    assert.equal(invariantOk(sim), false)
  })

  test('SET v = v + d: второй ждёт, перепроверяет свежую версию и ничего не теряет', () => {
    const sim = run(withConfig(lostUpdate, { rmw: 'atomic' }))
    assert.ok(sim.countEvents('lock.wait') >= 1)
    assert.ok(sim.countEvents('row.recheck') >= 1)
    assert.equal(committedValue(sim.world, 'acc'), 180)
    assert.equal(invariantOk(sim), true)
  })

  test('SELECT … FOR UPDATE: второй ждёт уже на чтении и читает свежее', () => {
    const sim = run(withConfig(lostUpdate, { rmw: 'for-update' }))
    assert.deepEqual(readsOf(sim, 'T2', 'acc'), [150])
    assert.equal(committedValue(sim.world, 'acc'), 180)
  })

  test('REPEATABLE READ: вместо потери — ошибка сериализации, а повтор даёт верный итог', () => {
    const rr = run(withConfig(lostUpdate, { isolation: 'repeatable-read' }))
    assert.equal(rr.world.stats.serializationFailures, 1)
    assert.equal(committedValue(rr.world, 'acc'), 150, 'упавшая транзакция не должна ничего оставить')
    assert.equal(invariantOk(rr), true)

    const retried = run(withConfig(lostUpdate, { isolation: 'repeatable-read', retry: true }))
    assert.equal(retried.world.stats.retries, 1)
    assert.equal(committedValue(retried.world, 'acc'), 180)
  })
})

describe('перекос записи и SERIALIZABLE', () => {
  test('REPEATABLE READ пропускает обе транзакции — и правило нарушено', () => {
    const sim = run(writeSkew)
    assert.equal(sim.world.stats.commits, 2)
    assert.equal(sim.world.stats.aborts, 0)
    assert.equal(invariantOk(sim), false)
  })

  test('SERIALIZABLE видит цикл rw-зависимостей и прерывает вторую', () => {
    const sim = run(withConfig(writeSkew, { isolation: 'serializable' }))
    assert.ok(sim.countEvents('ssi.conflict') >= 2)
    const abort = sim.firstEvent('txn.abort')
    assert.equal(abort?.payload.reason, 'ssi')
    assert.equal(invariantOk(sim), true)
  })

  test('повтор упавшей транзакции видит новое состояние и честно отказывается', () => {
    const sim = run(ssi)
    assert.equal(sim.world.stats.retries, 1)
    assert.equal(sim.countEvents('stmt.skip'), 1)
    assert.equal(invariantOk(sim), true)
    assert.equal(sim.world.stats.commits, 2)
  })

  test('SERIALIZABLE не мешает транзакциям, у которых нет цикла', () => {
    for (const sc of [nonrepeatable, phantom, rowLock]) {
      const sim = run(withConfig(sc, { isolation: 'serializable' }))
      assert.equal(sim.firstEvent('txn.abort')?.payload.reason === 'ssi', false, `${sc.id}: ложная ошибка SSI`)
    }
  })
})

describe('взаимная блокировка', () => {
  test('цикл находится после deadlock_timeout, и прерывается ровно одна транзакция', () => {
    const sim = run(deadlock)
    const found = sim.firstEvent('deadlock.found')
    assert.ok(found)
    const firstWait = sim.firstEvent('lock.wait')!
    assert.equal(found.tick - firstWait.tick, deadlock.config.deadlockTimeout)
    assert.equal(sim.world.stats.deadlocks, 1)
    assert.equal(sim.world.stats.aborts, 1)
    assert.equal(invariantOk(sim), true)
  })

  test('повтор жертвы встаёт в очередь за выжившей, а не устраивает новый дедлок', () => {
    const sim = run(withConfig(deadlock, { retry: true }))
    assert.equal(sim.world.stats.deadlocks, 1)
    assert.equal(sim.world.stats.commits, 2)
    assert.equal(committedValue(sim.world, 'a'), 110)
    assert.equal(committedValue(sim.world, 'b'), 90)
  })

  test('больше deadlock_timeout — дольше висят обе', () => {
    const slow = run(withConfig(deadlock, { deadlockTimeout: 8 }))
    assert.ok(slow.world.stats.waitTicks > run(deadlock).world.stats.waitTicks)
  })
})

describe('версии строк и очистка', () => {
  test('UPDATE добавляет версию, а транзакция со старым снимком читает прошлое', () => {
    const sim = run(versions)
    assert.deepEqual(readsOf(sim, 'R', 'acc'), [100, 100])
    const peak = Math.max(...sim.history.map((f) => f.world.tuples.length))
    assert.equal(peak, 3, 'три версии одной строки')
  })

  test('VACUUM не трогает то, что видит открытый снимок, и убирает после его конца', () => {
    const sim = run(versions)
    const blocked = sim.firstEvent('vacuum.blocked')
    assert.ok(blocked)
    assert.equal(blocked.payload.holder, 'R')
    const last = sim.events.filter((e) => e.type === 'vacuum.run').at(-1)!
    assert.equal(last.payload.removed, 2)
    assert.equal(sim.world.tuples.length, 1)
  })

  test('очередь к строке: все прибавки на месте, порядок — по приходу', () => {
    const sim = run(rowLock)
    assert.equal(committedValue(sim.world, 'acc'), 130)
    assert.equal(sim.countEvents('lock.wait'), 2)
    const order = sim.events.filter((e) => e.type === 'txn.commit').map((e) => e.payload.txn)
    assert.deepEqual(order, ['T1', 'T2', 'T3'])
  })

  test('забытая транзакция со снимком раздувает таблицу; без снимка — нет', () => {
    const sim = run(bloat)
    const whileOpen = sim.history.filter((f) => f.world.txns[1]!.state === 'active').map((f) => f.world.tuples.length)
    assert.ok(Math.max(...whileOpen) >= 10, `версий накопилось всего ${Math.max(...whileOpen)}`)
    assert.ok(sim.countEvents('vacuum.blocked') >= 3)
    assert.ok(sim.world.tuples.length <= 2, 'после конца транзакции очистка не справилась')

    const rc = run({ ...bloat, sessions: bloat.sessions.map((s) => (s.name === 'R' ? { ...s, isolation: 'read-committed' } : s)) })
    assert.equal(rc.countEvents('vacuum.blocked'), 0, 'READ COMMITTED между запросами не держит горизонт')
    assert.ok(Math.max(...rc.history.map((f) => f.world.tuples.length)) < Math.max(...whileOpen))
  })

  test('мёртвые версии не видит никто — иначе их нельзя было бы убирать', () => {
    for (const sc of TXN_SCENARIOS) {
      for (const f of run(sc).history) {
        const w = f.world
        const now = takeSnapshot(w, null)
        for (const t of w.tuples) {
          if (isDead(w, t)) assert.equal(visibleTo(w, t, now, null), false, `${sc.id}: мёртвая версия видна новому снимку`)
        }
      }
    }
  })
})

describe('долговечность', () => {
  test('synchronous_commit = on: падение не теряет ни одного подтверждённого коммита', () => {
    const sim = run(durability)
    assert.equal(sim.countEvents('db.crash'), 1)
    assert.equal(sim.world.stats.lostCommits, 0)
    assert.equal(invariantOk(sim), true)
  })

  test('групповой коммит: один fsync подтверждает сразу несколько транзакций', () => {
    const sim = run(durability)
    const group = sim.events.filter((e) => e.type === 'wal.flush' && Number(e.payload.commits) >= 2)
    assert.ok(group.length >= 2, 'коммиты ни разу не уехали вместе')
  })

  test('synchronous_commit = off: коммиты быстрее, но падение уносит подтверждённые', () => {
    const on = run(durability)
    const off = run(withConfig(durability, { syncCommit: false }))
    const acked = (s: Simulation, tick: number) =>
      s.events.filter((e) => e.type === 'txn.commit' && e.tick < tick).length
    assert.ok(acked(off, durability.config.crashAt!) > acked(on, durability.config.crashAt!) * 2, 'асинхронный коммит не дал прироста')
    assert.ok(off.world.stats.lostCommits > 0)
    assert.equal(invariantOk(off), false)
    const lost = off.firstEvent('commit.lost')!
    assert.ok(Number(lost.payload.lsn) > Number(lost.payload.flushed))
  })

  test('коммит, чью запись сбросила контрольная точка, не остаётся ждать вечно', () => {
    for (const every of [2, 3, 4, 6]) {
      const sim = run(withConfig(durability, { checkpointEvery: every, crashAt: 0 }))
      assert.equal(sim.world.finishReason, 'all-done', `контрольная точка каждые ${every}: прогон не закончился`)
      assert.equal(sim.world.stats.commits, 24)
    }
  })

  test('частые контрольные точки сокращают восстановление', () => {
    const down = (s: Simulation) => Number(s.firstEvent('db.crash')!.payload.down)
    assert.ok(down(run(withConfig(durability, { checkpointEvery: 6 }))) < down(run(durability)))
  })

  test('после восстановления видно ровно то, что было в WAL на диске', () => {
    const sim = run(withConfig(durability, { syncCommit: false }))
    const crashTick = durability.config.crashAt!
    const after = sim.at(crashTick)!.world
    for (const t of after.tuples) assert.ok(t.lsnIn <= after.flushedLsn, `версия ${t.key}=${t.value} пережила падение без записи в WAL`)
    for (const t of [...after.history, ...after.txns]) {
      if (t.lost) assert.equal(after.xact[t.xid!], 'aborted', `${labelOf(t)} потеряна, но считается закоммиченной`)
    }
  })
})
