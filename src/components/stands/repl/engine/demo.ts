/**
 * Текстовый прогон сценария — чтобы смотреть на движок до того, как появится UI.
 *
 *   pnpm demo:repl failover
 *   pnpm demo:repl failover standbys=any-1
 */
import { Simulation } from './simulation.ts'
import { REPL_SCENARIOS, replScenarioById } from './scenarios.ts'
import type { Frame, ReplConfig } from './types.ts'
import { lagOf, lastLsn, retainedOf } from './world.ts'

function renderTick(f: Frame): string {
  const w = f.world
  const nodes = w.nodes
    .map((n) => {
      const role = n.id === w.primary ? 'P' : n.broken ? 'X' : n.up ? 'R' : '-'
      const pos = n.id === w.primary ? `lsn ${lastLsn(w)}/${n.flushLsn}` : `w${n.writeLsn} f${n.flushLsn} r${n.replayLsn} lag ${lagOf(w, n)}`
      const q = n.queries.length > 0 ? ` q${n.queries.map((x) => x.snapshot).join(',')}` : ''
      const c = n.conflictSince !== null ? ' КОНФЛИКТ' : ''
      return `${role}:${n.name}[${pos}${q}${c}]`
    })
    .join(' ')
  const cl = w.clients
    .map((c) => {
      const m = w.marks[c.id]
      return `${c.name}:${!m ? '·' : m.kind === 'read' ? `R@${w.nodes[m.node]?.name}${m.bad ? '!' : ''}` : m.kind}`
    })
    .join(' ')
  const evs = f.events
    .filter((e) => !['replica.receive', 'replica.replay', 'wal.write', 'read.ok'].includes(e.type))
    .map((e) => e.type)
    .join(' ')
  return `t=${String(w.tick).padStart(3)} ${nodes} wal=${retainedOf(w)} dead=${w.dead.length}\n      ${cl}   ${evs}`
}

const id = process.argv[2] ?? 'stream'
const found = replScenarioById(id)
if (!found) {
  console.error(`Нет сценария «${id}». Доступны: ${REPL_SCENARIOS.map((s) => s.id).join(', ')}`)
  process.exit(1)
}
const patch: Record<string, unknown> = {}
for (const arg of process.argv.slice(3)) {
  const [k, v] = arg.split('=')
  if (!k || v === undefined) continue
  patch[k] = v === 'true' ? true : v === 'false' ? false : Number.isFinite(Number(v)) ? Number(v) : v
}
const sc = { ...found, config: { ...found.config, ...(patch as Partial<ReplConfig>) } }
const sim = new Simulation(sc)
console.log(`\n${sc.title}\n${sc.claim}\n${'─'.repeat(78)}`)
for (;;) {
  const f = sim.step()
  if (!f) break
  console.log(renderTick(f))
}
const s = sim.world.stats
console.log('─'.repeat(78))
console.log(
  `итог: тиков ${sim.tick} (${sim.world.finishReason}), коммитов ${s.commits}, подтверждено ${s.acked}, потеряно ${s.lost}, неизвестно ${s.unknown}, ` +
    `ошибок записи ${s.writeErrors}, чтений ${s.reads} (устаревших ${s.staleReads}, своё не видно ${s.ownStale}, назад ${s.backwards}, ошибок ${s.readErrors}), ` +
    `ожидание коммита ${s.commitWaitTicks}т (макс ${s.maxCommitWait}), макс отставание ${s.maxLag}, WAL макс ${s.maxRetained}, мусор макс ${s.maxDead}, отмен ${s.cancels}, без ведущего ${s.downTicks}т`,
)
