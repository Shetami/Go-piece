/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:cache stampede
 *   pnpm demo:cache stampede coalesce=true
 */
import { Simulation } from './simulation.ts'
import { CACHE_SCENARIOS, cacheScenarioById } from './scenarios.ts'
import type { CacheConfig } from './types.ts'
import { idealHitRatio, percentile, readLatencies } from './world.ts'

const id = process.argv[2] ?? 'hot-keys'
const found = cacheScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${CACHE_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}
const patch: Record<string, unknown> = {}
for (const arg of process.argv.slice(3)) {
  const [k, v] = arg.split('=')
  if (!k || v === undefined) continue
  patch[k] = v === 'true' ? true : v === 'false' ? false : Number.isFinite(Number(v)) ? Number(v) : v
}
const sc = { ...found, config: { ...found.config, ...(patch as Partial<CacheConfig>) } }
const sim = new Simulation(sc).runToEnd()
console.log(`\n${sc.title}\n${sc.claim}\n${'─'.repeat(78)}`)
const quiet = new Set(['req.hit', 'req.miss', 'db.done', 'write.done', 'req.join', 'req.swr', 'req.stale'])
for (let t = 1; t < sim.history.length; t++) {
  const f = sim.history[t]!
  const s = sim.log.series[t]!
  const evs = f.events.filter((e) => !quiet.has(e.type) && !(e.type === 'cache.expired' && e.payload.key !== 0)).map((e) => `${e.type}${e.payload.key !== undefined ? `(${e.payload.key})` : ''}`)
  console.log(
    `t=${String(t).padStart(3)} +${s.arrived} hit ${s.hits} miss ${s.misses}${s.stale ? ` stale ${s.stale}` : ''} db→${s.dbStarted} busy ${s.dbBusy} q ${String(s.dbQueue).padStart(3)} ${'█'.repeat(Math.min(50, s.dbQueue))} ${evs.join(' ')}`,
  )
}
const w = sim.world
const st = w.stats
const lat = readLatencies(sim.log.done, w.tick)
console.log('─'.repeat(78))
console.log(
  `чтений ${st.reads}, попаданий ${st.hits} (${Math.round((st.hits / Math.max(1, st.reads)) * 100)}%, идеал ${Math.round(idealHitRatio(w.config) * 100)}%), промахов ${st.misses}, присоединились ${st.joins}, просроченных отдано ${st.swr}, устаревших ${st.stale}\n` +
    `запросов в базу: чтений ${st.dbReads}, записей ${st.dbWrites}; макс очередь ${st.maxDbQueue}, макс за один ключ ${st.maxSameKey}\n` +
    `задержка чтения p50 ${percentile(lat, 50)} p90 ${percentile(lat, 90)} p99 ${percentile(lat, 99)} max ${percentile(lat, 100)}`,
)
