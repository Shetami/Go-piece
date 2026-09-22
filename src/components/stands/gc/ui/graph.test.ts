import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from '../engine/simulation.ts'
import { GC_SCENARIOS, writeBarrier } from '../engine/scenarios.ts'
import type { GcScenario, GcWorld } from '../engine/types.ts'
import { MAX_NODES, MIN_GAP, NODE_R, buildGraph } from './graph.ts'

/**
 * Граф — не украшение, а утверждение о мире: «живой объект это тот, до которого
 * можно дойти от корней». Тесты закрепляют ровно это, плюс требования к укладке,
 * без которых по картинке нельзя следить глазами: детерминизм и отсутствие
 * наложений.
 */

const run = (s: GcScenario, seed = 1) => new Simulation(s, seed).runToEnd(s.stopAfter)
const noBarrier: GcScenario = { ...writeBarrier, config: { ...writeBarrier.config, writeBarrier: false } }

/** Достижимость, посчитанная в лоб — эталон для проверки укладки. */
function reachable(w: GcWorld): Set<number> {
  const roots = w.muts
    .filter((m) => m.state !== 'done')
    .flatMap((m) => m.stack)
    .concat(w.globals)
    .filter((x): x is number => x !== null)
  const seen = new Set<number>()
  const stack = [...roots]
  while (stack.length > 0) {
    const id = stack.pop()!
    if (seen.has(id) || !w.cells[id]?.used) continue
    seen.add(id)
    for (const s of w.cells[id]!.slots) if (s !== null) stack.push(s)
  }
  return seen
}

describe('укладка графа кучи', () => {
  test('одинаковый снимок даёт одинаковые координаты', () => {
    for (const sc of GC_SCENARIOS) {
      const sim = run(sc)
      const snap = sim.history[Math.floor(sim.history.length / 2)]!
      assert.deepEqual(buildGraph(snap.world), buildGraph(snap.world), `${sc.id}: укладка не детерминирована`)
    }
  })

  test('мусор — это в точности недостижимое от корней', () => {
    for (const sc of GC_SCENARIOS) {
      const sim = run(sc)
      for (const snap of sim.history) {
        const g = buildGraph(snap.world)
        const live = reachable(snap.world)
        const used = snap.world.cells.filter((c) => c.used).length
        assert.equal(g.garbage, used - live.size, `${sc.id} тик ${snap.world.tick}: мусор посчитан неверно`)
        for (const n of g.nodes) {
          assert.equal(n.depth >= 0, live.has(n.id), `${sc.id}: блок #${n.id} отнесён не в ту группу`)
        }
      }
    }
  })

  test('узлы не наезжают друг на друга', () => {
    for (const sc of GC_SCENARIOS) {
      const sim = run(sc)
      for (const snap of sim.history) {
        const seen = new Set<string>()
        for (const n of buildGraph(snap.world).nodes) {
          const key = `${Math.round(n.x)}:${Math.round(n.y)}`
          assert.ok(!seen.has(key), `${sc.id} тик ${snap.world.tick}: два блока в одной точке ${key}`)
          seen.add(key)
        }
      }
    }
  })

  test('на схему попадает не больше MAX_NODES блоков, и режется именно мусор', () => {
    for (const sc of GC_SCENARIOS) {
      const sim = run(sc)
      for (const snap of sim.history) {
        const g = buildGraph(snap.world)
        assert.ok(g.shown <= MAX_NODES, `${sc.id}: на схеме ${g.shown} узлов`)
        assert.equal(g.shown + g.hidden, g.total, `${sc.id}: показано плюс скрыто не сходится с занятым`)
        // Достижимое прячем только если его самого больше лимита — иначе смысл картинки теряется.
        const live = reachable(snap.world).size
        if (live <= MAX_NODES) {
          assert.equal(
            g.nodes.filter((n) => n.depth >= 0).length,
            live,
            `${sc.id} тик ${snap.world.tick}: спрятали достижимый блок`,
          )
        }
      }
    }
  })

  test('ребро из чёрного в белое — только при выключенном барьере', () => {
    for (const snap of run(writeBarrier).history) {
      assert.equal(
        buildGraph(snap.world).violations,
        0,
        `тик ${snap.world.tick}: барьер включён, а инвариант нарушен`,
      )
    }
    const broken = run(noBarrier).history.filter((s) => buildGraph(s.world).violations > 0)
    assert.ok(broken.length > 0, 'без барьера нарушение инварианта так и не показалось на графе')
  })

  test('нарушение инварианта предшествует потере живой памяти', () => {
    const sim = run(noBarrier)
    const firstViolation = sim.history.find((s) => buildGraph(s.world).violations > 0)?.world.tick
    const firstLost = sim.events.find((e) => e.type === 'heap.lost')?.tick
    assert.ok(firstViolation !== undefined, 'нарушения не было')
    assert.ok(firstLost !== undefined, 'память так и не потерялась')
    assert.ok(
      firstViolation <= firstLost,
      `сначала потеря (${firstLost}), потом нарушение (${firstViolation}) — граф показывает следствие раньше причины`,
    )
  })

  test('цепочка без ветвлений идёт вертикально', () => {
    let exact = 0
    let total = 0
    let worst = 0
    for (const sc of GC_SCENARIOS) {
      for (const snap of run(sc).history) {
        const w = snap.world
        const g = buildGraph(w)
        const at = new Map(g.nodes.map((n) => [n.id, n]))
        for (const n of g.nodes) {
          if (n.depth < 0) continue
          const kids = (w.cells[n.id]?.slots ?? [])
            .filter((s): s is number => s !== null && at.has(s))
            .filter((s) => at.get(s)!.depth === n.depth + 1)
          if (kids.length !== 1) continue
          const kid = at.get(kids[0]!)!
          // У единственного потомка должен быть единственный родитель — иначе
          // он честно тянется к середине нескольких и стоять под нами не обязан.
          const parents = w.cells.filter(
            (c) => c.used && at.has(c.id) && at.get(c.id)!.depth === n.depth && c.slots.includes(kid.id),
          )
          if (parents.length !== 1) continue
          total++
          const dx = Math.abs(kid.x - n.x)
          if (dx < 0.5) exact++
          worst = Math.max(worst, dx)
          assert.ok(
            dx <= MIN_GAP * 2,
            `${sc.id} тик ${w.tick}: #${n.id} → #${kid.id} снесло на ${dx.toFixed(0)}px — это уже не цепочка`,
          )
        }
      }
    }
    assert.ok(total > 1000, `проверять было почти нечего: ${total} цепочек`)
    // Идеально вертикальной цепочка быть может не всегда: если на уровне тесно,
    // место приходится делить. Но это должно быть исключением, а не правилом.
    assert.ok(
      exact / total >= 0.9,
      `строго вертикальных всего ${Math.round((exact / total) * 100)}% (худший снос ${worst.toFixed(0)}px)`,
    )
  })

  test('рёбра приходят в узел по линии центров, а не в край', () => {
    for (const sc of GC_SCENARIOS) {
      const sim = run(sc)
      for (const snap of sim.history) {
        const g = buildGraph(snap.world)
        const at = new Map(g.nodes.map((n) => [n.id, n]))
        for (const e of g.edges) {
          if (e.from === null) continue
          const a = at.get(e.from)!
          const b = at.get(e.to)!
          const len = Math.hypot(b.x - a.x, b.y - a.y)
          if (len < 0.001) continue
          // Конец ребра обязан лежать ровно на окружности целевого узла.
          assert.ok(
            Math.abs(Math.hypot(e.x2 - b.x, e.y2 - b.y) - NODE_R) < 0.01,
            `${sc.id}: ребро #${e.from}→#${e.to} приходит не на границу узла`,
          )
          assert.ok(
            Math.abs(Math.hypot(e.x1 - a.x, e.y1 - a.y) - NODE_R) < 0.01,
            `${sc.id}: ребро #${e.from}→#${e.to} выходит не с границы узла`,
          )
        }
      }
    }
  })

  test('корни подписаны: стек каждой живой горутины плюс глобальные', () => {
    const sim = run(GC_SCENARIOS[0]!)
    const snap = sim.history[sim.history.length - 1]!
    const g = buildGraph(snap.world)
    const alive = snap.world.muts.filter((m) => m.state !== 'done').length
    assert.equal(g.roots.filter((r) => r.mut !== null).length, alive)
    assert.ok(g.roots.some((r) => r.mut === null && r.label === 'глоб.'), 'глобальные корни не показаны')
  })
})
