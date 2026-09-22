import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { CHAN_SCENARIOS, closing, deadlock } from '../engine/scenarios.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import { explain } from './events.ts'
import type { ChanScenario } from '../engine/types.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

/** Пресеты плюс две ветки, до которых обычный прогон не доходит. */
const ALL: ChanScenario[] = [
  ...CHAN_SCENARIOS,
  // Повторное закрытие: кто-то закрывает канал, который уже закрыт.
  {
    ...closing,
    id: 'double-close',
    workloads: [
      ...closing.workloads,
      { name: 'second-boss', count: 1, spawnAt: 0, phases: [{ kind: 'cpu', ticks: 14 }, { kind: 'close', chan: 'work' }], repeat: 1 },
    ],
  },
  // Закрытие nil-канала — тоже паника.
  {
    ...deadlock,
    id: 'close-nil',
    workloads: [
      { name: 'closer', count: 1, spawnAt: 0, phases: [{ kind: 'close', chan: 'never' }], repeat: 1 },
    ],
  },
]

test('у каждого события каждого пресета есть разбор без дыр', () => {
  for (const sc of ALL) {
    const sim = new Simulation(sc).runToEnd(sc.stopAfter ?? 300)
    for (const e of sim.events) {
      const x = explain(e)
      const text = [x.title, ...x.body].join(' ')
      assert.ok(x.body.length > 0, `${sc.id}: пустой разбор ${e.type}`)
      assert.doesNotMatch(text, /undefined|NaN|\bG\?|Infinity/, `${sc.id}: дыра в разборе ${e.type}: ${x.title}`)
    }
  }
})

test('разборы ссылаются только на существующие термины справочника', () => {
  for (const sc of ALL) {
    const sim = new Simulation(sc).runToEnd(sc.stopAfter ?? 300)
    for (const e of sim.events) {
      for (const id of explain(e).terms) {
        assert.ok(glossary.has(id), `${e.type} ссылается на термин «${id}», которого нет в src/content/glossary`)
      }
    }
  }
})

test('каждый тип события встречается хотя бы в одном прогоне', () => {
  const seen = new Set<string>()
  for (const sc of ALL) {
    for (const e of new Simulation(sc).runToEnd(sc.stopAfter ?? 300).events) seen.add(e.type)
  }
  for (const t of Object.keys(EVENT_IMPORTANCE)) {
    assert.ok(seen.has(t), `событие ${t} не случается ни в одном сценарии — его некому проверить`)
  }
})

test('обе ветки паники разбираются отдельно и по делу', () => {
  const double = new Simulation(ALL.find((s) => s.id === 'double-close')!).runToEnd(80)
  const again = double.events.find((e) => e.type === 'chan.close' && e.payload.again === true)
  assert.ok(again, 'повторное закрытие не случилось')
  assert.match(explain(again).title, /уже закрытый/)

  const nil = new Simulation(ALL.find((s) => s.id === 'close-nil')!).runToEnd(40)
  const ev = nil.events.find((e) => e.type === 'chan.nil' && e.payload.op === 'close')
  assert.ok(ev, 'закрытие nil-канала не случилось')
  assert.match(explain(ev).title, /Паника/)
})
