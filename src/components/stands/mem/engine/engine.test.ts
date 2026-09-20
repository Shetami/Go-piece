import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import {
  MEM_SCENARIOS,
  contention,
  escapeAnalysis,
  fastPath,
  largeObjects,
  sizeClasses,
  stackGrowth,
} from './scenarios.ts'
import { PAGE, SIZE_CLASSES, classFor } from './types.ts'
import type { MemScenario } from './types.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const withConfig = (s: MemScenario, config: Partial<MemScenario['config']>): MemScenario => ({
  ...s,
  config: { ...s.config, ...config },
})

const run = (s: MemScenario, seed = 1) => new Simulation(s, seed).runToEnd(s.stopAfter)

const only = (s: MemScenario, size: number, extra: Partial<MemScenario> = {}): MemScenario => ({
  ...s,
  ...extra,
  workloads: [
    {
      name: 'x',
      count: 2,
      spawnAt: 0,
      phases: [
        { kind: 'alloc', size, count: 6, lifetime: [15, 30] },
        { kind: 'cpu', ticks: 2 },
      ],
      repeat: 'forever',
    },
  ],
})

describe('быстрый путь', () => {
  test('подавляющее большинство аллокаций не доходит до общего списка спанов', () => {
    const sim = run(fastPath)
    const s = sim.world.stats
    const total = s.fast + s.tiny + s.refills + s.large + s.onStack
    assert.ok(total > 300, `аллокаций слишком мало: ${total}`)
    assert.ok(
      sim.fastShare > 0.95,
      `быстрым путём прошло ${Math.round(sim.fastShare * 100)}%, а должно больше 95%`,
    )
    assert.ok(s.refills < total / 20, `пополнений кэша слишком много: ${s.refills} на ${total}`)
  })

  test('мелочь без указателей складывается в общий блок, а без него — в отдельные слоты', () => {
    const on = run(fastPath)
    const off = run(withConfig(fastPath, { tinyAllocator: false }))

    assert.ok(on.world.stats.tiny > 200, `мелких объектов ${on.world.stats.tiny}`)
    assert.ok(
      on.world.stats.tinyBlocks < on.world.stats.tiny,
      `блоков ${on.world.stats.tinyBlocks} на ${on.world.stats.tiny} объектов — упаковки не случилось`,
    )
    assert.equal(off.world.stats.tiny, 0, 'с выключенным аллокатором мелочи общий блок не используется')
    assert.ok(off.world.stats.fast > on.world.stats.fast, 'без общего блока отдельных слотов должно стать больше')
  })
})

describe('классы размеров', () => {
  test('размер округляется вверх до класса, и лишний байт стоит трети слота', () => {
    const exact = run(only(sizeClasses, 32))
    const plusOne = run(only(sizeClasses, 33))

    assert.equal(Math.round(exact.waste * 100), 0, 'ровно 32 байта должны ложиться без потерь')
    assert.equal(Math.round(plusOne.waste * 100), 31, '33 байта уезжают в класс 48 — это 31% потерь')
    assert.equal(exact.countEvents('class.waste'), 0)
    assert.ok(plusOne.countEvents('class.waste') > 50, 'о потерях стенд обязан сообщать')
  })

  test('классы берутся из таблицы рантайма и идут по возрастанию', () => {
    for (let i = 1; i < SIZE_CLASSES.length; i++) {
      assert.ok(SIZE_CLASSES[i]! > SIZE_CLASSES[i - 1]!, `класс ${i} не больше предыдущего`)
    }
    assert.equal(classFor(1)!.size, 8)
    assert.equal(classFor(33)!.size, 48)
    assert.equal(classFor(48)!.size, 48)
    assert.equal(classFor(32768)!.size, 32768)
    assert.equal(classFor(32769), null, 'больше 32 КБ классов нет — это уже большой объект')
  })
})

describe('общий список спанов', () => {
  test('чем больше процессоров, тем чаще встречаются у блокировки', () => {
    const one = run(withConfig(contention, { gomaxprocs: 1 }))
    const four = run(withConfig(contention, { gomaxprocs: 4 }))
    const eight = run(withConfig(contention, { gomaxprocs: 8 }))

    assert.equal(one.world.stats.contended, 0, 'одному процессору не с кем делить блокировку')
    assert.ok(
      four.world.stats.contended > 0 && eight.world.stats.contended > four.world.stats.contended,
      `ожиданий: 4P — ${four.world.stats.contended}, 8P — ${eight.world.stats.contended}`,
    )
  })

  test('спан с освободившимися слотами возвращается в список частичных, а не создаётся заново', () => {
    const sim = run(contention)
    assert.ok(
      sim.world.stats.refills > sim.countEvents('central.empty') * 3,
      `спаны не переиспользуются: пополнений ${sim.world.stats.refills}, новых спанов ${sim.countEvents('central.empty')}`,
    )
  })
})

describe('большие объекты', () => {
  test('объект больше 32 КБ идёт мимо кэша, занимает целые страницы и округляется до страницы', () => {
    const sim = run(largeObjects)
    const large = sim.events.filter((e) => e.type === 'alloc.large')
    assert.ok(large.length > 20, `больших аллокаций ${large.length}`)
    for (const e of large) {
      const size = Number(e.payload.size)
      const slot = Number(e.payload.slotSize)
      assert.ok(size > 32768, 'большим считается объект больше последнего класса')
      assert.equal(slot % PAGE, 0, 'большой объект занимает целые страницы')
      assert.ok(slot >= size && slot - size < PAGE, 'округление идёт до страницы, а не больше')
    }
  })

  test('когда свободные страницы разбросаны, непрерывный кусок приходится брать у ОС', () => {
    const sim = run(largeObjects)
    assert.ok(sim.countEvents('frag.external') > 0, 'внешняя фрагментация должна проявиться')
    assert.ok(sim.rss > sim.heapPages, 'у ОС взято больше, чем занято прямо сейчас')
  })

  test('возврат памяти ОС уменьшает то, что держит процесс', () => {
    const on = run(largeObjects)
    const off = run(withConfig(largeObjects, { scavengeAfter: 0 }))
    assert.ok(on.countEvents('scavenge') > 0)
    assert.equal(off.countEvents('scavenge'), 0)
    assert.ok(off.rss >= on.rss, `без возврата у ОС должно остаться не меньше: ${off.rss} против ${on.rss}`)
  })
})

describe('побег в кучу', () => {
  test('пока указатель не убегает, объект живёт на стеке и не трогает кучу', () => {
    const on = run(escapeAnalysis)
    const off = run(withConfig(escapeAnalysis, { escapeAnalysis: false }))

    assert.ok(on.world.stats.onStack > 200, `на стеке ${on.world.stats.onStack} объектов`)
    assert.equal(off.world.stats.onStack, 0, 'без анализа побега всё уезжает в кучу')
    assert.ok(
      off.world.stats.fast > on.world.stats.fast * 1.5,
      `аллокаций в куче: с анализом ${on.world.stats.fast}, без него ${off.world.stats.fast}`,
    )
    assert.ok(
      off.world.stats.refills > on.world.stats.refills,
      'без анализа побега кэши пустеют чаще',
    )
  })
})

describe('стек горутины', () => {
  test('стек растёт удвоением и копированием, а потом ужимается', () => {
    const sim = run(stackGrowth)
    const grows = sim.events.filter((e) => e.type === 'stack.grow')
    assert.ok(grows.length >= 6, `ростов стека ${grows.length}`)
    for (const e of grows) {
      assert.equal(Number(e.payload.to), Number(e.payload.from) * 2, 'стек растёт ровно вдвое')
      assert.ok(Number(e.payload.ticks) >= 1, 'копирование стоит времени')
    }
    assert.ok(sim.countEvents('stack.shrink') > 0, 'неиспользуемый стек должен ужиматься')
  })

  test('больший стартовый стек экономит переезды, но стоит памяти', () => {
    const small = run(stackGrowth)
    const big = run(withConfig(stackGrowth, { stackStart: 8192 }))
    assert.ok(big.world.stats.stackGrows < small.world.stats.stackGrows, 'переездов должно стать меньше')
    assert.ok(big.heapPages > small.heapPages, 'зато памяти под стеки уходит больше')
  })

  test('четыре стартовых стека делят одну страницу', () => {
    const sim = new Simulation(stackGrowth, 1).runTo(1)
    const chunk = sim.world.stackChunks.find((c) => c.size === 2048)
    assert.ok(chunk, 'маленькие стеки должны нарезаться из общей страницы')
    assert.equal(chunk!.slots, 4, 'на странице 8 КБ помещается четыре стека по 2 КБ')
  })
})

describe('устройство кучи', () => {
  test('ни один слот спана не выдаётся дважды и страницы не делятся между спанами', () => {
    for (const sc of MEM_SCENARIOS) {
      const sim = run(sc)
      const w = sim.world
      for (const span of w.spans) {
        assert.ok(span.used >= 0 && span.used <= span.slots, `${sc.id}: спан #${span.id} вне границ`)
      }
      const owners = new Map<number, number>()
      for (const span of w.spans) {
        if (span.owner.kind === 'free') continue
        for (const pg of span.pages) {
          assert.equal(owners.get(pg), undefined, `${sc.id}: страница ${pg} принадлежит двум спанам`)
          owners.set(pg, span.id)
        }
      }
    }
  })

  test('ни один пресет не упирается в потолок памяти', () => {
    for (const sc of MEM_SCENARIOS) {
      const sim = run(sc)
      assert.notEqual(sim.world.finishReason, 'oom', `${sc.id} закончился нехваткой памяти`)
    }
  })
})

describe('детерминизм', () => {
  test('один seed и конфиг дают одинаковую ленту событий', () => {
    const a = run(fastPath, 42)
    const b = run(fastPath, 42)
    assert.equal(JSON.stringify(a.events), JSON.stringify(b.events))
    const c = run(fastPath, 43)
    assert.notEqual(JSON.stringify(a.events), JSON.stringify(c.events), 'разный seed — разная лента')
  })

  test('перемотка на тик N совпадает с прогоном заново до тика N', () => {
    const full = run(sizeClasses, 7)
    for (const n of [1, 5, 20, 60]) {
      const fresh = new Simulation(sizeClasses, 7).runTo(n)
      assert.deepEqual(full.at(n)!.world, fresh.world, `состояния разошлись на тике ${n}`)
    }
  })
})
