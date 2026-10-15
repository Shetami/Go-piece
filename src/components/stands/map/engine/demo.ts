/**
 * Текстовый прогон сценария — чтобы смотреть на движок без UI.
 *
 *   pnpm demo:map growth
 *   pnpm demo:map legacy-evac 80
 */
import { Simulation } from './simulation.ts'
import { MAP_SCENARIOS, mapScenarioById } from './scenarios.ts'
import { CTRL_DELETED, CTRL_EMPTY, type MapWorld } from './types.ts'
import { hex2 } from './hash.ts'

function renderMap(w: MapWorld): string {
  if (w.swiss) {
    const m = w.swiss
    const ctrl = (c: number) => (c === CTRL_EMPTY ? '..' : c === CTRL_DELETED ? 'xx' : hex2(c))
    if (m.dir.length === 0) return `small[${m.small ? m.small.slots.map((s) => ctrl(s.ctrl)).join(' ') : '—'}]`
    const tables = Object.values(m.tables).map(
      (t) =>
        `T${t.id}${t.replaced ? '(old)' : ''} d${t.localDepth} cap${t.capacity} used${t.used} left${t.growthLeft}: ` +
        t.groups.map((g) => g.slots.map((s) => ctrl(s.ctrl)).join('')).join('|'),
    )
    return `dir[${m.dir.join(',')}] gd${m.globalDepth}\n      ${tables.join('\n      ')}`
  }
  if (w.legacy) {
    const h = w.legacy
    const arr = (id: number | null) =>
      id === null ? '—' : h.arrays[id]!.map((b) => b.pages.map((p) => p.filter((s) => s.top > 4).length).join('+') + (b.evacuated ? '✓' : '')).join(' ')
    return `B${h.B} count${h.count} nov${h.noverflow} [${arr(h.buckets)}]${h.oldbuckets !== null ? ` old[${arr(h.oldbuckets)}] nevac${h.nevacuate}` : ''}`
  }
  return 'nil'
}

const id = process.argv[2] ?? 'small'
const limit = Number(process.argv[3] ?? 200)
const found = mapScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${MAP_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}

const sim = new Simulation(found)
if (sim.errors.length) console.error(sim.errors)
console.log(`\n${found.title}\n${found.claim}\n${'─'.repeat(78)}`)
for (let i = 0; i < limit; i++) {
  const snap = sim.step()
  if (!snap) break
  const evs = snap.events.map((e) => e.type).join(' ')
  console.log(`t=${String(snap.world.tick).padStart(3)} ${snap.trace?.code.padEnd(28)} ${snap.trace?.outcome}  [${evs}]`)
  console.log(`      ${renderMap(snap.world)}`)
}
const s = sim.world.stats
console.log('─'.repeat(78))
console.log(`итог: ${sim.world.finishReason}${sim.world.panic ? `: ${sim.world.panic}` : ''}; len ${sim.len}, ёмкость ${sim.capacity}; ` +
  `проб ${s.probes}, сравнений ${s.compares} (ложных ${s.falseMatches}), макс. проб ${s.maxProbe}, переложено ${s.moved}, ` +
  `рост ${s.grows}, деление ${s.splits}, каталог ${s.dirGrows}, прополка ${s.prunes}, переполнений ${s.overflows}, эвакуировано ${s.evacuated}`)
