/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:resil cascade
 *   pnpm demo:resil cascade timeout=8 breaker=true
 */
import { Simulation } from './simulation.ts'
import { RESIL_SCENARIOS, resilScenarioById } from './scenarios.ts'
import type { ResilConfig } from './types.ts'
import { latencies, percentile, successShare } from './world.ts'

const id = process.argv[2] ?? 'cascade'
const found = resilScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${RESIL_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}
const patch: Record<string, unknown> = {}
for (const arg of process.argv.slice(3)) {
  const [k, v] = arg.split('=')
  if (!k || v === undefined) continue
  patch[k] = v === 'true' ? true : v === 'false' ? false : Number.isFinite(Number(v)) ? Number(v) : v
}
const sc = { ...found, config: { ...found.config, ...(patch as Partial<ResilConfig>) } }
const sim = new Simulation(sc).runToEnd()
console.log(`\n${sc.title}\n${sc.claim}\n${'─'.repeat(78)}`)
const quiet = new Set(['req.accept', 'req.done', 'dep.call', 'dep.ok', 'dep.timeout', 'dep.error', 'dep.retry', 'req.fallback', 'req.fail', 'req.reject', 'bulkhead.block'])
for (let t = 1; t < sim.history.length; t++) {
  const f = sim.history[t]!
  const s = sim.log.series[t]!
  const evs = f.events.filter((e) => !quiet.has(e.type)).map((e) => e.type)
  console.log(
    `t=${String(t).padStart(3)} +${s.arrived} ok ${s.ok}${s.degraded ? ` degr ${s.degraded}` : ''}${s.errors ? ` err ${s.errors}` : ''}${s.rejected ? ` rej ${s.rejected}` : ''} | A занято ${s.busy}/${f.world.config.workers} из них ждут B ${s.waiting} очередь ${String(s.queue).padStart(3)} | B ${String(s.depQueue).padStart(3)} | ${f.world.breaker} ${evs.join(' ')}`,
  )
}
const w = sim.world
const st = w.stats
const all = latencies(sim.log.done, w.tick)
const local = latencies(sim.log.done, w.tick, 0, (c) => !c.needsDep)
console.log('─'.repeat(78))
console.log(
  `запросов ${st.arrived}: ответов ${st.ok}, заглушек ${st.degraded}, ошибок ${st.errors}, отказов ${st.rejected} — успешных ${(successShare(w) * 100).toFixed(0)}%\n` +
    `вызовов B ${st.depCalls} (таймаутов ${st.depTimeouts}, ошибок ${st.depErrors}, повторов ${st.retries}, отсечено переборкой ${st.blocked})\n` +
    `задержка p50 ${percentile(all, 50)} p99 ${percentile(all, 99)}; без зависимости p50 ${percentile(local, 50)} p99 ${percentile(local, 99)}; медленных локальных ${st.localSlow}\n` +
    `воркеры A заняты ${((st.busyTicks / (w.tick * w.config.workers)) * 100).toFixed(0)}%, из них ждут B ${((st.waitTicks / Math.max(1, st.busyTicks)) * 100).toFixed(0)}%; макс очередь ${st.maxQueue}`,
)
