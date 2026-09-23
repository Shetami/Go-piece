import type { CacheConfig, Fault, Invalidation, Phase } from '../engine/types.ts'
import { idealHitRatio } from '../engine/world.ts'

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

export function ConfigKnobs({ config: c, onChange }: { config: CacheConfig; onChange: (c: CacheConfig) => void }) {
  const set = <K extends keyof CacheConfig>(k: K, v: CacheConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <>
      <h4 className="settings-sub">Кэш</h4>
      <div className="knobs">
        <Range
          label="Размер кэша"
          hint="Сколько ключей помещается. В скобках — потолок попаданий при такой популярности ключей"
          value={c.cacheSize}
          min={0}
          max={Math.max(c.keys, 200)}
          step={5}
          display={c.cacheSize === 0 ? 'без кэша' : `${c.cacheSize} (${Math.round(idealHitRatio(c) * 100)}%)`}
          onChange={(v) => set('cacheSize', Math.min(v, c.keys))}
        />
        <Range label="TTL" hint="Сколько живёт запись в кэше. Крайнее левое — вечно" value={c.ttl} min={0} max={120} step={5} display={c.ttl === 0 ? 'вечно' : t(c.ttl)} onChange={(v) => set('ttl', v)} />
        <Range
          label="Разброс TTL"
          hint="Случайная добавка к TTL каждой записи: не даёт ключам истекать одновременно"
          value={c.ttlJitter}
          min={0}
          max={100}
          step={10}
          display={`${c.ttlJitter}%`}
          onChange={(v) => set('ttlJitter', v)}
        />
        <Toggle
          label="Объединять запросы"
          hint="Пока один запрос за ключом идёт в базу, остальные промахи ждут его ответа (singleflight)"
          checked={c.coalesce}
          onChange={(v) => set('coalesce', v)}
        />
        <Toggle
          label="Отдавать просроченное"
          hint="Просроченная запись отдаётся сразу, а обновление идёт в фоне (stale-while-revalidate)"
          checked={c.staleWhileRevalidate}
          onChange={(v) => set('staleWhileRevalidate', v)}
        />
        <Choice<Invalidation>
          label="При записи"
          hint="Что делать с кэшем, когда значение меняется в базе"
          value={c.invalidation}
          options={[
            { value: 'none', label: 'ничего' },
            { value: 'delete', label: 'удалить ключ' },
            { value: 'update', label: 'положить новое' },
          ]}
          onChange={(v) => set('invalidation', v)}
          wide
        />
      </div>

      <h4 className="settings-sub">Нагрузка</h4>
      <div className="knobs">
        <Range label="Запросов за тик" hint="Сколько запросов приходит в среднем за тик" value={c.rate} min={0.5} max={10} step={0.5} display={`${c.rate}/т.`} onChange={(v) => set('rate', v)} />
        <Range label="Всего ключей" hint="Сколько разных ключей бывает в запросах" value={c.keys} min={20} max={1000} step={10} onChange={(v) => set('keys', v)} />
        <Range
          label="Неравномерность"
          hint="Показатель распределения Ципфа: 0 — все ключи одинаково популярны, 1 — обычная для жизни неравномерность"
          value={c.zipf}
          min={0}
          max={1.5}
          step={0.1}
          display={c.zipf === 0 ? 'все поровну' : c.zipf.toFixed(1)}
          onChange={(v) => set('zipf', v)}
        />
        <Range label="Доля горячего ключа" hint="Какая доля запросов идёт в самый горячий ключ сверх распределения" value={c.hotShare} min={0} max={60} step={5} display={`${c.hotShare}%`} onChange={(v) => set('hotShare', v)} />
        <Range label="Доля записей" hint="Какая доля запросов меняет значение в базе" value={c.writeShare} min={0} max={50} step={5} display={`${c.writeShare}%`} onChange={(v) => set('writeShare', v)} />
        <Range label="Зерно случайности" hint="Другое зерно — другой, но тоже воспроизводимый прогон" value={c.seed} min={1} max={40} onChange={(v) => set('seed', v)} />
      </div>

      <h4 className="settings-sub">База</h4>
      <div className="knobs">
        <Range
          label="Воркеров"
          hint="Сколько запросов база обрабатывает одновременно"
          value={c.dbWorkers}
          min={1}
          max={16}
          display={`${c.dbWorkers} · ${(c.dbWorkers / c.dbTime).toFixed(2)} зап./т.`}
          onChange={(v) => set('dbWorkers', v)}
        />
        <Range label="Время запроса" hint="Среднее время запроса в базу" value={c.dbTime} min={1} max={12} display={t(c.dbTime)} onChange={(v) => set('dbTime', v)} />
        <Range
          label="Запрос горячего ключа"
          hint="Тяжёлый агрегат за самым горячим ключом. Крайнее левое — такой же, как остальные"
          value={c.hotTime}
          min={0}
          max={30}
          display={c.hotTime === 0 ? 'как обычный' : t(c.hotTime)}
          onChange={(v) => set('hotTime', v)}
        />
      </div>
    </>
  )
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

export function FaultsEditor({ faults, onChange }: { faults: Fault[]; onChange: (f: Fault[]) => void }) {
  const set = (i: number, f: Fault) => onChange(faults.map((x, j) => (j === i ? f : x)))
  return (
    <div className="chan-editor">
      {faults.map((f, i) => (
        <div className="chan-edit" key={i}>
          <select
            value={f.kind}
            onChange={(e) => set(i, e.target.value === 'flush' ? { kind: 'flush', at: f.at } : { kind: 'warm', at: f.at, count: 40 })}
            aria-label="Что случается"
          >
            <option value="flush">кэш опустел</option>
            <option value="warm">прогрев кэша</option>
          </select>
          {f.kind === 'warm' && (
            <label>
              ключей
              <input type="number" min={1} max={1000} value={f.count} onChange={(e) => set(i, { ...f, count: Math.max(1, Number(e.target.value)) })} />
            </label>
          )}
          <label>
            на тике
            <input type="number" min={1} max={290} value={f.at} onChange={(e) => set(i, { ...f, at: Math.max(1, Number(e.target.value)) })} />
          </label>
          <button type="button" className="btn-ghost" onClick={() => onChange(faults.filter((_, j) => j !== i))} title="Убрать событие">
            ✕
          </button>
        </div>
      ))}
      <button type="button" className="btn-ghost" onClick={() => onChange([...faults, { kind: 'flush', at: 60 }])}>
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
