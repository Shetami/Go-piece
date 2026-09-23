import { useMemo } from 'react'
import type { CacheEvent, CacheWorld } from '../engine/types.ts'
import type { Simulation } from '../engine/simulation.ts'
import { cachedCount, idealHitRatio, zipfCdf } from '../engine/world.ts'
import { Chart, FLOW_WINDOW, LAT_WINDOW, chartData } from './Charts.tsx'

/**
 * Кэш и база на одном тике. Чистая проекция: получает кадр, ничего не хранит.
 *
 * Сверху — графики прогона до курсора. Под ними — полка ключей: тёмная
 * клетка лежит в кэше, штрихованная просрочена, пустая в кэше отсутствует.
 * Справа — база с очередью и воркерами.
 */

export interface Highlight {
  key: Set<number>
  query: Set<number>
}

export function highlightOf(e: CacheEvent | null): Highlight {
  return { key: new Set(e?.actors.key ?? []), query: new Set(e?.actors.query ?? []) }
}

/** Столько ключей показываем поштучно — самые популярные. */
const KEYS_SHOWN = 60
const QUEUE_SHOWN = 40

function Keys({ w, hl }: { w: CacheWorld; hl: Highlight }) {
  const c = w.config
  const shown = Math.min(c.keys, KEYS_SHOWN)
  const share = useMemo(() => zipfCdf(c.keys, c.zipf), [c.keys, c.zipf])
  return (
    <section className="ca-keys" aria-label="Ключи">
      <header>
        <strong>Ключи</strong>
        <span className="kf-muted">
          в кэше {cachedCount(w)} из {c.cacheSize === 0 ? 'нуля' : c.cacheSize} мест, всего ключей {c.keys}
        </span>
      </header>
      <div className="ca-grid">
        {Array.from({ length: shown }, (_, k) => {
          const e = w.cache[k]
          const state = !e ? 'is-absent' : e.expiresAt > w.tick ? 'is-fresh' : 'is-expired'
          const inflight = (w.inflight[k] ?? 0) > 0
          const title = [
            `k${k}: доля запросов ${((k === 0 ? share[0]! : share[k]! - share[k - 1]!) * 100).toFixed(1)}%`,
            `в базе версия ${w.db[k]}`,
            e ? `в кэше версия ${e.version}${e.expiresAt === Number.POSITIVE_INFINITY ? ', без TTL' : e.expiresAt > w.tick ? `, истечёт через ${e.expiresAt - w.tick} т.` : ', просрочена'}` : 'в кэше нет',
            inflight ? `в базу летит запросов: ${w.inflight[k]}` : '',
          ]
            .filter(Boolean)
            .join(' · ')
          return (
            <span
              key={k}
              className={`ca-key ${state} ${e && e.version < w.db[k]! ? 'is-stale' : ''} ${inflight ? 'is-inflight' : ''} ${hl.key.has(k) ? 'is-hl' : ''}`}
              title={title}
            >
              {k === 0 && c.hotShare > 0 ? '★' : ''}
            </span>
          )
        })}
        {c.keys > shown && <span className="ld-more">+{c.keys - shown}</span>}
      </div>
    </section>
  )
}

function Db({ w, hl }: { w: CacheWorld; hl: Highlight }) {
  const shown = w.queue.slice(0, QUEUE_SHOWN)
  return (
    <section className="ca-db" aria-label="База">
      <header>
        <strong>База</strong>
        <span className="kf-muted">
          занято {w.workers.filter((q) => q !== null).length} из {w.config.dbWorkers} · в очереди {w.queue.length}
        </span>
      </header>
      <div className="ld-chips">
        {w.queue.length === 0 && <span className="kf-muted">очередь пуста</span>}
        {shown.map((q) => (
          <span
            key={q.id}
            className={`ld-chip ${q.kind === 'write' ? 'is-write' : ''} ${q.waiters.length > 1 ? 'is-joined' : ''} ${hl.query.has(q.id) || hl.key.has(q.key) ? 'is-hl' : ''}`}
            title={`${q.kind === 'write' ? 'запись' : q.kind === 'refresh' ? 'фоновое обновление' : 'чтение'} k${q.key}, ждёт ${w.tick - q.arriveAt} т.${q.waiters.length > 1 ? `, ответа ждут ${q.waiters.length}` : ''}`}
          >
            {q.kind === 'write' ? 'W' : q.waiters.length > 1 ? q.waiters.length : ''}
          </span>
        ))}
        {w.queue.length > QUEUE_SHOWN && <span className="ld-more">+{w.queue.length - QUEUE_SHOWN}</span>}
      </div>
      <div className="lb-workers">
        {w.workers.map((q, i) => (
          <span
            key={i}
            className={`lb-worker ${q ? 'is-busy' : ''}`}
            title={q ? `k${q.key}: ${w.tick - (q.startAt ?? w.tick)} из ${(q.finishAt ?? w.tick) - (q.startAt ?? w.tick)} т.` : 'свободен'}
          >
            {q && <span style={{ width: `${Math.min(100, ((w.tick - (q.startAt ?? w.tick)) / Math.max(1, (q.finishAt ?? w.tick) - (q.startAt ?? w.tick))) * 100)}%` }} />}
          </span>
        ))}
      </div>
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
          title="Чтения"
          note={`за последние ${FLOW_WINDOW} тиков`}
          series={[
            { label: 'попаданий', color: 'work', values: d.hits },
            { label: 'промахов', color: 'wait', values: d.misses },
            ...(d.anyStale ? [{ label: 'устаревших', color: 'waste' as const, values: d.stale }] : []),
          ]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
        <Chart
          title="Задержка чтений"
          note={`тиков, по ответам за последние ${LAT_WINDOW}`}
          series={[
            { label: 'p50', color: 'work', values: d.p50 },
            { label: 'p99', color: 'wait', values: d.p99 },
          ]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
        <Chart
          title="База"
          note="запросов в работе и в очереди"
          stacked
          series={[
            { label: 'в работе', color: 'work', values: d.dbBusy, area: true },
            { label: 'в очереди', color: 'wait', values: d.dbQueue, area: true },
          ]}
          refs={[{ y: c.dbWorkers, label: `воркеров ${c.dbWorkers}` }]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
      </div>

      <p className="lb-summary">
        кэш на <b>{c.cacheSize}</b> {c.cacheSize === 1 ? 'ключ' : 'ключей'} из {c.keys} · потолок попаданий при этом размере{' '}
        <b>{Math.round(idealHitRatio(c) * 100)}%</b> · TTL <b>{c.ttl === 0 ? 'нет' : `${c.ttl} т.`}</b>
        {c.ttlJitter > 0 ? ` (разброс до ${c.ttlJitter}%)` : ''} · база успевает <b>{(c.dbWorkers / c.dbTime).toFixed(2)}</b> запроса за тик
      </p>
      <div className="ca-view">
        <Keys w={w} hl={hl} />
        <Db w={w} hl={hl} />
      </div>
      <p className="rp-legend" aria-hidden="true">
        <span className="ca-key is-fresh" /> в кэше <span className="ca-key is-expired" /> просрочен <span className="ca-key is-stale is-fresh" /> устарел
        <span className="ca-key is-absent is-inflight" /> запрошен в базе <span className="ca-key is-absent" /> нет в кэше
        <span className="ld-chip is-write">W</span> запись в очереди базы
      </p>
    </div>
  )
}
