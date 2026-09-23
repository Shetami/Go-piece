import type { Fault, Phase, Scheme, ShardConfig } from '../engine/types.ts'
import { movedShare } from '../engine/world.ts'

/**
 * Ручки стенда. Всё, что здесь меняется, превращается в новый сценарий —
 * движок о существовании UI не знает.
 */

function Range({ label, hint, value, min, max, step = 1, display, onChange }: {
  label: string
  hint: string
  value: number
  min: number
  max: number
  step?: number
  display?: string
  onChange: (v: number) => void
}) {
  return (
    <label className="knob" title={hint}>
      <span className="knob-head">
        <span>{label}</span>
        <b>{display ?? value}</b>
      </span>
      <input type="range" min={min} max={max} step={step} value={value} onChange={(e) => onChange(Number(e.target.value))} />
    </label>
  )
}

function Choice<T extends string>({ label, hint, value, options, onChange, wide = false }: {
  label: string
  hint: string
  value: T
  options: { value: T; label: string }[]
  onChange: (v: T) => void
  /** Много вариантов — занять всю строку, чтобы переключатель не переносился. */
  wide?: boolean
}) {
  return (
    <div className="knob" title={hint} style={wide ? { gridColumn: '1 / -1' } : undefined}>
      <span className="knob-head">
        <span>{label}</span>
      </span>
      <span className="kf-choice db-choice" role="radiogroup" aria-label={label}>
        {options.map((o) => (
          <button key={o.value} type="button" role="radio" aria-checked={o.value === value} className={o.value === value ? 'is-now' : ''} onClick={() => onChange(o.value)}>
            {o.label}
          </button>
        ))}
      </span>
    </div>
  )
}

const t = (n: number) => `${n} т.`

export const SCHEME_LABEL: Record<Scheme, string> = {
  mod: 'остаток от деления',
  ring: 'кольцо',
  range: 'диапазоны',
}

export function ConfigKnobs({ config: c, onChange }: { config: ShardConfig; onChange: (c: ShardConfig) => void }) {
  const set = <K extends keyof ShardConfig>(k: K, v: ShardConfig[K]) => onChange({ ...c, [k]: v })
  const cap = (c.shards * c.workers) / c.serviceTime
  return (
    <>
      <h4 className="settings-sub">Раскладка</h4>
      <div className="knobs">
        <Choice<Scheme>
          label="Схема"
          hint="Остаток от деления хеша; кольцо согласованного хеширования; диапазоны подряд идущих ключей"
          value={c.scheme}
          options={(['mod', 'ring', 'range'] as Scheme[]).map((v) => ({ value: v, label: SCHEME_LABEL[v] }))}
          onChange={(v) => set('scheme', v)}
          wide
        />
        <Range
          label="Шардов"
          hint="Сколько шардов работает в начале прогона"
          value={c.shards}
          min={1}
          max={8}
          display={`${c.shards} · ёмкость ${cap.toFixed(2)} зап./т.`}
          onChange={(v) => set('shards', v)}
        />
        <Range
          label="Виртуальных узлов"
          hint="Только для кольца: сколько точек на окружности у каждого шарда. Больше — ровнее раскладка"
          value={c.vnodes}
          min={1}
          max={128}
          display={c.scheme === 'ring' ? String(c.vnodes) : 'только для кольца'}
          onChange={(v) => set('vnodes', v)}
        />
        <Range
          label="Переезд ключей"
          hint="Сколько ключей переезжает за тик при решардировании. Быстрее — сильнее мешает боевым запросам"
          value={c.migrateRate}
          min={1}
          max={20}
          display={`${c.migrateRate} кл./т.`}
          onChange={(v) => set('migrateRate', v)}
        />
      </div>

      <h4 className="settings-sub">Нагрузка</h4>
      <div className="knobs">
        <Range label="Запросов за тик" hint="Сколько запросов приходит в среднем за тик" value={c.rate} min={0.2} max={4} step={0.2} display={`${c.rate.toFixed(1)}/т.`} onChange={(v) => set('rate', v)} />
        <Range label="Всего ключей" hint="Сколько разных ключей бывает в запросах" value={c.keys} min={20} max={480} step={20} onChange={(v) => set('keys', v)} />
        <Range
          label="Неравномерность"
          hint="Показатель распределения Ципфа: 0 — все ключи одинаково популярны"
          value={c.zipf}
          min={0}
          max={1.5}
          step={0.1}
          display={c.zipf === 0 ? 'все поровну' : c.zipf.toFixed(1)}
          onChange={(v) => set('zipf', v)}
        />
        <Range label="Доля горячего ключа" hint="Какая доля запросов идёт в один самый горячий ключ" value={c.hotShare} min={0} max={60} step={5} display={`${c.hotShare}%`} onChange={(v) => set('hotShare', v)} />
        <Range
          label="Без ключа шардирования"
          hint="Доля запросов, которые приходится задавать всем шардам сразу"
          value={c.scatterShare}
          min={0}
          max={50}
          step={5}
          display={`${c.scatterShare}%`}
          onChange={(v) => set('scatterShare', v)}
        />
        <Range label="Зерно случайности" hint="Другое зерно — другой, но тоже воспроизводимый прогон" value={c.seed} min={1} max={40} onChange={(v) => set('seed', v)} />
      </div>

      <h4 className="settings-sub">Шард</h4>
      <div className="knobs">
        <Range label="Воркеров в шарде" hint="Сколько запросов шард обрабатывает одновременно" value={c.workers} min={1} max={6} onChange={(v) => set('workers', v)} />
        <Range label="Время запроса" hint="Среднее время обработки запроса шардом" value={c.serviceTime} min={1} max={10} display={t(c.serviceTime)} onChange={(v) => set('serviceTime', v)} />
      </div>
      <p className="kf-muted" style={{ fontSize: '0.75rem' }}>
        При добавлении одного шарда переедет: остаток от деления —{' '}
        {Math.round(movedShare({ ...c, scheme: 'mod' }, c.shards, c.shards + 1) * 100)}%, кольцо —{' '}
        {Math.round(movedShare({ ...c, scheme: 'ring' }, c.shards, c.shards + 1) * 100)}%, диапазоны —{' '}
        {Math.round(movedShare({ ...c, scheme: 'range' }, c.shards, c.shards + 1) * 100)}% ключей.
      </p>
    </>
  )
}

/* ─────────────────────────────── события ─────────────────────────────── */

export function FaultsEditor({ faults, onChange }: { faults: Fault[]; onChange: (f: Fault[]) => void }) {
  const set = (i: number, f: Fault) => onChange(faults.map((x, j) => (j === i ? f : x)))
  return (
    <div className="chan-editor">
      {faults.map((f, i) => (
        <div className="chan-edit" key={i}>
          <select value={f.kind} onChange={(e) => set(i, { kind: e.target.value as Fault['kind'], at: f.at })} aria-label="Что случается">
            <option value="add">добавить шард</option>
            <option value="remove">убрать шард</option>
          </select>
          <label>
            на тике
            <input type="number" min={1} max={290} value={f.at} onChange={(e) => set(i, { ...f, at: Math.max(1, Number(e.target.value)) })} />
          </label>
          <button type="button" className="btn-ghost" onClick={() => onChange(faults.filter((_, j) => j !== i))} title="Убрать событие">
            ✕
          </button>
        </div>
      ))}
      <button type="button" className="btn-ghost" onClick={() => onChange([...faults, { kind: 'add', at: 60 }])}>
        + событие
      </button>
    </div>
  )
}

/* ─────────────────────────────── фазы ─────────────────────────────── */

export function PhasesEditor({ phases, onChange }: { phases: Phase[]; onChange: (p: Phase[]) => void }) {
  const set = (i: number, p: Phase) => onChange(phases.map((x, j) => (j === i ? p : x)))
  return (
    <div className="chan-editor">
      {phases.map((p, i) => (
        <div className="chan-edit" key={i}>
          <label>
            с тика
            <input type="number" min={1} max={290} value={p.at} onChange={(e) => set(i, { ...p, at: Math.max(1, Number(e.target.value)) })} />
          </label>
          <label>
            поток ×
            <input type="number" min={0} max={4} step={0.1} value={p.x} onChange={(e) => set(i, { ...p, x: Math.max(0, Number(e.target.value)) })} />
          </label>
          <button type="button" className="btn-ghost" onClick={() => onChange(phases.filter((_, j) => j !== i))} title="Убрать фазу">
            ✕
          </button>
        </div>
      ))}
      <button
        type="button"
        className="btn-ghost"
        onClick={() => {
          const last = phases[phases.length - 1]
          onChange([...phases, last ? { at: last.at + 20, x: last.x === 1 ? 2 : 1 } : { at: 40, x: 2 }])
        }}
      >
        + фаза
      </button>
    </div>
  )
}
