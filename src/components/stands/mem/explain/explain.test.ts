import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { MEM_SCENARIOS } from '../engine/scenarios.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import { explain } from './events.ts'
import type { MemScenario } from '../engine/types.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

/** Пресеты плюс тесная арена — до нехватки памяти обычный прогон не доходит. */
const ALL: MemScenario[] = [
  ...MEM_SCENARIOS,
  { ...MEM_SCENARIOS[3]!, id: 'tight', config: { ...MEM_SCENARIOS[3]!.config, heapPages: 40 } },
]

test('у каждого события каждого пресета есть разбор без дыр', () => {
  for (const sc of ALL) {
    const sim = new Simulation(sc).runToEnd(sc.stopAfter ?? 300)
    for (const e of sim.events) {
      const x = explain(e)
      const text = [x.title, ...x.body].join(' ')
      assert.ok(x.body.length > 0, `${sc.id}: пустой разбор ${e.type}`)
      assert.doesNotMatch(text, /undefined|NaN|\bG\?|\bP\?/, `${sc.id}: дыра в разборе ${e.type}: ${x.title}`)
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
