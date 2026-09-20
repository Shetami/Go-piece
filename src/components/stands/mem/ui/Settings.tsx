import type { MemConfig, MemPhase, MemWorkload } from '../engine/types.ts'

/**
 * Ручки стенда. Всё, что здесь меняется, превращается в новый сценарий —
 * движок о существовании UI не знает.
 */

interface KnobProps {
  config: MemConfig
  onChange: (c: MemConfig) => void
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
      <input type="range" min={min} max={max} step={step} value={value} onChange={(e) => onChange(Number(e.target.value))} />
    </label>
  )
}

const kb = (b: number) => (b >= 1024 ? `${b / 1024} КБ` : `${b} Б`)

export function ConfigKnobs({ config: c, onChange }: KnobProps) {
  const set = <K extends keyof MemConfig>(k: K, v: MemConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <div className="knobs">
      <Range
        label="GOMAXPROCS"
        hint="Сколько процессоров, столько и кэшей аллокатора — и столько же претендентов на общую блокировку"
        value={c.gomaxprocs}
        min={1}
        max={8}
        onChange={(v) => set('gomaxprocs', v)}
      />
      <Range
        label="Страниц в арене"
        hint="Физический потолок модели: страница — 8 КБ"
        value={c.heapPages}
        min={32}
        max={256}
        step={8}
        display={`${c.heapPages} · ${(c.heapPages * 8) / 1024} МБ`}
        onChange={(v) => set('heapPages', v)}
      />
      <Range
        label="Страниц в спане"
        hint="Сколько страниц рантайм отдаёт под один спан мелкого класса. Больше спан — реже пополнения"
        value={c.spanPages}
        min={1}
        max={8}
        onChange={(v) => set('spanPages', v)}
      />
      <Range
        label="Стартовый стек"
        hint="С какого размера начинается стек горутины. В Go — 2 КБ"
        value={c.stackStart}
        min={1024}
        max={16384}
        step={1024}
        display={kb(c.stackStart)}
        onChange={(v) => set('stackStart', v)}
      />
      <Range
        label="Скорость копирования"
        hint="Сколько байт стека переезжает за тик при росте"
        value={c.stackCopyRate}
        min={512}
        max={16384}
        step={512}
        display={`${kb(c.stackCopyRate)}/тик`}
        onChange={(v) => set('stackCopyRate', v)}
      />
      <Range
        label="Возврат ОС через"
        hint="Сколько тиков страница должна простоять свободной, чтобы рантайм вернул её системе. 0 — не возвращать"
        value={c.scavengeAfter}
        min={0}
        max={120}
        step={5}
        display={c.scavengeAfter === 0 ? 'никогда' : `${c.scavengeAfter} тиков`}
        onChange={(v) => set('scavengeAfter', v)}
      />
      <Range
        label="Проверка стеков каждые"
        hint="Как часто рантайм смотрит, не пора ли ужать стек. В Go — при сборке мусора. 0 — не ужимать"
        value={c.shrinkEvery}
        min={0}
        max={200}
        step={10}
        display={c.shrinkEvery === 0 ? 'никогда' : `${c.shrinkEvery} тиков`}
        onChange={(v) => set('shrinkEvery', v)}
      />
      <Toggle
        label="Анализ побега"
        hint="Выключите — и объекты, которые могли остаться на стеке, уедут в кучу"
        checked={c.escapeAnalysis}
        onChange={(v) => set('escapeAnalysis', v)}
      />
      <Toggle
        label="Общий блок для мелочи"
        hint="Объекты меньше 16 байт без указателей складываются вплотную в один блок"
        checked={c.tinyAllocator}
        onChange={(v) => set('tinyAllocator', v)}
      />
      <Toggle
        label="Блокировка mcentral"
        hint="Выключите, чтобы увидеть, сколько стоит ожидание общей блокировки"
        checked={c.centralLock}
        onChange={(v) => set('centralLock', v)}
      />
    </div>
  )
}

const PHASE_LABEL: Record<MemPhase['kind'], string> = {
  cpu: 'считает',
  alloc: 'выделяет',
  recurse: 'рекурсия',
}

function defaultPhase(kind: MemPhase['kind']): MemPhase {
  switch (kind) {
    case 'cpu':
      return { kind, ticks: [2, 4] }
    case 'alloc':
      return { kind, size: 64, count: 4, lifetime: [10, 25] }
    case 'recurse':
      return { kind, depth: 8, frame: 256 }
  }
}

function PhaseFields({ phase, onChange }: { phase: MemPhase; onChange: (p: MemPhase) => void }) {
  if (phase.kind === 'cpu') {
    const [lo, hi] = typeof phase.ticks === 'number' ? [phase.ticks, phase.ticks] : phase.ticks
    const set = (a: number, b: number) => onChange({ ...phase, ticks: a === b ? a : [Math.min(a, b), Math.max(a, b)] })
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
    const size = typeof phase.size === 'number' ? phase.size : phase.size[0]
    return (
      <span className="phase-ticks">
        <input
          type="number"
          min={1}
          max={262144}
          value={size}
          onChange={(e) => onChange({ ...phase, size: Math.max(1, Number(e.target.value)) })}
          title="Размер объекта в байтах"
        />
        Б ×
        <input
          type="number"
          min={1}
          max={32}
          value={phase.count ?? 1}
          onChange={(e) => onChange({ ...phase, count: Math.max(1, Math.min(32, Number(e.target.value))) })}
          title="Сколько объектов за один проход"
        />
        <label title="Указатель уезжает из функции — объект обязан жить в куче">
          <input
            type="checkbox"
            checked={phase.escapes !== false}
            onChange={(e) => onChange({ ...phase, escapes: e.target.checked })}
          />
          в кучу
        </label>
        <label title="Внутри объекта есть указатели — в общий блок для мелочи такой не положить">
          <input
            type="checkbox"
            checked={phase.pointers !== false}
            onChange={(e) => onChange({ ...phase, pointers: e.target.checked })}
          />
          с указателями
        </label>
      </span>
    )
  }
  return (
    <span className="phase-ticks">
      глубина
      <input
        type="number"
        min={1}
        max={64}
        value={phase.depth}
        onChange={(e) => onChange({ ...phase, depth: Math.max(1, Math.min(64, Number(e.target.value))) })}
      />
      кадр
      <input
        type="number"
        min={16}
        max={8192}
        step={16}
        value={phase.frame}
        onChange={(e) => onChange({ ...phase, frame: Math.max(16, Number(e.target.value)) })}
      />
      Б
    </span>
  )
}

function WorkloadEditor({ wl, onChange, onRemove }: {
  wl: MemWorkload
  onChange: (w: MemWorkload) => void
  onRemove: () => void
}) {
  const setPhase = (i: number, p: MemPhase) => onChange({ ...wl, phases: wl.phases.map((x, j) => (j === i ? p : x)) })
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
        <label>
          повторов
          <select
            value={String(wl.repeat ?? 1)}
            onChange={(e) => onChange({ ...wl, repeat: e.target.value === 'forever' ? 'forever' : Number(e.target.value) })}
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
            <select value={ph.kind} onChange={(e) => setPhase(i, defaultPhase(e.target.value as MemPhase['kind']))}>
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
      <button type="button" className="btn-ghost" onClick={() => onChange({ ...wl, phases: [...wl.phases, defaultPhase('cpu')] })}>
        + фаза
      </button>
    </div>
  )
}

export function WorkloadsEditor({ workloads, onChange }: {
  workloads: MemWorkload[]
  onChange: (w: MemWorkload[]) => void
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
            { name: `load${workloads.length + 1}`, count: 2, spawnAt: 0, phases: [defaultPhase('alloc'), defaultPhase('cpu')], repeat: 'forever' },
          ])
        }
      >
        + нагрузка
      </button>
    </div>
  )
}
