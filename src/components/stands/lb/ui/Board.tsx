import { useMemo } from 'react'
import type { LbEvent, LbWorld, Replica } from '../engine/types.ts'
import type { Simulation } from '../engine/simulation.ts'
import { busyOf, routable } from '../engine/world.ts'
import { Chart, FLOW_WINDOW, LAT_WINDOW, chartData } from './Charts.tsx'

/**
 * Балансировщик и реплики на одном тике. Чистая проекция: получает кадр,
 * ничего не хранит. Сверху — графики прогона до курсора, под ними — карточки
 * реплик: что с ней на самом деле, что о ней думает балансировщик, очередь
 * и воркеры.
 */

export interface Highlight {
  replica: Set<number>
  req: Set<number>
}

export function highlightOf(e: LbEvent | null): Highlight {
  return { replica: new Set(e?.actors.replica ?? []), req: new Set(e?.actors.req ?? []) }
}

const QUEUE_SHOWN = 24

const STATE: Record<Replica['state'], { text: string; cls: string }> = {
  up: { text: 'работает', cls: '' },
  slow: { text: 'тормозит', cls: 'is-warn' },
  errors: { text: 'отвечает 500', cls: 'is-bad' },
  crash: { text: 'упала', cls: 'is-bad' },
  hang: { text: 'зависла', cls: 'is-bad' },
  booting: { text: 'запускается', cls: 'is-boot' },
}

function rotation(w: LbWorld, r: Replica): { text: string; cls: string } {
  if (r.state === 'booting') return { text: `будет готова на тике ${r.bootUntil}`, cls: 'is-off' }
  if (!r.inRotation) return { text: 'исключена проверкой здоровья', cls: 'is-off' }
  if (r.ejectedUntil > w.tick) return { text: `исключена по ошибкам до тика ${r.ejectedUntil}`, cls: 'is-off' }
  return { text: 'в ротации', cls: 'is-on' }
}

function ReplicaCard({ w, r, hl }: { w: LbWorld; r: Replica; hl: Highlight }) {
  const st = STATE[r.state]
  const rot = rotation(w, r)
  const shown = r.queue.slice(0, QUEUE_SHOWN)
  return (
    <section className={`lb-replica ld-c-r${r.id % 8} ${st.cls} ${routable(w, r) ? '' : 'is-out'} ${hl.replica.has(r.id) ? 'is-hl' : ''}`}>
      <header>
        <i className="lb-swatch" aria-hidden="true" />
        <strong>{r.name}</strong>
        <span className={`lb-badge ${st.cls}`}>
          {st.text}
          {r.state === 'slow' ? ` ×${r.factor}` : ''}
        </span>
      </header>
      <span className={`lb-rot ${rot.cls}`}>{rot.text}</span>
      <div className="ld-chips" aria-label={`Очередь ${r.name}`}>
        {shown.map((q) => (
          <span key={q.id} className={`ld-chip ${q.abandoned ? 'is-zombie' : ''} ${q.attempt > 0 ? 'is-retry' : ''} ${hl.req.has(q.id) ? 'is-hl' : ''}`} title={`#${q.id}, ждёт ${w.tick - q.arriveAt} т.${q.abandoned ? ' — клиент ушёл' : ''}`}>
            {q.attempt > 0 ? '↻' : ''}
          </span>
        ))}
        {r.queue.length > QUEUE_SHOWN && <span className="ld-more">+{r.queue.length - QUEUE_SHOWN}</span>}
        {r.queue.length === 0 && <span className="kf-muted">очередь пуста</span>}
      </div>
      <div className="lb-workers" aria-label={`Воркеры ${r.name}`}>
        {r.workers.map((q, i) => (
          <span
            key={i}
            className={`lb-worker ${q ? 'is-busy' : ''} ${q?.abandoned ? 'is-zombie' : ''} ${r.state === 'hang' && q ? 'is-stuck' : ''}`}
            title={q ? `#${q.id}: ${q.startAt === null ? '' : `${w.tick - q.startAt} из ${q.service} т.`}${r.state === 'hang' ? ' — завис' : ''}` : 'свободен'}
          >
            {q && <span style={{ width: `${Math.min(100, ((w.tick - (q.startAt ?? w.tick)) / Math.max(1, q.service)) * 100)}%` }} />}
          </span>
        ))}
      </div>
      <span className="lb-stats">
        ждёт ответов <b>{r.outstanding}</b> · отправлено <b>{r.routed}</b> · ответов <b>{r.served}</b>
        {r.errors > 0 && (
          <>
            {' '}
            · <span className="ld-bad">ошибок {r.errors}</span>
          </>
        )}
      </span>
    </section>
  )
}

export function Board({ sim, cursor, highlight: hl, onSeek }: { sim: Simulation; cursor: number; highlight: Highlight; onSeek: (t: number) => void }) {
  const w = sim.history[cursor]!.world
  const last = sim.history.length - 1
  const d = useMemo(() => chartData(sim.log), [sim])
  const c = w.config
  const busy = w.replicas.reduce((a, r) => a + (routable(w, r) ? busyOf(r) : 0), 0)
  const cap = w.replicas.reduce((a, r) => a + (routable(w, r) ? r.workers.length : 0), 0)

  return (
    <div className="board ld-board">
      <div className="ld-charts">
        <Chart
          title="Запросы в репликах"
          note="в очереди и в работе"
          series={d.load.map((values, i) => ({ label: `r${i + 1}`, color: `r${i % 8}` as const, values }))}
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
        <Chart
          title="Ответы и ошибки"
          note={`за последние ${FLOW_WINDOW} тиков; таймауты — тоже ошибки`}
          series={[
            { label: 'ответов', color: 'work', values: d.ok },
            ...(d.anyErrors ? [{ label: 'ошибок', color: 'waste' as const, values: d.errors }] : []),
          ]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
      </div>

      <p className="lb-summary">
        балансировщик <b>{ALGO_TITLE[c.algo]}</b> · в ротации <b>{w.replicas.filter((r) => routable(w, r)).length}</b> из {w.replicas.length} · воркеры в ротации заняты{' '}
        <b>{cap === 0 ? 0 : Math.round((busy / cap) * 100)}%</b>
      </p>
      <div className="lb-replicas">
        {w.replicas.map((r) => (
          <ReplicaCard key={r.id} w={w} r={r} hl={hl} />
        ))}
      </div>
      <p className="rp-legend" aria-hidden="true">
        <span className="ld-chip" /> ждёт в очереди <span className="ld-chip is-retry">↻</span> повтор <span className="ld-chip is-zombie" /> клиент ушёл
        <span className="lb-worker is-busy" /> воркер занят
      </p>
    </div>
  )
}

const ALGO_TITLE: Record<string, string> = {
  'round-robin': 'по кругу',
  random: 'случайно',
  'least-conn': 'наименьшее число соединений',
  p2c: 'два случайных',
}
