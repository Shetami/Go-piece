import type { KafkaScenario, KafkaWorld, Rec, StepKind } from '../engine/types.ts'
import { BAD_STEPS, STEP_INFO, canonicalPath, describeStep } from '../explain/journey.ts'
import { recColor } from './Board.tsx'

/**
 * Путь выбранного сообщения на текущем тике.
 *
 * Пройденные шаги — из trace сообщения в снимке, поэтому перемотка назад
 * честно «отматывает» путь. Непройденные — из обычного маршрута для текущего
 * acks: видно не только где сообщение сейчас, но и что его ждёт дальше.
 */

const STATE: Record<Rec['state'], { label: string; cls: string }> = {
  batched: { label: 'у продюсера', cls: '' },
  'in-flight': { label: 'в пути', cls: '' },
  acked: { label: 'подтверждено', cls: 'is-ok' },
  failed: { label: 'не доставлено', cls: 'is-bad' },
  lost: { label: 'потеряно', cls: 'is-bad' },
}

export function Journey({ world: w, scenario, focus, onFocus }: {
  world: KafkaWorld
  scenario: KafkaScenario
  focus: number | null
  onFocus: (id: number) => void
}) {
  const rec = focus !== null ? w.recs[focus - 1] : undefined

  const picker = (
    <select
      className="kf-journey-pick"
      value={rec ? String(rec.id) : ''}
      onChange={(e) => onFocus(Number(e.target.value))}
      aria-label="Сообщение"
    >
      {!rec && <option value="">—</option>}
      {w.recs.map((r) => (
        <option key={r.id} value={r.id}>
          m{r.id}
          {r.key === null ? '' : ` · ${r.key}`} · p{r.partition}
        </option>
      ))}
    </select>
  )

  if (!rec) {
    return (
      <section className="kf-journey">
        <header>
          <h4>Путь сообщения</h4>
          {w.recs.length > 0 && picker}
        </header>
        <p className="kf-muted">
          {focus !== null
            ? `Сообщение m${focus} ещё не создано: продюсер вызовет send() позже. Нажмите ▶ или шагните вперёд.`
            : 'Кликните любое сообщение на схеме — здесь появится его путь от send() до коммита оффсета.'}
        </p>
      </section>
    )
  }

  const done = new Set<StepKind>(rec.trace.map((t) => t.step))
  const final = rec.state === 'lost' || rec.state === 'failed'
  const groups = w.groups.map((g) => g.name)
  // Что ещё впереди. Для каждой группы обработка своя, но в маршруте показываем общий шаг.
  // С фактором репликации 1 копировать некуда — этой остановки на пути просто нет.
  const replicas = w.partitions[rec.partition]?.replicas.length ?? 1
  const ahead = final ? [] : canonicalPath(w.config.acks).filter((s) => !done.has(s) && !(s === 'replicate' && replicas <= 1))

  return (
    <section className="kf-journey">
      <header>
        <h4>Путь сообщения</h4>
        {picker}
        <span className="kf-journey-meta">
          <span className="kf-journey-dot" style={{ background: recColor(rec, scenario) }} />
          {rec.key === null ? 'без ключа' : `ключ «${rec.key}»`} · p{rec.partition}
          {rec.offset !== null && ` · оффсет ${rec.offset}`}
        </span>
        <span className={`kf-journey-state ${STATE[rec.state].cls}`}>{STATE[rec.state].label}</span>
        {groups.map((g) => {
          const n = rec.processed[g] ?? 0
          const skipped = rec.skipped.includes(g)
          return (
            <span key={g} className={`kf-journey-state ${n > 1 || skipped ? 'is-bad' : n === 1 ? 'is-ok' : ''}`}>
              {g}: {skipped ? 'пропущено' : n === 0 ? 'не обработано' : n === 1 ? 'обработано' : `обработано ×${n}`}
            </span>
          )
        })}
        <span className="kf-journey-nav">
          <button type="button" className="btn-ghost" disabled={rec.id <= 1} onClick={() => onFocus(rec.id - 1)} title="Предыдущее сообщение">
            ←
          </button>
          <button type="button" className="btn-ghost" disabled={rec.id >= w.recs.length} onClick={() => onFocus(rec.id + 1)} title="Следующее сообщение">
            →
          </button>
        </span>
      </header>

      <ol className="kf-steps">
        {rec.trace.map((t, i) => {
          const d = describeStep(t, rec)
          const info = STEP_INFO[t.step]
          const now = t.tick === w.tick
          return (
            <li key={i} className={`kf-step is-done ${BAD_STEPS.has(t.step) ? 'is-bad' : ''} ${now ? 'is-now' : ''}`}>
              <span className="kf-step-tick">{t.tick}</span>
              <span className="kf-step-where">{info.where}</span>
              <div>
                <b>{d.title}</b>
                <p>{d.body}</p>
              </div>
            </li>
          )
        })}
        {ahead.map((s) => {
          const info = STEP_INFO[s]
          return (
            <li key={`ahead-${s}`} className="kf-step is-ahead">
              <span className="kf-step-tick">…</span>
              <span className="kf-step-where">{info.where}</span>
              <div>
                <b>{info.label}</b>
                <p>{info.ahead}</p>
              </div>
            </li>
          )
        })}
      </ol>
    </section>
  )
}
