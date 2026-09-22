import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

import { Simulation } from '../engine/simulation.ts'
import { KAFKA_SCENARIOS, acks, group, retry } from '../engine/scenarios.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import type { KafkaScenario } from '../engine/types.ts'
import { explain } from './events.ts'
import { describeStep } from './journey.ts'

const glossaryDir = fileURLToPath(new URL('../../../../content/glossary/', import.meta.url))
const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

/** Пресеты плюс ветки, до которых обычные прогоны не доходят. */
const ALL: KafkaScenario[] = [
  ...KAFKA_SCENARIOS,
  // Тот же потерянный ответ, но с идемпотентностью: брокер узнаёт повтор.
  { ...retry, id: 'retry-idempotent', config: { ...retry.config, idempotence: true } },
  // Автокоммит: упавший потребитель уносит с собой необработанное.
  { ...group, id: 'group-auto', config: { ...group.config, commitMode: 'auto' } },
  // Нечистые выборы и партиция без лидера.
  {
    ...acks,
    id: 'offline',
    config: { ...acks.config, replicationFactor: 2, brokers: 2 },
    faults: [
      { at: 5, kind: 'broker.down', broker: 1 },
      { at: 12, kind: 'broker.down', broker: 0 },
      { at: 30, kind: 'broker.up', broker: 0 },
      { at: 34, kind: 'broker.up', broker: 1 },
    ],
  },
  // acks=0: запрос в мёртвого брокера пропадает молча.
  {
    ...acks,
    id: 'acks0',
    config: { ...acks.config, acks: 0 },
  },
]

const HOLES = /undefined|NaN|\bm\?|\bB\?|Infinity|\[object/

test('у каждого события каждого пресета есть разбор без дыр', () => {
  for (const sc of ALL) {
    const sim = new Simulation(sc).runToEnd(sc.stopAfter ?? 300)
    for (const e of sim.events) {
      const x = explain(e)
      const text = [x.title, ...x.body].join(' ')
      assert.ok(x.body.length > 0, `${sc.id}: пустой разбор ${e.type}`)
      assert.doesNotMatch(text, HOLES, `${sc.id}: дыра в разборе ${e.type}: ${x.title}`)
    }
  }
})

test('у каждого шага пути каждого сообщения есть описание без дыр', () => {
  for (const sc of ALL) {
    const sim = new Simulation(sc).runToEnd(sc.stopAfter ?? 300)
    for (const r of sim.world.recs) {
      for (const t of r.trace) {
        const d = describeStep(t, r)
        assert.doesNotMatch(`${d.title} ${d.body}`, HOLES, `${sc.id}: m${r.id} шаг ${t.step}: ${d.title}`)
      }
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

test('партиция без лидера разбирается отдельно', () => {
  const sim = new Simulation(ALL.find((s) => s.id === 'offline')!).runToEnd(140)
  const offline = sim.events.find((e) => e.type === 'leader.elect' && e.payload.to === null)
  assert.ok(offline, 'партиция так и не осталась без лидера')
  assert.match(explain(offline).title, /без лидера/)
})
