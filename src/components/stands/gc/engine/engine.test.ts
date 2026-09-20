import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import {
  GC_SCENARIOS,
  assistPressure,
  firstCycle,
  garbageVsLive,
  gogcDial,
  memLimit,
  writeBarrier,
} from './scenarios.ts'
import type { GcPhase, GcScenario } from './types.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const withConfig = (s: GcScenario, config: Partial<GcScenario['config']>): GcScenario => ({
  ...s,
  config: { ...s.config, ...config },
})

const run = (s: GcScenario, seed = 1, ticks?: number) => new Simulation(s, seed).runToEnd(ticks ?? s.stopAfter)

const scanWork = (sim: Simulation): number =>
  sim.events.filter((e) => e.type === 'mark.scan').reduce((sum, e) => sum + Number(e.payload.count), 0)

describe('цикл сборки', () => {
  test('фазы идут по кругу: пауза → разметка → пауза → подметание', () => {
    const allowed: Record<GcPhase, GcPhase[]> = {
      off: ['off', 'stw-start'],
      'stw-start': ['stw-start', 'mark'],
      mark: ['mark', 'stw-end'],
      'stw-end': ['stw-end', 'sweep'],
      sweep: ['sweep', 'off'],
    }
    const sim = run(firstCycle)
    let prev: GcPhase = 'off'
    for (const s of sim.history) {
      assert.ok(allowed[prev].includes(s.world.phase), `недопустимый переход ${prev} → ${s.world.phase}`)
      prev = s.world.phase
    }
    assert.ok(sim.cycles >= 3, `циклов должно быть несколько, а их ${sim.cycles}`)
  })

  test('на цикл приходится ровно две остановки мира', () => {
    const sim = run(firstCycle)
    const enters = sim.events.filter((e) => e.type === 'stw.enter')
    const exits = sim.events.filter((e) => e.type === 'stw.exit')
    assert.equal(enters.length, exits.length, 'каждая пауза должна закончиться')

    for (let cycle = 1; cycle <= sim.cycles; cycle++) {
      const kinds = enters.filter((e) => e.payload.cycle === cycle).map((e) => e.payload.kind)
      assert.deepEqual(kinds, ['sweep-termination', 'mark-termination'], `цикл ${cycle}: не две паузы`)
    }
  })

  test('разметка идёт одновременно с программой, а в паузах не работает никто', () => {
    const sim = run(firstCycle)
    let mutatorTicksDuringMark = 0
    for (const s of sim.history) {
      if (s.world.phase === 'mark') {
        mutatorTicksDuringMark += s.world.slots.filter((x) => x.kind === 'mutator').length
      }
      if (s.world.phase === 'stw-start' || s.world.phase === 'stw-end') {
        assert.ok(
          s.world.slots.every((x) => x.kind === 'stw'),
          `тик ${s.world.tick}: во время паузы кто-то работает`,
        )
      }
    }
    assert.ok(mutatorTicksDuringMark > 50, `программа почти не работала во время разметки: ${mutatorTicksDuringMark}`)
  })

  test('паузы не зависят от размера кучи', () => {
    const small = run(withConfig(firstCycle, { initialGoal: 24 }))
    const big = run(withConfig(firstCycle, { initialGoal: 90 }))
    const pause = (sim: Simulation) => sim.world.stats.maxPause
    assert.equal(pause(small), pause(big), 'длительность паузы должна быть одинаковой')
    assert.ok(big.peakHeap > small.peakHeap, 'куча при большей цели должна быть больше')
  })
})

describe('барьер записи', () => {
  test('с барьером ни один достижимый объект не освобождается — во всех пресетах', () => {
    for (const sc of GC_SCENARIOS) {
      for (const seed of [1, 2, 3]) {
        const sim = run(sc, seed)
        assert.equal(sim.lost, 0, `${sc.id} (seed ${seed}): сборщик освободил живую память`)
        assert.equal(sim.countEvents('barrier.missed'), 0, `${sc.id}: барьер что-то пропустил`)
      }
    }
  })

  test('без барьера объект прячется за чёрным и теряется', () => {
    const off = run(withConfig(writeBarrier, { writeBarrier: false }))
    assert.ok(off.countEvents('barrier.missed') > 0, 'опасных записей не случилось — сценарий перестал работать')
    assert.ok(off.lost > 0, 'без барьера живая память обязана теряться')
    assert.ok(off.countEvents('heap.lost') > 0)

    const on = run(writeBarrier)
    assert.ok(on.countEvents('barrier.shade') > 0, 'барьер должен закрашивать объекты')
    assert.equal(on.lost, 0)
  })

  test('без аллокации чёрным теряются объекты, рождённые во время разметки', () => {
    const off = run(withConfig(writeBarrier, { allocBlack: false }))
    assert.ok(off.lost > 0, 'объекты, созданные под разметкой, должны оказаться белыми')
  })
})

describe('помощь в разметке', () => {
  test('без помощи куча улетает за цель дальше', () => {
    const on = run(assistPressure)
    const off = run(withConfig(assistPressure, { markAssist: false }))
    assert.ok(on.world.stats.assistTicks > 50, `помощи почти не было: ${on.world.stats.assistTicks}`)
    assert.equal(off.world.stats.assistTicks, 0)
    assert.ok(
      off.peakHeap > on.peakHeap,
      `без помощи куча должна быть больше: с помощью ${on.peakHeap}, без ${off.peakHeap}`,
    )
  })

  test('чем больше процессоров у сборщика, тем меньше помощи берут с горутин', () => {
    const slow = run(withConfig(assistPressure, { gcCpuShare: 0.25 }))
    const fast = run(withConfig(assistPressure, { gcCpuShare: 0.5 }))
    assert.ok(
      fast.world.stats.assistTicks < slow.world.stats.assistTicks,
      `помощь не уменьшилась: ${slow.world.stats.assistTicks} → ${fast.world.stats.assistTicks}`,
    )
  })
})

describe('GOGC', () => {
  test('вдвое больший GOGC — вдвое реже циклы и заметно больше куча', () => {
    const low = run(withConfig(gogcDial, { gogc: 50 }))
    const mid = run(withConfig(gogcDial, { gogc: 100 }))
    const high = run(withConfig(gogcDial, { gogc: 400 }))

    assert.ok(low.cycles > mid.cycles && mid.cycles > high.cycles, `циклы: ${low.cycles}, ${mid.cycles}, ${high.cycles}`)
    assert.ok(low.peakHeap < mid.peakHeap && mid.peakHeap < high.peakHeap, 'куча должна расти вместе с GOGC')
    assert.ok(low.gcCpuShare > high.gcCpuShare, 'при большом GOGC сборщику должно доставаться меньше CPU')
    assert.ok(high.mutatorProgress > low.mutatorProgress, 'при большом GOGC программа должна успевать больше')
  })

  test('GOGC=off отключает сборку совсем', () => {
    const off = run(withConfig(gogcDial, { gogc: false }))
    assert.equal(off.cycles, 0)
    assert.equal(off.countEvents('gc.trigger'), 0)
    assert.ok(off.peakHeap > run(gogcDial).peakHeap, 'без сборки куча должна только расти')
  })
})

describe('GOMEMLIMIT', () => {
  test('лимит удерживает кучу ценой процессорного времени', () => {
    const limited = run(memLimit)
    const free = run(withConfig(memLimit, { memLimit: false }))

    assert.ok(limited.countEvents('limit.hit') > 0, 'цель должна упираться в лимит')
    assert.ok(limited.peakHeap < free.peakHeap * 0.8, `куча: с лимитом ${limited.peakHeap}, без ${free.peakHeap}`)
    assert.ok(limited.gcCpuShare > free.gcCpuShare, 'за удержание памяти платят процессором')
    assert.ok(
      limited.mutatorProgress < free.mutatorProgress,
      'полезной работы под лимитом должно стать меньше',
    )
  })

  test('лимит работает и при выключенном GOGC', () => {
    const sim = run(withConfig(memLimit, { gogc: false }))
    assert.ok(sim.cycles > 5, `сборка должна идти по лимиту, а циклов ${sim.cycles}`)
    assert.ok(sim.peakHeap < 100, `куча должна держаться у лимита, а пик ${sim.peakHeap}`)
  })
})

describe('мусор и живое', () => {
  test('разметка стоит по живому: вдвое больше мусора не добавляет работы на цикл', () => {
    const rate = (pause: number): GcScenario => ({
      ...garbageVsLive,
      workloads: [
        {
          ...garbageVsLive.workloads[0]!,
          phases: [{ kind: 'alloc', blocks: 3 }, { kind: 'cpu', ticks: pause }, { kind: 'drop' }],
        },
        garbageVsLive.workloads[1]!,
      ],
    })
    const fast = run(rate(1))
    const slow = run(rate(4))

    assert.ok(
      fast.world.stats.allocated > slow.world.stats.allocated * 1.3,
      `быстрый мусорщик должен выделить заметно больше: ${fast.world.stats.allocated} против ${slow.world.stats.allocated}`,
    )
    const perCycle = (sim: Simulation) => scanWork(sim) / sim.cycles
    const ratio = perCycle(fast) / perCycle(slow)
    assert.ok(ratio > 0.85 && ratio < 1.15, `работа разметки на цикл разъехалась: отношение ${ratio.toFixed(2)}`)
    assert.ok(fast.cycles > slow.cycles, 'зато циклов при быстром мусоре должно быть больше')
  })
})

describe('детерминизм', () => {
  test('один seed и конфиг дают одинаковую ленту событий', () => {
    const a = run(firstCycle, 42)
    const b = run(firstCycle, 42)
    assert.equal(JSON.stringify(a.events), JSON.stringify(b.events))

    const c = run(firstCycle, 43)
    assert.notEqual(JSON.stringify(a.events), JSON.stringify(c.events), 'разный seed — разная лента')
  })

  test('перемотка на тик N совпадает с прогоном заново до тика N', () => {
    const full = run(firstCycle, 11)
    for (const n of [1, 9, 30, 77]) {
      const fresh = new Simulation(firstCycle, 11).runTo(n)
      assert.deepEqual(full.at(n)!.world, fresh.world, `состояния разошлись на тике ${n}`)
    }
  })
})

describe('куча', () => {
  test('подметание освобождает ровно то, что разметка признала мусором', () => {
    const sim = run(firstCycle)
    const begins = sim.events.filter((e) => e.type === 'sweep.begin')
    const dones = sim.events.filter((e) => e.type === 'sweep.done')
    assert.ok(begins.length > 0)
    for (let i = 0; i < dones.length; i++) {
      assert.equal(dones[i]!.payload.freed, begins[i]!.payload.garbage, `цикл ${i + 1}: подмели не весь мусор`)
    }
  })

  test('блоки переиспользуются: освобождённое место занимают новые объекты', () => {
    const sim = run(firstCycle)
    assert.ok(sim.world.stats.freed > 40, 'мусор должен собираться')
    assert.ok(
      sim.world.stats.allocated > sim.world.config.heapCapacity,
      'выделено должно быть больше, чем физически влезает в кучу',
    )
  })
})
