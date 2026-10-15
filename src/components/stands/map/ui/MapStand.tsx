import { useEffect, useMemo, useState } from 'react'
import type { MapEvent, MapWorld } from '../engine/types.ts'
import { MAP_SCENARIOS, mapScenarioById } from '../engine/scenarios.ts'
import { mapCapacity, mapLen } from '../engine/world.ts'
import { Board } from './Board.tsx'
import { Console } from './Console.tsx'
import { ConfigKnobs, ScriptEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { HARD_CAP, useSimulation } from './useSimulation.ts'

export interface MapStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор команд, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [1, 2, 4, 8, 16]

function finishLabel(w: MapWorld): string {
  switch (w.finishReason) {
    case 'all-done':
      return 'Сценарий выполнен. Добавьте команду в консоли — стенд продолжит с этого места.'
    case 'panic':
      return 'Паника — программа завершилась бы аварийно, если её не перехватить recover.'
    case 'fatal':
      return 'fatal error — процесс завершён. recover не помогает, defer не выполняются.'
    default:
      return `Прогон обрезан на ${HARD_CAP} операциях: история хранит снимок карты на каждую.`
  }
}

export default function MapStand({ scenario = 'small', terms = {}, mode = 'embed' }: MapStandProps) {
  const lab = mode === 'lab'
  const preset = mapScenarioById(scenario) ?? MAP_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<MapEvent | null>(null)
  const [settingsOpen, setSettingsOpen] = useState(lab)

  const pb = useSimulation(sc, st.seed, { autoPause })
  const snap = pb.sim.history[pb.cursor]!

  useEffect(() => setSelected(null), [pb.sim])
  useEffect(() => {
    if (pb.pausedOn) setSelected(pb.pausedOn)
  }, [pb.pausedOn])

  const summary = useMemo(() => {
    const w = snap.world
    const s = w.stats
    const cap = mapCapacity(w)
    const len = mapLen(w)
    return {
      len,
      cap,
      load: cap === 0 ? 0 : Math.round((len / cap) * 100),
      probesPerOp: s.ops <= 1 ? 0 : Math.round((s.probes / (s.ops - 1)) * 100) / 100,
    }
  }, [snap])

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
  const labHref = `/lab/map/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(mapScenarioById(st.base) ?? preset))

  const ended = pb.sim.world.finishReason
  const consoleBlocked =
    ended === 'panic' || ended === 'fatal'
      ? 'Прогон закончился аварийно — команды после этого не выполнятся. Уберите последнюю строку сценария в настройках или начните заново.'
      : ended === 'stop-after'
        ? `Прогон уже упёрся в предел ${HARD_CAP} операций.`
        : null

  const runCommand = (cmd: string) => {
    pb.resumeAt(pb.last, true)
    setSt((s) => ({ ...s, script: `${(s.script ?? sc.script).trimEnd()}\n${cmd}` }))
  }

  return (
    <div className="stand map-stand" tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(mapScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {MAP_SCENARIOS.map((s) => (
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
          операция <b>{pb.cursor}</b> / {pb.last}
        </span>
        <label className="controls-inline">
          скорость
          <select value={pb.speed} onChange={(e) => pb.setSpeed(Number(e.target.value))}>
            {SPEEDS.map((s) => (
              <option key={s} value={s}>
                {s} оп/с
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

      <Board world={snap.world} trace={snap.trace} />

      <div className="metrics">
        <span title="len(m) — число записей">
          len <b>{summary.len}</b>
        </span>
        <span title="Слотов во всех живых таблицах (корзинах с цепочками)">
          ёмкость <b>{summary.cap}</b>
        </span>
        <span title="Доля занятых слотов">
          заполнено <b>{summary.load}%</b>
        </span>
        <span title="Сколько групп (звеньев цепочки) в среднем проверяет одна операция">
          проб на операцию <b>{summary.probesPerOp}</b>
        </span>
        <span title="Сколько раз ключ сравнивали целиком — дорогое сравнение, которое отсеивает H2 (tophash)">
          сравнений ключей <b>{snap.world.stats.compares}</b>
        </span>
        <span title="Из них впустую: 7 бит H2 (8 бит tophash) совпали, ключ — нет">
          ложных <b>{snap.world.stats.falseMatches}</b>
        </span>
        <span title="Сколько записей переложили при росте, делении, эвакуации">
          переложено <b>{snap.world.stats.moved}</b>
        </span>
        {snap.world.legacy ? (
          <>
            <span>
              переполнений <b>{snap.world.stats.overflows}</b>
            </span>
            <span>
              эвакуировано корзин <b>{snap.world.stats.evacuated}</b>
            </span>
          </>
        ) : (
          <>
            <span title="Удвоений таблиц">
              рост <b>{snap.world.stats.grows}</b>
            </span>
            <span title="Делений таблиц на потолке">
              делений <b>{snap.world.stats.splits}</b>
            </span>
            {snap.world.stats.prunes > 0 && (
              <span title="Прополок надгробий">
                прополок <b>{snap.world.stats.prunes}</b>
              </span>
            )}
          </>
        )}
      </div>

      {pb.cursor === pb.last && ended && (
        <p className={`finish finish-${ended}`}>
          {finishLabel(pb.sim.world)}
          {pb.sim.world.panic && <code> {ended === 'fatal' ? 'fatal error' : 'panic'}: {pb.sim.world.panic}</code>}
        </p>
      )}

      <Console onRun={runCommand} disabled={consoleBlocked} />

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
              Кликните событие в ленте — здесь появится разбор: что произошло с картой, почему она устроена именно
              так и какие функции рантайма за этим стоят.
            </p>
          )}
        </div>
      </div>

      <details className="settings" open={settingsOpen} onToggle={(e) => setSettingsOpen(e.currentTarget.open)}>
        <summary>Настройки и сценарий {modified && <span className="settings-mod">изменены</span>}</summary>
        <div className="settings-body">
          <div className="settings-row">
            <label className="controls-inline" title="Seed определяет соль хэша и точки старта обхода">
              seed
              <input
                type="number"
                min={1}
                value={st.seed}
                onChange={(e) => update({ seed: Math.max(1, Number(e.target.value) || 1) })}
              />
            </label>
            <button type="button" className="btn-ghost" onClick={() => update({ seed: 1 + Math.floor(Math.random() * 9999) })}>
              другая соль
            </button>
            {modified && (
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(mapScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            )}
          </div>
          <ConfigKnobs config={st.config} onChange={(config) => update({ config })} />
          <h4 className="settings-sub">Сценарий — по команде на строку</h4>
          <p className="map-help">
            <code>put k v</code> или <code>m["k"] = v</code> · <code>get k</code> · <code>del k</code> ·{' '}
            <code>put k1..k20</code> · <code>fill 20 [префикс]</code> · <code>collide 6 [группа]</code> ·{' '}
            <code>del @last</code>, <code>@last:3</code>, <code>@last:-1</code> · <code>range</code> ·{' '}
            <code>next [n|all]</code> · <code>clear</code> · <code>race [write|read|iter]</code> · комментарии после{' '}
            <code>#</code>
          </p>
          <ScriptEditor
            script={sc.script}
            errors={pb.sim.errors}
            onChange={(script) => {
              pb.resumeAt(Number.MAX_SAFE_INTEGER, false)
              update({ script })
            }}
          />
        </div>
      </details>
    </div>
  )
}
