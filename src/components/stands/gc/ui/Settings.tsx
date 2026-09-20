import type { GcConfig, MutPhase, MutWorkload } from '../engine/types.ts'

/**
 * Ручки стенда. Всё, что здесь меняется, превращается в новый сценарий —
 * движок о существовании UI не знает.
 */

interface KnobProps {
  config: GcConfig
  onChange: (c: GcConfig) => void
}

function Toggle({ label, hint, checked, onChange }: {
  label: string
  hint: string
  checked: boolean
  onChange: (v: boolean) => void
}) {
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
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
      />
    </label>
  )
}

export function ConfigKnobs({ config: c, onChange }: KnobProps) {
  const set = <K extends keyof GcConfig>(k: K, v: GcConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <div className="knobs">
      <Range
        label="GOGC"
        hint="Сколько мусора терпим рядом с живым: цель = живое × (1 + GOGC/100). 0 — GOGC=off"
        value={c.gogc === false ? 0 : c.gogc}
        min={0}
        max={400}
        step={10}
        display={c.gogc === false ? 'off' : String(c.gogc)}
        onChange={(v) => set('gogc', v === 0 ? false : v)}
      />
      <Range
        label="GOMEMLIMIT, блоков"
        hint="Мягкий лимит памяти: цель никогда не превысит его. 0 — не задан"
        value={c.memLimit === false ? 0 : c.memLimit}
        min={0}
        max={c.heapCapacity}
        step={10}
        display={c.memLimit === false ? 'нет' : String(c.memLimit)}
        onChange={(v) => set('memLimit', v === 0 ? false : v)}
      />
      <Range
        label="GOMAXPROCS"
        hint="Сколько процессоров делят между собой программа и сборщик"
        value={c.gomaxprocs}
        min={1}
        max={8}
        onChange={(v) => set('gomaxprocs', v)}
      />
      <Range
        label="Доля CPU сборщику"
        hint="Сколько процессоров уходит на фоновую разметку. В Go — 0.25"
        value={Math.round(c.gcCpuShare * 100)}
        min={0}
        max={75}
        step={5}
        display={`${Math.round(c.gcCpuShare * 100)}%`}
        onChange={(v) => set('gcCpuShare', v / 100)}
      />
      <Range
        label="Скорость разметки"
        hint="Сколько блоков просматривает маркер за тик"
        value={c.scanRate}
        min={1}
        max={10}
        onChange={(v) => set('scanRate', v)}
      />
      <Range
        label="Скорость подметания"
        hint="Сколько блоков освобождает подметальщик за тик"
        value={c.sweepRate}
        min={1}
        max={20}
        onChange={(v) => set('sweepRate', v)}
      />
      <Range
        label="Длина паузы, тиков"
        hint="Сколько длится каждая из двух остановок мира"
        value={c.stwTicks}
        min={1}
        max={12}
        onChange={(v) => set('stwTicks', v)}
      />
      <Range
        label="Минимальная цель"
        hint="Цель не опускается ниже этого значения. В Go минимум — 4 МБ"
        value={c.initialGoal}
        min={10}
        max={120}
        step={2}
        onChange={(v) => set('initialGoal', v)}
      />
      <Range
        label="Размер кучи, блоков"
        hint="Физический потолок: больше в модель не влезает"
        value={c.heapCapacity}
        min={60}
        max={400}
        step={20}
        onChange={(v) => set('heapCapacity', v)}
      />
      <Toggle
        label="Барьер записи"
        hint="Выключите — и увидите, как сборщик освобождает живую память. В Go выключить нельзя"
        checked={c.writeBarrier}
        onChange={(v) => set('writeBarrier', v)}
      />
      <Toggle
        label="Аллокация чёрным"
        hint="Объекты, созданные во время разметки, считаются живыми сразу"
        checked={c.allocBlack}
        onChange={(v) => set('allocBlack', v)}
      />
      <Toggle
        label="Помощь в разметке"
        hint="Аллокация оплачивается работой на разметке. Выключите — куча улетит за цель"
        checked={c.markAssist}
        onChange={(v) => set('markAssist', v)}
      />
      <Toggle
        label="Маркеры на простое"
        hint="Процессор, которому нечего исполнять, подрабатывает маркером"
        checked={c.idleWorkers}
        onChange={(v) => set('idleWorkers', v)}
      />
    </div>
  )
}

const PHASE_LABEL: Record<MutPhase['kind'], string> = {
  cpu: 'считает',
  alloc: 'выделяет',
  drop: 'бросает ссылку',
  move: 'прячет объект',
}

function defaultPhase(kind: MutPhase['kind']): MutPhase {
  switch (kind) {
    case 'cpu':
      return { kind, ticks: [2, 4] }
    case 'alloc':
      return { kind, blocks: 3 }
    case 'drop':
    case 'move':
      return { kind }
  }
}

function PhaseFields({ phase, onChange }: { phase: MutPhase; onChange: (p: MutPhase) => void }) {
  if (phase.kind === 'cpu') {
    const [lo, hi] = typeof phase.ticks === 'number' ? [phase.ticks, phase.ticks] : phase.ticks
    const set = (a: number, b: number) =>
      onChange({ ...phase, ticks: a === b ? a : [Math.min(a, b), Math.max(a, b)] })
    return (
      <span className="phase-ticks">
        <input type="number" min={1} max={99} value={lo} onChange={(e) => set(Number(e.target.value), hi)} />
        –
        <input type="number" min={1} max={99} value={hi} onChange={(e) => set(lo, Number(e.target.value))} />
        тиков
      </span>
    )
  }
  if (phase.kind === 'alloc') {
    return (
      <span className="phase-ticks">
        <input
          type="number"
          min={1}
          max={16}
          value={phase.blocks}
          onChange={(e) => onChange({ ...phase, blocks: Math.max(1, Math.min(16, Number(e.target.value))) })}
        />
        блоков
        <label title="Сохранить структуру в глобальной переменной — она станет долгоживущей">
          <input
            type="checkbox"
            checked={phase.retain === true}
            onChange={(e) => onChange({ ...phase, retain: e.target.checked })}
          />
          в глобальные
        </label>
      </span>
    )
  }
  return null
}

function WorkloadEditor({ wl, onChange, onRemove }: {
  wl: MutWorkload
  onChange: (w: MutWorkload) => void
  onRemove: () => void
}) {
  const setPhase = (i: number, p: MutPhase) =>
    onChange({ ...wl, phases: wl.phases.map((x, j) => (j === i ? p : x)) })

  return (
    <div className="wl">
      <div className="wl-head">
        <input
          className="wl-name"
          value={wl.name}
          onChange={(e) => onChange({ ...wl, name: e.target.value })}
          aria-label="Название нагрузки"
        />
        <label>
          ×
          <input
            type="number"
            min={1}
            max={32}
            value={wl.count}
            onChange={(e) => onChange({ ...wl, count: Math.max(1, Math.min(32, Number(e.target.value))) })}
          />
        </label>
        <label title="Сколько указателей горутина держит на стеке — это её корни">
          корней
          <input
            type="number"
            min={1}
            max={6}
            value={wl.stackSlots ?? 3}
            onChange={(e) => onChange({ ...wl, stackSlots: Math.max(1, Math.min(6, Number(e.target.value))) })}
          />
        </label>
        <label>
          повторов
          <select
            value={String(wl.repeat ?? 1)}
            onChange={(e) =>
              onChange({ ...wl, repeat: e.target.value === 'forever' ? 'forever' : Number(e.target.value) })
            }
          >
            {[1, 2, 5, 10, 20].map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
            <option value="forever">∞</option>
          </select>
        </label>
        <button type="button" className="btn-ghost" onClick={onRemove} title="Удалить нагрузку">
          ✕
        </button>
      </div>
      <ol className="wl-phases">
        {wl.phases.map((ph, i) => (
          <li key={i}>
            <select value={ph.kind} onChange={(e) => setPhase(i, defaultPhase(e.target.value as MutPhase['kind']))}>
              {Object.entries(PHASE_LABEL).map(([k, label]) => (
                <option key={k} value={k}>
                  {label}
                </option>
              ))}
            </select>
            <PhaseFields phase={ph} onChange={(p) => setPhase(i, p)} />
            {wl.phases.length > 1 && (
              <button
                type="button"
                className="btn-ghost"
                onClick={() => onChange({ ...wl, phases: wl.phases.filter((_, j) => j !== i) })}
                title="Удалить фазу"
              >
                ✕
              </button>
            )}
          </li>
        ))}
      </ol>
      <button
        type="button"
        className="btn-ghost"
        onClick={() => onChange({ ...wl, phases: [...wl.phases, defaultPhase('cpu')] })}
      >
        + фаза
      </button>
    </div>
  )
}

export function WorkloadsEditor({ workloads, onChange }: {
  workloads: MutWorkload[]
  onChange: (w: MutWorkload[]) => void
}) {
  return (
    <div className="wls">
      {workloads.map((wl, i) => (
        <WorkloadEditor
          key={i}
          wl={wl}
          onChange={(next) => onChange(workloads.map((x, j) => (j === i ? next : x)))}
          onRemove={() => onChange(workloads.filter((_, j) => j !== i))}
        />
      ))}
      <button
        type="button"
        className="btn"
        onClick={() =>
          onChange([
            ...workloads,
            {
              name: `load${workloads.length + 1}`,
              count: 2,
              spawnAt: 0,
              stackSlots: 3,
              phases: [defaultPhase('alloc'), defaultPhase('cpu')],
              repeat: 'forever',
            },
          ])
        }
      >
        + нагрузка
      </button>
    </div>
  )
}
