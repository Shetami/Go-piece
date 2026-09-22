import { useEffect, useMemo, useState } from 'react'
import type { TxnEvent, TxnWorld } from '../engine/types.ts'
import { TXN_SCENARIOS, txnScenarioById } from '../engine/scenarios.ts'
import { heapStats, invariantState } from '../engine/tick.ts'
import { Board, highlightOf } from './Board.tsx'
import { ConfigKnobs, RowsEditor, SessionsEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface TxnStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор сессий и таблицы, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [1, 2, 4, 8]

function plural(n: number, one: string, few: string, many: string): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few
  return many
}

/** Чем кончился прогон — одной фразой. */
function finishLabel(world: TxnWorld, invariantOk: boolean | undefined): string {
  const s = world.stats
  const parts: string[] = []
  if (world.finishReason === 'stop-after') {
    const open = world.txns.filter((t) => t.state === 'active' || t.state === 'waiting')
    parts.push(
      open.length > 0
        ? `Прогон остановлен по лимиту тиков: ${open.length} ${plural(open.length, 'транзакция так и осталась открытой', 'транзакции так и остались открытыми', 'транзакций так и остались открытыми')}.`
        : 'Прогон остановлен по лимиту тиков.',
    )
  } else {
    parts.push('Все сессии закончили работу.')
  }
  parts.push(
    `Коммитов ${s.commits}${s.aborts > 0 ? `, прервано ${s.aborts}` : ''}${s.retries > 0 ? `, повторов ${s.retries}` : ''}${s.lostCommits > 0 ? `, потеряно подтверждённых ${s.lostCommits}` : ''}.`,
  )
  if (invariantOk === false) parts.push('Правило сценария нарушено — и ни одна транзакция не получила из-за этого ошибку.')
  else if (invariantOk === true && s.aborts > 0) parts.push('Правило соблюдено — ценой ошибок, которые база вернула клиенту.')
  else if (invariantOk === true) parts.push('Правило соблюдено.')
  return parts.join(' ')
}

export default function TxnStand({ scenario = 'lost-update', terms = {}, mode = 'embed' }: TxnStandProps) {
  const lab = mode === 'lab'
  const preset = txnScenarioById(scenario) ?? TXN_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<TxnEvent | null>(null)
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
      if (e.type === 'snap.take' || e.type === 'txn.xid' || e.type === 'wal.flush') continue
      e.actors.txn?.forEach((x) => h.txn.add(x))
      e.actors.tuple?.forEach((x) => h.tuple.add(x))
      e.actors.key?.forEach((x) => h.key.add(x))
    }
    return h
  }, [hlEvent, frame])

  const w = frame.world
  const inv = invariantState(w, sc.invariant)
  const hs = heapStats(w)
  const finalInv = invariantState(pb.sim.world, sc.invariant)

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
  const labHref = `/lab/txn/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(txnScenarioById(st.base) ?? preset))
  const keys = [...new Set(sc.rows.map((r) => r.key))]

  return (
    <div className="stand" tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(txnScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {TXN_SCENARIOS.map((s) => (
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
        <span title="Транзакции, получившие «COMMIT»">
          коммитов <b>{w.stats.commits}</b>
        </span>
        <span className={w.stats.aborts > 0 ? 'metric-bad' : undefined} title="Ошибки сериализации, дедлоки, нарушения ключа, обрывы при падении">
          прервано <b>{w.stats.aborts}</b>
        </span>
        <span title="Сколько раз и сколько тиков в сумме транзакции ждали блокировку строки">
          ожиданий <b>{w.stats.waits}</b> · <b>{w.stats.waitTicks}</b> т.
        </span>
        {w.config.syncCommit && w.stats.commitWaitTicks > 0 && (
          <span title="Сколько тиков в сумме коммиты ждали сброса WAL">
            ждали fsync <b>{w.stats.commitWaitTicks}</b> т.
          </span>
        )}
        <span className={w.stats.anomalies > 0 ? 'metric-bad' : undefined} title="Аномалии изоляции, которые заметил стенд">
          аномалий <b>{w.stats.anomalies}</b>
        </span>
        <span title="Мёртвые версии строк, которые ещё лежат в таблице">
          мёртвых версий <b>{hs.dead}</b> из {hs.total}
        </span>
        {w.stats.lostCommits > 0 && (
          <span className="metric-bad" title="Коммиты, подтверждённые клиенту, но не пережившие падения">
            потеряно <b>{w.stats.lostCommits}</b>
          </span>
        )}
        {inv && (
          <span className={inv.ok ? undefined : 'metric-bad'} title={sc.invariant!.label}>
            правило <b>{inv.ok ? 'соблюдено' : 'нарушено'}</b>
          </span>
        )}
      </div>

      {pb.cursor === pb.last && pb.sim.world.finishReason && (
        <p className={`finish ${finalInv?.ok === false || pb.sim.world.stats.lostCommits > 0 ? 'finish-panic' : ''}`}>
          {finishLabel(pb.sim.world, finalInv?.ok)}
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
              Кликните событие в ленте — здесь появится разбор: что сделала база, почему именно так и какие функции
              PostgreSQL за этим стоят.
            </p>
          )}
        </div>
      </div>

      <details className="settings" open={settingsOpen} onToggle={(e) => setSettingsOpen(e.currentTarget.open)}>
        <summary>Настройки {modified && <span className="settings-mod">изменены</span>}</summary>
        <div className="settings-body">
          {modified && (
            <div className="settings-row">
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(txnScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            </div>
          )}
          <ConfigKnobs config={st.config} scenario={sc} onChange={(config) => update({ config })} />
          {lab && (
            <>
              <h4 className="settings-sub">Таблица t</h4>
              <RowsEditor rows={sc.rows} onChange={(rows) => update({ rows })} />
              <h4 className="settings-sub">Сессии</h4>
              <SessionsEditor sessions={sc.sessions} keys={keys} onChange={(sessions) => update({ sessions })} />
            </>
          )}
        </div>
      </details>
    </div>
  )
}
