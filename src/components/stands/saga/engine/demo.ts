/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:saga dual-write
 *   pnpm demo:saga dual-write publish=outbox
 */
import { Simulation } from './simulation.ts'
import { SAGA_SCENARIOS, sagaScenarioById } from './scenarios.ts'
import type { SagaConfig } from './types.ts'
import { consistency, latencies, outcome, percentile } from './world.ts'

const id = process.argv[2] ?? 'dual-write'
const found = sagaScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${SAGA_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}
const patch: Record<string, unknown> = {}
for (const arg of process.argv.slice(3)) {
  const [k, v] = arg.split('=')
  if (!k || v === undefined) continue
  patch[k] = v === 'true' ? true : v === 'false' ? false : Number.isFinite(Number(v)) ? Number(v) : v
}
const sc = { ...found, config: { ...found.config, ...(patch as Partial<SagaConfig>) } }
const sim = new Simulation(sc).runToEnd()
console.log(`\n${sc.title}\n${sc.claim}\n${'─'.repeat(78)}`)
const quiet = new Set(['order.created', 'outbox.write', 'event.publish', 'event.deliver', 'pay.ok', 'ship.ok', 'relay.send', 'pay.retry', 'pay.fail', 'ship.fail', 'event.duplicate'])
for (let t = 1; t < sim.history.length; t++) {
  const f = sim.history[t]!
  const s = sim.log.series[t]!
  const evs = f.events.filter((e) => !quiet.has(e.type)).map((e) => `${e.type}${e.payload.order !== undefined ? `(#${e.payload.order})` : ''}`)
  if (evs.length === 0 && t % 10 !== 0) continue
  console.log(`t=${String(t).padStart(3)} новых ${s.created} доставлено ${s.shipped} отменено ${s.cancelled} | в пути ${s.inFlight} брокер ${s.inBroker} outbox ${s.outbox} | ${evs.join(' ')}`)
}
const w = sim.world
const st = w.stats
const oc = outcome(w)
const lat = latencies(sim.log.done, w.tick)
console.log('─'.repeat(78))
console.log(
  `заказов ${st.orders}: доставлено ${st.shipped}, отменено ${st.cancelled}, не дошли до конца ${oc.stuck} (застряли навсегда ${st.stuck}) — завершено ${(consistency(w) * 100).toFixed(0)}%\n` +
    `потеряно событий ${st.lost}, дублей доставки ${st.duplicates}, двойных списаний ${st.doubleCharges}, повторов ${st.retries}, сдались ${st.giveups}, компенсаций ${st.compensations}\n` +
    `время заказа до конечного состояния p50 ${percentile(lat, 50)} p99 ${percentile(lat, 99)}; тиков в состоянии «оплачен, но не доставлен» ${st.inconsistentTicks}`,
)
