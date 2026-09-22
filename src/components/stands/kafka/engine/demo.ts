/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:kafka journey
 *   pnpm demo:kafka acks 80
 *   pnpm demo:kafka journey 120 7    — путь сообщения m7
 */
import { Simulation } from './simulation.ts'
import { KAFKA_SCENARIOS, kafkaScenarioById } from './scenarios.ts'
import type { Snapshot } from './types.ts'

const QUIET = new Set(['rec.send', 'repl.fetch', 'consumer.process', 'consumer.fetch'])

function renderTick(snap: Snapshot): string {
  const w = snap.world
  const parts = w.partitions
    .map((p) => {
      const reps = p.replicas
        .map((b) => {
          const alive = w.brokers[b]!.alive
          const tag = p.leader === b ? '*' : p.isr.includes(b) ? '' : '~'
          return `B${b}${tag}${alive ? '' : '✕'}:${p.logs[b]!.log.length}`
        })
        .join(' ')
      return `p${p.id}[hw ${p.hw}] ${reps}`
    })
    .join(' | ')
  const evs = snap.events
    .filter((e) => !QUIET.has(e.type))
    .map((e) => e.type)
    .join(' ')
  return `t=${String(w.tick).padStart(3)} ${parts}\n      ${evs}`
}

const id = process.argv[2] ?? 'journey'
const limit = Number(process.argv[3] ?? 80)
const focus = process.argv[4] ? Number(process.argv[4]) : undefined
const found = kafkaScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${KAFKA_SCENARIOS.map((s) => s.id).join(', ')}`)
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
  `итог: тиков ${sim.tick} (${sim.world.finishReason ?? 'идёт'}), отправлено ${s.produced}, подтверждено ${s.acked}, ` +
    `закоммичено ${s.committed}, потеряно ${s.lost}, неудач ${s.failed}, дублей в логе ${s.duplicates}, отброшено повторов ${s.dedups}, ` +
    `запросов ${s.requests} (в среднем ${sim.avgBatch.toFixed(1)} в пакете), повторов ${s.retries}, ` +
    `ack ${sim.avgAck.toFixed(1)} т., e2e ${sim.avgE2e.toFixed(1)} т., обработано ${s.processed} (повторно ${s.reprocessed}, пропущено ${s.skipped}), ` +
    `ребалансировок ${s.rebalances}, выборов ${s.elections}`,
)
const rec = sim.world.recs[(focus ?? found.focus ?? 1) - 1]
if (rec) {
  console.log(`\nпуть m${rec.id} (ключ ${rec.key ?? '—'}, p${rec.partition}):`)
  for (const t of rec.trace) console.log(`  t=${String(t.tick).padStart(3)} ${t.step.padEnd(10)} ${JSON.stringify(t.data)}`)
}
