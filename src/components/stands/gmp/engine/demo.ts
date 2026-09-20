/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo skew-and-stealing
 *   pnpm demo blocking-syscalls 60
 */
import { Simulation } from './simulation.ts'
import { SCENARIOS, scenarioById } from './scenarios.ts'
import type { Snapshot, World } from './types.ts'

const STATE_MARK: Record<string, string> = {
  running: '▶',
  runnable: '·',
  waiting: '~',
  syscall: '!',
  dead: ' ',
}

function renderP(w: World, pid: number): string {
  const p = w.ps[pid]!
  const m = p.m === null ? '  —' : `M${p.m}`.padStart(3)
  const g = p.m !== null && w.ms[p.m - 1]?.g ? `G${w.ms[p.m - 1]!.g}` : '—'
  const next = p.runnext === null ? '' : `[G${p.runnext}]`
  return `P${pid}(${p.state[0]}) ${m} ${g.padEnd(4)} ${next.padEnd(6)} q:${String(p.runq.length).padStart(2)}`
}

function renderTick(snap: Snapshot): string {
  const w = snap.world
  const head = `t=${String(w.tick).padStart(4)} | ` + w.ps.map((p) => renderP(w, p.id)).join(' | ')
  const gs = w.gs
    .filter((g) => g.state !== 'dead')
    .slice(0, 16)
    .map((g) => `${STATE_MARK[g.state]}G${g.id}`)
    .join(' ')
  const evs = snap.events
    .filter((e) => e.type !== 'p.idle')
    .map((e) => `${e.type}${e.actors.g?.length ? `(G${e.actors.g[0]})` : ''}`)
    .join(' ')
  const tail = [
    `glob:${w.globrunq.length}`,
    `net:${w.netpoll.length}`,
    `M:${w.ms.length}`,
  ].join(' ')
  return `${head}\n        ${tail}  ${gs}\n        ${evs || '—'}`
}

const id = process.argv[2] ?? 'skew-and-stealing'
const limit = Number(process.argv[3] ?? 40)
const found = scenarioById(id)

if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}
const scenario = found

const sim = new Simulation(scenario)
console.log(`\n${scenario.title}\n${scenario.claim}\n${'─'.repeat(78)}`)
for (let i = 0; i < limit; i++) {
  const snap = sim.step()
  if (!snap) break
  console.log(renderTick(snap))
}
console.log('─'.repeat(78))
console.log(
  `итог: тиков ${sim.tick}, причина ${sim.world.finishReason ?? 'ещё идёт'}, ` +
    `потоков пик ${sim.peakThreads}, краж ${sim.countEvents('p.stole')}, ` +
    `вытеснений ${sim.countEvents('sysmon.preempt')}, отборов P ${sim.countEvents('sysmon.retake')}`,
)
