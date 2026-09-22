import { useMemo } from 'react'
import type { LoadEvent, LoadWorld, Request } from '../engine/types.ts'
import type { Simulation } from '../engine/simulation.ts'
import { Chart, FLOW_WINDOW, LAT_WINDOW, chartData } from './Charts.tsx'

/**
 * Сервис на одном тике. Чистая проекция: получает кадр, ничего не хранит.
 *
 * Сверху — графики всего прогона до курсора. Под ними — сам сервис:
 * очередь слева направо от головы к хвосту, дальше воркеры с полосой
 * прогресса. Запрос, чей клиент уже ушёл, перечёркнут: его обработают
 * впустую, если сервис не узнает об уходе.
 */

export interface Highlight {
  req: Set<number>
  worker: Set<number>
}

export function highlightOf(e: LoadEvent | null): Highlight {
  return { req: new Set(e?.actors.req ?? []), worker: new Set(e?.actors.worker ?? []) }
}

/** Больше стольких запросов очередь не рисует поштучно. */
const QUEUE_SHOWN = 48

function reqTitle(w: LoadWorld, r: Request): string {
  const parts = [`#${r.id}`]
  if (r.attempt > 0) parts.push(`${r.attempt + 1}-я попытка`)
  parts.push(r.startAt === null ? `ждёт ${w.tick - r.arriveAt} т.` : `в работе ${w.tick - r.startAt} из ${r.service} т.`)
  if (r.slow) parts.push('медленный')
  if (r.abandoned) parts.push('клиент уже ушёл — ответ никому не нужен')
  else if (r.deadline !== null) parts.push(`клиент ждёт ещё ${Math.max(0, r.deadline - w.tick)} т.`)
  return parts.join(' · ')
}

function Queue({ w, hl }: { w: LoadWorld; hl: Highlight }) {
  const shown = w.queue.slice(0, QUEUE_SHOWN)
  const zombies = w.queue.filter((r) => r.abandoned).length
  const limit = w.config.queueLimit
  return (
    <section className="ld-queue" aria-label="Очередь">
      <header>
        <strong>Очередь</strong>
        <span className="kf-muted">
          {w.queue.length}
          {limit > 0 ? ` из ${limit}` : ''}
          {zombies > 0 && <span className="ld-bad"> · {zombies} никому не нужны</span>}
        </span>
      </header>
      <div className="ld-chips">
        {w.queue.length === 0 && <span className="kf-muted">пусто — новый запрос сразу попадёт к воркеру</span>}
        {shown.map((r) => (
          <span
            key={r.id}
            className={`ld-chip ${r.slow ? 'is-slow' : ''} ${r.abandoned ? 'is-zombie' : ''} ${r.attempt > 0 ? 'is-retry' : ''} ${hl.req.has(r.id) ? 'is-hl' : ''}`}
            title={reqTitle(w, r)}
          >
            {r.attempt > 0 ? '↻' : ''}
          </span>
        ))}
        {w.queue.length > QUEUE_SHOWN && <span className="ld-more">+{w.queue.length - QUEUE_SHOWN}</span>}
        {limit > 0 &&
          Array.from({ length: Math.max(0, Math.min(limit, QUEUE_SHOWN) - shown.length) }, (_, i) => <span key={`f${i}`} className="ld-chip is-free" />)}
      </div>
    </section>
  )
}

function Worker({ w, i, hl }: { w: LoadWorld; i: number; hl: Highlight }) {
  const r = w.workers[i] ?? null
  const done = r && r.startAt !== null ? w.tick - r.startAt : 0
  return (
    <div className={`ld-worker ${r ? 'is-busy' : ''} ${r?.abandoned ? 'is-zombie' : ''} ${r?.slow ? 'is-slow' : ''} ${hl.worker.has(i) ? 'is-hl' : ''}`} title={r ? reqTitle(w, r) : 'свободен'}>
      <span className="ld-worker-name">в{i + 1}</span>
      {r ? (
        <>
          <span className={`ld-worker-req ${hl.req.has(r.id) ? 'is-hl' : ''}`}>
            #{r.id}
            {r.abandoned && <small> ушёл</small>}
          </span>
          <span className="ld-bar" aria-hidden="true">
            <span style={{ width: `${Math.min(100, (done / Math.max(1, r.service)) * 100)}%` }} />
          </span>
        </>
      ) : (
        <span className="kf-muted">свободен</span>
      )}
    </div>
  )
}

export function Board({ sim, cursor, highlight: hl, onSeek }: { sim: Simulation; cursor: number; highlight: Highlight; onSeek: (t: number) => void }) {
  const w = sim.history[cursor]!.world
  const last = sim.history.length - 1
  const d = useMemo(() => chartData(sim.log), [sim])
  const c = w.config

  const flow = [
    { label: 'пришло', color: 'in' as const, values: d.arrived },
    { label: 'ответов', color: 'work' as const, values: d.ok },
    ...(d.anyWasted ? [{ label: 'впустую', color: 'waste' as const, values: d.wasted }] : []),
  ]

  return (
    <div className="board ld-board">
      <div className="ld-charts">
        <Chart
          title="В сервисе"
          note="запросов"
          stacked
          series={[
            { label: 'в работе', color: 'work', values: d.busy, area: true },
            { label: 'в очереди', color: 'wait', values: d.queue, area: true },
          ]}
          refs={[{ y: c.workers, label: `воркеров ${c.workers}` }]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
        <Chart
          title="Задержка ответов"
          note={`тиков, по ответам за последние ${LAT_WINDOW}`}
          series={[
            { label: 'p50', color: 'work', values: d.p50 },
            { label: 'p99', color: 'wait', values: d.p99 },
          ]}
          refs={c.timeout > 0 ? [{ y: c.timeout, label: `таймаут ${c.timeout}` }] : []}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
        <Chart title="Поток" note={`запросов за последние ${FLOW_WINDOW} тиков`} series={flow} last={last} cursor={cursor} onSeek={onSeek} />
      </div>

      <div className="ld-service">
        <Queue w={w} hl={hl} />
        <section className="ld-workers" aria-label="Воркеры">
          <header>
            <strong>Воркеры</strong>
            <span className="kf-muted">
              заняты {w.workers.filter((r) => r !== null).length} из {c.workers}
            </span>
          </header>
          <div className="ld-worker-list">
            {w.workers.map((_, i) => (
              <Worker key={i} w={w} i={i} hl={hl} />
            ))}
          </div>
        </section>
        {w.retryQ.length > 0 && (
          <p className="ld-retry-wait">
            ↻ ждут повтора у клиентов: <b>{w.retryQ.length}</b>
          </p>
        )}
      </div>
      <p className="rp-legend" aria-hidden="true">
        <span className="ld-chip" /> ждёт <span className="ld-chip is-slow" /> медленный <span className="ld-chip is-retry">↻</span> повтор
        <span className="ld-chip is-zombie" /> клиент ушёл {c.queueLimit > 0 && (<><span className="ld-chip is-free" /> свободное место</>)}
      </p>
    </div>
  )
}
