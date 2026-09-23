import { useEffect, useMemo, useState } from 'react'
import type { SagaEvent, SagaWorld } from '../engine/types.ts'
import { SAGA_SCENARIOS, sagaScenarioById } from '../engine/scenarios.ts'
import { consistency, latencies, outcome, percentile } from '../engine/world.ts'
import { Board, highlightOf } from './Board.tsx'
import { ConfigKnobs, FaultsEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface SagaStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор сбоев, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [2, 5, 10, 20]

/** Чем кончился прогон — одной фразой. */
function finishLabel(w: SagaWorld, p50: number, p99: number): string {
  const s = w.stats
  const oc = outcome(w)
  const parts = [`За ${w.tick} тиков создано ${s.orders} заказов: доставлено ${s.shipped}${s.cancelled > 0 ? `, отменено ${s.cancelled}` : ''}, без конечного состояния ${oc.stuck}.`]
  parts.push(`Заказ доходил до конца за ${p50} тиков в среднем, p99 — ${p99}.`)
  if (s.lost > 0) parts.push(`Потеряно событий: ${s.lost}.`)
  if (s.doubleCharges > 0) parts.push(`Двойных списаний: ${s.doubleCharges}.`)
  if (s.stuck > 0) parts.push(`Застряли навсегда: ${s.stuck}.`)
  if (s.compensations > 0) parts.push(`Компенсаций: ${s.compensations}.`)
  return parts.join(' ')
}

export default function SagaStand({ scenario = 'dual-write', terms = {}, mode = 'embed' }: SagaStandProps) {
  const lab = mode === 'lab'
  const preset = sagaScenarioById(scenario) ?? SAGA_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<SagaEvent | null>(null)
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
      if (e.type === 'order.created' || e.type === 'event.deliver' || e.type === 'pay.ok' || e.type === 'ship.ok') continue
      e.actors.order?.forEach((x) => h.order.add(x))
      e.actors.msg?.forEach((x) => h.msg.add(x))
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
  const labHref = `/lab/saga/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(sagaScenarioById(st.base) ?? preset))
  const s = w.stats

  return (
    <div className="stand" tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(sagaScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {SAGA_SCENARIOS.map((x) => (
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
        <span title="Доля заказов, дошедших до конечного состояния: доставлен или отменён">
          завершено <b>{Math.round(consistency(w) * 100)}%</b>
        </span>
        <span title="Заказы, доведённые до доставки">
          доставлено <b>{s.shipped}</b> из {s.orders}
        </span>
        <span title="Сколько тиков заказ в среднем шёл до конечного состояния">
          p50 <b>{p50}</b> · p99 <b>{p99}</b>
        </span>
        {s.lost > 0 && (
          <span className="metric-bad" title="События, которые не дошли до брокера и уже не дойдут">
            потеряно событий <b>{s.lost}</b>
          </span>
        )}
        {s.duplicates > 0 && (
          <span title="Повторные доставки одного и того же события">
            дублей <b>{s.duplicates}</b>
          </span>
        )}
        {s.doubleCharges > 0 && (
          <span className="metric-bad" title="Списания сверх одного на заказ — настоящие потерянные деньги">
            двойных списаний <b>{s.doubleCharges}</b>
          </span>
        )}
        {s.compensations > 0 && (
          <span title="Компенсирующие действия саги: возврат денег">
            компенсаций <b>{s.compensations}</b>
          </span>
        )}
        {s.stuck > 0 && (
          <span className="metric-bad" title="Заказы, застрявшие в промежуточном состоянии навсегда">
            застряло <b>{s.stuck}</b>
          </span>
        )}
      </div>

      {pb.cursor === pb.last && pb.sim.world.finishReason && (
        <p className={`finish ${pb.sim.world.stats.stuck > 0 || pb.sim.world.stats.doubleCharges > 0 ? 'finish-panic' : ''}`}>{finishLabel(w, p50, p99)}</p>
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
              Кликните событие в ленте — здесь появится разбор: что случилось с заказом и событием о нём,
              почему сервисы разошлись и как это делается в коде на Go.
            </p>
          )}
        </div>
      </div>

      <details className="settings" open={settingsOpen} onToggle={(e) => setSettingsOpen(e.currentTarget.open)}>
        <summary>Настройки {modified && <span className="settings-mod">изменены</span>}</summary>
        <div className="settings-body">
          {modified && (
            <div className="settings-row">
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(sagaScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            </div>
          )}
          <ConfigKnobs config={st.config} onChange={(config) => update({ config })} />
          <h4 className="settings-sub">Сбои</h4>
          <FaultsEditor faults={sc.faults} onChange={(faults) => update({ faults })} />
        </div>
      </details>
    </div>
  )
}
