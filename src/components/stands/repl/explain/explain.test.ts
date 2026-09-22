import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { REPL_SCENARIOS, conflict, failover, readYourWrites, slot, stream } from '../engine/scenarios.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import type { ReplScenario } from '../engine/types.ts'
import { explain } from './events.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

/** Пресеты плюс ветки, до которых обычные прогоны не доходят. */
const ALL: ReplScenario[] = [
  ...REPL_SCENARIOS,
  { ...conflict, id: 'feedback', config: { ...conflict.config, hotStandbyFeedback: true } },
  { ...stream, id: 'slow', faults: [{ kind: 'replica-slow', replica: 1, at: 5, until: 15, rate: 1 }] },
  {
    ...slot,
    id: 'no-slot-read',
    config: { ...slot.config, slots: false },
    clients: [...slot.clients, { name: 'R', at: 38, ops: [{ kind: 'read', key: 'a', from: 1 }] }],
  },
  {
    ...readYourWrites,
    id: 'primary-read',
    clients: readYourWrites.clients.map((c) => ({ ...c, ops: c.ops.map((o) => (o.kind === 'read' ? { ...o, from: 'primary' as const } : o)) })),
  },
  failover,
]

test('у каждого события каждого пресета есть разбор без дыр', () => {
  for (const sc of ALL) {
    for (const e of new Simulation(sc).runToEnd(sc.stopAfter ?? 200).events) {
      const x = explain(e)
      const text = [x.title, ...x.body].join(' ')
      assert.ok(x.body.length > 0, `${sc.id}: пустой разбор ${e.type}`)
      assert.doesNotMatch(text, /undefined|NaN|\[object|Infinity/, `${sc.id}: дыра в разборе ${e.type}: ${x.title}`)
    }
  }
})

test('разборы ссылаются только на существующие термины справочника', () => {
  for (const sc of ALL) {
    for (const e of new Simulation(sc).runToEnd(sc.stopAfter ?? 200).events) {
      for (const id of explain(e).terms) assert.ok(glossary.has(id), `${e.type} ссылается на термин «${id}», которого нет в справочнике`)
    }
  }
})

test('каждый тип события встречается хотя бы в одном прогоне', () => {
  const seen = new Set<string>()
  for (const sc of ALL) for (const e of new Simulation(sc).runToEnd(sc.stopAfter ?? 200).events) seen.add(e.type)
  for (const t of Object.keys(EVENT_IMPORTANCE)) assert.ok(seen.has(t), `событие ${t} не случается ни в одном сценарии`)
})
