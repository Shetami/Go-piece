import type { Config, Phase, Workload } from '../engine/types.ts'

/**
 * Ручки стенда. Всё, что здесь меняется, превращается в новый Scenario —
 * движок о существовании UI не знает.
 */

interface KnobProps {
  config: Config
  onChange: (c: Config) => void
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

function Range({ label, hint, value, min, max, step = 1, onChange }: {
  label: string
  hint: string
  value: number
  min: number
  max: number
  step?: number
  onChange: (v: number) => void
}) {
  return (
    <label className="knob" title={hint}>
      <span className="knob-head">
        <span>{label}</span>
        <b>{value}</b>
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
  const set = <K extends keyof Config>(k: K, v: Config[K]) => onChange({ ...c, [k]: v })
  return (
    <div className="knobs">
      <Range
        label="GOMAXPROCS"
        hint="Сколько P — сколько горутин исполняется одновременно"
        value={c.gomaxprocs}
        min={1}
        max={8}
        onChange={(v) => set('gomaxprocs', v)}
      />
      <Range
        label="Квант, тиков"
        hint="Сколько горутина работает подряд, прежде чем sysmon её вытеснит (в Go — 10 мс)"
        value={c.quantum}
        min={2}
        max={100}
        onChange={(v) => set('quantum', v)}
      />
      <Range
        label="Порог retake, тиков"
        hint="Сколько тиков sysmon терпит системный вызов, прежде чем отобрать P"
        value={c.retakeThreshold}
        min={1}
        max={40}
        onChange={(v) => set('retakeThreshold', v)}
      />
      <Range
        label="Ёмкость runq"
        hint="Размер локальной очереди P. В Go — 256"
        value={c.runqCapacity}
        min={4}
        max={256}
        step={4}
        onChange={(v) => set('runqCapacity', v)}
      />
      <Range
        label="Правило 61: каждые N"
        hint="Как часто P заглядывает в глобальную очередь раньше локальной. 0 — никогда"
        value={c.globalCheckEvery === false ? 0 : c.globalCheckEvery}
        min={0}
        max={101}
        onChange={(v) => set('globalCheckEvery', v === 0 ? false : v)}
      />
      <Range
        label="Максимум потоков"
        hint="Предел числа M (debug.SetMaxThreads). В Go — 10 000"
        value={Math.min(c.maxThreads, 64)}
        min={1}
        max={64}
        onChange={(v) => set('maxThreads', v === 64 ? 10_000 : v)}
      />
      <Toggle
        label="Кража работы"
        hint="Пустой P забирает половину очереди у соседа"
        checked={c.workStealing}
        onChange={(v) => set('workStealing', v)}
      />
      <Toggle
        label="Слот runnext"
        hint="Разбуженная горутина идёт следующей, минуя очередь"
        checked={c.runnext}
        onChange={(v) => set('runnext', v)}
      />
      <Toggle
        label="Асинхронное вытеснение"
        hint="Go 1.14+: sysmon прерывает долгую горутину сигналом. Выключите — вернётесь в Go 1.13"
        checked={c.asyncPreemption}
        onChange={(v) => set('asyncPreemption', v)}
      />
    </div>
  )
}

const PHASE_LABEL: Record<Phase['kind'], string> = {
  cpu: 'CPU',
  syscall: 'syscall (файл)',
  net: 'сеть',
  sleep: 'sleep',
  mutex: 'мьютекс',
  chanSend: 'ch <- v',
  chanRecv: '<-ch',
}

function defaultPhase(kind: Phase['kind']): Phase {
  switch (kind) {
    case 'cpu':
    case 'syscall':
    case 'net':
    case 'sleep':
      return { kind, ticks: [5, 10] }
    case 'mutex':
      return { kind, lock: 'mu', ticks: [3, 6] }
    case 'chanSend':
    case 'chanRecv':
      return { kind, chan: 'ch' }
  }
}

function TicksInput({ phase, onChange }: { phase: Phase; onChange: (p: Phase) => void }) {
  if (!('ticks' in phase)) return null
  const t = phase.ticks
  const forever = t === 'forever'
  const [lo, hi] = forever ? [10, 10] : typeof t === 'number' ? [t, t] : t
  const set = (a: number, b: number) =>
    onChange({ ...phase, ticks: a === b ? a : [Math.min(a, b), Math.max(a, b)] } as Phase)
  return (
    <span className="phase-ticks">
      {!forever && (
        <>
          <input type="number" min={1} max={500} value={lo} onChange={(e) => set(Number(e.target.value), hi)} />
          –
          <input type="number" min={1} max={500} value={hi} onChange={(e) => set(lo, Number(e.target.value))} />
          тиков
        </>
      )}
      {phase.kind === 'cpu' && (
        <label title="Бесконечный цикл без точек безопасности">
          <input
            type="checkbox"
            checked={forever}
            onChange={(e) => onChange({ ...phase, ticks: e.target.checked ? 'forever' : [5, 10] })}
          />
          вечно
        </label>
      )}
    </span>
  )
}

function WorkloadEditor({ wl, gomaxprocs, onChange, onRemove }: {
  wl: Workload
  gomaxprocs: number
  onChange: (w: Workload) => void
  onRemove: () => void
}) {
  const setPhase = (i: number, p: Phase) =>
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
            max={64}
            value={wl.count}
            onChange={(e) => onChange({ ...wl, count: Math.max(1, Math.min(64, Number(e.target.value))) })}
          />
        </label>
        <label>
          на
          <select
            value={wl.spawnOn ?? -1}
            onChange={(e) => {
              const v = Number(e.target.value)
              onChange({ ...wl, spawnOn: v < 0 ? undefined : v })
            }}
          >
            <option value={-1}>все P по кругу</option>
            {Array.from({ length: gomaxprocs }, (_, i) => (
              <option key={i} value={i}>
                P{i}
              </option>
            ))}
          </select>
        </label>
        <label>
          старт
          <select
            value={wl.spawnAt === 'staggered' ? 'staggered' : String(wl.spawnAt)}
            onChange={(e) =>
              onChange({ ...wl, spawnAt: e.target.value === 'staggered' ? 'staggered' : Number(e.target.value) })
            }
          >
            <option value="0">сразу</option>
            <option value="staggered">по одной в тик</option>
            {typeof wl.spawnAt === 'number' && wl.spawnAt > 0 && (
              <option value={wl.spawnAt}>на тике {wl.spawnAt}</option>
            )}
          </select>
        </label>
        <label>
          повторов
          <select
            value={String(wl.repeat ?? 1)}
            onChange={(e) =>
              onChange({ ...wl, repeat: e.target.value === 'forever' ? 'forever' : Number(e.target.value) })
            }
          >
            {[1, 2, 3, 5, 10].map((n) => (
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
            <select
              value={ph.kind}
              onChange={(e) => setPhase(i, defaultPhase(e.target.value as Phase['kind']))}
            >
              {Object.entries(PHASE_LABEL).map(([k, label]) => (
                <option key={k} value={k}>
                  {label}
                </option>
              ))}
            </select>
            {(ph.kind === 'chanSend' || ph.kind === 'chanRecv') && (
              <input
                className="phase-name"
                value={ph.chan}
                onChange={(e) => setPhase(i, { ...ph, chan: e.target.value })}
                aria-label="Имя канала"
              />
            )}
            {ph.kind === 'mutex' && (
              <input
                className="phase-name"
                value={ph.lock}
                onChange={(e) => setPhase(i, { ...ph, lock: e.target.value })}
                aria-label="Имя мьютекса"
              />
            )}
            <TicksInput phase={ph} onChange={(p) => setPhase(i, p)} />
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

export function WorkloadsEditor({ workloads, gomaxprocs, onChange }: {
  workloads: Workload[]
  gomaxprocs: number
  onChange: (w: Workload[]) => void
}) {
  return (
    <div className="wls">
      {workloads.map((wl, i) => (
        <WorkloadEditor
          key={i}
          wl={wl}
          gomaxprocs={gomaxprocs}
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
            { name: `load${workloads.length + 1}`, count: 4, spawnAt: 0, phases: [defaultPhase('cpu')] },
          ])
        }
      >
        + нагрузка
      </button>
    </div>
  )
}
