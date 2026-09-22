/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:load retry-storm
 *   pnpm demo:load retry-storm cancelOnTimeout=true
 */
import { Simulation } from './simulation.ts'
import { LOAD_SCENARIOS, loadScenarioById } from './scenarios.ts'
import type { LoadConfig } from './types.ts'
import { capacityOf, little, okLatencies, percentile } from './world.ts'

const id = process.argv[2] ?? 'little'
const found = loadScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${LOAD_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}
const patch: Record<string, unknown> = {}
for (const arg of process.argv.slice(3)) {
  const [k, v] = arg.split('=')
  if (!k || v === undefined) continue
  patch[k] = v === 'true' ? true : v === 'false' ? false : Number.isFinite(Number(v)) ? Number(v) : v
}
const sc = { ...found, config: { ...found.config, ...(patch as Partial<LoadConfig>) } }
const sim = new Simulation(sc).runToEnd()
console.log(`\n${sc.title}\n${sc.claim}\n${'─'.repeat(78)}`)

const quiet = new Set(['req.arrive', 'req.start', 'req.done', 'req.retry', 'req.timeout', 'req.wasted', 'req.reject', 'req.fail'])
for (let t = 1; t < sim.history.length; t++) {
  const f = sim.history[t]!
  const s = sim.log.series[t]!
  const bar = '█'.repeat(Math.min(60, s.queue))
  const evs = f.events.filter((e) => !quiet.has(e.type)).map((e) => e.type)
  const lat = okLatencies(sim.log.done, t, t - 10)
  console.log(
    `t=${String(t).padStart(3)} x${s.x} +${s.arrived}${s.retried ? `(${s.retried}↻)` : ''}${s.rejected ? ` ✕${s.rejected}` : ''} ` +
      `busy ${s.busy} ok ${s.ok}${s.wasted ? ` зря ${s.wasted}` : ''}${s.timeouts ? ` ⏱${s.timeouts}` : ''} p99₁₀ ${percentile(lat, 99)} q ${String(s.queue).padStart(3)} ${bar} ${evs.join(' ')}`,
  )
}
const w = sim.world
const st = w.stats
const all = okLatencies(sim.log.done, w.tick)
const l = little(w)
console.log('─'.repeat(78))
console.log(
  `ёмкость ${capacityOf(w.config).toFixed(2)}/т, загрузка ${((st.busyTicks / (w.tick * w.config.workers)) * 100).toFixed(0)}%, ` +
    `запросов ${st.logical}, попыток ${st.arrivals} (повторов ${st.retries}), ответов ${st.ok}, впустую ${st.wasted}, отменено ${st.cancelled}, ` +
    `таймаутов ${st.timeouts}, отказов ${st.rejected}, без ответа ${st.failed}, макс очередь ${st.maxQueue}\n` +
    `задержка p50 ${percentile(all, 50)} p90 ${percentile(all, 90)} p99 ${percentile(all, 99)} max ${percentile(all, 100)}; ` +
    `Литтл: L ${l.L.toFixed(2)} = λ ${l.lambda.toFixed(2)} × W ${l.W.toFixed(2)} = ${(l.lambda * l.W).toFixed(2)}`,
)
