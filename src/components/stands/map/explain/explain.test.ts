import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { MAP_SCENARIOS } from '../engine/scenarios.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import type { MapScenario } from '../engine/types.ts'
import { explain } from './events.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

/** Пресеты плюс ветки, до которых они не доходят: обе реализации на всех командах. */
const extra = (id: string, script: string, config: MapScenario['config']): MapScenario => ({
  id,
  title: id,
  claim: id,
  config,
  script,
  watchFor: [],
})
const ALL: MapScenario[] = [
  ...MAP_SCENARIOS,
  extra('legacy-all', 'fill 30\nget k1\nget nope\ndel k1\ndel nope\nput k2 5\nput NaN 1\nget NaN\nrange\nnext all\nclear\nrace read', { impl: 'legacy' }),
  extra('legacy-nil', 'get a\nclear\nput a 1', { impl: 'legacy', nilMap: true }),
  extra('swiss-split-iter', 'fill 30\nrange\nnext 2\nfill 60 z\nnext all\nrace iter', { maxTableSlots: 16 }),
  extra('prune-fail', 'collide 8 0\ndel @last:2\ncollide 8 1\ndel @last:-1\nfill 20', { hint: 12, maxTableSlots: 16 }),
]

test('у каждого события каждого прогона есть разбор без дыр', () => {
  const seen = new Set<string>()
  for (const sc of ALL) {
    for (const seed of [1, 2, 3]) {
      const sim = new Simulation(sc, seed).runToEnd()
      for (const e of sim.events) {
        seen.add(e.type)
        const x = explain(e)
        const text = [x.title, ...x.body].join(' ')
        assert.ok(x.body.length > 0, `${sc.id}: пустой разбор ${e.type}`)
        assert.doesNotMatch(text, /undefined|null|Infinity|\[object/, `${sc.id}: дыра в разборе ${e.type}: ${x.title}`)
        // Слово NaN законно только там, где речь о ключе NaN; иначе это сломанное число.
        const aboutNaN = e.payload.nan || Object.values(e.payload).includes('NaN') || String(e.payload.order ?? '').split(' ').includes('NaN')
        if (!aboutNaN) assert.doesNotMatch(text, /NaN/, `${sc.id}: NaN в разборе ${e.type}: ${x.title}`)
      }
    }
  }
  const never = Object.keys(EVENT_IMPORTANCE).filter((t) => !seen.has(t))
  assert.deepEqual(never, [], 'события, которые ни разу не случились, — разбор не проверен')
})

test('разборы ссылаются только на существующие термины справочника', () => {
  for (const sc of ALL) {
    const sim = new Simulation(sc).runToEnd()
    for (const e of sim.events) {
      for (const t of explain(e).terms) assert.ok(glossary.has(t), `нет термина ${t} (${e.type})`)
    }
  }
})
