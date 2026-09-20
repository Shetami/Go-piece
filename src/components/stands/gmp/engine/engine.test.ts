import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { channelPingPong, greedyLoop, skewAndStealing } from './scenarios.ts'
import type { Scenario } from './types.ts'

/**
 * Критерии готовности из документа «Стенд „Планировщик Go“: модель».
 * Каждый тест — один пункт чек-листа.
 */

const withConfig = (s: Scenario, config: Partial<Scenario['config']>): Scenario => ({
  ...s,
  config: { ...s.config, ...config },
})

describe('кража работы', () => {
  test('при GOMAXPROCS=1 события p.stole не возникают никогда', () => {
    const sim = new Simulation(withConfig(skewAndStealing, { gomaxprocs: 1 }), 3)
    sim.runToEnd(600)
    assert.equal(sim.countEvents('p.stole'), 0)
    assert.ok(sim.world.gs.every((g) => g.state === 'dead'), 'все горутины должны завершиться')
  })

  test('без кражи остальные P простаивают до конца прогона', () => {
    const sim = new Simulation(withConfig(skewAndStealing, { workStealing: false }), 3)
    sim.runToEnd(1200)
    const usedPs = new Set(
      sim.events.filter((e) => e.type === 'g.scheduled').flatMap((e) => e.actors.p ?? []),
    )
    assert.deepEqual([...usedPs], [0], 'работать должен только P0')
    for (const p of sim.world.ps.slice(1)) {
      assert.equal(p.runq.length, 0)
    }
  })

  test('с кражей разброс очередей сходится к 50-му тику', () => {
    const sim = new Simulation(skewAndStealing, 3)
    sim.runTo(50)
    const lens = sim.runqLengths()
    const max = Math.max(...lens)
    const min = Math.min(...lens)
    assert.ok(
      max - min <= Math.max(2, Math.ceil(max / 2)),
      `очереди разъехались: ${JSON.stringify(lens)}`,
    )
  })
})

describe('кража runnext', () => {
  test('горутину из runnext занятого P крадут на последней попытке', () => {
    const sc: Scenario = {
      id: 'runnext-steal',
      title: 'Кража runnext',
      claim: 'runnext тоже можно украсть.',
      config: { gomaxprocs: 2, asyncPreemption: false },
      workloads: [
        { name: 'hog', count: 1, spawnAt: 0, spawnOn: 0, phases: [{ kind: 'cpu', ticks: 'forever' }] },
        { name: 'worker', count: 1, spawnAt: 1, spawnOn: 0, phases: [{ kind: 'cpu', ticks: 3 }] },
      ],
      watchFor: [],
    }
    const sim = new Simulation(sc, 1)
    sim.runToEnd(200)
    const stolen = sim.events.find((e) => e.type === 'p.stole' && e.payload.runnext === true)
    assert.ok(stolen, 'должна быть кража из runnext')
    const worker = sim.world.gs.find((g) => g.name === 'worker')!
    assert.equal(worker.state, 'dead', 'украденный worker должен доработать на P1')
  })
})

describe('очереди', () => {
  test('при переполнении локальной очереди половина уезжает в глобальную', () => {
    const flood: Scenario = {
      id: 'flood',
      title: 'Переполнение',
      claim: 'Локальная очередь не резиновая.',
      config: { gomaxprocs: 1, runqCapacity: 256 },
      workloads: [
        { name: 'w', count: 300, spawnAt: 0, spawnOn: 0, phases: [{ kind: 'cpu', ticks: 5 }] },
      ],
      watchFor: ['p.runqOverflow'],
    }
    const sim = new Simulation(flood, 1)
    sim.runTo(1)

    const overflow = sim.events.find((e) => e.type === 'p.runqOverflow')
    assert.ok(overflow, 'должно быть событие p.runqOverflow')
    assert.equal(overflow!.payload.fromRunq, 128, 'уезжает ровно половина вместимости')
    assert.equal(overflow!.payload.total, 129, 'половина очереди плюс сама новая горутина')
  })

  test('без правила 61-го тика глобальная очередь ждёт, пока локальная не опустеет', () => {
    const sc: Scenario = {
      id: 'starve',
      title: 'Голодание глобальной очереди',
      claim: 'Локальная очередь всегда идёт первой.',
      config: { gomaxprocs: 1, globalCheckEvery: false, quantum: 100 },
      workloads: [
        { name: 'w', count: 10, spawnAt: 0, spawnOn: 0, phases: [{ kind: 'cpu', ticks: 5 }] },
      ],
      watchFor: [],
    }
    const sim = new Simulation(sc, 1)
    sim.runTo(1)

    const p0 = sim.world.ps[0]!
    const victim = p0.runq.pop()!
    sim.world.globrunq.push(victim)

    sim.runToEnd(400)
    assert.equal(sim.countEvents('p.globalCheck'), 0, 'правило выключено — проверок быть не должно')

    const sched = sim.events.find((e) => e.type === 'g.scheduled' && (e.actors.g ?? [])[0] === victim)
    assert.ok(sched, 'горутина из глобальной очереди всё-таки должна дождаться')
    assert.equal(sched!.payload.source, 'global')

    const before = sim.at(sched!.tick - 1)!
    assert.equal(before.world.ps[0]!.runq.length, 0, 'её очередь дошла только на пустой локальной')
    assert.equal(before.world.ps[0]!.runnext, null)
  })

  test('с правилом 61-го тика глобальная очередь получает управление, не дожидаясь пустой локальной', () => {
    const sc: Scenario = {
      id: 'fair',
      title: 'Правило 61-го тика',
      claim: 'Глобальная очередь не голодает.',
      config: { gomaxprocs: 1, globalCheckEvery: 3, quantum: 100 },
      workloads: [
        { name: 'w', count: 10, spawnAt: 0, spawnOn: 0, phases: [{ kind: 'cpu', ticks: 5 }] },
      ],
      watchFor: [],
    }
    const sim = new Simulation(sc, 1)
    sim.runTo(1)
    const p0 = sim.world.ps[0]!
    const victim = p0.runq.pop()!
    sim.world.globrunq.push(victim)
    sim.runToEnd(400)

    const sched = sim.events.find((e) => e.type === 'g.scheduled' && (e.actors.g ?? [])[0] === victim)
    assert.ok(sched)
    assert.equal(sched!.payload.source, 'global-61')
    const before = sim.at(sched!.tick - 1)!
    assert.ok(before.world.ps[0]!.runq.length > 0, 'локальная очередь была ещё не пуста')
  })
})

describe('системные вызовы и потоки', () => {
  const syscallScenario = (ticks: number, count: number): Scenario => ({
    id: `syscall-${ticks}-${count}`,
    title: 'Системные вызовы',
    claim: 'Долгий вызов стоит потока.',
    config: { gomaxprocs: 4, retakeThreshold: 5 },
    workloads: [
      { name: 'io', count, spawnAt: 0, phases: [{ kind: 'syscall', ticks }] },
    ],
    watchFor: [],
  })

  test('короткий вызов не порождает sysmon.retake, длинный — ровно один', () => {
    const short = new Simulation(syscallScenario(3, 1), 1)
    short.runToEnd(200)
    assert.equal(short.countEvents('sysmon.retake'), 0)
    assert.equal(short.events.find((e) => e.type === 'm.syscallExit')?.payload.keptP, true)

    const long = new Simulation(syscallScenario(30, 1), 1)
    long.runToEnd(200)
    assert.equal(long.countEvents('sysmon.retake'), 1)
    assert.equal(long.events.find((e) => e.type === 'm.syscallExit')?.payload.keptP, false)
  })

  test('20 долгих системных вызовов дают не меньше 20 потоков при GOMAXPROCS=4', () => {
    const sim = new Simulation(syscallScenario(60, 20), 1)
    sim.runToEnd(400)
    assert.ok(sim.peakThreads >= 20, `потоков всего ${sim.peakThreads}`)
  })

  test('20 сетевых ожиданий не увеличивают число потоков вообще', () => {
    const sc: Scenario = {
      id: 'net-only',
      title: 'Только сеть',
      claim: 'Сетевое ожидание не стоит потока.',
      config: { gomaxprocs: 4, retakeThreshold: 5 },
      workloads: [
        { name: 'net', count: 20, spawnAt: 0, phases: [{ kind: 'net', ticks: 60 }] },
      ],
      watchFor: [],
    }
    const sim = new Simulation(sc, 1)
    sim.runToEnd(400)
    assert.ok(sim.peakThreads <= 4, `потоков ${sim.peakThreads}, а P всего 4`)
    assert.equal(sim.countEvents('sysmon.retake'), 0)
  })
})

describe('вытеснение', () => {
  test('без асинхронного вытеснения вечный цикл не пускает никого на процессор', () => {
    const off = new Simulation(withConfig(greedyLoop, { asyncPreemption: false }), 1)
    off.runToEnd(400)
    const ranWorkloads = new Set(
      off.events
        .filter((e) => e.type === 'g.scheduled')
        .map((e) => off.world.gs[(e.actors.g ?? [])[0]! - 1]!.workload),
    )
    assert.deepEqual([...ranWorkloads], [0], 'должен исполняться только hog')
    assert.equal(off.countEvents('sysmon.preempt'), 0)

    const on = new Simulation(greedyLoop, 1)
    on.runToEnd(400)
    const finished = on.world.gs.filter((g) => g.state === 'dead').length
    assert.ok(finished >= 5, `с вытеснением воркеры должны доработать, завершилось ${finished}`)
  })
})

describe('runnext', () => {
  const exchanges = (sim: Simulation): number =>
    sim.events.filter(
      (e) => e.type === 'g.ready' && (e.payload.by === 'chanSend' || e.payload.by === 'chanRecv'),
    ).length

  test('со слотом runnext пинг-понг делает заметно больше обменов', () => {
    const on = new Simulation(channelPingPong, 5)
    on.runToEnd(300)
    const off = new Simulation(withConfig(channelPingPong, { runnext: false }), 5)
    off.runToEnd(300)

    assert.ok(
      exchanges(on) > exchanges(off) * 1.5,
      `с runnext ${exchanges(on)}, без него ${exchanges(off)}`,
    )
  })
})

describe('детерминизм', () => {
  test('один seed и конфиг дают одинаковую ленту событий', () => {
    const a = new Simulation(skewAndStealing, 42).runToEnd(300)
    const b = new Simulation(skewAndStealing, 42).runToEnd(300)
    assert.equal(JSON.stringify(a.events), JSON.stringify(b.events))

    const c = new Simulation(skewAndStealing, 43).runToEnd(300)
    assert.notEqual(JSON.stringify(a.events), JSON.stringify(c.events), 'разный seed — разная лента')
  })

  test('перемотка на тик N совпадает с прогоном заново до тика N', () => {
    const full = new Simulation(skewAndStealing, 11).runToEnd(300)
    for (const n of [1, 7, 23, 50]) {
      const fresh = new Simulation(skewAndStealing, 11).runTo(n)
      assert.deepEqual(full.at(n)!.world, fresh.world, `состояния разошлись на тике ${n}`)
    }
  })
})

describe('взаимная блокировка', () => {
  test('все горутины в ожидании и будить некому — событие deadlock и остановка', () => {
    const sc: Scenario = {
      id: 'deadlock',
      title: 'Приём без отправителя',
      claim: 'Если будить некому, рантайм это замечает.',
      config: { gomaxprocs: 1 },
      workloads: [
        { name: 'reader', count: 2, spawnAt: 0, phases: [{ kind: 'chanRecv', chan: 'c' }] },
      ],
      watchFor: ['deadlock'],
    }
    const sim = new Simulation(sc, 1)
    sim.runToEnd(200)
    assert.equal(sim.countEvents('deadlock'), 1)
    assert.equal(sim.world.finishReason, 'deadlock')
    assert.ok(sim.tick < 30, 'обнаружиться должно сразу, а не по потолку тиков')
  })
})
