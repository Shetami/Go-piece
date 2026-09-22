import type { ClientOp, Frame, Mark, NodeState, ReplEvent, ReplScenario, ReplWorld } from '../engine/types.ts'
import { commitReady, isSyncCandidate } from '../engine/tick.ts'
import { lagOf, lastLsn, primaryOf, recordAt, replicasOf, retainedOf, timeLagOf } from '../engine/world.ts'

/**
 * Схема кластера на одном тике. Чистая проекция: получает кадр, ничего не хранит.
 *
 * Сверху — расписание клиентов. Под ним — узлы: у каждого полоса последних
 * записей WAL, закрашенная по тому, докуда узел их получил, сбросил и проиграл.
 * Ниже — клиенты: что делают, чего ждут и что прочитали.
 */

const COLORS = 8
const color = (i: number) => ({ '--c': `var(--wl-${i % COLORS})` }) as React.CSSProperties

export interface Highlight {
  node: Set<number>
  client: Set<number>
  lsn: Set<number>
}

export function highlightOf(e: ReplEvent | null): Highlight {
  return { node: new Set(e?.actors.node ?? []), client: new Set(e?.actors.client ?? []), lsn: new Set(e?.actors.lsn ?? []) }
}

/* ───────────────────────────── расписание ───────────────────────────── */

function markClass(m: Mark | null): string {
  if (!m) return 'is-empty'
  switch (m.kind) {
    case 'write':
      return 'op-write'
    case 'wait':
      return 'is-fsync'
    case 'ack':
      return 'op-commit'
    case 'read':
      return `op-read${m.bad ? ' is-bad' : ''}`
    case 'query':
      return 'is-wait'
    case 'error':
      return 'op-abort'
  }
}

function markGlyph(w: ReplWorld, m: Mark | null): string {
  if (!m) return ''
  switch (m.kind) {
    case 'write':
      return 'W'
    case 'ack':
      return 'C'
    case 'read':
      return m.node === w.primary ? 'P' : String(m.node)
    case 'error':
      return '!'
    default:
      return ''
  }
}

const MARK_TEXT: Record<Mark['kind'], string> = {
  write: 'запись на ведущем',
  wait: 'COMMIT ждёт подтверждения',
  ack: 'COMMIT подтверждён',
  read: 'чтение',
  query: 'долгий запрос на реплике',
  error: 'ошибка',
}

function Schedule({ history, cursor, scenario, hl, onSeek }: {
  history: readonly Frame[]
  cursor: number
  scenario: ReplScenario
  hl: Highlight
  onSeek: (t: number) => void
}) {
  const last = history.length - 1
  const ticks = Array.from({ length: last }, (_, i) => i + 1)
  return (
    <div className="db-sched" role="group" aria-label="Расписание клиентов">
      <div className="db-sched-grid" style={{ gridTemplateColumns: `4.5rem repeat(${Math.max(1, last)}, 16px)` }}>
        <span className="db-sched-corner">тик</span>
        {ticks.map((t) => (
          <button key={t} type="button" className={`db-sched-tick ${t === cursor ? 'is-now' : ''}`} onClick={() => onSeek(t)} aria-label={`Тик ${t}`}>
            {t % 5 === 0 || t === 1 ? t : ''}
          </button>
        ))}
        {scenario.clients.map((c, ci) => (
          <Row key={ci} ci={ci} name={c.name} history={history} ticks={ticks} cursor={cursor} hl={hl} onSeek={onSeek} />
        ))}
        <span className="db-sched-name rp-sched-primary">ведущий</span>
        {ticks.map((t) => {
          const w = history[t]!.world
          const down = !primaryOf(w).up
          const promoted = history[t]!.events.some((e) => e.type === 'failover.promote')
          return (
            <button
              key={t}
              type="button"
              className={`db-cell ${down ? 'is-down' : promoted ? 'op-commit' : 'is-empty'} ${t > cursor ? 'is-future' : ''} ${t === cursor ? 'is-now' : ''}`}
              title={`тик ${t}: ${down ? 'ведущего нет — запись невозможна' : promoted ? `новый ведущий — ${primaryOf(w).name}` : `ведущий ${primaryOf(w).name}`}`}
              onClick={() => onSeek(t)}
            >
              {promoted ? '↑' : ''}
            </button>
          )
        })}
      </div>
      <p className="db-sched-legend" aria-hidden="true">
        <span className="db-cell op-write">W</span> запись <span className="db-cell is-fsync" /> ждёт подтверждения
        <span className="db-cell op-commit">C</span> подтверждено <span className="db-cell op-read">1</span> чтение с реплики 1
        <span className="db-cell op-read">P</span> чтение с ведущего <span className="db-cell op-read is-bad">2</span> устаревшее своё или назад
        <span className="db-cell is-wait" /> отчёт <span className="db-cell op-abort">!</span> ошибка <span className="db-cell is-down" /> ведущего нет
      </p>
    </div>
  )
}

function Row({ ci, name, history, ticks, cursor, hl, onSeek }: {
  ci: number
  name: string
  history: readonly Frame[]
  ticks: number[]
  cursor: number
  hl: Highlight
  onSeek: (t: number) => void
}) {
  return (
    <>
      <span className={`db-sched-name ${hl.client.has(ci) ? 'is-hl' : ''}`} style={color(ci)}>
        {name}
      </span>
      {ticks.map((t) => {
        const w = history[t]!.world
        const m = w.marks[ci] ?? null
        return (
          <button
            key={t}
            type="button"
            className={`db-cell ${markClass(m)} ${t > cursor ? 'is-future' : ''} ${t === cursor ? 'is-now' : ''}`}
            style={color(ci)}
            title={`тик ${t} · ${name}: ${m ? MARK_TEXT[m.kind] + (m.kind === 'read' ? ` с ${w.nodes[m.node]?.name}${m.bad ? ' — устаревшее' : ''}` : '') : '—'}`}
            onClick={() => onSeek(t)}
          >
            {markGlyph(w, m)}
          </button>
        )
      })}
    </>
  )
}

/* ─────────────────────────────── узлы ─────────────────────────────── */

const WINDOW = 18

function inflightTo(w: ReplWorld, id: number): Set<number> {
  return new Set(w.net.filter((m) => m.kind === 'wal' && m.to === id).map((m) => (m.kind === 'wal' ? m.lsn : 0)))
}

function roleLabel(w: ReplWorld, n: NodeState): { text: string; cls: string } {
  if (n.id === w.primary) return n.up ? { text: 'ведущий', cls: 'is-primary' } : { text: 'ведущий упал', cls: 'is-down' }
  if (n.broken && n.promotedAt !== null) return { text: 'бывший ведущий — нужен pg_rewind', cls: 'is-down' }
  if (n.broken) return { text: 'отстала безнадёжно — нужна новая копия', cls: 'is-down' }
  if (!n.up) return { text: 'недоступна', cls: 'is-down' }
  if (isSyncCandidate(w, n)) return { text: 'синхронная реплика', cls: 'is-sync' }
  return { text: 'асинхронная реплика', cls: '' }
}

function WalStrip({ w, n, hl }: { w: ReplWorld; n: NodeState; hl: Highlight }) {
  const last = lastLsn(w)
  const from = Math.max(1, last - WINDOW + 1)
  const flight = inflightTo(w, n.id)
  const isPrimary = n.id === w.primary
  const cells = []
  for (let lsn = from; lsn <= Math.max(last, from + WINDOW - 1); lsn++) {
    const r = recordAt(w, lsn)
    let cls = 'is-none'
    let state = 'ещё не существует'
    if (lsn <= last) {
      if (isPrimary) {
        cls = lsn <= n.flushLsn ? 'is-replayed' : 'is-written'
        state = lsn <= n.flushLsn ? 'на диске ведущего' : 'записана, ещё не сброшена'
        if (lsn < w.walOldest) {
          cls += ' is-gone'
          state += ', WAL уже удалён'
        }
      } else if (lsn <= n.replayLsn) {
        cls = 'is-replayed'
        state = 'проиграна — видна запросам'
      } else if (lsn <= n.flushLsn) {
        cls = 'is-flushed'
        state = 'на диске реплики, не проиграна'
      } else if (lsn <= n.writeLsn) {
        cls = 'is-written'
        state = 'получена, не сброшена'
      } else if (flight.has(lsn)) {
        cls = 'is-flight'
        state = 'в пути по сети'
      } else {
        cls = 'is-missing'
        state = 'не получена'
      }
    }
    cells.push(
      <span
        key={lsn}
        className={`rp-lsn ${cls} ${r?.kind === 'cleanup' ? 'is-cleanup' : ''} ${hl.lsn.has(lsn) ? 'is-hl' : ''}`}
        style={r?.client !== undefined ? color(r.client) : undefined}
        title={
          r
            ? `LSN ${lsn}: ${r.kind === 'cleanup' ? `очистка, удалено версий: ${r.removed}` : `${r.key} = ${r.value}`} — ${state}`
            : `LSN ${lsn}: ${state}`
        }
      >
        {r?.kind === 'cleanup' ? 'v' : ''}
      </span>,
    )
  }
  return (
    <span className="rp-strip" aria-label={`WAL ${n.name}`}>
      {cells}
    </span>
  )
}

function NodeCard({ w, n, hl, scenario }: { w: ReplWorld; n: NodeState; hl: Highlight; scenario: ReplScenario }) {
  const role = roleLabel(w, n)
  const isPrimary = n.id === w.primary
  const lag = lagOf(w, n)
  const tlag = timeLagOf(w, n)
  return (
    <section className={`rp-node ${role.cls} ${hl.node.has(n.id) ? 'is-hl' : ''}`}>
      <header>
        <strong>{n.name}</strong>
        <span className="rp-role">{role.text}</span>
        {!isPrimary && <span className="kf-muted">сеть {n.latency} т.</span>}
      </header>
      <WalStrip w={w} n={n} hl={hl} />
      {isPrimary ? (
        <span className="rp-pos">
          WAL до <b>{lastLsn(w)}</b>, на диске до <b>{n.flushLsn}</b> · хранит <b>{retainedOf(w)}</b> записей
          {w.config.slots ? ' (слоты)' : ` (wal_keep ${w.config.walKeep})`}
        </span>
      ) : (
        <span className="rp-pos" title="write — получено, flush — на диске, replay — проиграно">
          write <b>{n.writeLsn}</b> · flush <b>{n.flushLsn}</b> · replay <b>{n.replayLsn}</b>
          {!n.broken && n.up && (
            <span className={lag > 0 ? 'db-warn' : undefined}>
              {' '}
              · отстаёт на {lag} зап.{tlag > 0 ? `, ${tlag} т.` : ''}
            </span>
          )}
        </span>
      )}
      <span className="rp-kv">
        {Object.entries(n.kv).map(([k, v]) => (
          <span key={k} className="rp-kv-item" title={`${k} = ${v.value}, изменено записью WAL ${v.lsn}`}>
            {k}=<b>{v.value}</b>
          </span>
        ))}
      </span>
      {n.queries.length > 0 && (
        <span className="rp-query">
          отчёт {n.queries.map((q) => w.clients[q.client]?.name).join(', ')} · снимок на LSN {n.queries.map((q) => q.snapshot).join(', ')}
          {n.conflictSince !== null && <b className="rp-conflict"> · проигрывание стоит {w.tick - n.conflictSince} т.</b>}
        </span>
      )}
      {isPrimary && scenario.config.vacuumEvery !== undefined && w.config.vacuumEvery > 0 && (
        <span className="rp-pos">
          мёртвых версий <b className={w.dead.length > 6 ? 'db-warn' : undefined}>{w.dead.length}</b> · вакуум каждые {w.config.vacuumEvery} т.
        </span>
      )}
      {isPrimary && replicasOf(w).length > 0 && (
        <table className="rp-stat" title="pg_stat_replication — что ведущий знает о репликах из их ответов">
          <thead>
            <tr>
              <th>pg_stat_replication</th>
              <th>sent</th>
              <th>write</th>
              <th>flush</th>
              <th>replay</th>
              {w.config.hotStandbyFeedback && <th>xmin</th>}
            </tr>
          </thead>
          <tbody>
            {replicasOf(w).map((r) => {
              const v = w.view[r.id]
              if (!v) return null
              return (
                <tr key={r.id} className={!r.up || r.broken ? 'is-off' : ''}>
                  <td>
                    {r.name}
                    {isSyncCandidate(w, r) && <small> sync</small>}
                  </td>
                  <td>{v.sent}</td>
                  <td>{v.write}</td>
                  <td>{v.flush}</td>
                  <td>{v.replay}</td>
                  {w.config.hotStandbyFeedback && <td>{v.feedback ?? '—'}</td>}
                </tr>
              )
            })}
          </tbody>
        </table>
      )}
    </section>
  )
}

/* ─────────────────────────────── клиенты ─────────────────────────────── */

function opText(w: ReplWorld, op: ClientOp): string {
  switch (op.kind) {
    case 'write':
      return `UPDATE t SET v = v + 1 WHERE k = '${op.key}'`
    case 'read':
      return `SELECT v FROM t WHERE k = '${op.key}' — ${op.from === 'primary' ? 'на ведущем' : op.from === 'any' ? 'на любой реплике' : `на ${w.nodes[op.from]?.name ?? `r${op.from}`}`}`
    case 'query':
      return `отчёт по '${op.key}' на ${w.nodes[op.from]?.name ?? `r${op.from}`}, ${op.ticks} т.`
  }
}

function ClientCard({ w, ci, scenario, hl }: { w: ReplWorld; ci: number; scenario: ReplScenario; hl: Highlight }) {
  const c = w.clients[ci]!
  const spec = scenario.clients[ci]!
  const op = spec.ops[c.opIdx]
  const waiting = c.state === 'commit-wait' && c.waitLsn !== null ? commitReady(w, c.waitLsn) : null
  const r = c.lastRead
  const lost = w.ackedLsns.filter((a) => a.client === ci && a.lost).length
  return (
    <section className={`rp-client ${hl.client.has(ci) ? 'is-hl' : ''}`} style={color(ci)}>
      <header>
        <strong>{c.name}</strong>
        <span className="db-sess-state">
          {c.state === 'done'
            ? 'закончил'
            : c.state === 'commit-wait'
              ? `ждёт COMMIT ${w.tick - (c.waitSince ?? w.tick)} т.`
              : c.state === 'reading'
                ? 'выполняет отчёт'
                : 'работает'}
        </span>
        {spec.repeat !== undefined && spec.repeat !== 1 && (
          <span className="kf-muted">
            круг {c.iter + 1}
            {spec.repeat === 'forever' ? '' : ` из ${spec.repeat}`}
          </span>
        )}
      </header>
      {op && c.state !== 'done' && <code className="rp-op">{opText(w, op)}</code>}
      {waiting && !waiting.ok && <span className="rp-wait">ждёт: {waiting.waitingFor}</span>}
      <span className="db-sess-foot">
        <span>
          подтверждено <b>{c.acked}</b>
        </span>
        {lost > 0 && (
          <span className="metric-bad">
            потеряно <b>{lost}</b>
          </span>
        )}
        {c.errors > 0 && (
          <span>
            ошибок <b>{c.errors}</b>
          </span>
        )}
        {r && (
          <span>
            прочитал {r.key}=<b>{r.value ?? 'NULL'}</b> на {w.nodes[r.node]?.name}
          </span>
        )}
      </span>
    </section>
  )
}

/* ─────────────────────────────── целиком ─────────────────────────────── */

export function Board({ history, cursor, scenario, highlight: hl, onSeek }: {
  history: readonly Frame[]
  cursor: number
  scenario: ReplScenario
  highlight: Highlight
  onSeek: (t: number) => void
}) {
  const w = history[cursor]!.world
  const order = [w.nodes[w.primary]!, ...w.nodes.filter((n) => n.id !== w.primary)]
  return (
    <div className="board db-board">
      <Schedule history={history} cursor={cursor} scenario={scenario} hl={hl} onSeek={onSeek} />
      <div className="rp-nodes">
        {order.map((n) => (
          <NodeCard key={n.id} w={w} n={n} hl={hl} scenario={scenario} />
        ))}
      </div>
      <p className="rp-legend" aria-hidden="true">
        <span className="rp-lsn is-replayed" /> проиграна <span className="rp-lsn is-flushed" /> на диске
        <span className="rp-lsn is-written" /> получена <span className="rp-lsn is-flight" /> в пути
        <span className="rp-lsn is-missing" /> не получена <span className="rp-lsn is-replayed is-cleanup">v</span> очистка
        <span className="rp-lsn is-replayed is-gone" /> WAL удалён с ведущего
      </p>
      <div className="rp-clients">
        {w.clients.map((c) => (
          <ClientCard key={c.id} w={w} ci={c.id} scenario={scenario} hl={hl} />
        ))}
      </div>
    </div>
  )
}
