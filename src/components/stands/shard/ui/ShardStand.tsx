import { useEffect, useMemo, useState } from 'react'
import type { ShardEvent, ShardWorld } from '../engine/types.ts'
import { SHARD_SCENARIOS, shardScenarioById } from '../engine/scenarios.ts'
import { imbalance, latencies, percentile } from '../engine/world.ts'
import { Board, highlightOf } from './Board.tsx'
import { ConfigKnobs, FaultsEditor, PhasesEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface ShardStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор событий и фаз нагрузки, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [2, 5, 10, 20]

/** Чем кончился прогон — одной фразой. */
function finishLabel(w: ShardWorld, p50: number, p99: number): string {
  const s = w.stats
  const parts = [`За ${w.tick} тиков пришло ${s.requests} запросов, в шарды ушло ${s.subRequests}, ответов ${s.done}.`]
  parts.push(`Задержка: медиана ${p50}, p99 ${p99} тиков.`)
  if (s.moved > 0) parts.push(`Переехало ${s.moved} ключей, из них ${s.migratingHits} запросов застали ключ в пути.`)
  const hot = [...w.shards].sort((a, b) => b.served - a.served)[0]
  if (hot && imbalance(w) > 1.3) parts.push(`Перекос ${imbalance(w).toFixed(2)}×: больше всех работал ${hot.name} — ${hot.served} запросов.`)
  return parts.join(' ')
}

export default function ShardStand({ scenario = 'add-shard', terms = {}, mode = 'embed' }: ShardStandProps) {
  const lab = mode === 'lab'
  const preset = shardScenarioById(scenario) ?? SHARD_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<ShardEvent | null>(null)
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
      if (e.type === 'req.route' || e.type === 'req.done' || e.type === 'scatter.start' || e.type === 'reshard.move') continue
      e.actors.key?.forEach((x) => h.key.add(x))
      e.actors.shard?.forEach((x) => h.shard.add(x))
    }
    return h
  }, [hlEvent, frame])

  const w = frame.world
  const lat = useMemo(() => latencies(pb.sim.log.done, pb.cursor), [pb.sim, pb.cursor])
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
  const labHref = `/lab/shard/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(shardScenarioById(st.base) ?? preset))
  const s = w.stats

  return (
    <div className="stand" tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(shardScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {SHARD_SCENARIOS.map((x) => (
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
        <span title="Запросы пользователей и во что они превратились для шардов">
          в шарды <b>{s.subRequests}</b> из {s.requests}
        </span>
        <span title="Задержка ответов: медиана и 99-й перцентиль, тиков">
          p50 <b>{p50}</b> · p99 <b>{p99}</b>
        </span>
        <span className={imbalance(w) > 1.5 ? 'metric-bad' : undefined} title="Во сколько раз самый загруженный шард обработал больше среднего">
          перекос <b>{imbalance(w).toFixed(2)}×</b>
        </span>
        <span title="Самая длинная очередь на одном шарде за прогон">
          макс. очередь <b>{s.maxQueue}</b>
        </span>
        {s.scatter > 0 && (
          <span title="Запросы без ключа шардирования: ушли во все шарды сразу">
            веерных <b>{s.scatter}</b>
          </span>
        )}
        {s.moved > 0 && (
          <span title="Ключей переехало при решардировании">
            переехало <b>{s.moved}</b>
          </span>
        )}
        {s.migratingHits > 0 && (
          <span className="metric-bad" title="Запросы к ключам, которые ещё не доехали: они стоят вдвое дороже">
            к переезжающим <b>{s.migratingHits}</b>
          </span>
        )}
      </div>

      {pb.cursor === pb.last && pb.sim.world.finishReason && (
        <p className={`finish ${imbalance(pb.sim.world) > 1.8 ? 'finish-panic' : ''}`}>{finishLabel(w, p50, p99)}</p>
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
              Кликните событие в ленте — здесь появится разбор: куда ушёл запрос, почему ключи переезжают именно
              так и что с этим делают в реальных хранилищах.
            </p>
          )}
        </div>
      </div>

      <details className="settings" open={settingsOpen} onToggle={(e) => setSettingsOpen(e.currentTarget.open)}>
        <summary>Настройки {modified && <span className="settings-mod">изменены</span>}</summary>
        <div className="settings-body">
          {modified && (
            <div className="settings-row">
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(shardScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            </div>
          )}
          <ConfigKnobs config={st.config} onChange={(config) => update({ config })} />
          <h4 className="settings-sub">События</h4>
          <FaultsEditor faults={sc.faults} onChange={(faults) => update({ faults })} />
          <h4 className="settings-sub">Фазы нагрузки</h4>
          <PhasesEditor phases={sc.phases} onChange={(phases) => update({ phases })} />
        </div>
      </details>
    </div>
  )
}
