import { useEffect, useMemo, useRef, useState } from 'react'
import type { KafkaEvent, KafkaWorld } from '../engine/types.ts'
import { KAFKA_SCENARIOS, kafkaScenarioById } from '../engine/scenarios.ts'
import { msgs } from '../explain/events.ts'
import { Board, highlightOf } from './Board.tsx'
import { Journey } from './Journey.tsx'
import { ConfigKnobs, ConsumersEditor, FaultsEditor, ProducersEditor } from './Settings.tsx'
import { EventDetail, EventFeed, Scrubber, type TermInfo } from './Timeline.tsx'
import { decodeState, encodeState, initialState, toScenario, type StandState } from './url.ts'
import { useSimulation } from './useSimulation.ts'

export interface KafkaStandProps {
  /** id пресета из engine/scenarios.ts. */
  scenario?: string
  /** Краткие определения терминов — приходят со страницы, из коллекции glossary. */
  terms?: Record<string, TermInfo>
  /**
   * embed — стенд внутри лекции: сценарий фиксирован, настройки свёрнуты.
   * lab — отдельная страница: выбор сценария, редактор нагрузки и сбоев, состояние в адресе.
   */
  mode?: 'embed' | 'lab'
}

const SPEEDS = [1, 2, 4, 8, 16]

/** Чем кончился прогон — одной фразой, с тем, что пошло не так. */
function finishLabel(w: KafkaWorld): string {
  const s = w.stats
  const bad: string[] = []
  if (s.lost > 0) bad.push(`${msgs(s.lost)} потеряно после подтверждения`)
  if (s.failed > 0) bad.push(`${msgs(s.failed)} не доставлено`)
  if (s.duplicates > 0) bad.push(`дублей в логе: ${s.duplicates}`)
  if (s.reprocessed > 0) bad.push(`обработано повторно: ${s.reprocessed}`)
  if (s.skipped > 0) bad.push(`пропущено группой: ${s.skipped}`)
  const head =
    w.finishReason === 'all-done'
      ? 'Прогон завершён: всё записанное прочитано и закоммичено.'
      : 'Прогон остановлен по лимиту тиков.'
  return bad.length === 0 ? `${head} Потерь и дублей нет.` : `${head} Но: ${bad.join(', ')}.`
}

export default function KafkaStand({ scenario = 'journey', terms = {}, mode = 'embed' }: KafkaStandProps) {
  const lab = mode === 'lab'
  const preset = kafkaScenarioById(scenario) ?? KAFKA_SCENARIOS[0]!

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
  const [selected, setSelected] = useState<KafkaEvent | null>(null)
  const [settingsOpen, setSettingsOpen] = useState(lab)
  const [focus, setFocus] = useState<number | null>(sc.focus ?? 1)

  const pb = useSimulation(sc, st.seed, { autoPause })
  const snap = pb.sim.history[pb.cursor]!

  useEffect(() => setSelected(null), [pb.sim])
  useEffect(() => setFocus(sc.focus ?? 1), [sc.id, sc.focus])
  useEffect(() => {
    if (pb.pausedOn) setSelected(pb.pausedOn)
  }, [pb.pausedOn])

  const hlEvent = selected && selected.tick === pb.cursor ? selected : null
  const highlight = useMemo(() => {
    if (hlEvent) return highlightOf(hlEvent)
    const h = highlightOf(null)
    for (const e of snap.events) {
      if (e.type === 'repl.fetch' || e.type === 'consumer.process') continue
      e.actors.broker?.forEach((x) => h.brokers.add(x))
      e.actors.partition?.forEach((x) => h.partitions.add(x))
      e.actors.producer?.forEach((x) => h.producers.add(x))
      e.actors.consumer?.forEach((x) => h.consumers.add(x))
    }
    return h
  }, [hlEvent, snap])

  const summary = useMemo(() => {
    const w = snap.world
    const s = w.stats
    const lag = w.groups.reduce(
      (n, g) => n + w.partitions.reduce((m, p) => m + Math.max(0, p.hw - (g.committed[p.id] ?? 0)), 0),
      0,
    )
    return {
      avgBatch: s.requests === 0 ? 0 : Math.round((s.recsSent / s.requests) * 10) / 10,
      avgAck: s.ackCount === 0 ? 0 : Math.round((s.ackLatency / s.ackCount) * 10) / 10,
      avgE2e: s.e2eCount === 0 ? 0 : Math.round((s.e2eLatency / s.e2eCount) * 10) / 10,
      lag,
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
  const labHref = `/lab/kafka/#${encodeState(st)}`
  const modified = JSON.stringify(st) !== JSON.stringify(initialState(kafkaScenarioById(st.base) ?? preset))
  const s = snap.world.stats

  return (
    <div className="stand" ref={rootRef} tabIndex={-1} onKeyDown={onKey}>
      <header className="stand-head">
        <div>
          {lab ? (
            <select
              className="stand-select"
              value={st.base}
              onChange={(e) => setSt(initialState(kafkaScenarioById(e.target.value) ?? preset))}
              aria-label="Сценарий"
            >
              {KAFKA_SCENARIOS.map((x) => (
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

      <Scrubber
        history={pb.sim.history}
        cursor={pb.cursor}
        last={pb.last}
        watchFor={sc.watchFor}
        onSeek={(t) => { pb.setPlaying(false); pb.setCursor(t) }}
      />

      <Board world={snap.world} events={snap.events} scenario={sc} highlight={highlight} focus={focus} onFocus={setFocus} />

      <Journey world={snap.world} scenario={sc} focus={focus} onFocus={setFocus} />

      <div className="metrics">
        <span title="Вызовов send()">отправлено <b>{s.produced}</b></span>
        <span title="Продюсер получил подтверждение">подтверждено <b>{s.acked}</b></span>
        <span title="High watermark перешёл сообщение: оно на всех репликах ISR">закоммичено <b>{s.committed}</b></span>
        <span title="Запросов записи и сколько сообщений в среднем в одном">
          запросов <b>{s.requests}</b> · по <b>{summary.avgBatch}</b>
        </span>
        <span title="Средняя задержка от send() до подтверждения">ack <b>{summary.avgAck}</b> т.</span>
        <span title="Средняя задержка от send() до первой обработки потребителем">send → обработка <b>{summary.avgE2e}</b> т.</span>
        <span className={summary.lag > 8 ? 'metric-bad' : undefined} title="Лаг: сумма по группам и партициям HW минус закоммиченный оффсет">
          лаг <b>{summary.lag}</b>
        </span>
        {s.lost > 0 && (
          <span className="metric-bad" title="Подтверждённые сообщения, которых больше нет в логе">
            потеряно <b>{s.lost}</b>
          </span>
        )}
        {s.failed > 0 && (
          <span className="metric-bad" title="Продюсер отдал приложению ошибку: истёк delivery.timeout.ms">
            не доставлено <b>{s.failed}</b>
          </span>
        )}
        {s.duplicates > 0 && (
          <span className="metric-bad" title="Лишние копии в логе от повторов без идемпотентности">
            дублей <b>{s.duplicates}</b>
          </span>
        )}
        {s.dedups > 0 && (
          <span title="Повторы, которые брокер узнал и не записал">
            повторов отброшено <b>{s.dedups}</b>
          </span>
        )}
        {s.reprocessed > 0 && (
          <span className="metric-bad" title="Сообщения, обработанные группой больше одного раза">
            повторно <b>{s.reprocessed}</b>
          </span>
        )}
        {s.skipped > 0 && (
          <span className="metric-bad" title="Сообщения, которые группа не обработает никогда">
            пропущено <b>{s.skipped}</b>
          </span>
        )}
      </div>

      {pb.cursor === pb.last && pb.sim.world.finishReason && (
        <p className={`finish finish-${pb.sim.world.stats.lost + pb.sim.world.stats.failed + pb.sim.world.stats.skipped > 0 ? 'panic' : pb.sim.world.finishReason}`}>
          {finishLabel(pb.sim.world)}
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
              const r = e.actors.rec
              if (r && r.length > 0 && (focus === null || !r.includes(focus))) setFocus(r[0]!)
            }}
          />
        </div>
        <div className="stand-detail">
          {selected ? (
            <EventDetail event={selected} terms={terms} onClose={() => setSelected(null)} />
          ) : (
            <p className="detail-hint">
              Кликните событие в ленте — здесь появится разбор: что произошло в кластере, почему Kafka устроена
              именно так и какие классы за этим стоят.
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
              <button type="button" className="btn-ghost" onClick={() => setSt(initialState(kafkaScenarioById(st.base) ?? preset))}>
                сбросить к пресету
              </button>
            )}
          </div>
          <ConfigKnobs config={st.config} lab={lab} onChange={(config) => update({ config })} />
          {lab && (
            <>
              <h4 className="settings-sub">Продюсеры</h4>
              <ProducersEditor producers={sc.producers} onChange={(producers) => update({ producers })} />
              <h4 className="settings-sub">Потребители</h4>
              <ConsumersEditor consumers={sc.consumers} onChange={(consumers) => update({ consumers })} />
              <h4 className="settings-sub">Сбои</h4>
              <FaultsEditor
                faults={sc.faults}
                brokers={sc.config.brokers ?? 3}
                consumers={sc.consumers}
                producers={sc.producers}
                onChange={(faults) => update({ faults })}
              />
            </>
          )}
        </div>
      </details>
    </div>
  )
}
