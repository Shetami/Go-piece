import type { Fault, Publish, SagaConfig } from '../engine/types.ts'

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

export function ConfigKnobs({ config: c, onChange }: { config: SagaConfig; onChange: (c: SagaConfig) => void }) {
  const set = <K extends keyof SagaConfig>(k: K, v: SagaConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <>
      <h4 className="settings-sub">Публикация событий</h4>
      <div className="knobs">
        <Choice<Publish>
          label="Как отправляется событие"
          hint="Две записи подряд: сначала база, потом брокер. Outbox: событие пишется в ту же транзакцию, отправляет отдельный процесс"
          value={c.publish}
          options={[
            { value: 'dual-write', label: 'две записи' },
            { value: 'outbox', label: 'outbox' },
          ]}
          onChange={(v) => set('publish', v)}
          wide
        />
        <Range label="Отправщик выгребает outbox" hint="Как часто процесс-отправщик забирает накопленные строки" value={c.relayEvery} min={1} max={15} display={`раз в ${t(c.relayEvery)}`} onChange={(v) => set('relayEvery', v)} />
        <Range label="Доставка события" hint="Сколько тиков событие идёт от публикации до потребителя" value={c.brokerDelay} min={1} max={10} display={t(c.brokerDelay)} onChange={(v) => set('brokerDelay', v)} />
      </div>

      <h4 className="settings-sub">Потребители</h4>
      <div className="knobs">
        <Toggle label="Ключ идемпотентности" hint="Потребитель распознаёт повторную доставку и не делает работу дважды" checked={c.idempotent} onChange={(v) => set('idempotent', v)} />
        <Toggle label="Компенсация" hint="Если доставка не удалась, деньги возвращаются отдельным шагом саги" checked={c.compensate} onChange={(v) => set('compensate', v)} />
        <Range label="Неудач списания" hint="Доля обработок события об оплате, которые не удаются" value={c.payFail} min={0} max={60} step={5} display={`${c.payFail}%`} onChange={(v) => set('payFail', v)} />
        <Range label="Неудач доставки" hint="Доля обработок события о доставке, которые не удаются" value={c.shipFail} min={0} max={80} step={5} display={`${c.shipFail}%`} onChange={(v) => set('shipFail', v)} />
        <Range label="Повторов обработки" hint="Сколько раз потребитель повторяет неудачную обработку" value={c.retries} min={0} max={5} onChange={(v) => set('retries', v)} />
        <Range label="Пауза перед повтором" hint="Через сколько тиков потребитель попробует снова" value={c.retryAfter} min={1} max={15} display={t(c.retryAfter)} onChange={(v) => set('retryAfter', v)} />
        <Range label="Обрабатывают одновременно" hint="Сколько событий каждый потребитель обрабатывает параллельно" value={c.workers} min={1} max={8} onChange={(v) => set('workers', v)} />
        <Range label="Время обработки" hint="Среднее время обработки события потребителем" value={c.workTime} min={1} max={8} display={t(c.workTime)} onChange={(v) => set('workTime', v)} />
      </div>

      <h4 className="settings-sub">Поток</h4>
      <div className="knobs">
        <Range label="Заказов за тик" hint="Сколько заказов приходит в среднем за тик" value={c.rate} min={0.1} max={2} step={0.1} display={`${c.rate.toFixed(1)}/т.`} onChange={(v) => set('rate', v)} />
        <Range label="Зерно случайности" hint="Другое зерно — другой, но тоже воспроизводимый прогон" value={c.seed} min={1} max={40} onChange={(v) => set('seed', v)} />
      </div>
    </>
  )
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

const FAULT_LABEL: Record<Fault['kind'], string> = {
  crash: 'сервис заказов падает',
  'broker-down': 'брокер недоступен',
  duplicates: 'доставка дублирует события',
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
              set(i, kind === 'crash' ? { kind, at: f.at } : { kind, at: f.at, until: 'until' in f && f.until ? f.until : f.at + 20 })
            }}
            aria-label="Что случается"
          >
            {(Object.keys(FAULT_LABEL) as Fault['kind'][]).map((k) => (
              <option key={k} value={k}>
                {FAULT_LABEL[k]}
              </option>
            ))}
          </select>
          <label>
            с тика
            <input type="number" min={1} max={290} value={f.at} onChange={(e) => set(i, { ...f, at: Math.max(1, Number(e.target.value)) })} />
          </label>
          {f.kind !== 'crash' && (
            <label>
              до
              <input type="number" min={f.at + 1} max={300} value={f.until} onChange={(e) => set(i, { ...f, until: Math.max(f.at + 1, Number(e.target.value)) })} />
            </label>
          )}
          <button type="button" className="btn-ghost" onClick={() => onChange(faults.filter((_, j) => j !== i))} title="Убрать сбой">
            ✕
          </button>
        </div>
      ))}
      <button type="button" className="btn-ghost" onClick={() => onChange([...faults, { kind: 'crash', at: 60 }])}>
        + сбой
      </button>
    </div>
  )
}
