import { useMemo } from 'react'
import type { Shard, ShardEvent, ShardWorld } from '../engine/types.ts'
import type { Simulation } from '../engine/simulation.ts'
import { imbalance, movedShare, zipfCdf } from '../engine/world.ts'
import { Chart, LAT_WINDOW, chartData } from './Charts.tsx'

/**
 * Шарды на одном тике. Чистая проекция: получает кадр, ничего не хранит.
 *
 * Сверху — графики прогона до курсора. Под ними — полоса ключей, раскрашенная
 * по владельцу: видно, как ключи раскиданы по шардам и какие из них переезжают.
 * Ниже — карточки шардов с очередью и воркерами.
 */

export interface Highlight {
  shard: Set<number>
  key: Set<number>
}

export function highlightOf(e: ShardEvent | null): Highlight {
  return { shard: new Set(e?.actors.shard ?? []), key: new Set(e?.actors.key ?? []) }
}

/** Столько ключей показываем поштучно — самые популярные идут первыми. */
const KEYS_SHOWN = 120
const QUEUE_SHOWN = 24

function KeyMap({ w, hl }: { w: ShardWorld; hl: Highlight }) {
  const shown = Math.min(w.config.keys, KEYS_SHOWN)
  const share = useMemo(() => zipfCdf(w.config.keys, w.config.zipf), [w.config.keys, w.config.zipf])
  const moving = Object.keys(w.moving).length
  return (
    <section className="sh-keys" aria-label="Раскладка ключей">
      <header>
        <strong>Ключи</strong>
        <span className="kf-muted">
          цвет — владелец; всего {w.config.keys}
          {moving > 0 && <span className="ld-bad"> · переезжает {moving}</span>}
        </span>
      </header>
      <div className="sh-keymap">
        {Array.from({ length: shown }, (_, k) => {
          const owner = w.owner[k]!
          const from = w.moving[k]
          const pctShare = ((k === 0 ? share[0]! : share[k]! - share[k - 1]!) * 100).toFixed(1)
          return (
            <span
              key={k}
              className={`sh-key ld-c-r${owner % 8} ${from !== undefined ? 'is-moving' : ''} ${hl.key.has(k) ? 'is-hl' : ''}`}
              title={`k${k}: ${w.shards[owner]?.name ?? ''}${from !== undefined ? ` (переезжает с ${w.shards[from]?.name ?? `s${from + 1}`})` : ''} · доля запросов ${pctShare}%`}
            />
          )
        })}
        {w.config.keys > shown && <span className="ld-more">+{w.config.keys - shown}</span>}
      </div>
    </section>
  )
}

function ShardCard({ w, s, hl }: { w: ShardWorld; s: Shard; hl: Highlight }) {
  const shown = s.queue.slice(0, QUEUE_SHOWN)
  const total = w.shards.reduce((a, x) => a + x.served, 0)
  return (
    <section className={`lb-replica ld-c-r${s.id % 8} ${hl.shard.has(s.id) ? 'is-hl' : ''}`}>
      <header>
        <i className="lb-swatch" aria-hidden="true" />
        <strong>{s.name}</strong>
        <span className="lb-badge">{s.owns} ключей</span>
      </header>
      <div className="ld-chips" aria-label={`Очередь ${s.name}`}>
        {s.queue.length === 0 && <span className="kf-muted">очередь пуста</span>}
        {shown.map((q) => (
          <span
            key={q.id}
            className={`ld-chip ${q.group !== null ? 'is-joined' : ''} ${q.migrating ? 'is-zombie' : ''} ${hl.key.has(q.key) ? 'is-hl' : ''}`}
            title={`${q.group !== null ? 'часть веерного запроса' : `k${q.key}`}, ждёт ${w.tick - q.arriveAt} т.${q.migrating ? ' — ключ ещё переезжает' : ''}`}
          >
            {q.group !== null ? '∗' : ''}
          </span>
        ))}
        {s.queue.length > QUEUE_SHOWN && <span className="ld-more">+{s.queue.length - QUEUE_SHOWN}</span>}
      </div>
      <div className="lb-workers">
        {s.workers.map((q, i) => (
          <span key={i} className={`lb-worker ${q ? 'is-busy' : ''} ${q?.migrating ? 'is-zombie' : ''}`} title={q ? `${q.group !== null ? 'веерный' : `k${q.key}`}: ${w.tick - (q.startAt ?? w.tick)} из ${(q.finishAt ?? w.tick) - (q.startAt ?? w.tick)} т.` : 'свободен'}>
            {q && <span style={{ width: `${Math.min(100, ((w.tick - (q.startAt ?? w.tick)) / Math.max(1, (q.finishAt ?? w.tick) - (q.startAt ?? w.tick))) * 100)}%` }} />}
          </span>
        ))}
      </div>
      <span className="lb-stats">
        обработал <b>{s.served}</b>
        {total > 0 && ` · ${Math.round((s.served / total) * 100)}% всех запросов`}
      </span>
    </section>
  )
}

const SCHEME_LABEL: Record<string, string> = {
  mod: 'остаток от деления',
  ring: 'кольцо',
  range: 'диапазоны',
}

export function Board({ sim, cursor, highlight: hl, onSeek }: { sim: Simulation; cursor: number; highlight: Highlight; onSeek: (t: number) => void }) {
  const w = sim.history[cursor]!.world
  const last = sim.history.length - 1
  const d = useMemo(() => chartData(sim.log), [sim])
  const c = w.config
  const next = useMemo(() => movedShare(c, w.shards.length, w.shards.length + 1), [c, w.shards.length])

  return (
    <div className="board ld-board">
      <div className="ld-charts">
        <Chart
          title="Запросы в шардах"
          note="в очереди и в работе"
          series={d.load.map((values, i) => ({ label: `s${i + 1}`, color: `r${i % 8}` as const, values }))}
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
            ...(d.anyScatter ? [{ label: 'веерные, p50', color: 'in' as const, values: d.scatter }] : []),
          ]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
        {d.anyMigrating && (
          <Chart
            title="Переезд ключей"
            note="осталось перевезти"
            series={[{ label: 'ключей', color: 'in', values: d.migrating, area: true }]}
            last={last}
            cursor={cursor}
            onSeek={onSeek}
          />
        )}
      </div>

      <p className="lb-summary">
        раскладка <b>{SCHEME_LABEL[c.scheme]}</b>
        {c.scheme === 'ring' ? ` (${c.vnodes} виртуальных узлов на шард)` : ''} · шардов <b>{w.shards.length}</b> · перекос нагрузки{' '}
        <b>{imbalance(w).toFixed(2)}×</b> · добавить ещё один шард сейчас — переехало бы <b>{Math.round(next * 100)}%</b> ключей
      </p>
      <KeyMap w={w} hl={hl} />
      <div className="lb-replicas">
        {w.shards.map((s) => (
          <ShardCard key={s.id} w={w} s={s} hl={hl} />
        ))}
      </div>
      <p className="rp-legend" aria-hidden="true">
        <span className="sh-key ld-c-r0" /> ключ шарда s1 <span className="sh-key ld-c-r1 is-moving" /> ключ переезжает
        <span className="ld-chip is-joined">∗</span> часть веерного запроса <span className="ld-chip is-zombie" /> запрос к переезжающему ключу
      </p>
    </div>
  )
}
