import type { Frame, Mark, Op, Tuple, Txn, TxnEvent, TxnScenario, TxnWorld } from '../engine/types.ts'
import { effectiveOp, heapStats, invariantState } from '../engine/tick.ts'
import { BOOT_XID, horizon, isDead, isLive, labelOf, nowSnapshot, statusOf, txnByXid, viewOf, visibleTo } from '../engine/world.ts'
import { ISOLATION_LABEL } from '../explain/events.ts'

/**
 * Схема мира на одном тике. Чистая проекция: получает кадр, ничего не хранит.
 *
 * Сверху — расписание прогона, как в учебнике: кто что делал на каждом тике.
 * Под ним — сессии с их SQL, а ниже — сама таблица: все версии строк с xmin и
 * xmax и отметками, какой транзакции какая версия видна. Справа — то, от чего
 * зависит видимость и долговечность: pg_xact, WAL и очистка.
 */

const COLORS = 8
const color = (spec: number) => ({ '--c': `var(--wl-${spec % COLORS})` }) as React.CSSProperties

export interface Highlight {
  txn: Set<number>
  tuple: Set<number>
  key: Set<string>
}

export function highlightOf(e: TxnEvent | null): Highlight {
  return {
    txn: new Set(e?.actors.txn ?? []),
    tuple: new Set(e?.actors.tuple ?? []),
    key: new Set(e?.actors.key ?? []),
  }
}

/* ─────────────────────────────── SQL ─────────────────────────────── */

const q = (k: string) => `'${k}'`

interface SqlLine {
  code: string
  note?: string
}

/** Операция сценария в виде SQL, который её выполняет. */
export function sqlOf(op: Op, t: Txn | null): SqlLine {
  switch (op.kind) {
    case 'read':
      return { code: `SELECT v FROM t WHERE k = ${q(op.key)}${op.forUpdate ? ' FOR UPDATE' : ''}` }
    case 'scan':
      return { code: `SELECT count(*) FROM t${op.min === undefined ? '' : ` WHERE v >= ${op.min}`}` }
    case 'update': {
      const when = op.when ? `если ${op.when.sumOf.join(' + ')} ≥ ${op.when.gte}` : undefined
      if (op.set.kind === 'const') return { code: `UPDATE t SET v = ${op.set.value} WHERE k = ${q(op.key)}`, note: when }
      if (op.set.kind === 'delta') {
        return { code: `UPDATE t SET v = v ${op.set.d >= 0 ? '+' : '−'} ${Math.abs(op.set.d)} WHERE k = ${q(op.key)}`, note: when }
      }
      const read = t?.reads[op.key]
      const sign = op.set.d >= 0 ? '+' : '−'
      const calc =
        read === undefined || read === null
          ? `$1 = прочитанное ${sign} ${Math.abs(op.set.d)}, считает приложение`
          : `$1 = ${read} ${sign} ${Math.abs(op.set.d)} = ${read + op.set.d}, считает приложение`
      return { code: `UPDATE t SET v = $1 WHERE k = ${q(op.key)}`, note: when ? `${when}; ${calc}` : calc }
    }
    case 'insert':
      return { code: `INSERT INTO t VALUES (${q(op.key)}, ${op.value})` }
    case 'delete':
      return { code: `DELETE FROM t WHERE k = ${q(op.key)}` }
    case 'commit':
      return { code: 'COMMIT' }
    case 'rollback':
      return { code: 'ROLLBACK' }
  }
}

const ERROR_TEXT: Record<string, string> = {
  serialization: 'could not serialize access due to concurrent update',
  ssi: 'could not serialize access due to read/write dependencies among transactions',
  deadlock: 'deadlock detected',
  unique: 'duplicate key value violates unique constraint',
  crash: 'server closed the connection unexpectedly',
}

const STATE_LABEL: Record<Txn['state'], string> = {
  idle: 'ещё не начата',
  active: 'идёт',
  waiting: 'ждёт блокировку',
  committing: 'ждёт fsync',
  committed: 'COMMIT',
  aborted: 'прервана',
  done: 'закончила',
}

/* ───────────────────────────── расписание ───────────────────────────── */

const MARK_GLYPH: Record<string, string> = {
  read: 'R',
  scan: 'R',
  update: 'W',
  insert: 'W',
  delete: 'W',
  commit: 'C',
  rollback: 'A',
}

function markClass(m: Mark | null): string {
  if (!m) return 'is-empty'
  switch (m.kind) {
    case 'op':
      return `op-${m.op === 'read' || m.op === 'scan' ? 'read' : m.op === 'commit' ? 'commit' : m.op === 'rollback' ? 'abort' : 'write'}${m.bad ? ' is-bad' : ''}`
    case 'wait':
      return 'is-wait'
    case 'commit-wait':
      return 'is-fsync'
    case 'abort':
      return 'op-abort'
    case 'down':
      return 'is-down'
  }
}

function markTitle(m: Mark | null, label: string, tick: number): string {
  const head = `тик ${tick} · ${label}`
  if (!m) return `${head}: —`
  switch (m.kind) {
    case 'op':
      return `${head}: ${m.op.toUpperCase()}${m.key ? ` ${m.key}` : ''}${m.bad ? ' — аномалия' : ''}`
    case 'wait':
      return `${head}: ждёт блокировку строки`
    case 'commit-wait':
      return `${head}: COMMIT ждёт сброса WAL`
    case 'abort':
      return `${head}: ошибка, транзакция прервана`
    case 'down':
      return `${head}: сервер недоступен`
  }
}

function Schedule({ history, cursor, scenario, hl, onSeek }: {
  history: readonly Frame[]
  cursor: number
  scenario: TxnScenario
  hl: Highlight
  onSeek: (t: number) => void
}) {
  const last = history.length - 1
  const ticks = Array.from({ length: last }, (_, i) => i + 1)
  return (
    <div className="db-sched" role="group" aria-label="Расписание прогона">
      <div className="db-sched-grid" style={{ gridTemplateColumns: `4.5rem repeat(${Math.max(1, last)}, 16px)` }}>
        <span className="db-sched-corner">тик</span>
        {ticks.map((t) => (
          <button
            key={t}
            type="button"
            className={`db-sched-tick ${t === cursor ? 'is-now' : ''}`}
            onClick={() => onSeek(t)}
            aria-label={`Тик ${t}`}
          >
            {t % 5 === 0 || t === 1 ? t : ''}
          </button>
        ))}
        {scenario.sessions.map((s, si) => (
          <Row key={si} spec={si} name={s.name} history={history} ticks={ticks} cursor={cursor} hl={hl} onSeek={onSeek} />
        ))}
      </div>
      <p className="db-sched-legend" aria-hidden="true">
        <span className="db-cell op-read">R</span> чтение <span className="db-cell op-write">W</span> запись
        <span className="db-cell op-commit">C</span> коммит <span className="db-cell op-abort">A</span> откат или ошибка
        <span className="db-cell is-wait" /> ждёт строку <span className="db-cell is-fsync" /> ждёт fsync
        <span className="db-cell op-write is-bad">W</span> аномалия
      </p>
    </div>
  )
}

function Row({ spec, name, history, ticks, cursor, hl, onSeek }: {
  spec: number
  name: string
  history: readonly Frame[]
  ticks: number[]
  cursor: number
  hl: Highlight
  onSeek: (t: number) => void
}) {
  return (
    <>
      <span className={`db-sched-name ${hl.txn.has(spec) ? 'is-hl' : ''}`} style={color(spec)}>
        {name}
      </span>
      {ticks.map((t) => {
        const f = history[t]!
        const m = f.world.marks[spec] ?? null
        const txn = f.world.txns[spec]!
        const prev = history[t - 1]?.world.txns[spec]
        const fresh = prev !== undefined && prev.run !== txn.run
        return (
          <button
            key={t}
            type="button"
            className={`db-cell ${markClass(m)} ${t > cursor ? 'is-future' : ''} ${t === cursor ? 'is-now' : ''} ${fresh ? 'is-fresh' : ''}`}
            style={color(spec)}
            title={markTitle(m, labelOf(txn), t)}
            onClick={() => onSeek(t)}
          >
            {m?.kind === 'op' ? MARK_GLYPH[m.op] : m?.kind === 'abort' ? 'A' : ''}
          </button>
        )
      })}
    </>
  )
}

/* ─────────────────────────────── сессии ─────────────────────────────── */

function snapText(w: TxnWorld, t: Txn): string {
  const s = viewOf(w, t)
  if (!s) return '—'
  const xip = s.xip.length > 0 ? s.xip.join(',') : ''
  return `${s.xmin}:${s.xmax}:${xip}`
}

function SessionCard({ w, t, scenario, hl }: { w: TxnWorld; t: Txn; scenario: TxnScenario; hl: Highlight }) {
  const spec = scenario.sessions[t.spec]!
  const perStatement = t.isolation === 'read-committed' || t.isolation === 'read-uncommitted'
  const holder = t.wait ? w.txns.find((x) => `${x.spec}:${x.run}` === t.wait!.on) : undefined
  const reads = Object.entries(t.reads)
  const prev = [...w.history].reverse().find((x) => x.spec === t.spec)
  const idleInTx = isLive(t) && t.state === 'active' && t.opIdx >= spec.ops.length

  const lineState = (i: number): { cls: string; tag?: string } => {
    if (t.state === 'idle') return { cls: 'is-next', tag: i === 0 ? `тик ${t.nextAt}` : undefined }
    if (i < t.opIdx) return { cls: 'is-done' }
    if (i > t.opIdx) return { cls: t.state === 'aborted' ? 'is-skip' : 'is-next' }
    // Текущая строка.
    switch (t.state) {
      case 'waiting':
        return { cls: 'is-wait', tag: `ждёт ${holder ? labelOf(holder) : 'строку'}` }
      case 'committing':
        return { cls: 'is-wait', tag: 'ждёт fsync' }
      case 'committed':
        return { cls: 'is-done', tag: t.lost ? 'потерян' : undefined }
      case 'aborted':
        return t.error ? { cls: 'is-err', tag: 'ошибка' } : { cls: 'is-done' }
      default:
        return { cls: w.tick >= t.nextAt ? 'is-now' : 'is-next' }
    }
  }

  return (
    <section className={`db-sess state-${t.state} ${hl.txn.has(t.spec) ? 'is-hl' : ''}`} style={color(t.spec)}>
      <header>
        <strong>{labelOf(t)}</strong>
        <span className="db-sess-state">{idleInTx ? 'idle in transaction' : STATE_LABEL[t.state]}</span>
        <span className="db-sess-level" title="Уровень изоляции этой транзакции">
          {ISOLATION_LABEL[t.isolation]}
        </span>
      </header>
      <ol className="db-sql">
        <li className={t.state === 'idle' ? 'is-next' : 'is-done'}>
          <code>BEGIN{t.isolation === 'read-committed' ? '' : ` ISOLATION LEVEL ${ISOLATION_LABEL[t.isolation]}`}</code>
        </li>
        {spec.ops.map((raw, i) => {
          const op = effectiveOp(w.config.rmw, raw)
          const line = sqlOf(op, t)
          const st = lineState(i)
          return (
            <li key={i} className={st.cls}>
              <code>{line.code}</code>
              {st.tag && <b className="db-sql-tag">{st.tag}</b>}
              {line.note && <small>{line.note}</small>}
            </li>
          )
        })}
        {!spec.ops.some((o) => o.kind === 'commit' || o.kind === 'rollback') && (
          <li className="is-skip">
            <code>-- COMMIT так и не приходит</code>
          </li>
        )}
      </ol>
      <footer className="db-sess-foot">
        <span title="Номер транзакции. PostgreSQL выдаёт его только при первой записи">
          xid <b>{t.xid ?? '—'}</b>
        </span>
        <span
          title={
            perStatement
              ? 'READ COMMITTED берёт снимок на каждый оператор — здесь тот, что взял бы следующий'
              : 'Снимок на всю транзакцию: xmin:xmax:идущие'
          }
        >
          снимок <b>{snapText(w, t)}</b>
          {perStatement && isLive(t) && <i> на оператор</i>}
        </span>
        {reads.length > 0 && (
          <span title="Что приложение прочитало и помнит">
            прочитано{' '}
            {reads.map(([k, v]) => (
              <b key={k}>
                {k}={v ?? 'NULL'}{' '}
              </b>
            ))}
          </span>
        )}
      </footer>
      {t.state === 'aborted' && t.error && (
        <p className="db-sess-err">
          ERROR: {ERROR_TEXT[t.error] ?? t.error}
          {t.restartAt !== null && t.retryNext && <i> — клиент повторит на тике {t.restartAt}</i>}
        </p>
      )}
      {t.lost && <p className="db-sess-err">Клиент получил «COMMIT», но запись о коммите не пережила падения.</p>}
      {t.retries > 0 && prev && <p className="db-sess-prev">повтор после ошибки {labelOf(prev)}</p>}
    </section>
  )
}

/* ─────────────────────────────── куча ─────────────────────────────── */

function XidCell({ w, xid, lockOnly }: { w: TxnWorld; xid: number | null; lockOnly?: boolean }) {
  if (xid === null) return <span className="db-xid is-none">—</span>
  const st = statusOf(w, xid)
  const t = xid === BOOT_XID ? undefined : txnByXid(w, xid)
  return (
    <span
      className={`db-xid st-${st}`}
      style={t ? color(t.spec) : undefined}
      title={`xid ${xid}${t ? ` — ${labelOf(t)}` : ' — исходные данные'}: ${st === 'committed' ? 'закоммичена' : st === 'aborted' ? 'откачена' : 'идёт'}${lockOnly ? '\nтолько блокировка (FOR UPDATE), версия жива' : ''}`}
    >
      {xid}
      {t && <small>{labelOf(t)}</small>}
      {lockOnly && <em>lock</em>}
    </span>
  )
}

function tupleStatus(w: TxnWorld, x: Tuple, hz: number): { cls: string; text: string } {
  if (statusOf(w, x.xmin) === 'aborted') return { cls: 'is-dead', text: 'мёртвая: откат' }
  if (statusOf(w, x.xmin) === 'in-progress') return { cls: 'is-new', text: 'новая, не закоммичена' }
  if (x.xmax !== null && statusOf(w, x.xmax) === 'in-progress') {
    return x.lockOnly ? { cls: 'is-live is-locked', text: 'живая, заблокирована' } : { cls: 'is-live is-locked', text: 'живая, заменяется' }
  }
  if (isDead(w, x)) return x.xmax! < hz ? { cls: 'is-dead', text: 'мёртвая, можно убрать' } : { cls: 'is-dead is-kept', text: 'мёртвая, но кому-то видна' }
  return { cls: 'is-live', text: 'живая' }
}

function HeapTable({ w, hl }: { w: TxnWorld; hl: Highlight }) {
  const hz = horizon(w).xid
  const viewers = w.txns.filter(isLive)
  const views = viewers.map((t) => ({ t, snap: viewOf(w, t)! }))
  const now = nowSnapshot(w)
  const pages = w.tuples.reduce((m, t) => Math.max(m, t.page + 1), 0)
  const bySlot = new Map(w.tuples.map((t) => [t.page * 1000 + t.slot, t]))

  const rows: React.ReactNode[] = []
  for (let p = 0; p < pages; p++) {
    rows.push(
      <tr key={`p${p}`} className="db-page-row">
        <td colSpan={6 + views.length + 1}>страница {p}</td>
      </tr>,
    )
    for (let s = 0; s < w.config.pageSlots; s++) {
      const x = bySlot.get(p * 1000 + s)
      if (!x) {
        rows.push(
          <tr key={`e${p}.${s}`} className="db-free">
            <td className="db-ctid">({p},{s})</td>
            <td colSpan={5 + views.length + 1}>свободно</td>
          </tr>,
        )
        continue
      }
      const st = tupleStatus(w, x, hz)
      rows.push(
        <tr key={x.id} className={`db-tuple ${st.cls} ${hl.tuple.has(x.id) ? 'is-hl' : hl.key.has(x.key) ? 'is-soft' : ''}`}>
          <td className="db-ctid">({x.page},{x.slot})</td>
          <td className="db-k">{x.key}</td>
          <td className="db-v">{x.value}</td>
          <td>
            <XidCell w={w} xid={x.xmin} />
          </td>
          <td>
            <XidCell w={w} xid={x.xmax} lockOnly={x.lockOnly} />
          </td>
          <td className="db-status">{st.text}</td>
          {views.map(({ t, snap }) => {
            const vis = visibleTo(w, x, snap, t.xid, t.isolation === 'read-uncommitted')
            return (
              <td key={t.spec} className={`db-vis ${vis ? 'is-yes' : ''}`} style={color(t.spec)} title={`${labelOf(t)} ${vis ? 'видит' : 'не видит'} эту версию`}>
                {vis ? '●' : '·'}
              </td>
            )
          })}
          <td className={`db-vis db-vis-now ${visibleTo(w, x, now, null) ? 'is-yes' : ''}`} title="Что увидел бы новый запрос прямо сейчас">
            {visibleTo(w, x, now, null) ? '●' : '·'}
          </td>
        </tr>,
      )
    }
  }

  return (
    <div className="db-heap">
      <table>
        <thead>
          <tr>
            <th title="Адрес версии: страница и слот">ctid</th>
            <th>k</th>
            <th>v</th>
            <th title="Кто создал версию">xmin</th>
            <th title="Кто удалил или заменил версию — или держит строку">xmax</th>
            <th>состояние</th>
            {views.map(({ t }) => (
              <th key={t.spec} className={`db-vis-h ${hl.txn.has(t.spec) ? 'is-hl' : ''}`} style={color(t.spec)} title={`Видна ли версия транзакции ${labelOf(t)}`}>
                {labelOf(t)}
              </th>
            ))}
            <th className="db-vis-h" title="Что увидел бы новый запрос прямо сейчас">
              новый
            </th>
          </tr>
        </thead>
        <tbody>{rows}</tbody>
      </table>
    </div>
  )
}

/* ─────────────────────────── боковая панель ─────────────────────────── */

function Side({ w, scenario, hl }: { w: TxnWorld; scenario: TxnScenario; hl: Highlight }) {
  const inv = invariantState(w, scenario.invariant)
  const hs = heapStats(w)
  const hz = horizon(w)
  const xids = Object.keys(w.xact)
    .map(Number)
    .filter((x) => x !== BOOT_XID)
    .slice(-18)
  const wal = w.wal.slice(-10)
  const down = w.tick < w.downUntil

  return (
    <aside className="db-side">
      {down && (
        <div className="db-down" role="status">
          Сервер лежит: восстановление по WAL, ещё {w.downUntil - w.tick} т.
        </div>
      )}
      {inv && (
        <div className={`db-box db-inv ${inv.ok ? 'is-ok' : 'is-bad'}`}>
          <span className="board-label">правило</span>
          <b>{scenario.invariant!.label}</b>
          <span>{inv.ok ? '✓ соблюдено' : '✕ нарушено'} · {inv.detail}</span>
        </div>
      )}
      <div className="db-box">
        <span className="board-label">куча и очистка</span>
        <span>
          версий <b>{hs.total}</b>, мёртвых <b className={hs.dead > 0 ? 'db-warn' : undefined}>{hs.dead}</b>
          {hs.dead > 0 && <> (убрать можно {hs.removable})</>}, страниц <b>{hs.pages}</b>
        </span>
        <span title="Мёртвые версии моложе горизонта VACUUM убрать не может">
          горизонт <b>{hz.xid}</b>
          {hz.holder && (
            <>
              {' '}— держит{' '}
              <b className={hl.txn.has(hz.holder.spec) ? 'db-hl-text' : undefined} style={color(hz.holder.spec)}>
                {labelOf(hz.holder)}
              </b>
            </>
          )}
        </span>
        {w.config.autovacuum > 0 ? (
          <span className="kf-muted">автовакуум каждые {w.config.autovacuum} т., прошёл {w.stats.vacuumRuns} раз, убрал {w.stats.vacuumRemoved}</span>
        ) : (
          <span className="kf-muted">автовакуум выключен</span>
        )}
      </div>
      <div className="db-box">
        <span className="board-label" title="Статусы транзакций: по два бита на xid">pg_xact</span>
        <span className="db-xact">
          {xids.length === 0 && <i className="kf-empty">ни одной транзакции с xid</i>}
          {xids.map((x) => {
            const t = txnByXid(w, x)
            return (
              <span
                key={x}
                className={`db-xid st-${statusOf(w, x)} ${t && hl.txn.has(t.spec) ? 'is-hl' : ''}`}
                style={t ? color(t.spec) : undefined}
                title={`${x}${t ? ` — ${labelOf(t)}` : ''}: ${statusOf(w, x)}`}
              >
                {x}
                {t && <small>{labelOf(t)}</small>}
              </span>
            )
          })}
        </span>
      </div>
      <div className="db-box">
        <span className="board-label" title="Журнал упреждающей записи: сплошные записи уже на диске">
          WAL · на диске до {w.flushedLsn}
          {w.flushing && ` · fsync до ${w.flushing.upTo}…`}
        </span>
        <ol className="db-wal">
          {wal.length === 0 && <li className="kf-empty">пусто</li>}
          {wal.map((r) => {
            const t = r.xid === null ? undefined : txnByXid(w, r.xid)
            return (
              <li key={r.lsn} className={`${r.lsn <= w.flushedLsn ? 'is-flushed' : ''} ${r.kind === 'commit' ? 'is-commit' : ''} ${r.kind === 'checkpoint' ? 'is-ckpt' : ''}`}>
                <span className="db-lsn">{r.lsn}</span>
                <span>{r.kind}</span>
                {r.key && <span className="db-k">{r.key}</span>}
                {t && (
                  <small style={color(t.spec)}>{labelOf(t)}</small>
                )}
              </li>
            )
          })}
        </ol>
      </div>
    </aside>
  )
}

/* ─────────────────────────────── целиком ─────────────────────────────── */

export function Board({ history, cursor, scenario, highlight: hl, onSeek }: {
  history: readonly Frame[]
  cursor: number
  scenario: TxnScenario
  highlight: Highlight
  onSeek: (t: number) => void
}) {
  const w = history[cursor]!.world
  return (
    <div className="board db-board">
      <Schedule history={history} cursor={cursor} scenario={scenario} hl={hl} onSeek={onSeek} />
      <div className="db-sessions">
        {w.txns.map((t) => (
          <SessionCard key={t.spec} w={w} t={t} scenario={scenario} hl={hl} />
        ))}
      </div>
      <div className="db-main">
        <HeapTable w={w} hl={hl} />
        <Side w={w} scenario={scenario} hl={hl} />
      </div>
    </div>
  )
}
