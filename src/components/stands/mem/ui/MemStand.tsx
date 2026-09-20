import { useEffect, useMemo, useRef, useState } from 'react'
import type { MemEvent } from '../engine/types.ts'
import { MEM_SCENARIOS, memScenarioById } from '../engine/scenarios.ts'
import { Board, highlightOf } from './Board.tsx'
import { ConfigKnobs, WorkloadsEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface MemStandProps {
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
  oom: 'Память кончилась: в арене не нашлось непрерывного куска нужного размера.',
  'stop-after': 'Прогон остановлен по лимиту тиков — нагрузка в сценарии бесконечная.',
}

export default function MemStand({ scenario = 'fast-path', terms = {}, mode = 'embed' }: MemStandProps) {
  const lab = mode === 'lab'
  const preset = memScenarioById(scenario) ?? MEM_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<MemEvent | null>(null)
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
      if (e.type === 'alloc.fast' || e.type === 'obj.free') continue
      e.actors.pages?.forEach((x) => h.pages.add(x))
      e.actors.span?.forEach((x) => h.spans.add(x))
      e.actors.g?.forEach((x) => h.g.add(x))
    }
    return h
  }, [hlEvent, snap])

  /** Сколько аллокаций обошлись без похода в общий список — к текущему тику. */
  const summary = useMemo(() => {
    const s = snap.world.stats
    const total = s.onStack + s.tiny + s.fast + s.refills + s.large
    const fast = s.onStack + s.tiny + s.fast
    return {
      total,
      fastShare: total === 0 ? 0 : Math.round((fast / total) * 100),
      waste: s.roundSlot === 0 ? 0 : Math.round(((s.roundSlot - s.roundAsked) / s.roundSlot) * 100),
      used: snap.world.pages.filter((p) => p.kind === 'span' || p.kind === 'large' || p.kind === 'stack').length,
      rss: snap.world.pages.filter((p) => p.kind !== 'returned').length,
    }
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
  const labHref = `/lab/mem/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(memScenarioById(st.base) ?? preset))
  const stalls = snap.world.stats.contended

  return (
    <div className="stand" ref={rootRef} tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(memScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {MEM_SCENARIOS.map((s) => (
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

      <Board world={snap.world} events={snap.events} highlight={highlight} />

      <div className="metrics">
        <span title="Все аллокации: на стеке, мелочью, из кэша, через общий список и большие">
          аллокаций <b>{summary.total}</b>
        </span>
        <span title="Доля аллокаций, обошедшихся без общей блокировки: стек, общий блок и кэш процессора">
          быстрым путём <b>{summary.fastShare}%</b>
        </span>
        <span title="Объекты, которые компилятор оставил на стеке — куча их не видит">
          на стеке <b>{snap.world.stats.onStack}</b>
        </span>
        <span title="Сколько выделенной памяти съело округление до класса размеров">
          потери на классе <b>{summary.waste}%</b>
        </span>
        <span title="Страницы под спаны, большие объекты и стеки">
          занято <b>{summary.used}</b> стр.
        </span>
        <span title="Сколько страниц рантайм держит у операционной системы — это и есть RSS процесса">
          у ОС <b>{summary.rss}</b>
        </span>
        <span className={stalls > 0 ? 'metric-bad' : undefined} title="Сколько раз горутина ждала общую блокировку mcentral">
          ожиданий <b>{stalls}</b>
        </span>
        <span title="Сколько раз стек горутины переезжал на новое место">
          ростов стека <b>{snap.world.stats.stackGrows}</b>
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
              Кликните событие в ленте — здесь появится разбор: что произошло, почему аллокатор так устроен
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
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(memScenarioById(st.base) ?? preset))}>
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
