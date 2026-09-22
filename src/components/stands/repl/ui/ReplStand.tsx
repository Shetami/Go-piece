import { useEffect, useMemo, useState } from 'react'
import type { ReplEvent, ReplWorld } from '../engine/types.ts'
import { REPL_SCENARIOS, replScenarioById } from '../engine/scenarios.ts'
import { lagOf, replicasOf, retainedOf } from '../engine/world.ts'
import { Board, highlightOf } from './Board.tsx'
import { ClientsEditor, ConfigKnobs, FaultsEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface ReplStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор клиентов и сбоев, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [1, 2, 4, 8]

/** Чем кончился прогон — одной фразой. */
function finishLabel(world: ReplWorld): string {
  const s = world.stats
  const parts: string[] = [world.finishReason === 'stop-after' ? 'Прогон остановлен по лимиту тиков.' : 'Клиенты закончили, реплики догнали.']
  parts.push(`Подтверждено коммитов ${s.acked}${s.lost > 0 ? `, потеряно подтверждённых ${s.lost}` : ''}${s.unknown > 0 ? `, с неизвестным исходом ${s.unknown}` : ''}.`)
  if (s.ownStale + s.backwards > 0) parts.push(`Чтений со своим-не-видно или назад во времени: ${s.ownStale + s.backwards}.`)
  if (s.cancels > 0) parts.push(`Отменено запросов на репликах: ${s.cancels}.`)
  if (s.downTicks > 0) parts.push(`Без ведущего — ${s.downTicks} т.`)
  return parts.join(' ')
}

export default function ReplStand({ scenario = 'failover', terms = {}, mode = 'embed' }: ReplStandProps) {
  const lab = mode === 'lab'
  const preset = replScenarioById(scenario) ?? REPL_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<ReplEvent | null>(null)
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
      if (e.type === 'replica.receive' || e.type === 'replica.replay' || e.type === 'read.ok') continue
      e.actors.node?.forEach((x) => h.node.add(x))
      e.actors.client?.forEach((x) => h.client.add(x))
      e.actors.lsn?.forEach((x) => h.lsn.add(x))
    }
    return h
  }, [hlEvent, frame])

  const w = frame.world
  const maxLag = Math.max(0, ...replicasOf(w).filter((n) => n.up && !n.broken).map((n) => lagOf(w, n)))

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
  const labHref = `/lab/repl/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(replScenarioById(st.base) ?? preset))

  return (
    <div className="stand" tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(replScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {REPL_SCENARIOS.map((s) => (
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

      <Scrubber history={pb.sim.history} cursor={pb.cursor} last={pb.last} watchFor={sc.watchFor} onSeek={seek} />

      <Board history={pb.sim.history} cursor={pb.cursor} scenario={sc} highlight={highlight} onSeek={seek} />

      <div className="metrics">
        <span title="Коммиты, о которых клиенты получили подтверждение">
          подтверждено <b>{w.stats.acked}</b>
        </span>
        <span className={w.stats.lost > 0 ? 'metric-bad' : undefined} title="Подтверждённые коммиты, которых нет на новом ведущем">
          потеряно <b>{w.stats.lost}</b>
        </span>
        <span title="Сколько тиков в среднем коммит ждал подтверждения">
          ожидание коммита <b>{w.stats.acked === 0 ? 0 : Math.round((w.stats.commitWaitTicks / Math.max(1, w.stats.commits)) * 10) / 10}</b> т.
        </span>
        <span title="Самое большое отставание реплики прямо сейчас, в записях WAL">
          отставание <b>{maxLag}</b> зап.
        </span>
        {w.stats.reads > 0 && (
          <span className={w.stats.ownStale + w.stats.backwards > 0 ? 'metric-bad' : undefined} title="Чтения, которые не увидели своей записи или ушли назад во времени">
            плохих чтений <b>{w.stats.ownStale + w.stats.backwards}</b> из {w.stats.reads}
          </span>
        )}
        <span title="Сколько записей WAL хранит ведущий">
          WAL на ведущем <b>{retainedOf(w)}</b>
        </span>
        {w.stats.cancels > 0 && (
          <span className="metric-bad" title="Запросы на репликах, отменённые из-за конфликта с восстановлением">
            отменено <b>{w.stats.cancels}</b>
          </span>
        )}
        {w.stats.downTicks > 0 && (
          <span className="metric-bad" title="Тики, когда писать было некуда">
            без ведущего <b>{w.stats.downTicks}</b> т.
          </span>
        )}
      </div>

      {pb.cursor === pb.last && pb.sim.world.finishReason && (
        <p className={`finish ${pb.sim.world.stats.lost > 0 ? 'finish-panic' : ''}`}>{finishLabel(pb.sim.world)}</p>
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
              Кликните событие в ленте — здесь появится разбор: что сделал кластер, почему именно так и какие
              механизмы PostgreSQL за этим стоят.
            </p>
          )}
        </div>
      </div>

      <details className="settings" open={settingsOpen} onToggle={(e) => setSettingsOpen(e.currentTarget.open)}>
        <summary>Настройки {modified && <span className="settings-mod">изменены</span>}</summary>
        <div className="settings-body">
          {modified && (
            <div className="settings-row">
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(replScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            </div>
          )}
          <ConfigKnobs config={st.config} onChange={(config) => update({ config })} />
          <h4 className="settings-sub">Сбои</h4>
          <FaultsEditor faults={sc.faults} replicas={st.config.replicas} onChange={(faults) => update({ faults })} />
          {lab && (
            <>
              <h4 className="settings-sub">Клиенты</h4>
              <ClientsEditor clients={sc.clients} keys={sc.keys} replicas={st.config.replicas} onChange={(clients) => update({ clients })} />
            </>
          )}
        </div>
      </details>
    </div>
  )
}
