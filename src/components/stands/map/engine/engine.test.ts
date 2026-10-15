import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import { parseScript } from './script.ts'
import { Rng } from './rng.ts'
import { hashKey, probeSeq, h2 } from './hash.ts'
import { isFull, maxGrowth, tombstones } from './swiss.ts'
import { isFullTop } from './legacy.ts'
import {
  MAP_SCENARIOS,
  growth,
  hint,
  iterGrow,
  legacyEvac,
  legacySameSize,
  noShrink,
  order,
  promote,
  prune,
  tombs,
} from './scenarios.ts'
import type { Impl, MapConfig, MapScenario, MapWorld } from './types.ts'
import { CTRL_EMPTY } from './types.ts'

/**
 * Каждый тест закрепляет либо правило рантайма, либо claim сценария: если стенд
 * однажды начнёт показывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: MapScenario, seed?: number) => new Simulation(s, seed).runToEnd()
const scenario = (script: string, config: Partial<MapConfig> = {}): MapScenario => ({
  id: 'test',
  title: 'test',
  claim: 'test',
  config,
  script,
  watchFor: [],
})

/** Содержимое карты — по живым слотам, без поиска. */
function contents(w: MapWorld): Map<string, number> {
  const out = new Map<string, number>()
  if (w.swiss) {
    const m = w.swiss
    const groups = m.dir.length === 0 ? (m.small ? [m.small] : []) : Object.values(m.tables).filter((t) => !t.replaced).flatMap((t) => t.groups)
    for (const g of groups) for (const s of g.slots) if (isFull(s.ctrl)) out.set(s.key!, s.val!)
  }
  if (w.legacy) {
    const h = w.legacy
    const arrays = [h.buckets, h.oldbuckets].filter((x): x is number => x !== null)
    for (const id of arrays) {
      h.arrays[id]!.forEach((b) => {
        if (id === h.oldbuckets && b.evacuated) return
        for (const p of b.pages) for (const s of p) if (isFullTop(s.top)) out.set(s.key!, s.val!)
      })
    }
  }
  return out
}

/** Инварианты Swiss table, которые рантайм держит после каждой операции. */
function checkSwiss(w: MapWorld, where: string) {
  const m = w.swiss
  if (!m) return
  assert.equal(m.dir.length, m.dir.length === 0 ? 0 : 1 << m.globalDepth, `${where}: длина каталога не 2^globalDepth`)
  let used = 0
  for (const t of Object.values(m.tables)) {
    if (t.replaced) continue
    const full = t.groups.flatMap((g) => g.slots).filter((s) => isFull(s.ctrl)).length
    assert.equal(full, t.used, `${where}: T${t.id} used расходится со слотами`)
    assert.ok(tombstones(t) >= 0, `${where}: T${t.id} отрицательное число надгробий`)
    assert.equal(
      t.groups.flatMap((g) => g.slots).filter((s) => s.ctrl === 0xfe).length,
      tombstones(t),
      `${where}: T${t.id} надгробия не сходятся с growthLeft`,
    )
    assert.ok(t.capacity <= w.config.maxTableSlots, `${where}: T${t.id} больше потолка`)
    assert.ok(t.used + tombstones(t) <= maxGrowth(t.capacity), `${where}: T${t.id} заполнена больше 7/8`)
    // Строки каталога, указывающие на таблицу, — ровно 2^(globalDepth - localDepth) подряд.
    const rows: number[] = m.dir.flatMap((id, i) => (id === t.id ? [i] : []))
    assert.equal(rows.length, 1 << (m.globalDepth - t.localDepth), `${where}: T${t.id} не та доля каталога`)
    assert.equal(rows[0], t.index, `${where}: T${t.index} index не первая строка`)
    // Каждый ключ достижим по своей последовательности проб, не перешагивая групп с empty.
    t.groups.forEach((g, gi) => {
      g.slots.forEach((s) => {
        if (!isFull(s.ctrl)) return
        assert.equal(s.ctrl, h2(s.hash), `${where}: контрольный байт не H2`)
        for (const off of probeSeq(s.hash, t.groups.length - 1)) {
          if (off === gi) break
          assert.ok(
            !t.groups[off]!.slots.some((x) => x.ctrl === CTRL_EMPTY),
            `${where}: ключ ${s.key} спрятан за группой с empty`,
          )
        }
      })
    })
    used += t.used
  }
  if (m.dir.length > 0) assert.equal(used, m.used, `${where}: used карты не равен сумме таблиц`)
}

describe('общие свойства движка', () => {
  test('у всех пресетов сценарий разбирается без ошибок', () => {
    for (const sc of MAP_SCENARIOS) {
      assert.deepEqual(parseScript(sc.script).errors, [], `${sc.id}: ошибки в сценарии`)
    }
  })

  test('один seed даёт посимвольно одинаковую ленту событий', () => {
    for (const sc of MAP_SCENARIOS) {
      assert.equal(JSON.stringify(run(sc, 7).events), JSON.stringify(run(sc, 7).events), `${sc.id}: прогоны разошлись`)
    }
  })

  test('каждый пресет доходит до конца и показывает то, ради чего заведён', () => {
    for (const sc of MAP_SCENARIOS) {
      for (const seed of [1, 2, 3, 4, 5]) {
        const sim = run(sc, seed)
        assert.ok(sim.finished && sim.world.finishReason !== 'stop-after', `${sc.id}/${seed}: не закончился`)
        const seen = sc.watchFor.filter((t) => sim.countEvents(t) > 0)
        assert.ok(seen.length > 0, `${sc.id}/${seed}: ни одного события из watchFor`)
      }
    }
  })

  test('инварианты Swiss table держатся на каждом тике каждого пресета', () => {
    for (const sc of MAP_SCENARIOS) {
      const sim = run(sc, 3)
      for (const snap of sim.history) checkSwiss(snap.world, `${sc.id} t=${snap.world.tick}`)
    }
  })
})

describe('фаззинг против эталона', () => {
  /** Случайный сценарий из put/get/del/clear по небольшому набору ключей — чтобы были и перезаписи. */
  function randomScript(seed: number, n: number, keys: number): string {
    const r = new Rng(seed)
    const lines: string[] = []
    for (let i = 0; i < n; i++) {
      const k = `key${Math.floor(r.next() * keys)}`
      const x = r.next()
      if (x < 0.55) lines.push(`put ${k} ${i}`)
      else if (x < 0.75) lines.push(`get ${k}`)
      else if (x < 0.995) lines.push(`del ${k}`)
      else lines.push('clear')
    }
    return lines.join('\n')
  }

  const cases: { impl: Impl; max: number; hint: number }[] = [
    { impl: 'swiss', max: 16, hint: 0 },
    { impl: 'swiss', max: 32, hint: 0 },
    { impl: 'swiss', max: 64, hint: 30 },
    { impl: 'legacy', max: 32, hint: 0 },
    { impl: 'legacy', max: 32, hint: 20 },
  ]

  for (const c of cases) {
    test(`${c.impl}, потолок ${c.max}, hint ${c.hint}: содержимое и чтения совпадают с Map`, () => {
      for (let seed = 1; seed <= 12; seed++) {
        const script = randomScript(seed, 260, 90)
        const sim = new Simulation(scenario(script, { impl: c.impl, maxTableSlots: c.max, hint: c.hint }), seed).runToEnd(1000)
        assert.equal(sim.world.finishReason, 'all-done')

        const ref = new Map<string, number>()
        const { ops } = parseScript(script)
        sim.history.slice(2).forEach((snap, i) => {
          const op = ops[i]!
          if (op.kind === 'put') ref.set(op.key, op.val!)
          if (op.kind === 'del') ref.delete(op.key)
          if (op.kind === 'clear') ref.clear()
          if (op.kind === 'get') {
            const e = snap.events.find((x) => x.type === 'get.hit' || x.type === 'get.miss')!
            assert.equal(e.type === 'get.hit' ? e.payload.val : undefined, ref.get(op.key), `seed ${seed}, t=${snap.world.tick}: get ${op.key}`)
          }
          const got = contents(snap.world)
          assert.deepEqual(new Map([...got].sort()), new Map([...ref].sort()), `seed ${seed}, t=${snap.world.tick}: содержимое разошлось`)
          checkSwiss(snap.world, `seed ${seed}, t=${snap.world.tick}`)
        })
      }
    })

    test(`${c.impl}, потолок ${c.max}: обход по правилам спецификации, даже когда карту меняют`, () => {
      for (let seed = 1; seed <= 25; seed++) {
        const r = new Rng(seed * 31)
        const lines = [`fill ${5 + Math.floor(r.next() * 40)}`, 'range']
        for (let i = 0; i < 40; i++) {
          const x = r.next()
          if (x < 0.4) lines.push('next')
          else if (x < 0.7) lines.push(`put n${i} ${1000 + i}`)
          else if (x < 0.85) lines.push(`put k${1 + Math.floor(r.next() * 40)} ${2000 + i}`)
          else lines.push(`del k${1 + Math.floor(r.next() * 40)}`)
        }
        lines.push('next all')
        const sim = new Simulation(scenario(lines.join('\n'), { impl: c.impl, maxTableSlots: c.max, hint: c.hint }), seed).runToEnd(1000)

        const startTick = sim.history.findIndex((s) => s.events.some((e) => e.type === 'iter.start'))
        const atStart = contents(sim.history[startTick]!.world)
        const atEnd = contents(sim.world)
        const yielded = new Map<string, number>()
        const everDeleted = new Set<string>()
        for (const snap of sim.history.slice(startTick + 1)) {
          for (const e of snap.events) {
            if (e.type === 'del.hit') everDeleted.add(String(e.payload.key))
            if (e.type !== 'iter.next' && e.type !== 'iter.reread') continue
            const key = String(e.payload.key)
            assert.ok(!yielded.has(key), `seed ${seed}: ${key} отдан дважды`)
            yielded.set(key, Number(e.payload.val))
            // Отдано текущее значение — то, что лежит в карте на этом тике.
            assert.equal(contents(snap.world).get(key), Number(e.payload.val), `seed ${seed}: ${key} отдан со старым значением`)
          }
        }
        // Всё, что было в начале и дожило до конца, не удаляясь, отдано ровно один раз.
        for (const k of atStart.keys()) {
          if (atEnd.has(k) && !everDeleted.has(k)) assert.ok(yielded.has(k), `seed ${seed}: ${k} потерян обходом`)
        }
      }
    })
  }
})

describe('Swiss table: правила рантайма', () => {
  test('восемь ключей — одна группа, девятый превращает карту в таблицу на 16 слотов', () => {
    const sim = run(promote)
    const promoteAt = sim.history.find((s) => s.events.some((e) => e.type === 'small.promote'))!
    assert.equal(promoteAt.world.swiss!.used, 9)
    const t = Object.values(promoteAt.world.swiss!.tables)[0]!
    assert.equal(t.capacity, 16)
  })

  test('перезапись в полной маленькой карте тоже превращает её в таблицу — как в рантайме', () => {
    const sim = run(scenario('fill 8\nput k1 100'))
    assert.equal(sim.countEvents('small.promote'), 1)
    assert.equal(sim.countEvents('put.update'), 1)
    assert.equal(sim.len, 8)
  })

  test('удаление из полной группы — надгробие, из группы с empty — empty; надгробие переиспользуется', () => {
    for (let seed = 1; seed <= 8; seed++) {
      const sim = run(tombs, seed)
      assert.equal(sim.countEvents('tomb.set'), 2, `seed ${seed}`)
      const dels = sim.events.filter((e) => e.type === 'del.hit')
      assert.deepEqual(dels.map((e) => e.payload.tomb), [true, true, false], `seed ${seed}`)
      assert.equal(sim.countEvents('tomb.reuse'), 1, `seed ${seed}`)
    }
  })

  test('прополка освобождает место, и таблица не растёт, пока может', () => {
    for (let seed = 1; seed <= 8; seed++) {
      const sim = run(prune, seed)
      const pr = sim.events.find((e) => e.type === 'tomb.prune')!
      assert.ok(pr, `seed ${seed}: прополки не было`)
      assert.equal(pr.payload.ok, true, `seed ${seed}`)
      const firstSplit = sim.events.findIndex((e) => e.type === 'table.split' || e.type === 'table.grow')
      assert.ok(firstSplit === -1 || firstSplit > sim.events.indexOf(pr), `seed ${seed}: рост раньше прополки`)
    }
  })

  test('рост: удвоение до потолка, дальше деление; таблицы не больше потолка', () => {
    const sim = run(growth)
    assert.ok(sim.countEvents('table.grow') >= 1)
    assert.ok(sim.countEvents('table.split') >= 1)
    assert.ok(sim.countEvents('dir.grow') >= 1)
    for (const e of sim.events.filter((x) => x.type === 'table.split')) assert.equal(e.payload.capacity, 32)
  })

  test('make с подсказкой: те же вставки без единого роста и перекладывания', () => {
    const a = run(growth)
    const b = run(hint)
    assert.equal(a.len, b.len)
    assert.equal(b.world.stats.moved, 0)
    assert.equal(b.countEvents('table.grow') + b.countEvents('table.split'), 0)
    assert.ok(a.world.stats.moved > a.len, 'без подсказки каждую запись перекладывают больше раза в среднем')
  })

  test('одна вставка перестраивает не больше одной таблицы', () => {
    const sim = new Simulation(scenario('fill 200', { maxTableSlots: 16 })).runToEnd()
    for (const e of sim.events.filter((x) => x.type === 'table.split' || x.type === 'table.grow')) {
      assert.ok(Number(e.payload.moved ?? Number(e.payload.leftN) + Number(e.payload.rightN)) <= 16)
    }
  })

  test('карта не уменьшается; обход пустой карты проверяет каждый слот; соль меняется', () => {
    const sim = run(noShrink)
    const done = sim.events.find((e) => e.type === 'iter.done')!
    assert.equal(done.payload.yielded, 2)
    assert.equal(done.payload.visited, done.payload.capacity)
    assert.ok(Number(done.payload.capacity) >= 64)
    assert.equal(sim.capacity, done.payload.capacity, 'clear не вернул таблицы')
    assert.ok(sim.countEvents('seed.reset') >= 1)
  })

  test('порядок обхода — циклический сдвиг, и он меняется от обхода к обходу', () => {
    const sim = run(order)
    const orders = sim.events.filter((e) => e.type === 'iter.done').map((e) => String(e.payload.order))
    assert.equal(orders.length, 3)
    assert.ok(new Set(orders).size > 1, `все обходы одинаковы: ${orders.join(' | ')}`)
    const base = orders[0]!.split(' ')
    for (const o of orders) {
      const ks = o.split(' ')
      const shift = ks.indexOf(base[0]!)
      assert.deepEqual([...ks.slice(shift), ...ks.slice(0, shift)], base, 'обход — не сдвиг')
    }
  })

  test('рост под итератором: значения перечитываются, удалённое пропускается, новое в новой таблице не видно', () => {
    const sim = run(iterGrow)
    assert.ok(sim.countEvents('iter.reread') > 0)
    assert.equal(sim.countEvents('iter.skip'), 1)
    const keys = sim.events.filter((e) => e.type === 'iter.next' || e.type === 'iter.reread').map((e) => String(e.payload.key))
    assert.ok(!keys.includes('k2'))
    const k1 = sim.events.find((e) => (e.type === 'iter.next' || e.type === 'iter.reread') && e.payload.key === 'k1')
    if (k1) assert.equal(k1.payload.val, 100)
  })

  test('NaN: каждая запись новая, найти и удалить нельзя, обход видит все', () => {
    const sim = run(MAP_SCENARIOS.find((s) => s.id === 'nan')!)
    assert.equal(sim.countEvents('put.new'), 3)
    assert.equal(sim.countEvents('get.miss'), 1)
    assert.equal(sim.countEvents('del.miss'), 1)
    assert.equal(sim.events.find((e) => e.type === 'iter.done')!.payload.yielded, 3)
    assert.equal(sim.len, 0)
  })

  test('nil-карта читается и паникует на записи; гонка — fatal error', () => {
    const nil = run(MAP_SCENARIOS.find((s) => s.id === 'nil')!)
    assert.equal(nil.world.finishReason, 'panic')
    assert.equal(nil.world.panic, 'assignment to entry in nil map')
    assert.equal(nil.countEvents('nil.read'), 3)
    const race = run(MAP_SCENARIOS.find((s) => s.id === 'race')!)
    assert.equal(race.world.finishReason, 'fatal')
    assert.equal(race.world.panic, 'concurrent map writes')
  })

  test('collide действительно подбирает ключи с одной стартовой группой', () => {
    const sim = run(scenario('collide 12', { hint: 40 }))
    const m = sim.world.swiss!
    const groups = new Set(
      sim.world.lastKeys.map((k) => {
        const h = hashKey(k, m.seed)
        const t = m.tables[m.dir[m.globalDepth === 0 ? 0 : h >>> (32 - m.globalDepth)]!]!
        return `${t.id}:${(h >>> 7) & (t.groups.length - 1)}`
      }),
    )
    assert.equal(groups.size, 1)
    assert.equal(sim.world.stats.maxProbe, 2)
  })
})

describe('старая карта: правила рантайма', () => {
  test('рост при средней заполненности 6,5 на корзину', () => {
    const sim = new Simulation(scenario('fill 200', { impl: 'legacy' })).runToEnd()
    for (const e of sim.events.filter((x) => x.type === 'grow.start')) {
      assert.equal(e.payload.sameSize, false)
      const old = Number(e.payload.old)
      assert.ok(Number(e.payload.count) + 1 > Math.max(8, 6.5 * old), `рост раньше времени: ${e.payload.count} на ${old}`)
    }
  })

  test('эвакуация постепенная: не больше двух корзин на запись, чтение во время роста верное', () => {
    const sim = run(legacyEvac)
    for (const snap of sim.history) assert.ok(snap.events.filter((e) => e.type === 'evac').length <= 2)
    assert.ok(sim.countEvents('grow.done') >= 1)
    const start = sim.history.findIndex((s) => s.events.some((e) => e.type === 'grow.start'))
    const done = sim.history.findIndex((s) => s.events.some((e) => e.type === 'grow.done'))
    assert.ok(done > start, 'переезд не растянулся на несколько записей')
    for (const e of sim.events.filter((x) => x.type === 'get.hit' || x.type === 'get.miss')) assert.equal(e.type, 'get.hit')
  })

  test('цепочки переполнения без роста нагрузки ведут к росту того же размера', () => {
    for (let seed = 1; seed <= 6; seed++) {
      const sim = run(legacySameSize, seed)
      const g = sim.events.find((e) => e.type === 'grow.start')!
      assert.ok(g, `seed ${seed}: роста нет`)
      assert.equal(g.payload.sameSize, true, `seed ${seed}`)
      assert.equal(sim.world.legacy!.B, 2, `seed ${seed}: B изменился`)
    }
  })
})
