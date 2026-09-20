import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { SCENARIOS } from '../engine/scenarios.ts'
import { explain } from './events.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

test('у каждого события каждого пресета есть разбор без дыр', () => {
  for (const sc of SCENARIOS) {
    const sim = new Simulation(sc).runToEnd(sc.stopAfter ?? 1000)
    for (const e of sim.events) {
      const x = explain(e)
      const text = [x.title, ...x.body].join(' ')
      assert.ok(x.body.length > 0, `${sc.id}: пустой разбор ${e.type}`)
      assert.doesNotMatch(text, /undefined|NaN|\bG\?|\bM\?|\bP\?/, `${sc.id}: дыра в разборе ${e.type}: ${x.title}`)
    }
  }
})

test('разборы ссылаются только на существующие термины справочника', () => {
  for (const sc of SCENARIOS) {
    const sim = new Simulation(sc).runToEnd(sc.stopAfter ?? 1000)
    for (const e of sim.events) {
      for (const id of explain(e).terms) {
        assert.ok(glossary.has(id), `${e.type} ссылается на термин «${id}», которого нет в src/content/glossary`)
      }
    }
  }
})
