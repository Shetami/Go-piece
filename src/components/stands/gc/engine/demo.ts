/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:gc write-barrier
 *   pnpm demo:gc mem-limit 120
 */
import { Simulation } from './simulation.ts'
import { GC_SCENARIOS, gcScenarioById } from './scenarios.ts'
import type { Snapshot } from './types.ts'

const MARK: Record<string, string> = { white: '·', grey: '▒', black: '█' }
const PHASE: Record<string, string> = {
  off: 'выкл ',
  'stw-start': 'STW▶ ',
  mark: 'разм ',
  'stw-end': 'STW◀ ',
  sweep: 'подм ',
}

function renderTick(snap: Snapshot): string {
  const w = snap.world
  const heap = w.cells.map((c) => (c.used ? (c.lost ? 'X' : MARK[c.color]!) : ' ')).join('')
  const used = w.cells.filter((c) => c.used).length
  const head =
    `t=${String(w.tick).padStart(3)} ${PHASE[w.phase]} ` +
    `куча ${String(used).padStart(3)}/${String(w.goal).padStart(3)} ` +
    `порог ${String(w.trigger).padStart(3)} живое ${String(w.heapMarked).padStart(3)} ` +
    `серых ${String(w.greyq.length).padStart(3)} помощь×${w.assistRatio.toFixed(1)}`
  const slots = w.slots.map((s) => (s.kind === 'mutator' ? `G${s.mut}` : s.kind.slice(0, 4))).join(' ')
  const evs = snap.events
    .filter((e) => e.type !== 'alloc' && e.type !== 'mark.scan' && e.type !== 'assist.work')
    .map((e) => e.type)
    .join(' ')
  return `${head}\n      [${heap}]\n      ${slots}   ${evs}`
}

const id = process.argv[2] ?? 'first-cycle'
const limit = Number(process.argv[3] ?? 80)
const found = gcScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${GC_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}

const sim = new Simulation(found)
console.log(`\n${found.title}\n${found.claim}\n${'─'.repeat(78)}`)
for (let i = 0; i < limit; i++) {
  const snap = sim.step()
  if (!snap) break
  console.log(renderTick(snap))
}
console.log('─'.repeat(78))
console.log(
  `итог: тиков ${sim.tick}, причина ${sim.world.finishReason ?? 'ещё идёт'}, циклов ${sim.cycles}, ` +
    `пик кучи ${sim.peakHeap}, потеряно ${sim.lost}, CPU сборщику ${Math.round(sim.gcCpuShare * 100)}%, ` +
    `полезной работы ${sim.mutatorProgress} тиков`,
)
