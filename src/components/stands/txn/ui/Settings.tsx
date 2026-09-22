import type { Isolation, Op, RmwStyle, RowSpec, TxnConfig, TxnScenario, TxnSpec } from '../engine/types.ts'
import { ISOLATIONS } from '../engine/types.ts'
import { ISOLATION_LABEL } from '../explain/events.ts'

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
          <button
            key={o.value}
            type="button"
            role="radio"
            aria-checked={o.value === value}
            className={o.value === value ? 'is-now' : ''}
            onClick={() => onChange(o.value)}
          >
            {o.label}
          </button>
        ))}
      </span>
    </div>
  )
}

const SHORT_LEVEL: Record<Isolation, string> = {
  'read-uncommitted': 'RU',
  'read-committed': 'RC',
  'repeatable-read': 'RR',
  serializable: 'SER',
}

const t = (n: number) => `${n} т.`
const off = (n: number, text = 'выкл') => (n === 0 ? text : t(n))

/** Есть ли в сценарии «прочитать — изменить — записать» — иначе ручка стиля ничего не меняет. */
export const hasRmw = (sc: TxnScenario) =>
  sc.sessions.some((s) => s.ops.some((o) => (o.kind === 'read' && o.rmw) || (o.kind === 'update' && o.set.kind === 'read')))

export function ConfigKnobs({ config: c, scenario, onChange }: {
  config: TxnConfig
  scenario: TxnScenario
  onChange: (c: TxnConfig) => void
}) {
  const set = <K extends keyof TxnConfig>(k: K, v: TxnConfig[K]) => onChange({ ...c, [k]: v })
  const overridden = scenario.sessions.filter((s) => s.isolation)
  return (
    <>
      <h4 className="settings-sub">Изоляция и блокировки</h4>
      <div className="knobs">
        <Choice<Isolation>
          label="Уровень изоляции"
          hint="default_transaction_isolation: RU — READ UNCOMMITTED (учебный, в PostgreSQL его нет), RC — READ COMMITTED, RR — REPEATABLE READ, SER — SERIALIZABLE"
          value={c.isolation}
          options={ISOLATIONS.map((v) => ({ value: v, label: SHORT_LEVEL[v] }))}
          onChange={(v) => set('isolation', v)}
        />
        {hasRmw(scenario) && (
          <Choice<RmwStyle>
            label="Прочитать — изменить — записать"
            hint="Как приложение прибавляет к балансу: посчитать у себя и записать число, заблокировать строку при чтении или отдать расчёт базе"
            value={c.rmw}
            options={[
              { value: 'app', label: 'SELECT, потом UPDATE' },
              { value: 'for-update', label: 'FOR UPDATE' },
              { value: 'atomic', label: 'SET v = v + d' },
            ]}
            onChange={(v) => set('rmw', v)}
          />
        )}
        <Toggle
          label="Повторять при ошибке"
          hint="Клиент повторяет транзакцию целиком, если она упала с ошибкой сериализации (40001), дедлоком (40P01) или из-за падения сервера"
          checked={c.retry}
          onChange={(v) => set('retry', v)}
        />
        <Range
          label="deadlock_timeout"
          hint="Сколько тиков ждать блокировку, прежде чем искать цикл. В PostgreSQL — 1 секунда"
          value={c.deadlockTimeout}
          min={1}
          max={10}
          display={t(c.deadlockTimeout)}
          onChange={(v) => set('deadlockTimeout', v)}
        />
      </div>
      {overridden.length > 0 && (
        <p className="settings-note">
          Уровень задан в самом сценарии для {overridden.map((s) => `${s.name} (${ISOLATION_LABEL[s.isolation!]})`).join(', ')} — общая
          настройка на {overridden.length === 1 ? 'неё' : 'них'} не действует.
        </p>
      )}

      <h4 className="settings-sub">Очистка</h4>
      <div className="knobs">
        <Range
          label="Автовакуум"
          hint="Как часто приходит VACUUM. В PostgreSQL его запускает число изменённых строк, а не таймер"
          value={c.autovacuum}
          min={0}
          max={20}
          display={off(c.autovacuum)}
          onChange={(v) => set('autovacuum', v)}
        />
        <Range
          label="Слотов в странице"
          hint="Сколько версий помещается в одну страницу кучи. В PostgreSQL страница — 8 КБ, и строк в ней обычно десятки"
          value={c.pageSlots}
          min={2}
          max={8}
          onChange={(v) => set('pageSlots', v)}
        />
      </div>

      <h4 className="settings-sub">WAL и долговечность</h4>
      <div className="knobs">
        <Toggle
          label="synchronous_commit"
          hint="Включено — COMMIT ждёт, пока запись о нём окажется на диске. Выключено — ответ сразу, а WAL сбросит фоновый процесс"
          checked={c.syncCommit}
          onChange={(v) => set('syncCommit', v)}
        />
        <Range
          label="Длительность fsync"
          hint="Сколько тиков занимает сброс WAL на диск. Все коммиты, подошедшие за это время, уедут следующим сбросом вместе"
          value={c.fsyncTicks}
          min={1}
          max={6}
          display={t(c.fsyncTicks)}
          onChange={(v) => set('fsyncTicks', v)}
        />
        <Range
          label="wal_writer_delay"
          hint="Как часто фоновый walwriter сбрасывает WAL сам, без ждущих коммитов"
          value={c.walWriterDelay}
          min={1}
          max={12}
          display={t(c.walWriterDelay)}
          onChange={(v) => set('walWriterDelay', v)}
        />
        <Range
          label="Контрольная точка"
          hint="Как часто все изменённые страницы пишутся на диск. После неё восстановлению не нужен WAL до этой точки"
          value={c.checkpointEvery}
          min={0}
          max={20}
          display={off(c.checkpointEvery, 'только в начале')}
          onChange={(v) => set('checkpointEvery', v)}
        />
        <Range
          label="Падение сервера"
          hint="На каком тике сервер падает. Выживает только то, что успело попасть в WAL на диске"
          value={c.crashAt}
          min={0}
          max={60}
          display={off(c.crashAt, 'не падает')}
          onChange={(v) => set('crashAt', v)}
        />
      </div>
    </>
  )
}

/* ─────────────────────────────── таблица ─────────────────────────────── */

export function RowsEditor({ rows, onChange }: { rows: RowSpec[]; onChange: (r: RowSpec[]) => void }) {
  const set = (i: number, next: RowSpec) => onChange(rows.map((x, j) => (j === i ? next : x)))
  return (
    <div className="chan-editor">
      {rows.map((r, i) => (
        <div className="chan-edit" key={i}>
          <input className="wl-name" value={r.key} onChange={(e) => set(i, { ...r, key: e.target.value })} aria-label="Ключ строки" />
          <label>
            v
            <input type="number" value={r.value} onChange={(e) => set(i, { ...r, value: Number(e.target.value) })} />
          </label>
          {rows.length > 1 && (
            <button type="button" className="btn-ghost" onClick={() => onChange(rows.filter((_, j) => j !== i))} title="Удалить строку">
              ✕
            </button>
          )}
        </div>
      ))}
      <button type="button" className="btn-ghost" onClick={() => onChange([...rows, { key: `k${rows.length + 1}`, value: 0 }])}>
        + строка
      </button>
    </div>
  )
}

/* ─────────────────────────────── сессии ─────────────────────────────── */

const OP_LABEL: Record<Op['kind'], string> = {
  read: 'SELECT',
  scan: 'SELECT count(*)',
  update: 'UPDATE',
  insert: 'INSERT',
  delete: 'DELETE',
  commit: 'COMMIT',
  rollback: 'ROLLBACK',
}

function defaultOp(kind: Op['kind'], key: string): Op {
  switch (kind) {
    case 'read':
      return { kind, key, rmw: true }
    case 'scan':
      return { kind, min: 100 }
    case 'update':
      return { kind, key, set: { kind: 'read', d: 10 } }
    case 'insert':
      return { kind, key: `${key}2`, value: 100 }
    case 'delete':
      return { kind, key }
    case 'commit':
      return { kind }
    case 'rollback':
      return { kind }
  }
}

function KeyPicker({ keys, value, onChange }: { keys: string[]; value: string; onChange: (v: string) => void }) {
  return (
    <select value={value} onChange={(e) => onChange(e.target.value)} aria-label="Строка">
      {[...new Set([...keys, value])].map((k) => (
        <option key={k} value={k}>
          {k}
        </option>
      ))}
    </select>
  )
}

function OpFields({ op, keys, onChange }: { op: Op; keys: string[]; onChange: (o: Op) => void }) {
  switch (op.kind) {
    case 'read':
      return (
        <span className="phase-ticks">
          <KeyPicker keys={keys} value={op.key} onChange={(key) => onChange({ ...op, key })} />
          <label title="SELECT … FOR UPDATE — заблокировать строку при чтении">
            <input type="checkbox" checked={op.forUpdate === true} onChange={(e) => onChange({ ...op, forUpdate: e.target.checked })} />
            FOR UPDATE
          </label>
        </span>
      )
    case 'scan':
      return (
        <span className="phase-ticks">
          v ≥
          <input type="number" value={op.min ?? 0} onChange={(e) => onChange({ ...op, min: Number(e.target.value) })} />
        </span>
      )
    case 'update':
      return (
        <span className="phase-ticks">
          <KeyPicker keys={keys} value={op.key} onChange={(key) => onChange({ ...op, key })} />
          <select
            value={op.set.kind}
            onChange={(e) => {
              const kind = e.target.value as 'const' | 'delta' | 'read'
              onChange({ ...op, set: kind === 'const' ? { kind, value: 0 } : { kind, d: 10 } })
            }}
            title="Как считается новое значение"
          >
            <option value="read">v = прочитанное + d</option>
            <option value="delta">v = v + d</option>
            <option value="const">v = число</option>
          </select>
          <input
            type="number"
            value={op.set.kind === 'const' ? op.set.value : op.set.d}
            onChange={(e) => {
              const n = Number(e.target.value)
              onChange({ ...op, set: op.set.kind === 'const' ? { kind: 'const', value: n } : { ...op.set, d: n } })
            }}
          />
        </span>
      )
    case 'insert':
      return (
        <span className="phase-ticks">
          <input className="wl-name" value={op.key} onChange={(e) => onChange({ ...op, key: e.target.value })} aria-label="Ключ" />
          <input type="number" value={op.value} onChange={(e) => onChange({ ...op, value: Number(e.target.value) })} />
        </span>
      )
    case 'delete':
      return (
        <span className="phase-ticks">
          <KeyPicker keys={keys} value={op.key} onChange={(key) => onChange({ ...op, key })} />
        </span>
      )
    default:
      return null
  }
}

function SessionEditor({ s, keys, onChange, onRemove }: {
  s: TxnSpec
  keys: string[]
  onChange: (s: TxnSpec) => void
  onRemove: () => void
}) {
  const setOp = (i: number, o: Op) => onChange({ ...s, ops: s.ops.map((x, j) => (j === i ? o : x)) })
  const first = keys[0] ?? 'k1'
  return (
    <div className="wl">
      <div className="wl-head">
        <input className="wl-name" value={s.name} onChange={(e) => onChange({ ...s, name: e.target.value })} aria-label="Имя сессии" />
        <select
          value={s.isolation ?? ''}
          onChange={(e) => onChange({ ...s, isolation: e.target.value === '' ? null : (e.target.value as Isolation) })}
          title="Уровень изоляции этой сессии"
        >
          <option value="">общий уровень</option>
          {ISOLATIONS.map((v) => (
            <option key={v} value={v}>
              {ISOLATION_LABEL[v]}
            </option>
          ))}
        </select>
        <label title="На каком тике выполнится первая операция">
          с тика
          <input type="number" min={1} max={60} value={s.at} onChange={(e) => onChange({ ...s, at: Math.max(1, Number(e.target.value)) })} />
        </label>
        <label title="Сколько раз повторить транзакцию целиком">
          ×
          <input type="number" min={1} max={20} value={s.repeat ?? 1} onChange={(e) => onChange({ ...s, repeat: Math.max(1, Math.min(20, Number(e.target.value))) })} />
        </label>
        <button type="button" className="btn-ghost" onClick={onRemove} title="Удалить сессию">
          ✕
        </button>
      </div>
      <ol className="wl-phases">
        {s.ops.map((op, i) => (
          <li key={i}>
            <label title="Через сколько тиков после предыдущей операции">
              +
              <input
                type="number"
                min={1}
                max={40}
                value={op.after ?? 1}
                onChange={(e) => setOp(i, { ...op, after: Math.max(1, Number(e.target.value)) })}
              />
            </label>
            <select value={op.kind} onChange={(e) => setOp(i, { ...defaultOp(e.target.value as Op['kind'], first), after: op.after })}>
              {Object.entries(OP_LABEL).map(([k, label]) => (
                <option key={k} value={k}>
                  {label}
                </option>
              ))}
            </select>
            <OpFields op={op} keys={keys} onChange={(o) => setOp(i, o)} />
            {s.ops.length > 1 && (
              <button type="button" className="btn-ghost" onClick={() => onChange({ ...s, ops: s.ops.filter((_, j) => j !== i) })} title="Удалить операцию">
                ✕
              </button>
            )}
          </li>
        ))}
      </ol>
      <button type="button" className="btn-ghost" onClick={() => onChange({ ...s, ops: [...s.ops, defaultOp('read', first)] })}>
        + операция
      </button>
    </div>
  )
}

export function SessionsEditor({ sessions, keys, onChange }: {
  sessions: TxnSpec[]
  keys: string[]
  onChange: (s: TxnSpec[]) => void
}) {
  const first = keys[0] ?? 'k1'
  return (
    <div className="wls">
      {sessions.map((s, i) => (
        <SessionEditor
          key={i}
          s={s}
          keys={keys}
          onChange={(next) => onChange(sessions.map((x, j) => (j === i ? next : x)))}
          onRemove={() => onChange(sessions.filter((_, j) => j !== i))}
        />
      ))}
      {sessions.length < 6 && (
        <button
          type="button"
          className="btn"
          onClick={() =>
            onChange([
              ...sessions,
              { name: `T${sessions.length + 1}`, at: 1, ops: [defaultOp('read', first), defaultOp('update', first), defaultOp('commit', first)] },
            ])
          }
        >
          + сессия
        </button>
      )}
    </div>
  )
}
