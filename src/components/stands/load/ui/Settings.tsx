import type { Arrivals, LoadConfig, Phase, ServiceDist } from '../engine/types.ts'
import { capacityOf, offeredLoad } from '../engine/world.ts'

/**
 * Ручки стенда. Всё, что здесь меняется, превращается в новый сценарий —
 * движок о существовании UI не знает.
 */

function Toggle({ label, hint, checked, onChange }: { label: string; hint: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="knob knob-toggle" title={hint}>
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      <span>{label}</span>
    </label>
  )
}

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

function Choice<T extends string>({ label, hint, value, options, onChange }: {
  label: string
  hint: string
  value: T
  options: { value: T; label: string }[]
  onChange: (v: T) => void
}) {
  return (
    <div className="knob" title={hint}>
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
const pct = (v: number) => `${Math.round(v * 100)}%`

export function ConfigKnobs({ config: c, onChange }: { config: LoadConfig; onChange: (c: LoadConfig) => void }) {
  const set = <K extends keyof LoadConfig>(k: K, v: LoadConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <>
      <h4 className="settings-sub">Нагрузка</h4>
      <div className="knobs">
        <Range
          label="Поток запросов"
          hint="Сколько запросов приходит в среднем за тик. В скобках — загрузка: доля ёмкости сервиса, которую этот поток займёт"
          value={c.rate}
          min={0.1}
          max={2}
          step={0.05}
          display={`${c.rate.toFixed(2)}/т. (${pct(offeredLoad(c))})`}
          onChange={(v) => set('rate', v)}
        />
        <Choice<Arrivals>
          label="Приход"
          hint="Ровно — через равные промежутки. Случайно — пуассоновский поток: иногда пачкой, иногда тишина"
          value={c.arrivals}
          options={[
            { value: 'random', label: 'случайно' },
            { value: 'even', label: 'ровно' },
          ]}
          onChange={(v) => set('arrivals', v)}
        />
        <Range label="Зерно случайности" hint="Другое зерно — другой, но тоже воспроизводимый прогон" value={c.seed} min={1} max={40} onChange={(v) => set('seed', v)} />
      </div>

      <h4 className="settings-sub">Сервис</h4>
      <div className="knobs">
        <Range
          label="Воркеров"
          hint="Сколько запросов сервис обрабатывает одновременно: горутины, потоки, соединения с базой"
          value={c.workers}
          min={1}
          max={8}
          display={`${c.workers} · ёмкость ${capacityOf(c).toFixed(2)}/т.`}
          onChange={(v) => set('workers', v)}
        />
        <Range label="Время обработки" hint="Среднее время работы над обычным запросом" value={c.serviceTime} min={1} max={10} display={t(c.serviceTime)} onChange={(v) => set('serviceTime', v)} />
        <Choice<ServiceDist>
          label="Разброс времени"
          hint="Одинаково — каждый запрос ровно столько. Случайно — большинство короче среднего, редкие намного длиннее"
          value={c.serviceDist}
          options={[
            { value: 'random', label: 'случайно' },
            { value: 'fixed', label: 'одинаково' },
          ]}
          onChange={(v) => set('serviceDist', v)}
        />
        <Range label="Медленных запросов" hint="Какая доля запросов в разы медленнее обычных" value={c.slowShare} min={0} max={10} display={`${c.slowShare}%`} onChange={(v) => set('slowShare', v)} />
        <Range label="Медленный дольше в" hint="Во сколько раз медленный запрос дольше обычного" value={c.slowFactor} min={2} max={20} display={`${c.slowFactor} раз`} onChange={(v) => set('slowFactor', v)} />
        <Range
          label="Предел очереди"
          hint="Сколько запросов может ждать. Лишние сразу получают отказ. Крайнее левое — без предела"
          value={c.queueLimit}
          min={0}
          max={40}
          display={c.queueLimit === 0 ? 'без предела' : String(c.queueLimit)}
          onChange={(v) => set('queueLimit', v)}
        />
        <Toggle
          label="Отмена по context"
          hint="Сервис узнаёт, что клиент ушёл по таймауту, и бросает его запрос — в очереди и в работе"
          checked={c.cancelOnTimeout}
          onChange={(v) => set('cancelOnTimeout', v)}
        />
      </div>

      <h4 className="settings-sub">Клиенты</h4>
      <div className="knobs">
        <Range label="Таймаут" hint="Сколько клиент ждёт ответа. Крайнее левое — сколько угодно" value={c.timeout} min={0} max={40} display={c.timeout === 0 ? 'нет' : t(c.timeout)} onChange={(v) => set('timeout', v)} />
        <Range label="Повторов" hint="Сколько раз клиент повторяет запрос после таймаута или отказа" value={c.retries} min={0} max={3} onChange={(v) => set('retries', v)} />
        <Toggle
          label="Пауза перед повтором"
          hint="Экспоненциальная пауза со случайным разбросом (full jitter). Без неё повтор уходит сразу"
          checked={c.backoff}
          onChange={(v) => set('backoff', v)}
        />
      </div>
    </>
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
