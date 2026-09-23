import { useMemo } from 'react'
import type { ResilEvent, ResilWorld } from '../engine/types.ts'
import type { Simulation } from '../engine/simulation.ts'
import { depCapacityOf, waitingOf } from '../engine/world.ts'
import { Chart, FLOW_WINDOW, LAT_WINDOW, chartData } from './Charts.tsx'

/**
 * Сервис и его зависимость на одном тике. Чистая проекция: получает кадр,
 * ничего не хранит. Сверху — графики прогона до курсора, под ними — воркеры A
 * (видно, кто работает, а кто просто ждёт B), предохранитель и сам B.
 */

export interface Highlight {
  req: Set<number>
  worker: Set<number>
}

export function highlightOf(e: ResilEvent | null): Highlight {
  return { req: new Set(e?.actors.req ?? []), worker: new Set(e?.actors.worker ?? []) }
}

const QUEUE_SHOWN = 40

const DEP_STATE: Record<ResilWorld['dep'], { text: string; cls: string }> = {
  ok: { text: 'отвечает нормально', cls: '' },
  slow: { text: 'медленный', cls: 'is-warn' },
  errors: { text: 'отвечает ошибками', cls: 'is-bad' },
  hang: { text: 'не отвечает', cls: 'is-bad' },
}

const BREAKER: Record<ResilWorld['breaker'], { text: string; cls: string }> = {
  closed: { text: 'замкнут — вызовы идут', cls: 'is-on' },
  open: { text: 'разомкнут — вызовов нет', cls: 'is-off' },
  'half-open': { text: 'полуоткрыт — пробный вызов', cls: 'is-warn' },
}

function ServiceA({ w, hl }: { w: ResilWorld; hl: Highlight }) {
  const shown = w.queue.slice(0, QUEUE_SHOWN)
  return (
    <section className="rs-panel" aria-label="Сервис A">
      <header>
        <strong>Сервис A</strong>
        <span className="kf-muted">
          воркеров {w.config.workers}, из них ждут B <b>{waitingOf(w)}</b>
        </span>
      </header>
      <div className="rs-workers">
        {w.workers.map((id, i) => {
          const r = id === null ? null : w.requests[id]
          const stage = r?.stage
          return (
            <span
              key={i}
              className={`rs-worker ${stage === 'own' ? 'is-own' : ''} ${stage === 'dep' ? 'is-wait' : ''} ${hl.worker.has(i) || (id !== null && hl.req.has(id)) ? 'is-hl' : ''}`}
              title={
                r
                  ? stage === 'dep'
                    ? `#${r.id}: ждёт ответа B уже ${w.tick - (r.calledAt ?? w.tick)} т.${w.config.timeout > 0 ? ` из ${w.config.timeout}` : ''}`
                    : `#${r.id}: своя работа, осталось ${Math.max(0, r.until - w.tick)} т.`
                  : 'свободен'
              }
            >
              {stage === 'dep' ? '⏳' : stage === 'own' ? '●' : ''}
            </span>
          )
        })}
      </div>
      <div className="ld-chips" aria-label="Очередь на входе">
        {w.queue.length === 0 && <span className="kf-muted">очередь пуста</span>}
        {shown.map((id) => {
          const r = w.requests[id]
          return (
            <span key={id} className={`ld-chip ${r?.needsDep ? '' : 'is-local'} ${hl.req.has(id) ? 'is-hl' : ''}`} title={`#${id}: ${r?.needsDep ? 'нужен B' : 'обойдётся без B'}, ждёт ${w.tick - (r?.bornAt ?? w.tick)} т.`}>
              {r?.needsDep ? '' : '·'}
            </span>
          )
        })}
        {w.queue.length > QUEUE_SHOWN && <span className="ld-more">+{w.queue.length - QUEUE_SHOWN}</span>}
      </div>
      {w.config.rateLimit > 0 && (
        <span className="lb-stats">
          ограничитель: <b>{w.tokens.toFixed(1)}</b> токенов из {w.config.burst} · пропускает {w.config.rateLimit}/т.
        </span>
      )}
    </section>
  )
}

function ServiceB({ w, hl }: { w: ResilWorld; hl: Highlight }) {
  const st = DEP_STATE[w.dep]
  const br = BREAKER[w.breaker]
  return (
    <section className={`rs-panel ${st.cls}`} aria-label="Сервис B">
      <header>
        <strong>Сервис B</strong>
        <span className={`lb-badge ${st.cls}`}>
          {st.text}
          {w.dep === 'slow' ? ` ×${w.depFactor}` : ''}
        </span>
      </header>
      {w.config.breaker && (
        <span className={`lb-rot ${br.cls === 'is-warn' ? '' : br.cls}`}>
          предохранитель: {br.text}
          {w.breaker === 'open' ? ` (ещё ${Math.max(0, w.breakerUntil - w.tick)} т.)` : ''}
        </span>
      )}
      <div className="lb-workers">
        {w.depWorkers.map((c, i) => (
          <span key={i} className={`lb-worker ${c ? 'is-busy' : ''} ${c && hl.req.has(c.req) ? 'is-hl' : ''}`} title={c ? `#${c.req}: ${w.tick - (c.startAt ?? w.tick)} из ${(c.finishAt ?? w.tick) - (c.startAt ?? w.tick)} т.` : 'свободен'}>
            {c && <span style={{ width: `${Math.min(100, ((w.tick - (c.startAt ?? w.tick)) / Math.max(1, (c.finishAt ?? w.tick) - (c.startAt ?? w.tick))) * 100)}%` }} />}
          </span>
        ))}
      </div>
      <span className="lb-stats">
        в очереди <b>{w.depQueue.length}</b> · успевает <b>{depCapacityOf(w.config).toFixed(2)}</b> вызова за тик
        {w.config.bulkhead > 0 && ` · переборка: не больше ${w.config.bulkhead} ждущих`}
      </span>
    </section>
  )
}

export function Board({ sim, cursor, highlight: hl, onSeek }: { sim: Simulation; cursor: number; highlight: Highlight; onSeek: (t: number) => void }) {
  const w = sim.history[cursor]!.world
  const last = sim.history.length - 1
  const d = useMemo(() => chartData(sim.log), [sim])
  const c = w.config

  return (
    <div className="board ld-board">
      <div className="ld-charts">
        <Chart
          title="Воркеры A"
          note="своя работа, ожидание B и очередь"
          stacked
          series={[
            { label: 'работают', color: 'work', values: d.own, area: true },
            { label: 'ждут B', color: 'wait', values: d.waiting, area: true },
            { label: 'в очереди', color: 'waste', values: d.queue },
          ]}
          refs={[{ y: c.workers, label: `воркеров ${c.workers}` }]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
        <Chart
          title="Задержка"
          note={`тиков, по ответам за последние ${LAT_WINDOW}`}
          series={[
            { label: 'p50', color: 'work', values: d.p50 },
            { label: 'p99', color: 'wait', values: d.p99 },
            { label: 'p99 без B', color: 'in', values: d.localP99 },
          ]}
          refs={c.timeout > 0 ? [{ y: c.timeout, label: `таймаут ${c.timeout}` }] : []}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
        <Chart
          title="Ответы"
          note={`за последние ${FLOW_WINDOW} тиков`}
          series={[
            { label: 'полные', color: 'work', values: d.ok },
            ...(d.anyDegraded ? [{ label: 'заглушки', color: 'in' as const, values: d.degraded }] : []),
            ...(d.anyBad ? [{ label: 'ошибки и отказы', color: 'waste' as const, values: d.bad }] : []),
          ]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
      </div>

      <div className="rs-view">
        <ServiceA w={w} hl={hl} />
        <ServiceB w={w} hl={hl} />
      </div>
      <p className="rp-legend" aria-hidden="true">
        <span className="rs-worker is-own">●</span> воркер работает <span className="rs-worker is-wait">⏳</span> воркер ждёт B
        <span className="ld-chip" /> запрос к B <span className="ld-chip is-local">·</span> запрос без B
      </p>
    </div>
  )
}
