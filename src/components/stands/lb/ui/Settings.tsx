import type { Algo, Fault, LbConfig, Phase } from '../engine/types.ts'
import { replicaCapacity } from '../engine/world.ts'

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

export const ALGO_LABEL: Record<Algo, string> = {
  'round-robin': 'по кругу',
  random: 'случайно',
  'least-conn': 'меньше соединений',
  p2c: 'два случайных',
}

export function ConfigKnobs({ config: c, onChange }: { config: LbConfig; onChange: (c: LbConfig) => void }) {
  const set = <K extends keyof LbConfig>(k: K, v: LbConfig[K]) => onChange({ ...c, [k]: v })
  const cap = replicaCapacity(c) * c.replicas
  return (
    <>
      <h4 className="settings-sub">Балансировщик</h4>
      <div className="knobs">
        <Choice<Algo>
          label="Алгоритм"
          hint="По кругу и случайно не смотрят на реплики. Меньше соединений — на ту, от которой ждём меньше ответов. Два случайных — лучшая из двух случайных"
          value={c.algo}
          options={(['round-robin', 'random', 'least-conn', 'p2c'] as Algo[]).map((v) => ({ value: v, label: ALGO_LABEL[v] }))}
          onChange={(v) => set('algo', v)}
          wide
        />
        <Range
          label="Проверка здоровья"
          hint="Как часто балансировщик опрашивает /healthz каждой реплики. Крайнее левое — не проверяет"
          value={c.healthEvery}
          min={0}
          max={20}
          display={c.healthEvery === 0 ? 'выкл' : `каждые ${t(c.healthEvery)}`}
          onChange={(v) => set('healthEvery', v)}
        />
        <Range label="Исключать после провалов" hint="Сколько проверок подряд должна провалить реплика, чтобы её исключили" value={c.healthFails} min={1} max={5} onChange={(v) => set('healthFails', v)} />
        <Toggle
          label="Исключение по ошибкам"
          hint="Пассивная проверка: 3 ошибки подряд на настоящих запросах выводят реплику из ротации на 20 тиков"
          checked={c.outlier}
          onChange={(v) => set('outlier', v)}
        />
      </div>

      <h4 className="settings-sub">Нагрузка и реплики</h4>
      <div className="knobs">
        <Range
          label="Поток запросов"
          hint="Сколько запросов приходит в среднем за тик на весь сервис. В скобках — загрузка реплик, которые есть в начале"
          value={c.rate}
          min={0.2}
          max={3}
          step={0.1}
          display={`${c.rate.toFixed(1)}/т. (${Math.round((c.rate / cap) * 100)}%)`}
          onChange={(v) => set('rate', v)}
        />
        <Range label="Реплик в начале" hint="Сколько реплик работает с первого тика" value={c.replicas} min={1} max={c.maxReplicas} onChange={(v) => set('replicas', v)} />
        <Range label="Воркеров в реплике" hint="Сколько запросов реплика обрабатывает одновременно" value={c.workers} min={1} max={6} onChange={(v) => set('workers', v)} />
        <Range label="Время обработки" hint="Среднее время обработки запроса на здоровой реплике" value={c.serviceTime} min={1} max={10} display={t(c.serviceTime)} onChange={(v) => set('serviceTime', v)} />
        <Range label="Зерно случайности" hint="Другое зерно — другой, но тоже воспроизводимый прогон" value={c.seed} min={1} max={40} onChange={(v) => set('seed', v)} />
      </div>

      <h4 className="settings-sub">Клиенты</h4>
      <div className="knobs">
        <Range label="Таймаут" hint="Сколько клиент ждёт ответа. Крайнее левое — сколько угодно" value={c.timeout} min={0} max={40} display={c.timeout === 0 ? 'нет' : t(c.timeout)} onChange={(v) => set('timeout', v)} />
        <Range label="Повторов на другой реплике" hint="Сколько раз повторить запрос после ошибки или таймаута — в обход реплики, где он провалился" value={c.retries} min={0} max={2} onChange={(v) => set('retries', v)} />
      </div>

      <h4 className="settings-sub">Автомасштабирование</h4>
      <div className="knobs">
        <Toggle label="Включено" hint="Каждые 10 тиков добавлять реплики, если загрузка выше целевой" checked={c.autoscale} onChange={(v) => set('autoscale', v)} />
        <Range label="Целевая загрузка" hint="Какую загрузку реплик поддерживать. Меньше — больше запаса и реплик" value={c.scaleTarget} min={30} max={95} step={5} display={`${c.scaleTarget}%`} onChange={(v) => set('scaleTarget', v)} />
        <Range label="Запуск реплики" hint="Сколько тиков новая реплика запускается, прежде чем войти в ротацию" value={c.bootTime} min={1} max={60} display={t(c.bootTime)} onChange={(v) => set('bootTime', v)} />
        <Range label="Предел реплик" hint="Больше этого автомасштабирование не добавит" value={c.maxReplicas} min={c.replicas} max={8} onChange={(v) => set('maxReplicas', v)} />
      </div>
    </>
  )
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

const FAULT_LABEL: Record<Fault['kind'], string> = {
  slow: 'тормозит',
  errors: 'отвечает 500',
  crash: 'падает',
  hang: 'зависает',
}

export function FaultsEditor({ faults, replicas, onChange }: { faults: Fault[]; replicas: number; onChange: (f: Fault[]) => void }) {
  const set = (i: number, f: Fault) => onChange(faults.map((x, j) => (j === i ? f : x)))
  return (
    <div className="chan-editor">
      {faults.map((f, i) => (
        <div className="chan-edit" key={i}>
          <select value={f.replica} onChange={(e) => set(i, { ...f, replica: Number(e.target.value) })} aria-label="Реплика">
            {Array.from({ length: replicas }, (_, r) => (
              <option key={r} value={r}>
                r{r + 1}
              </option>
            ))}
          </select>
          <select
            value={f.kind}
            onChange={(e) => {
              const kind = e.target.value as Fault['kind']
              set(i, kind === 'slow' ? { kind, replica: f.replica, at: f.at, until: f.until, factor: 4 } : { kind, replica: f.replica, at: f.at, until: f.until })
            }}
            aria-label="Что случается"
          >
            {(Object.keys(FAULT_LABEL) as Fault['kind'][]).map((k) => (
              <option key={k} value={k}>
                {FAULT_LABEL[k]}
              </option>
            ))}
          </select>
          {f.kind === 'slow' && (
            <label>
              ×
              <input type="number" min={2} max={10} value={f.factor} onChange={(e) => set(i, { ...f, factor: Math.max(2, Number(e.target.value)) })} />
            </label>
          )}
          <label>
            с тика
            <input type="number" min={1} max={290} value={f.at} onChange={(e) => set(i, { ...f, at: Math.max(1, Number(e.target.value)) })} />
          </label>
          <label>
            до
            <input
              type="number"
              min={0}
              max={300}
              value={f.until ?? 0}
              title="0 — до конца прогона"
              onChange={(e) => {
                const v = Number(e.target.value)
                set(i, { ...f, until: v > f.at ? v : undefined })
              }}
            />
          </label>
          <button type="button" className="btn-ghost" onClick={() => onChange(faults.filter((_, j) => j !== i))} title="Убрать сбой">
            ✕
          </button>
        </div>
      ))}
      <button type="button" className="btn-ghost" onClick={() => onChange([...faults, { kind: 'slow', replica: 0, at: 20, factor: 4 }])}>
        + сбой
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
