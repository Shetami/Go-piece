/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:lb slow-replica
 *   pnpm demo:lb slow-replica algo=least-conn
 */
import { Simulation } from './simulation.ts'
import { LB_SCENARIOS, lbScenarioById } from './scenarios.ts'
import type { LbConfig } from './types.ts'
import { okLatencies, percentile } from './world.ts'

const id = process.argv[2] ?? 'slow-replica'
const found = lbScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${LB_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}
const patch: Record<string, unknown> = {}
for (const arg of process.argv.slice(3)) {
  const [k, v] = arg.split('=')
  if (!k || v === undefined) continue
  patch[k] = v === 'true' ? true : v === 'false' ? false : Number.isFinite(Number(v)) ? Number(v) : v
}
const sc = { ...found, config: { ...found.config, ...(patch as Partial<LbConfig>) } }
const sim = new Simulation(sc).runToEnd()
console.log(`\n${sc.title}\n${sc.claim}\n${'─'.repeat(78)}`)

const quiet = new Set(['req.route', 'req.done', 'req.error', 'req.timeout', 'req.retry', 'req.fail', 'health.fail', 'req.nobackend'])
for (let t = 1; t < sim.history.length; t++) {
  const f = sim.history[t]!
  const s = sim.log.series[t]!
  const reps = f.world.replicas
    .map((r) => {
      const mark = r.state === 'booting' ? 'b' : !r.inRotation ? 'x' : r.ejectedUntil > t ? 'o' : r.state === 'up' ? ' ' : r.state[0]
      return `${mark}${String(s.load[r.id] ?? 0).padStart(3)}`
    })
    .join(' ')
  const evs = f.events.filter((e) => !quiet.has(e.type)).map((e) => `${e.type}(${String(e.payload.replica ?? '')})`)
  console.log(`t=${String(t).padStart(3)} +${s.arrived} ok ${s.ok}${s.errors ? ` err ${s.errors}` : ''}${s.timeouts ? ` ⏱${s.timeouts}` : ''} | ${reps} | ${evs.join(' ')}`)
}
const w = sim.world
const st = w.stats
const lat = okLatencies(sim.log.done, w.tick)
console.log('─'.repeat(78))
console.log(
  `запросов ${st.logical}, попыток ${st.attempts}, ответов ${st.ok}, ошибок ${st.errors}, таймаутов ${st.timeouts}, без ответа ${st.failed}, впустую ${st.wasted}, макс очередь ${st.maxQueue}\n` +
    `задержка p50 ${percentile(lat, 50)} p90 ${percentile(lat, 90)} p99 ${percentile(lat, 99)} max ${percentile(lat, 100)}\n` +
    `по репликам: ${w.replicas.map((r) => `${r.name} отправлено ${r.routed} ответов ${r.served} ошибок ${r.errors}`).join('; ')}`,
)
