import { useEffect, useMemo, useRef, useState } from 'react'
import type { ChanEvent, ChanWorld } from '../engine/types.ts'
import { CHAN_SCENARIOS, chanScenarioById } from '../engine/scenarios.ts'
import { Board, highlightOf } from './Board.tsx'
import { ChansEditor, ConfigKnobs, WorkloadsEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface ChanStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор каналов и нагрузок, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [1, 2, 4, 8, 16]

const gs = (n: number) => `${n} ${n % 10 === 1 && n % 100 !== 11 ? 'горутина' : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 'горутины' : 'горутин'}`

/**
 * Чем кончился прогон.
 *
 * У лимита тиков три разных смысла, и путать их нельзя. Спящий отправитель на
 * полном буфере проснётся, как только получатель возьмёт значение, — это не
 * утечка, а обратное давление. Утечка — только то, что спит дольше leakAfter:
 * ровно тот же порог, по которому движок выдаёт событие «утечка горутин».
 */
function finishLabel(world: ChanWorld): string {
  switch (world.finishReason) {
    case 'all-done':
      return 'Все горутины завершились.'
    case 'deadlock':
      return 'Взаимная блокировка: все горутины спят, и разбудить их некому. Настоящая программа здесь падает.'
    case 'panic':
      return 'Паника — программа завершилась аварийно.'
    default: {
      const parked = world.gs.filter((g) => g.state === 'waiting' && g.wait !== null)
      const stuck = parked.filter((g) => world.tick - g.wait!.since >= world.config.leakAfter)
      if (stuck.length > 0) {
        return `Прогон остановлен по лимиту тиков. ${gs(stuck.length)} ${stuck.length === 1 ? 'спит' : 'спят'} дольше ${world.config.leakAfter} тиков — в настоящей программе ${stuck.length === 1 ? 'она осталась' : 'они остались'} бы в памяти навсегда.`
      }
      if (parked.length > 0) {
        return `Прогон остановлен по лимиту тиков — нагрузка в сценарии бесконечная. ${gs(parked.length)} сейчас ${parked.length === 1 ? 'спит' : 'спят'} на каналах, но ${parked.length === 1 ? 'её' : 'их'} разбудили бы: это обратное давление, а не утечка.`
      }
      return 'Прогон остановлен по лимиту тиков — нагрузка в сценарии бесконечная.'
    }
  }
}

export default function ChanStand({ scenario = 'rendezvous', terms = {}, mode = 'embed' }: ChanStandProps) {
  const lab = mode === 'lab'
  const preset = chanScenarioById(scenario) ?? CHAN_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<ChanEvent | null>(null)
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
      if (e.type === 'g.ready' || e.type === 'g.start') continue
      e.actors.g?.forEach((x) => h.g.add(x))
      e.actors.chan?.forEach((x) => h.chans.add(x))
    }
    return h
  }, [hlEvent, snap])

  const summary = useMemo(() => {
    const s = snap.world.stats
    return {
      transfers: s.transfers,
      directShare: s.transfers === 0 ? 0 : Math.round((s.direct / s.transfers) * 100),
      avgWait: s.parks === 0 ? 0 : Math.round((s.waitTicks / s.parks) * 10) / 10,
      parked: snap.world.gs.filter((g) => g.state === 'waiting').length,
      idle:
        s.idleSlots + s.busySlots === 0 ? 0 : Math.round((s.idleSlots / (s.idleSlots + s.busySlots)) * 100),
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
  const labHref = `/lab/chan/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(chanScenarioById(st.base) ?? preset))

  return (
    <div className="stand" ref={rootRef} tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(chanScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {CHAN_SCENARIOS.map((s) => (
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
        <span title="Все значения, дошедшие до получателя или до буфера">
          передач <b>{summary.transfers}</b>
        </span>
        <span title="Доля передач мимо буфера — прямо из стека отправителя в стек получателя">
          из рук в руки <b>{summary.directShare}%</b>
        </span>
        <span title="Сколько раз горутина засыпала на канале">
          парковок <b>{snap.world.stats.parks}</b>
        </span>
        <span title="Сколько тиков в среднем длится одна парковка">
          ожидание <b>{summary.avgWait}</b> т.
        </span>
        <span className={summary.parked > 0 ? 'metric-bad' : undefined} title="Горутины, которые спят на каналах прямо сейчас">
          спят <b>{summary.parked}</b>
        </span>
        <span title="Доля процессорных слотов, простоявших без работы: все готовые горутины кончились">
          простой <b>{summary.idle}%</b>
        </span>
        {snap.world.stats.selectBlocks > 0 && (
          <span title="Сколько раз select уснул, встав в очередь сразу ко всем своим каналам">
            select уснул <b>{snap.world.stats.selectBlocks}</b>
          </span>
        )}
        {snap.world.stats.selectDefaults > 0 && (
          <span title="Сколько раз сработала ветка default — никто не был готов">
            default <b>{snap.world.stats.selectDefaults}</b>
          </span>
        )}
      </div>

      {pb.cursor === pb.last && pb.sim.world.finishReason && (
        <p className={`finish finish-${pb.sim.world.finishReason}`}>
          {finishLabel(pb.sim.world)}
          {pb.sim.world.panic && <code> panic: {pb.sim.world.panic}</code>}
        </p>
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
              Кликните событие в ленте — здесь появится разбор: что произошло, почему канал устроен именно так
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
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(chanScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            )}
          </div>
          <ConfigKnobs config={st.config} onChange={(config) => update({ config })} />
          <h4 className="settings-sub">Каналы</h4>
          <ChansEditor chans={sc.chans} onChange={(chans) => update({ chans })} />
          {lab && (
            <>
              <h4 className="settings-sub">Нагрузки</h4>
              <WorkloadsEditor workloads={sc.workloads} chans={sc.chans} onChange={(workloads) => update({ workloads })} />
            </>
          )}
        </div>
      </details>
    </div>
  )
}
