import type { ClientOp, ClientSpec, Fault, ReplConfig, ReadFrom, Standbys, SyncLevel } from '../engine/types.ts'
import { SYNC_LEVELS } from '../engine/types.ts'

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

function Range({ label, hint, value, min, max, display, onChange }: {
  label: string
  hint: string
  value: number
  min: number
  max: number
  display?: string
  onChange: (v: number) => void
}) {
  return (
    <label className="knob" title={hint}>
      <span className="knob-head">
        <span>{label}</span>
        <b>{display ?? value}</b>
      </span>
      <input type="range" min={min} max={max} value={value} onChange={(e) => onChange(Number(e.target.value))} />
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

export function ConfigKnobs({ config: c, onChange }: { config: ReplConfig; onChange: (c: ReplConfig) => void }) {
  const set = <K extends keyof ReplConfig>(k: K, v: ReplConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <>
      <h4 className="settings-sub">Синхронность</h4>
      <div className="knobs">
        <Choice<Standbys>
          label="synchronous_standby_names"
          hint="Кого ждёт коммит: никого (асинхронно), только r1 или любую одну реплику"
          value={c.standbys}
          options={[
            { value: 'none', label: 'пусто' },
            { value: 'first-r1', label: 'FIRST 1 (r1)' },
            { value: 'any-1', label: 'ANY 1 (*)' },
          ]}
          onChange={(v) => set('standbys', v)}
        />
        <Choice<SyncLevel>
          label="synchronous_commit"
          hint="Чего ждать от синхронной реплики: remote_write — получила, on — сбросила на диск, remote_apply — проиграла. off и local реплик не ждут"
          value={c.syncCommit}
          options={SYNC_LEVELS.map((v) => ({ value: v, label: v }))}
          onChange={(v) => set('syncCommit', v)}
        />
      </div>

      <h4 className="settings-sub">Кластер</h4>
      <div className="knobs">
        <Range label="Реплик" hint="Сколько реплик у ведущего" value={c.replicas} min={1} max={3} onChange={(v) => set('replicas', v)} />
        <Range
          label="Проигрывание"
          hint="Сколько записей WAL реплика проигрывает за тик (если у реплики не задано своё)"
          value={c.replayRate}
          min={1}
          max={8}
          display={`${c.replayRate} зап./т.`}
          onChange={(v) => set('replayRate', v)}
        />
        <Toggle
          label="Автоматическое переключение"
          hint="Менеджер кластера (Patroni) сам делает ведущей самую свежую реплику, когда ведущий умирает"
          checked={c.failover}
          onChange={(v) => set('failover', v)}
        />
        <Range
          label="Смерть ведущего замечают через"
          hint="Сколько тиков менеджер кластера ждёт, прежде чем признать ведущего мёртвым. Меньше — быстрее переключение, но больше ложных"
          value={c.detectAfter}
          min={1}
          max={12}
          display={t(c.detectAfter)}
          onChange={(v) => set('detectAfter', v)}
        />
      </div>

      <h4 className="settings-sub">Хранение WAL</h4>
      <div className="knobs">
        <Toggle
          label="Слоты репликации"
          hint="Ведущий хранит WAL, пока его не заберут все реплики — даже выключенные"
          checked={c.slots}
          onChange={(v) => set('slots', v)}
        />
        <Range
          label="wal_keep_size"
          hint="Без слотов ведущий хранит только столько последних записей WAL"
          value={c.walKeep}
          min={4}
          max={40}
          display={`${c.walKeep} зап.`}
          onChange={(v) => set('walKeep', v)}
        />
      </div>

      <h4 className="settings-sub">Запросы на репликах</h4>
      <div className="knobs">
        <Toggle
          label="hot_standby_feedback"
          hint="Реплика сообщает ведущему свой самый старый снимок, и ведущий не чистит нужные ему версии"
          checked={c.hotStandbyFeedback}
          onChange={(v) => set('hotStandbyFeedback', v)}
        />
        <Range
          label="max_standby_streaming_delay"
          hint="Сколько реплика ждёт мешающий запрос, прежде чем отменить его. Крайнее левое значение — ждать сколько угодно"
          value={c.maxStandbyDelay}
          min={-1}
          max={12}
          display={c.maxStandbyDelay < 0 ? 'без предела' : t(c.maxStandbyDelay)}
          onChange={(v) => set('maxStandbyDelay', v)}
        />
        <Range
          label="Автовакуум на ведущем"
          hint="Как часто вакуум убирает мёртвые версии на ведущем. 0 — выключен"
          value={c.vacuumEvery}
          min={0}
          max={12}
          display={c.vacuumEvery === 0 ? 'выкл' : t(c.vacuumEvery)}
          onChange={(v) => set('vacuumEvery', v)}
        />
      </div>
    </>
  )
}

/* ─────────────────────────────── клиенты ─────────────────────────────── */

function FromPicker({ value, replicas, onChange, noPrimary }: { value: ReadFrom; replicas: number; onChange: (v: ReadFrom) => void; noPrimary?: boolean }) {
  return (
    <select
      value={String(value)}
      onChange={(e) => onChange(e.target.value === 'primary' || e.target.value === 'any' ? e.target.value : Number(e.target.value))}
      aria-label="Откуда читать"
    >
      {!noPrimary && <option value="primary">ведущий</option>}
      {!noPrimary && <option value="any">любая реплика</option>}
      {Array.from({ length: replicas }, (_, i) => (
        <option key={i} value={i + 1}>
          r{i + 1}
        </option>
      ))}
    </select>
  )
}

function OpEditor({ op, keys, replicas, onChange }: { op: ClientOp; keys: string[]; replicas: number; onChange: (o: ClientOp) => void }) {
  const keyPick = (
    <select value={op.key} onChange={(e) => onChange({ ...op, key: e.target.value })} aria-label="Строка">
      {keys.map((k) => (
        <option key={k} value={k}>
          {k}
        </option>
      ))}
    </select>
  )
  return (
    <span className="phase-ticks">
      <select
        value={op.kind}
        onChange={(e) => {
          const kind = e.target.value as ClientOp['kind']
          onChange(kind === 'write' ? { kind, key: op.key } : kind === 'read' ? { kind, key: op.key, from: 'any' } : { kind, key: op.key, from: 1, ticks: 10 })
        }}
      >
        <option value="write">UPDATE</option>
        <option value="read">SELECT</option>
        <option value="query">отчёт</option>
      </select>
      {keyPick}
      {op.kind === 'read' && <FromPicker value={op.from} replicas={replicas} onChange={(from) => onChange({ ...op, from })} />}
      {op.kind === 'query' && (
        <>
          <FromPicker value={op.from} replicas={replicas} noPrimary onChange={(from) => onChange({ ...op, from: typeof from === 'number' ? from : 1 })} />
          <input type="number" min={1} max={40} value={op.ticks} onChange={(e) => onChange({ ...op, ticks: Math.max(1, Number(e.target.value)) })} title="Сколько тиков идёт отчёт" />
          т.
        </>
      )}
    </span>
  )
}

export function ClientsEditor({ clients, keys, replicas, onChange }: {
  clients: ClientSpec[]
  keys: string[]
  replicas: number
  onChange: (c: ClientSpec[]) => void
}) {
  const setClient = (i: number, c: ClientSpec) => onChange(clients.map((x, j) => (j === i ? c : x)))
  return (
    <div className="wls">
      {clients.map((c, i) => (
        <div className="wl" key={i}>
          <div className="wl-head">
            <input className="wl-name" value={c.name} onChange={(e) => setClient(i, { ...c, name: e.target.value })} aria-label="Имя клиента" />
            <label>
              с тика
              <input type="number" min={1} max={60} value={c.at} onChange={(e) => setClient(i, { ...c, at: Math.max(1, Number(e.target.value)) })} />
            </label>
            <label>
              повторов
              <select
                value={String(c.repeat ?? 1)}
                onChange={(e) => setClient(i, { ...c, repeat: e.target.value === 'forever' ? 'forever' : Number(e.target.value) })}
              >
                {[1, 4, 8, 12, 20].map((n) => (
                  <option key={n} value={n}>
                    {n}
                  </option>
                ))}
                <option value="forever">∞</option>
              </select>
            </label>
            <button type="button" className="btn-ghost" onClick={() => onChange(clients.filter((_, j) => j !== i))} title="Удалить клиента">
              ✕
            </button>
          </div>
          <ol className="wl-phases">
            {c.ops.map((op, oi) => (
              <li key={oi}>
                <OpEditor op={op} keys={keys} replicas={replicas} onChange={(o) => setClient(i, { ...c, ops: c.ops.map((x, j) => (j === oi ? o : x)) })} />
                {c.ops.length > 1 && (
                  <button type="button" className="btn-ghost" onClick={() => setClient(i, { ...c, ops: c.ops.filter((_, j) => j !== oi) })} title="Удалить операцию">
                    ✕
                  </button>
                )}
              </li>
            ))}
          </ol>
          <button type="button" className="btn-ghost" onClick={() => setClient(i, { ...c, ops: [...c.ops, { kind: 'read', key: keys[0] ?? 'a', from: 'any' }] })}>
            + операция
          </button>
        </div>
      ))}
      {clients.length < 5 && (
        <button
          type="button"
          className="btn"
          onClick={() => onChange([...clients, { name: `C${clients.length + 1}`, at: 1, ops: [{ kind: 'write', key: keys[0] ?? 'a' }], repeat: 8 }])}
        >
          + клиент
        </button>
      )}
    </div>
  )
}

/* ─────────────────────────────── сбои ─────────────────────────────── */

export function FaultsEditor({ faults, replicas, onChange }: { faults: Fault[]; replicas: number; onChange: (f: Fault[]) => void }) {
  const set = (i: number, f: Fault) => onChange(faults.map((x, j) => (j === i ? f : x)))
  return (
    <div className="chan-editor">
      {faults.map((f, i) => (
        <div className="chan-edit" key={i}>
          <select
            value={f.kind}
            onChange={(e) => {
              const kind = e.target.value as Fault['kind']
              set(i, kind === 'primary-down' ? { kind, at: f.at } : kind === 'replica-down' ? { kind, at: f.at, replica: 1, until: f.at + 10 } : { kind, at: f.at, replica: 1, until: f.at + 10, rate: 1 })
            }}
          >
            <option value="primary-down">ведущий падает</option>
            <option value="replica-down">реплика недоступна</option>
            <option value="replica-slow">реплика тормозит</option>
          </select>
          {f.kind !== 'primary-down' && (
            <select value={f.replica} onChange={(e) => set(i, { ...f, replica: Number(e.target.value) })} aria-label="Реплика">
              {Array.from({ length: replicas }, (_, r) => (
                <option key={r} value={r + 1}>
                  r{r + 1}
                </option>
              ))}
            </select>
          )}
          <label>
            тик
            <input type="number" min={1} max={80} value={f.at} onChange={(e) => set(i, { ...f, at: Math.max(1, Number(e.target.value)) })} />
          </label>
          {f.kind !== 'primary-down' && (
            <label>
              до
              <input
                type="number"
                min={f.at + 1}
                max={100}
                value={f.until ?? f.at + 10}
                onChange={(e) => set(i, { ...f, until: Math.max(f.at + 1, Number(e.target.value)) } as Fault)}
              />
            </label>
          )}
          <button type="button" className="btn-ghost" onClick={() => onChange(faults.filter((_, j) => j !== i))} title="Убрать сбой">
            ✕
          </button>
        </div>
      ))}
      <button type="button" className="btn-ghost" onClick={() => onChange([...faults, { kind: 'replica-down', replica: 1, at: 10, until: 20 }])}>
        + сбой
      </button>
    </div>
  )
}
