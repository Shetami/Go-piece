import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { TXN_SCENARIOS, deadlock, durability, lostUpdate, phantom } from '../engine/scenarios.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import type { TxnScenario } from '../engine/types.ts'
import { explain } from './events.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

/** Пресеты плюс ветки, до которых обычные прогоны не доходят. */
const ALL: TxnScenario[] = [
  ...TXN_SCENARIOS,
  { ...lostUpdate, id: 'for-update', config: { ...lostUpdate.config, rmw: 'for-update' } },
  { ...lostUpdate, id: 'rr', config: { ...lostUpdate.config, isolation: 'repeatable-read', retry: true } },
  { ...deadlock, id: 'deadlock-retry', config: { ...deadlock.config, retry: true } },
  { ...durability, id: 'async', config: { ...durability.config, syncCommit: false, checkpointEvery: 6 } },
  // Удаление, вставка того же ключа и строка, которой нет.
  {
    ...phantom,
    id: 'delete-insert',
    sessions: [
      {
        name: 'T1',
        at: 1,
        ops: [
          { kind: 'delete', key: 'o1' },
          { kind: 'update', key: 'nope', set: { kind: 'const', value: 1 } },
          { kind: 'commit', after: 2 },
        ],
      },
      { name: 'T2', at: 2, ops: [{ kind: 'insert', key: 'o2', value: 1 }, { kind: 'commit' }] },
      { name: 'T3', at: 2, ops: [{ kind: 'update', key: 'o1', set: { kind: 'delta', d: 1 } }, { kind: 'commit' }] },
    ],
  },
]

test('у каждого события каждого пресета есть разбор без дыр', () => {
  for (const sc of ALL) {
    const sim = new Simulation(sc).runToEnd(sc.stopAfter ?? 200)
    for (const e of sim.events) {
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
      for (const id of explain(e).terms) {
        assert.ok(glossary.has(id), `${e.type} ссылается на термин «${id}», которого нет в src/content/glossary`)
      }
    }
  }
})

test('каждый тип события встречается хотя бы в одном прогоне', () => {
  const seen = new Set<string>()
  for (const sc of ALL) {
    for (const e of new Simulation(sc).runToEnd(sc.stopAfter ?? 200).events) seen.add(e.type)
  }
  for (const t of Object.keys(EVENT_IMPORTANCE)) {
    assert.ok(seen.has(t), `событие ${t} не случается ни в одном сценарии — его некому проверить`)
  }
})

test('у каждой причины прерывания свой разбор', () => {
  const reasons = new Map<string, string>()
  for (const sc of ALL) {
    for (const e of new Simulation(sc).runToEnd(sc.stopAfter ?? 200).events) {
      if (e.type === 'txn.abort') reasons.set(String(e.payload.reason), explain(e).title)
    }
  }
  for (const r of ['serialization', 'ssi', 'deadlock', 'unique', 'crash']) {
    assert.ok(reasons.has(r), `ни разу не случилось прерывание по причине ${r}`)
  }
  assert.equal(new Set(reasons.values()).size, reasons.size, 'у разных причин одинаковые заголовки')
})
