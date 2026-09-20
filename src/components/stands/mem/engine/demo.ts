/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:mem contention
 *   pnpm demo:mem large-objects 60
 */
import { Simulation } from './simulation.ts'
import { MEM_SCENARIOS, memScenarioById } from './scenarios.ts'
import type { Snapshot } from './types.ts'

const MARK: Record<string, string> = { free: '·', returned: ' ', span: '▒', large: '█', stack: 'S' }

function renderTick(snap: Snapshot): string {
  const w = snap.world
  const heap = w.pages.map((p) => MARK[p.kind] ?? '?').join('')
  const head =
    `t=${String(w.tick).padStart(3)} страниц ${String(w.pages.filter((p) => p.kind === 'span' || p.kind === 'large' || p.kind === 'stack').length).padStart(3)} ` +
    `спанов ${String(w.spans.filter((s) => s.owner.kind !== 'free').length).padStart(3)} ` +
    `объектов ${String(w.objects.length).padStart(4)} ` +
    `потери ${Math.round(((w.stats.slotBytes - w.stats.askedBytes) / Math.max(1, w.stats.slotBytes)) * 100)}%`
  const slots = w.slots.map((s) => (s.g === null ? '—' : `G${s.g}:${s.path}`)).join(' ')
  const evs = snap.events
    .filter((e) => e.type !== 'alloc.fast' && e.type !== 'obj.free' && e.type !== 'alloc.stack')
    .map((e) => e.type)
    .join(' ')
  return `${head}\n      [${heap}]\n      ${slots}   ${evs}`
}

const id = process.argv[2] ?? 'fast-path'
const limit = Number(process.argv[3] ?? 60)
const found = memScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${MEM_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}

const sim = new Simulation(found)
console.log(`\n${found.title}\n${found.claim}\n${'─'.repeat(78)}`)
for (let i = 0; i < limit; i++) {
  const snap = sim.step()
  if (!snap) break
  console.log(renderTick(snap))
}
const s = sim.world.stats
console.log('─'.repeat(78))
console.log(
  `итог: тиков ${sim.tick} (${sim.world.finishReason ?? 'идёт'}), на стеке ${s.onStack}, мелочью ${s.tiny}, ` +
    `быстрым путём ${s.fast}, refill ${s.refills}, больших ${s.large}, конфликтов ${s.contended}, ` +
    `потери ${Math.round(sim.waste * 100)}%, страниц пик ${s.peakPages}, у ОС ${sim.rss}, ` +
    `ростов стека ${s.stackGrows}`,
)
