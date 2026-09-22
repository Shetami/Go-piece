/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:txn lost-update
 *   pnpm demo:txn lost-update isolation=repeatable-read retry=true
 */
import { Simulation } from './simulation.ts'
import { TXN_SCENARIOS, txnScenarioById } from './scenarios.ts'
import { invariantState } from './tick.ts'
import type { Frame, TxnConfig } from './types.ts'
import { labelOf, visibleTo, xidLabel } from './world.ts'

function renderTick(f: Frame): string {
  const w = f.world
  const heap = w.tuples
    .map((t) => `${t.key}=${t.value}[${t.xmin}${t.xmax === null ? '' : `→${t.xmax}${t.lockOnly ? 'L' : ''}`}]`)
    .join(' ')
  const txns = w.txns
    .map((t) => {
      const m = w.marks[t.spec]
      const mk = !m ? '·' : m.kind === 'op' ? `${m.op}${m.key ? ` ${m.key}` : ''}${m.bad ? '!' : ''}` : m.kind
      const vis = t.snapshot ? w.tuples.filter((x) => visibleTo(w, x, t.snapshot!, t.xid)).map((x) => x.value).join('/') : ''
      return `${labelOf(t)}(${t.xid ?? '—'} ${t.state}${t.wait ? ` ждёт ${t.wait.key}` : ''}${vis ? ` видит ${vis}` : ''}): ${mk}`
    })
    .join(' | ')
  const evs = f.events
    .filter((e) => e.type !== 'snap.take' && e.type !== 'txn.xid')
    .map((e) => e.type)
    .join(' ')
  return `t=${String(w.tick).padStart(3)} ${heap}\n      ${txns}\n      ${evs}`
}

const id = process.argv[2] ?? 'lost-update'
const found = txnScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${TXN_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}

// Остальные аргументы — правки настроек: isolation=serializable syncCommit=false
const patch: Record<string, unknown> = {}
for (const arg of process.argv.slice(3)) {
  const [k, v] = arg.split('=')
  if (!k || v === undefined) continue
  patch[k] = v === 'true' ? true : v === 'false' ? false : Number.isFinite(Number(v)) ? Number(v) : v
}
const sc = { ...found, config: { ...found.config, ...(patch as Partial<TxnConfig>) } }

const sim = new Simulation(sc)
console.log(`\n${sc.title}\n${sc.claim}\n${'─'.repeat(78)}`)
for (;;) {
  const f = sim.step()
  if (!f) break
  console.log(renderTick(f))
  if (sim.tick >= (sc.stopAfter ?? 200)) break
}
const w = sim.world
const s = w.stats
const inv = invariantState(w, sc.invariant)
console.log('─'.repeat(78))
console.log(
  `итог: тиков ${sim.tick} (${w.finishReason ?? 'идёт'}), коммитов ${s.commits}, откатов ${s.rollbacks}, ошибок ${s.aborts} ` +
    `(сериализации ${s.serializationFailures}, дедлоков ${s.deadlocks}), повторов ${s.retries}, ожиданий ${s.waits}/${s.waitTicks}т, ` +
    `аномалий ${s.anomalies}, вакуум ${s.vacuumRuns}/${s.vacuumRemoved}, сбросов WAL ${s.flushes}, потеряно коммитов ${s.lostCommits}`,
)
if (inv) console.log(`инвариант: ${inv.ok ? 'соблюдён' : 'НАРУШЕН'} — ${inv.detail}`)
console.log(`xact: ${Object.entries(w.xact).map(([x, st]) => `${xidLabel(w, Number(x))}=${st}`).join(', ')}`)
