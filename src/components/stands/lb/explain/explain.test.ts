import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { LB_SCENARIOS, blackHole, hang, slowReplica } from '../engine/scenarios.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import type { LbScenario } from '../engine/types.ts'
import { explain } from './events.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

/** Пресеты плюс ветки, до которых обычные прогоны не доходят. */
const ALL: LbScenario[] = [
  ...LB_SCENARIOS,
  { ...blackHole, id: 'outlier', config: { ...blackHole.config, outlier: true, retries: 1 } },
  { ...blackHole, id: 'all-bad', config: { ...blackHole.config, replicas: 1, outlier: true, retries: 1 }, faults: [{ kind: 'errors', replica: 0, at: 5 }] },
  { ...hang, id: 'crash', faults: [{ kind: 'crash', replica: 2, at: 20, until: 60 }, { kind: 'hang', replica: 1, at: 30, until: 50 }] },
  { ...slowReplica, id: 'slow-recover', faults: [{ kind: 'slow', replica: 3, at: 20, until: 60, factor: 4 }] },
]

test('у каждого события каждого пресета есть разбор без дыр', () => {
  for (const sc of ALL) {
    for (const e of new Simulation(sc).runToEnd().events) {
      const x = explain(e)
      const text = [x.title, ...x.body].join(' ')
      assert.ok(x.body.length > 0, `${sc.id}: пустой разбор ${e.type}`)
      assert.doesNotMatch(text, /undefined|NaN|\[object|Infinity/, `${sc.id}: дыра в разборе ${e.type}: ${x.title}`)
    }
  }
})

test('разборы ссылаются только на существующие термины справочника', () => {
  for (const sc of ALL) {
    for (const e of new Simulation(sc).runToEnd().events) {
      for (const id of explain(e).terms) assert.ok(glossary.has(id), `${e.type} ссылается на термин «${id}», которого нет в справочнике`)
    }
  }
})

test('каждый тип события встречается хотя бы в одном прогоне', () => {
  const seen = new Set<string>()
  for (const sc of ALL) for (const e of new Simulation(sc).runToEnd().events) seen.add(e.type)
  for (const t of Object.keys(EVENT_IMPORTANCE)) assert.ok(seen.has(t), `событие ${t} не случается ни в одном сценарии`)
})
