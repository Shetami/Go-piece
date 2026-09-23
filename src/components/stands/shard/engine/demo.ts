/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:shard add-shard
 *   pnpm demo:shard add-shard scheme=ring
 */
import { Simulation } from './simulation.ts'
import { SHARD_SCENARIOS, shardScenarioById } from './scenarios.ts'
import type { ShardConfig } from './types.ts'
import { imbalance, latencies, movedShare, percentile, resolveConfig } from './world.ts'

const id = process.argv[2] ?? 'add-shard'
const found = shardScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${SHARD_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}
const patch: Record<string, unknown> = {}
for (const arg of process.argv.slice(3)) {
  const [k, v] = arg.split('=')
  if (!k || v === undefined) continue
  patch[k] = v === 'true' ? true : v === 'false' ? false : Number.isFinite(Number(v)) ? Number(v) : v
}
const sc = { ...found, config: { ...found.config, ...(patch as Partial<ShardConfig>) } }
const sim = new Simulation(sc).runToEnd()
console.log(`\n${sc.title}\n${sc.claim}\n${'─'.repeat(78)}`)
const quiet = new Set(['req.route', 'req.done', 'scatter.start', 'scatter.done', 'reshard.move', 'req.migrating'])
for (let t = 1; t < sim.history.length; t++) {
  const f = sim.history[t]!
  const s = sim.log.series[t]!
  const evs = f.events.filter((e) => !quiet.has(e.type)).map((e) => e.type)
  const bars = s.load.map((l) => String(l).padStart(3)).join(' ')
  console.log(`t=${String(t).padStart(3)} +${s.arrived} ok ${s.done} | ${bars} | переезжает ${String(s.migrating).padStart(3)} ${evs.join(' ')}`)
}
const w = sim.world
const st = w.stats
const c = resolveConfig(sc.config)
const all = latencies(sim.log.done, w.tick)
const point = latencies(sim.log.done, w.tick, 0, 'point')
const scat = latencies(sim.log.done, w.tick, 0, 'scatter')
console.log('─'.repeat(78))
console.log(
  `запросов ${st.requests}, в шарды ушло ${st.subRequests}, ответов ${st.done} (веерных ${st.scatter}), к переезжающим ключам ${st.migratingHits}, переехало ключей ${st.moved}, макс очередь ${st.maxQueue}\n` +
    `задержка p50 ${percentile(all, 50)} p99 ${percentile(all, 99)}; точечные p50 ${percentile(point, 50)} p99 ${percentile(point, 99)}; веерные p50 ${percentile(scat, 50)} p99 ${percentile(scat, 99)}\n` +
    `по шардам: ${w.shards.map((s) => `${s.name} ключей ${s.owns} запросов ${s.served}`).join('; ')}; перекос ${imbalance(w).toFixed(2)}×\n` +
    `переехало бы при +1 шарде: mod ${(movedShare({ ...c, scheme: 'mod' }, w.shards.length, w.shards.length + 1) * 100).toFixed(0)}%, ring ${(movedShare({ ...c, scheme: 'ring' }, w.shards.length, w.shards.length + 1) * 100).toFixed(0)}%, range ${(movedShare({ ...c, scheme: 'range' }, w.shards.length, w.shards.length + 1) * 100).toFixed(0)}%`,
)
