import type { Fault, Phase, ResilConfig } from '../engine/types.ts'
import { capacityOf, depCapacityOf } from '../engine/world.ts'

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

const t = (n: number) => `${n} т.`

export function ConfigKnobs({ config: c, onChange }: { config: ResilConfig; onChange: (c: ResilConfig) => void }) {
  const set = <K extends keyof ResilConfig>(k: K, v: ResilConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <>
      <h4 className="settings-sub">Защита вызовов</h4>
      <div className="knobs">
        <Range
          label="Таймаут вызова B"
          hint="Сколько воркер A готов ждать ответа. Крайнее левое — ждать сколько угодно"
          value={c.timeout}
          min={0}
          max={40}
          display={c.timeout === 0 ? 'нет' : t(c.timeout)}
          onChange={(v) => set('timeout', v)}
        />
        <Range label="Повторов вызова" hint="Сколько раз повторить вызов B после таймаута или ошибки" value={c.retries} min={0} max={3} onChange={(v) => set('retries', v)} />
        <Toggle label="Предохранитель" hint="После нескольких неудач подряд перестать ходить в B и раз в N тиков пробовать снова" checked={c.breaker} onChange={(v) => set('breaker', v)} />
        <Range label="Неудач до размыкания" hint="Сколько неудачных вызовов подряд размыкают предохранитель" value={c.breakerFails} min={2} max={15} onChange={(v) => set('breakerFails', v)} />
        <Range label="Держать разомкнутым" hint="Через сколько тиков предохранитель пропустит пробный вызов" value={c.breakerOpen} min={5} max={60} step={5} display={t(c.breakerOpen)} onChange={(v) => set('breakerOpen', v)} />
        <Range
          label="Переборка"
          hint="Сколько воркеров A могут одновременно ждать B. Крайнее левое — все"
          value={c.bulkhead}
          min={0}
          max={c.workers}
          display={c.bulkhead === 0 ? 'без предела' : `${c.bulkhead} из ${c.workers}`}
          onChange={(v) => set('bulkhead', v)}
        />
        <Toggle label="Ответ-заглушка" hint="Когда B недоступен, отвечать урезанно вместо ошибки" checked={c.fallback} onChange={(v) => set('fallback', v)} />
      </div>

      <h4 className="settings-sub">Вход</h4>
      <div className="knobs">
        <Range label="Запросов за тик" hint="Сколько запросов приходит в среднем за тик" value={c.rate} min={0.2} max={4} step={0.2} display={`${c.rate.toFixed(1)}/т.`} onChange={(v) => set('rate', v)} />
        <Range label="Доля запросов к B" hint="Какой части запросов нужна зависимость" value={c.depShare} min={0} max={100} step={10} display={`${c.depShare}%`} onChange={(v) => set('depShare', v)} />
        <Range
          label="Ограничитель частоты"
          hint="Сколько запросов в тик пропускать на вход. Крайнее левое — выключен"
          value={c.rateLimit}
          min={0}
          max={4}
          step={0.2}
          display={c.rateLimit === 0 ? 'выключен' : `${c.rateLimit.toFixed(1)}/т.`}
          onChange={(v) => set('rateLimit', v)}
        />
        <Range label="Запас ограничителя" hint="Сколько запросов можно принять сверх нормы разом" value={c.burst} min={1} max={30} onChange={(v) => set('burst', v)} />
        <Range label="Зерно случайности" hint="Другое зерно — другой, но тоже воспроизводимый прогон" value={c.seed} min={1} max={40} onChange={(v) => set('seed', v)} />
      </div>

      <h4 className="settings-sub">Сервисы</h4>
      <div className="knobs">
        <Range
          label="Воркеров в A"
          hint="Сколько запросов сервис A обрабатывает одновременно"
          value={c.workers}
          min={1}
          max={12}
          display={`${c.workers} · ${capacityOf(c).toFixed(1)} зап./т. без ожидания`}
          onChange={(v) => set('workers', v)}
        />
        <Range label="Своя работа A" hint="Сколько тиков A тратит на запрос сам" value={c.ownTime} min={1} max={8} display={t(c.ownTime)} onChange={(v) => set('ownTime', v)} />
        <Range
          label="Воркеров в B"
          hint="Сколько вызовов B обрабатывает одновременно"
          value={c.depWorkers}
          min={1}
          max={12}
          display={`${c.depWorkers} · ${depCapacityOf(c).toFixed(2)} выз./т.`}
          onChange={(v) => set('depWorkers', v)}
        />
        <Range label="Время ответа B" hint="Среднее время ответа здоровой зависимости" value={c.depTime} min={1} max={12} display={t(c.depTime)} onChange={(v) => set('depTime', v)} />
      </div>
    </>
  )
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

const FAULT_LABEL: Record<Fault['kind'], string> = {
  slow: 'B тормозит',
  errors: 'B отвечает ошибками',
  hang: 'B не отвечает',
}

export function FaultsEditor({ faults, onChange }: { faults: Fault[]; onChange: (f: Fault[]) => void }) {
  const set = (i: number, f: Fault) => onChange(faults.map((x, j) => (j === i ? f : x)))
  return (
    <div className="chan-editor">
      {faults.map((f, i) => (
        <div className="chan-edit" key={i}>
          <select
            value={f.kind}
            onChange={(e) => {
              const kind = e.target.value as Fault['kind']
              set(i, kind === 'slow' ? { kind, at: f.at, until: f.until, factor: 6 } : { kind, at: f.at, until: f.until })
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
              <input type="number" min={2} max={20} value={f.factor} onChange={(e) => set(i, { ...f, factor: Math.max(2, Number(e.target.value)) })} />
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
      <button type="button" className="btn-ghost" onClick={() => onChange([...faults, { kind: 'slow', at: 40, factor: 6 }])}>
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
