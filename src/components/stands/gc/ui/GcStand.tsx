import { useEffect, useMemo, useRef, useState } from 'react'
import type { GcEvent } from '../engine/types.ts'
import { GC_SCENARIOS, gcScenarioById } from '../engine/scenarios.ts'
import { Board, highlightOf } from './Board.tsx'
import { ConfigKnobs, WorkloadsEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface GcStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор нагрузок, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [1, 2, 4, 8, 16]

const FINISH_LABEL: Record<string, string> = {
  'all-done': 'Все горутины завершились.',
  oom: 'Память кончилась: куча упёрлась в потолок, сборщик не успел ничего освободить.',
  'stop-after': 'Прогон остановлен по лимиту тиков — нагрузка в сценарии бесконечная.',
}

export default function GcStand({ scenario = 'first-cycle', terms = {}, mode = 'embed' }: GcStandProps) {
  const lab = mode === 'lab'
  const preset = gcScenarioById(scenario) ?? GC_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<GcEvent | null>(null)
  const [settingsOpen, setSettingsOpen] = useState(lab)

  const pb = useSimulation(sc, st.seed, { autoPause })
  const snap = pb.sim.history[pb.cursor]!

  useEffect(() => setSelected(null), [pb.sim])
  useEffect(() => {
    if (pb.pausedOn) setSelected(pb.pausedOn)
  }, [pb.pausedOn])

  const hlEvent = selected && selected.tick === pb.cursor ? selected : null
  const highlight = useMemo(() => {
    if (hlEvent) return highlightOf(hlEvent)
    const h = highlightOf(null)
    for (const e of snap.events) {
      if (e.type === 'alloc' || e.type === 'mark.scan') continue
      e.actors.cells?.forEach((x) => h.cells.add(x))
      e.actors.mut?.forEach((x) => h.mut.add(x))
    }
    return h
  }, [hlEvent, snap])

  /** Доля процессорного времени, ушедшая сборщику к текущему тику. */
  const gcShare = useMemo(() => {
    const { gcSlotTicks, mutatorSlotTicks } = snap.world.stats
    const total = gcSlotTicks + mutatorSlotTicks
    return total === 0 ? 0 : Math.round((gcSlotTicks / total) * 100)
  }, [snap])

  const rootRef = useRef<HTMLDivElement>(null)
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
  const labHref = `/lab/gc/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(gcScenarioById(st.base) ?? preset))
  const lost = snap.world.stats.lost

  return (
    <div className="stand" ref={rootRef} tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(gcScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {GC_SCENARIOS.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.title}
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
          <button type="button" className="btn" onClick={() => { pb.setPlaying(false); pb.setCursor(0) }} title="В начало">
            ⏮
          </button>
          <button type="button" className="btn" onClick={() => { pb.setPlaying(false); pb.setCursor(pb.cursor - 1) }} title="Шаг назад (←)">
            ◀
          </button>
          <button type="button" className="btn btn-primary" onClick={() => pb.setPlaying(!pb.playing)} title="Пуск/пауза (пробел)">
            {pb.playing ? '❚❚ пауза' : '▶ пуск'}
          </button>
          <button type="button" className="btn" onClick={() => { pb.setPlaying(false); pb.setCursor(pb.cursor + 1) }} title="Шаг вперёд (→)">
            ▶
          </button>
          <button type="button" className="btn" onClick={() => { pb.setPlaying(false); pb.setCursor(pb.last) }} title="В конец">
            ⏭
          </button>
        </div>
        <span className="controls-tick">
          тик <b>{pb.cursor}</b> / {pb.last}
        </span>
        <label className="controls-inline">
          скорость
          <select value={pb.speed} onChange={(e) => pb.setSpeed(Number(e.target.value))}>
            {SPEEDS.map((s) => (
              <option key={s} value={s}>
                {s} тик/с
              </option>
            ))}
          </select>
        </label>
        <label className="controls-inline" title="Останавливаться при первом появлении ключевых для сценария событий">
          <input type="checkbox" checked={autoPause} onChange={(e) => setAutoPause(e.target.checked)} />
          пауза на важном
        </label>
      </div>

      <Scrubber
        history={pb.sim.history}
        cursor={pb.cursor}
        last={pb.last}
        watchFor={sc.watchFor}
        onSeek={(t) => { pb.setPlaying(false); pb.setCursor(t) }}
      />

      <Board world={snap.world} highlight={highlight} />

      <div className="metrics">
        <span>циклов <b>{snap.world.stats.cycles}</b></span>
        <span title="Суммарно тиков, когда мир был остановлен">пауз <b>{snap.world.stats.stwTicks}</b></span>
        <span title="Доля процессорного времени, ушедшая на разметку, помощь, паузы и подметание">
          CPU сборщику <b>{gcShare}%</b>
        </span>
        <span title="Тики, которые горутины отдали разметке вместо своей работы">
          помощь <b>{snap.world.stats.assistTicks}</b>
        </span>
        <span title="Тики, в которые горутины делали свою работу">
          полезной работы <b>{snap.world.muts.reduce((s, m) => s + m.cpuTicks, 0)}</b>
        </span>
        <span title="Самая большая куча за прогон">пик кучи <b>{snap.world.stats.peakHeap}</b></span>
        <span className={lost > 0 ? 'metric-bad' : undefined} title="Достижимые объекты, которые сборщик освободил. При исправном барьере записи — всегда 0">
          потеряно <b>{lost}</b>
        </span>
      </div>

      {pb.cursor === pb.last && pb.sim.world.finishReason && (
        <p className={`finish finish-${pb.sim.world.finishReason}`}>{FINISH_LABEL[pb.sim.world.finishReason]}</p>
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
              pb.setPlaying(false)
              pb.setCursor(e.tick)
              setSelected(e)
            }}
          />
        </div>
        <div className="stand-detail">
          {selected ? (
            <EventDetail event={selected} terms={terms} onClose={() => setSelected(null)} />
          ) : (
            <p className="detail-hint">
              Кликните событие в ленте — здесь появится разбор: что произошло, почему сборщик так устроен
              и какие функции рантайма за этим стоят.
            </p>
          )}
        </div>
      </div>

      <details className="settings" open={settingsOpen} onToggle={(e) => setSettingsOpen(e.currentTarget.open)}>
        <summary>Настройки {modified && <span className="settings-mod">изменены</span>}</summary>
        <div className="settings-body">
          <div className="settings-row">
            <label className="controls-inline">
              seed
              <input
                type="number"
                min={1}
                value={st.seed}
                onChange={(e) => update({ seed: Math.max(1, Number(e.target.value) || 1) })}
              />
            </label>
            <button type="button" className="btn-ghost" onClick={() => update({ seed: 1 + Math.floor(Math.random() * 9999) })}>
              другой seed
            </button>
            {modified && (
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(gcScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            )}
          </div>
          <ConfigKnobs config={st.config} onChange={(config) => update({ config })} />
          {lab && (
            <>
              <h4 className="settings-sub">Нагрузки</h4>
              <WorkloadsEditor workloads={sc.workloads} onChange={(workloads) => update({ workloads })} />
            </>
          )}
        </div>
      </details>
    </div>
  )
}
