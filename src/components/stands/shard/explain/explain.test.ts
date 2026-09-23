import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { SHARD_SCENARIOS, addShard, hotKey, scatter } from '../engine/scenarios.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import type { ShardScenario } from '../engine/types.ts'
import { explain } from './events.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

/** Пресеты плюс ветки, до которых обычные прогоны не доходят. */
const ALL: ShardScenario[] = [
  ...SHARD_SCENARIOS,
  { ...addShard, id: 'ring-add', config: { ...addShard.config, scheme: 'ring' } },
  { ...addShard, id: 'remove', faults: [{ kind: 'add', at: 40 }, { kind: 'remove', at: 100 }], phases: [{ at: 80, x: 1.6 }] },
  { ...hotKey, id: 'hot-scatter', config: { ...hotKey.config, scatterShare: 20 } },
  { ...scatter, id: 'scatter-wide', config: { ...scatter.config, shards: 8, workers: 1 } },
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
