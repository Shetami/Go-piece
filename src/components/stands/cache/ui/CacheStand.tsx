import { useEffect, useMemo, useState } from 'react'
import type { CacheEvent, CacheWorld } from '../engine/types.ts'
import { CACHE_SCENARIOS, cacheScenarioById } from '../engine/scenarios.ts'
import { percentile, readLatencies } from '../engine/world.ts'
import { Board, highlightOf } from './Board.tsx'
import { ConfigKnobs, FaultsEditor, PhasesEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface CacheStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор событий кэша и фаз нагрузки, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [2, 5, 10, 20]

/** Чем кончился прогон — одной фразой. */
function finishLabel(w: CacheWorld, p50: number, p99: number): string {
  const s = w.stats
  const ratio = Math.round(((s.hits + s.swr) / Math.max(1, s.reads)) * 100)
  const parts = [`За ${w.tick} тиков было ${s.reads} чтений, ${ratio}% из них обслужил кэш.`]
  parts.push(`Задержка чтения: медиана ${p50}, p99 ${p99} тиков.`)
  parts.push(`В базу ушло ${s.dbReads} чтений${s.dbWrites > 0 ? ` и ${s.dbWrites} записей` : ''}; самая длинная очередь — ${s.maxDbQueue}.`)
  if (s.stale > 0) parts.push(`Устаревших ответов: ${s.stale}.`)
  if (s.joins > 0) parts.push(`Промахов, присоединившихся к чужому походу в базу: ${s.joins}.`)
  return parts.join(' ')
}

export default function CacheStand({ scenario = 'hot-keys', terms = {}, mode = 'embed' }: CacheStandProps) {
  const lab = mode === 'lab'
  const preset = cacheScenarioById(scenario) ?? CACHE_SCENARIOS[0]!

  const [st, setSt] = useState<StandState>(() => initialState(preset))

  useEffect(() => {
    if (lab && window.location.hash.length > 1) setSt(decodeState(window.location.hash, preset))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (lab) history.replaceState(null, '', `#${encodeState(st)}`)
  }, [lab, st])

  const sc = useMemo(() => toScenario(st), [st])
  const [autoPause, setAutoPause] = useState(true)
  const [showMinor, setShowMinor] = useState(false)
  const [selected, setSelected] = useState<CacheEvent | null>(null)
  const [settingsOpen, setSettingsOpen] = useState(lab)

  const pb = useSimulation(sc, { autoPause })
  const frame = pb.sim.history[pb.cursor]!

  useEffect(() => setSelected(null), [pb.sim])
  useEffect(() => {
    if (pb.pausedOn) setSelected(pb.pausedOn)
  }, [pb.pausedOn])

  const hlEvent = selected && selected.tick === pb.cursor ? selected : null
  const highlight = useMemo(() => {
    if (hlEvent) return highlightOf(hlEvent)
    const h = highlightOf(null)
    for (const e of frame.events) {
      if (e.type === 'req.hit' || e.type === 'req.miss' || e.type === 'db.done') continue
      e.actors.key?.forEach((x) => h.key.add(x))
      e.actors.query?.forEach((x) => h.query.add(x))
    }
    return h
  }, [hlEvent, frame])

  const w = frame.world
  const lat = useMemo(() => readLatencies(pb.sim.log.done, pb.cursor), [pb.sim, pb.cursor])
  const p50 = percentile(lat, 50)
  const p99 = percentile(lat, 99)

  const onKey = (e: React.KeyboardEvent) => {
    const tag = (e.target as HTMLElement).tagName
    if (tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA') return
    if (e.key === ' ') {
      e.preventDefault()
      pb.setPlaying(!pb.playing)
    } else if (e.key === 'ArrowRight') {
      e.preventDefault()
      pb.setPlaying(false)
      pb.setCursor(pb.cursor + (e.shiftKey ? 10 : 1))
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      pb.setPlaying(false)
      pb.setCursor(pb.cursor - (e.shiftKey ? 10 : 1))
    }
  }

  const update = (patch: Partial<StandState>) => setSt((s) => ({ ...s, ...patch }))
  const seek = (t: number) => {
    pb.setPlaying(false)
    pb.setCursor(t)
  }
  const labHref = `/lab/cache/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(cacheScenarioById(st.base) ?? preset))
  const s = w.stats

  return (
    <div className="stand" tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(cacheScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {CACHE_SCENARIOS.map((x) => (
                <option key={x.id} value={x.id}>
                  {x.title}
                </option>
              ))}
            </select>
          ) : (
            <h3 className="stand-title">
              <span className="stand-kicker">Стенд</span>
              {sc.title}
            </h3>
          )}
          <p className="stand-claim">{sc.claim}</p>
        </div>
        {!lab && (
          <a className="btn-ghost stand-open" href={labHref} title="Открыть этот прогон в лаборатории">
            в лабораторию ↗
          </a>
        )}
      </header>

      <div className="controls">
        <div className="controls-group">
          <button type="button" className="btn" onClick={() => seek(0)} title="В начало">
            ⏮
          </button>
          <button type="button" className="btn" onClick={() => seek(pb.cursor - 1)} title="Шаг назад (←)">
            ◀
          </button>
          <button type="button" className="btn btn-primary" onClick={() => pb.setPlaying(!pb.playing)} title="Пуск/пауза (пробел)">
            {pb.playing ? '❚❚ пауза' : '▶ пуск'}
          </button>
          <button type="button" className="btn" onClick={() => seek(pb.cursor + 1)} title="Шаг вперёд (→)">
            ▶
          </button>
          <button type="button" className="btn" onClick={() => seek(pb.last)} title="В конец">
            ⏭
          </button>
        </div>
        <span className="controls-tick">
          тик <b>{pb.cursor}</b> / {pb.last}
        </span>
        <label className="controls-inline">
          скорость
          <select value={pb.speed} onChange={(e) => pb.setSpeed(Number(e.target.value))}>
            {SPEEDS.map((x) => (
              <option key={x} value={x}>
                {x} тик/с
              </option>
            ))}
          </select>
        </label>
        <label className="controls-inline" title="Останавливаться при первом появлении ключевых для сценария событий">
          <input type="checkbox" checked={autoPause} onChange={(e) => setAutoPause(e.target.checked)} />
          пауза на важном
        </label>
      </div>

      <Scrubber history={pb.sim.history} cursor={pb.cursor} last={pb.last} watchFor={sc.watchFor} onSeek={seek} />

      <Board sim={pb.sim} cursor={pb.cursor} highlight={highlight} onSeek={seek} />

      <div className="metrics">
        <span title="Доля чтений, которые обслужил кэш, — с начала прогона">
          попаданий <b>{s.reads === 0 ? 0 : Math.round(((s.hits + s.swr) / s.reads) * 100)}%</b>
        </span>
        <span title="Задержка чтений: медиана и 99-й перцентиль, тиков">
          p50 <b>{p50}</b> · p99 <b>{p99}</b>
        </span>
        <span title="Сколько чтений дошло до базы">
          в базу <b>{s.dbReads}</b> из {s.reads}
        </span>
        <span title="Самая длинная очередь в базе за прогон">
          макс. очередь <b>{s.maxDbQueue}</b>
        </span>
        {s.maxSameKey > 1 && (
          <span className={s.maxSameKey >= 4 ? 'metric-bad' : undefined} title="Самый большой всплеск одновременных запросов в базу за один и тот же ключ">
            дублей за ключ <b>{s.maxSameKey}</b>
          </span>
        )}
        {s.joins > 0 && (
          <span title="Промахи, которые присоединились к уже летящему запросу вместо своего">
            объединено <b>{s.joins}</b>
          </span>
        )}
        {s.swr > 0 && (
          <span title="Чтения, получившие просроченное значение сразу, без ожидания базы">
            просроченных <b>{s.swr}</b>
          </span>
        )}
        {s.stale > 0 && (
          <span className="metric-bad" title="Чтения, получившие значение старее, чем в базе">
            устаревших <b>{s.stale}</b>
          </span>
        )}
      </div>

      {pb.cursor === pb.last && pb.sim.world.finishReason && (
        <p className={`finish ${pb.sim.world.stats.maxDbQueue > 40 ? 'finish-panic' : ''}`}>{finishLabel(w, p50, p99)}</p>
      )}

      <div className="stand-log">
        <div className="stand-feed">
          <div className="feed-head">
            <span>События</span>
            <label className="controls-inline">
              <input type="checkbox" checked={showMinor} onChange={(e) => setShowMinor(e.target.checked)} />
              мелкие
            </label>
          </div>
          <EventFeed
            events={pb.eventsUpTo}
            selected={selected}
            showMinor={showMinor}
            onSelect={(e) => {
              seek(e.tick)
              setSelected(e)
            }}
          />
        </div>
        <div className="stand-detail">
          {selected ? (
            <EventDetail event={selected} terms={terms} onClose={() => setSelected(null)} />
          ) : (
            <p className="detail-hint">
              Кликните событие в ленте — здесь появится разбор: что случилось с ключом, почему кэш повёл себя
              именно так и как это делается в коде на Go.
            </p>
          )}
        </div>
      </div>

      <details className="settings" open={settingsOpen} onToggle={(e) => setSettingsOpen(e.currentTarget.open)}>
        <summary>Настройки {modified && <span className="settings-mod">изменены</span>}</summary>
        <div className="settings-body">
          {modified && (
            <div className="settings-row">
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(cacheScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            </div>
          )}
          <ConfigKnobs config={st.config} onChange={(config) => update({ config })} />
          <h4 className="settings-sub">События кэша</h4>
          <FaultsEditor faults={sc.faults} onChange={(faults) => update({ faults })} />
          <h4 className="settings-sub">Фазы нагрузки</h4>
          <PhasesEditor phases={sc.phases} onChange={(phases) => update({ phases })} />
        </div>
      </details>
    </div>
  )
}
