/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:chan rendezvous
 *   pnpm demo:chan buffer 80
 */
import { Simulation } from './simulation.ts'
import { CHAN_SCENARIOS, chanScenarioById } from './scenarios.ts'
import type { Snapshot } from './types.ts'

function renderTick(snap: Snapshot): string {
  const w = snap.world
  const chans = w.chans
    .map((c) => {
      const buf = c.isNil ? 'nil' : c.cap === 0 ? '—' : c.buf.map((x) => (x === null ? '·' : '▣')).join('')
      const s = c.sendq.length > 0 ? ` ←${c.sendq.map((x) => `G${x.g}`).join(',')}` : ''
      const r = c.recvq.length > 0 ? ` →${c.recvq.map((x) => `G${x.g}`).join(',')}` : ''
      return `${c.name}[${buf}]${c.closed ? '✕' : ''}${s}${r}`
    })
    .join('  ')
  const slots = w.slots.map((s) => (s.g === null ? '—' : `G${s.g}:${s.op}`)).join(' ')
  const evs = snap.events
    .filter((e) => e.type !== 'g.ready' && e.type !== 'send.buffer' && e.type !== 'recv.buffer')
    .map((e) => e.type)
    .join(' ')
  return `t=${String(w.tick).padStart(3)} ${chans}\n      ${slots}   ${evs}`
}

const id = process.argv[2] ?? 'rendezvous'
const limit = Number(process.argv[3] ?? 60)
const found = chanScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${CHAN_SCENARIOS.map((s) => s.id).join(', ')}`)
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
  `итог: тиков ${sim.tick} (${sim.world.finishReason ?? 'идёт'}${sim.world.panic ? `: ${sim.world.panic}` : ''}), ` +
    `передач ${s.transfers} (из рук в руки ${s.direct}, через буфер ${s.buffered}), ` +
    `парковок ${s.parks}, пробуждений ${s.wakeups}, среднее ожидание ${sim.avgWait.toFixed(1)} тиков, ` +
    `select ${s.selects}/default ${s.selectDefaults}/блокировок ${s.selectBlocks}, спят сейчас ${sim.parked}`,
)
