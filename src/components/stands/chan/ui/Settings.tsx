import type { ChanConfig, ChanPhase, ChanSpec, ChanWorkload, SelectCase } from '../engine/types.ts'

/**
 * Ручки стенда. Всё, что здесь меняется, превращается в новый сценарий —
 * движок о существовании UI не знает.
 */

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

export function ConfigKnobs({ config: c, onChange }: { config: ChanConfig; onChange: (c: ChanConfig) => void }) {
  const set = <K extends keyof ChanConfig>(k: K, v: ChanConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <div className="knobs">
      <Range
        label="GOMAXPROCS"
        hint="Сколько горутин исполняется одновременно. На каналы влияет косвенно: чем больше процессоров, тем чаще встречная сторона оказывается на месте"
        value={c.gomaxprocs}
        min={1}
        max={8}
        onChange={(v) => set('gomaxprocs', v)}
      />
      <Range
        label="Утечка после"
        hint="Сколько тиков горутина должна проспать, чтобы стенд назвал её утечкой. Мерка самого стенда: рантайм такой границы не проводит"
        value={c.leakAfter}
        min={10}
        max={120}
        step={5}
        display={`${c.leakAfter} тиков`}
        onChange={(v) => set('leakAfter', v)}
      />
      <Toggle
        label="Передача из рук в руки"
        hint="Выключите — и значение будет проходить через буфер даже тогда, когда получатель уже ждёт: лишнее копирование и лишнее пробуждение вместо одной передачи"
        checked={c.directHandoff}
        onChange={(v) => set('directHandoff', v)}
      />
      <Toggle
        label="runnext для разбуженной"
        hint="Разбуженная горутина исполняется следующей. Выключите — встанет в хвост очереди, и средняя задержка передачи вырастет"
        checked={c.runnext}
        onChange={(v) => set('runnext', v)}
      />
      <Toggle
        label="Обнаружение deadlock"
        hint="checkdead: рантайм падает, когда спят все горутины. Выключите, чтобы досмотреть, что происходит после"
        checked={c.deadlockDetect}
        onChange={(v) => set('deadlockDetect', v)}
      />
    </div>
  )
}

/* ──────────────────────────────── каналы ──────────────────────────────── */

export function ChansEditor({ chans, onChange }: { chans: ChanSpec[]; onChange: (c: ChanSpec[]) => void }) {
  const set = (i: number, next: ChanSpec) => onChange(chans.map((x, j) => (j === i ? next : x)))
  return (
    <div className="chan-editor">
      {chans.map((c, i) => (
        <div className="chan-edit" key={i}>
          <input
            className="wl-name"
            value={c.name}
            onChange={(e) => set(i, { ...c, name: e.target.value })}
            aria-label="Имя канала"
          />
          <label title="Вместимость буфера. 0 — небуферизованный канал, точка встречи">
            буфер
            <input
              type="number"
              min={0}
              max={16}
              value={c.cap}
              disabled={c.nil === true}
              onChange={(e) => set(i, { ...c, cap: Math.max(0, Math.min(16, Number(e.target.value))) })}
            />
          </label>
          <label title="Канал объявлен, но не создан через make. Любая операция на нём блокируется навсегда">
            <input type="checkbox" checked={c.nil === true} onChange={(e) => set(i, { ...c, nil: e.target.checked })} />
            nil
          </label>
        </div>
      ))}
    </div>
  )
}

/* ─────────────────────────────── нагрузки ─────────────────────────────── */

const PHASE_LABEL: Record<ChanPhase['kind'], string> = {
  cpu: 'считает',
  send: 'отправляет',
  recv: 'принимает',
  select: 'select',
  close: 'закрывает',
}

function defaultPhase(kind: ChanPhase['kind'], chan: string): ChanPhase {
  switch (kind) {
    case 'cpu':
      return { kind, ticks: [2, 4] }
    case 'send':
      return { kind, chan, count: 1 }
    case 'recv':
      return { kind, chan, count: 1 }
    case 'select':
      return { kind, cases: [{ chan, op: 'recv' }] }
    case 'close':
      return { kind, chan }
  }
}

function ChanPicker({ chans, value, onChange }: { chans: ChanSpec[]; value: string; onChange: (v: string) => void }) {
  return (
    <select value={value} onChange={(e) => onChange(e.target.value)} aria-label="Канал">
      {chans.map((c) => (
        <option key={c.name} value={c.name}>
          {c.name}
        </option>
      ))}
    </select>
  )
}

function PhaseFields({ phase, chans, onChange }: {
  phase: ChanPhase
  chans: ChanSpec[]
  onChange: (p: ChanPhase) => void
}) {
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

  if (phase.kind === 'send' || phase.kind === 'recv') {
    return (
      <span className="phase-ticks">
        <ChanPicker chans={chans} value={phase.chan} onChange={(chan) => onChange({ ...phase, chan })} />
        ×
        <input
          type="number"
          min={1}
          max={32}
          value={phase.count ?? 1}
          title="Сколько значений подряд"
          onChange={(e) => onChange({ ...phase, count: Math.max(1, Math.min(32, Number(e.target.value))) })}
        />
      </span>
    )
  }

  if (phase.kind === 'close') {
    return (
      <span className="phase-ticks">
        <ChanPicker chans={chans} value={phase.chan} onChange={(chan) => onChange({ ...phase, chan })} />
      </span>
    )
  }

  // select: по одному переключателю на канал — «не участвует», «приём» или «отправка».
  const opOf = (name: string): '' | SelectCase['op'] => phase.cases.find((c) => c.chan === name)?.op ?? ''
  const setOp = (name: string, op: '' | SelectCase['op']) => {
    const rest = phase.cases.filter((c) => c.chan !== name)
    onChange({ ...phase, cases: op === '' ? rest : [...rest, { chan: name, op }] })
  }
  return (
    <span className="phase-ticks phase-select">
      {chans.map((c) => (
        <label key={c.name} title={`Участвует ли «${c.name}» в этом select`}>
          {c.name}
          <select value={opOf(c.name)} onChange={(e) => setOp(c.name, e.target.value as '' | SelectCase['op'])}>
            <option value="">—</option>
            <option value="recv">приём</option>
            <option value="send">отправка</option>
          </select>
        </label>
      ))}
      <label title="Ветка default: не блокироваться, если ни один case не готов">
        <input
          type="checkbox"
          checked={phase.default === true}
          onChange={(e) => onChange({ ...phase, default: e.target.checked })}
        />
        default
      </label>
    </span>
  )
}

function WorkloadEditor({ wl, chans, onChange, onRemove }: {
  wl: ChanWorkload
  chans: ChanSpec[]
  onChange: (w: ChanWorkload) => void
  onRemove: () => void
}) {
  const setPhase = (i: number, p: ChanPhase) => onChange({ ...wl, phases: wl.phases.map((x, j) => (j === i ? p : x)) })
  const first = chans[0]?.name ?? 'ch'
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
            max={24}
            value={wl.count}
            onChange={(e) => onChange({ ...wl, count: Math.max(1, Math.min(24, Number(e.target.value))) })}
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
            <select
              value={ph.kind}
              onChange={(e) => setPhase(i, defaultPhase(e.target.value as ChanPhase['kind'], first))}
            >
              {Object.entries(PHASE_LABEL).map(([k, label]) => (
                <option key={k} value={k}>
                  {label}
                </option>
              ))}
            </select>
            <PhaseFields phase={ph} chans={chans} onChange={(p) => setPhase(i, p)} />
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
        onClick={() => onChange({ ...wl, phases: [...wl.phases, defaultPhase('cpu', first)] })}
      >
        + фаза
      </button>
    </div>
  )
}

export function WorkloadsEditor({ workloads, chans, onChange }: {
  workloads: ChanWorkload[]
  chans: ChanSpec[]
  onChange: (w: ChanWorkload[]) => void
}) {
  const first = chans[0]?.name ?? 'ch'
  return (
    <div className="wls">
      {workloads.map((wl, i) => (
        <WorkloadEditor
          key={i}
          wl={wl}
          chans={chans}
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
              count: 1,
              spawnAt: 0,
              phases: [defaultPhase('send', first), defaultPhase('cpu', first)],
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
