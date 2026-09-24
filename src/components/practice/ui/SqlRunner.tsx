/**
 * Остров SQL-задачи: редактор запроса, результат таблицей и вердикт.
 *
 * Postgres живёт прямо в браузере (PGlite в воркере, см. sql/session.ts), так
 * что запуск и проверка обходятся без сервера. Исключение — «найди баг»: номер
 * строки сверяет /api/check, как и у задач на Go.
 *
 * Эталон для «почини» и «реализуй» — результат solution.sql на тех же данных.
 * Секрета из него не сделать: решение и так открывается кнопкой, как у Go.
 * Для «что вернёт» эталон — результат нетронутой заготовки, а не того, что
 * сейчас в редакторе: иначе ответ можно было бы подогнать.
 */

import { useCallback, useEffect, useRef, useState } from 'react'
import { Text } from '@codemirror/state'
import type { TaskKindId } from '../../../data/practice.ts'
import { checkTask } from '../client.ts'
import type { Verdict } from '../protocol.ts'
import { compareTables, formatRows, parseAnswer, type SqlOutcome, type SqlTable } from '../sql/result.ts'
import { SqlSession } from '../sql/session.ts'
import Editor, { type EditorHandle } from './Editor.tsx'
import { sqlDiagnostic } from './editor.ts'

interface Props {
  taskId: string
  kind: TaskKindId
  prompt: string
  starter: string
  solution: string
  /** DDL и данные задачи — заливаются в базу перед первым запросом. */
  schema: string
  /** Важен ли порядок строк при сверке. */
  ordered: boolean
}

function storageKey(taskId: string, what: 'code' | 'solved') {
  return `go-piece:practice:${taskId}:${what}`
}

function load(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

function save(key: string, value: string): void {
  try {
    localStorage.setItem(key, value)
  } catch {
    // Приватный режим — черновик просто не переживёт перезагрузку.
  }
}

export default function SqlRunner({ taskId, kind, prompt, starter, solution, schema, ordered }: Props) {
  const picksLine = kind === 'bug'
  const editsCode = kind === 'fix' || kind === 'implement'

  const [code, setCode] = useState(starter)
  const [answer, setAnswer] = useState('')
  const [pickedLine, setPickedLine] = useState(0)
  const [unlocked, setUnlocked] = useState(editsCode)

  const [outcome, setOutcome] = useState<SqlOutcome | null>(null)
  const [verdict, setVerdict] = useState<Verdict | null>(null)
  const [failure, setFailure] = useState<string>()
  const [busy, setBusy] = useState<null | 'run' | 'check'>(null)
  const [solved, setSolved] = useState(false)

  const editor = useRef<EditorHandle>(null)
  const session = useRef<SqlSession | null>(null)
  /** Результат эталона не меняется за время жизни страницы — считаем его один раз. */
  const reference = useRef<SqlOutcome | null>(null)

  useEffect(() => {
    const s = new SqlSession(schema)
    session.current = s
    // Поднимаем базу сразу: пока человек читает условие, WASM успеет доехать.
    s.warmUp()
    return () => s.dispose()
  }, [schema])

  useEffect(() => {
    const draft = load(storageKey(taskId, 'code'))
    if (draft !== null && draft !== '' && draft !== starter) {
      setCode(draft)
      if (editsCode) setUnlocked(true)
    }
    if (load(storageKey(taskId, 'solved')) === '1') setSolved(true)
  }, [taskId, starter, editsCode])

  useEffect(() => {
    if (code === starter) return
    save(storageKey(taskId, 'code'), code)
  }, [code, starter, taskId])

  const dirty = code !== starter

  const markSolved = useCallback(() => {
    setSolved(true)
    save(storageKey(taskId, 'solved'), '1')
    document.querySelector<HTMLDetailsElement>('.reveal')?.setAttribute('open', '')
  }, [taskId])

  const showError = useCallback(
    (result: SqlOutcome) => {
      const doc = Text.of(code.split('\n'))
      editor.current?.setDiagnostics(result.ok ? [] : sqlDiagnostic(result.error, result.position, doc))
    },
    [code],
  )

  const execute = useCallback(async (sql: string) => {
    if (!session.current) throw new Error('база ещё не готова')
    return session.current.run(sql)
  }, [])

  const doRun = useCallback(async () => {
    setBusy('run')
    setFailure(undefined)
    setVerdict(null)
    try {
      const result = await execute(code)
      setOutcome(result)
      showError(result)
    } catch (err) {
      setFailure(err instanceof Error ? err.message : String(err))
      setOutcome(null)
    } finally {
      setBusy(null)
    }
  }, [code, execute, showError])

  const referenceTable = useCallback(
    async (sql: string): Promise<SqlTable> => {
      if (!reference.current) reference.current = await execute(sql)
      const ref = reference.current
      if (!ref.ok || !ref.table) {
        reference.current = null
        throw new Error(`эталон задачи не выполняется — это баг сайта: ${ref.ok ? 'нет строк' : ref.error}`)
      }
      return ref.table
    },
    [execute],
  )

  const doCheck = useCallback(async () => {
    setBusy('check')
    setFailure(undefined)
    try {
      let result: Verdict

      if (kind === 'bug') {
        result = (await checkTask({ task: taskId, kind, line: pickedLine })).verdict
      } else if (kind === 'output') {
        const expected = await referenceTable(starter)
        setOutcome(reference.current)
        const given = parseAnswer(answer, expected.columns)
        const cmp = compareTables(expected, { columns: expected.columns, rows: given }, ordered)
        const widthOk = given.every((r) => r.length === expected.columns.length)
        const pass = cmp.pass && widthOk
        result = {
          pass,
          title: pass ? 'Верно' : 'Не сходится',
          detail: pass
            ? undefined
            : `вы написали:\n${formatRows(given)}\n\nзапрос вернул:\n${formatRows(expected.rows)}` +
              (widthOk ? '' : `\n\nв каждой строке должно быть ячеек: ${expected.columns.length}, через «|»`),
        }
      } else {
        const expected = await referenceTable(solution)
        const got = await execute(code)
        setOutcome(got)
        showError(got)
        if (!got.ok) {
          result = { pass: false, title: 'Запрос падает', detail: got.error }
        } else if (!got.table) {
          result = { pass: false, title: 'Запрос не вернул строк', detail: 'последней командой должен быть SELECT (или команда с RETURNING)' }
        } else {
          const cmp = compareTables(expected, got.table, ordered)
          result = { pass: cmp.pass, title: cmp.pass ? 'Готово' : 'Результат не тот', detail: cmp.detail }
        }
      }

      setVerdict(result)
      if (result.pass) markSolved()
    } catch (err) {
      setFailure(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(null)
    }
  }, [answer, code, execute, kind, markSolved, ordered, pickedLine, referenceTable, showError, solution, starter, taskId])

  const reset = useCallback(() => {
    setCode(starter)
    setPickedLine(0)
    setVerdict(null)
    setOutcome(null)
    setFailure(undefined)
    editor.current?.setDiagnostics([])
    save(storageKey(taskId, 'code'), '')
  }, [starter, taskId])

  const revealSolution = useCallback(() => {
    setCode(solution)
    setUnlocked(true)
    setVerdict(null)
    editor.current?.setDiagnostics([])
  }, [solution])

  const blocker =
    kind === 'output' && answer.trim() === ''
      ? 'напишите, что вернёт запрос'
      : picksLine && pickedLine === 0
        ? 'отметьте строку слева от номеров'
        : picksLine && dirty
          ? 'запрос изменён, номера строк уехали — нажмите «Сбросить»'
          : null

  return (
    <div className="pr-runner not-prose">
      <div className="pr-toolbar">
        <button type="button" className="pr-btn pr-btn-main" onClick={doRun} disabled={busy !== null}>
          {busy === 'run' ? 'Выполняю…' : 'Выполнить'}
        </button>

        <button type="button" className="pr-btn" onClick={doCheck} disabled={busy !== null || blocker !== null} title={blocker ?? undefined}>
          {busy === 'check' ? 'Проверяю…' : 'Проверить'}
        </button>

        <span className="pr-toolbar-gap" />

        {!unlocked && (
          <button type="button" className="pr-btn pr-btn-ghost" onClick={() => setUnlocked(true)}>
            Открыть песочницу
          </button>
        )}

        {editsCode && (
          <button type="button" className="pr-btn pr-btn-ghost" onClick={revealSolution}>
            Решение
          </button>
        )}

        <button type="button" className="pr-btn pr-btn-ghost" onClick={reset} disabled={!dirty && pickedLine === 0}>
          Сбросить
        </button>
      </div>

      {solved && <div className="pr-solved">Задача засчитана. Разбор ниже открыт.</div>}

      <div className="pr-split">
        <div className="pr-pane">
          <div className="pr-pane-head">
            <span>query.sql</span>
            {!unlocked && <span className="pr-lock">только чтение</span>}
            {picksLine && <span className="pr-hint">кликните в колонку слева от номера строки</span>}
          </div>
          <Editor
            ref={editor}
            code={code}
            onChange={setCode}
            onRun={doRun}
            onFormat={() => undefined}
            onPick={picksLine ? setPickedLine : undefined}
            pickedLine={picksLine ? pickedLine : undefined}
            readOnly={!unlocked}
            language="sql"
          />
          <div className="pr-keys">
            <kbd>⌘</kbd>+<kbd>↵</kbd> выполнить{unlocked && <> · <kbd>Tab</kbd> отступ</>}
          </div>

          <details className="sql-schema">
            <summary>Схема и данные</summary>
            <pre>{schema.trim()}</pre>
          </details>
        </div>

        <div className="pr-pane">
          {kind === 'output' && (
            <div className="pr-answer">
              <label className="pr-answer-label" htmlFor="pr-answer">
                {prompt}
              </label>
              <textarea
                id="pr-answer"
                className="pr-answer-input"
                value={answer}
                onChange={(e) => setAnswer(e.target.value)}
                rows={5}
                spellCheck={false}
                placeholder={'строка результата на строку,\nячейки через |, например:\nАня | 3\nБоря | NULL\n\nнет строк — «пусто»'}
              />
            </div>
          )}

          {picksLine && (
            <div className="pr-answer">
              <div className="pr-answer-label">{prompt}</div>
              <div className="pr-picked">{pickedLine > 0 ? `Отмечена строка ${pickedLine}` : 'Строка не отмечена'}</div>
            </div>
          )}

          {verdict && (
            <div className={verdict.pass ? 'pr-verdict pr-verdict-pass' : 'pr-verdict pr-verdict-fail'}>
              <div className="pr-verdict-title">{verdict.title}</div>
              {verdict.detail && <pre className="pr-verdict-detail">{verdict.detail}</pre>}
            </div>
          )}

          <ResultPanel outcome={outcome} running={busy !== null} failure={failure} />
        </div>
      </div>
    </div>
  )
}

function ResultPanel({ outcome, running, failure }: { outcome: SqlOutcome | null; running: boolean; failure?: string }) {
  let status = ''
  if (running) status = 'выполняется…'
  else if (outcome?.ok) {
    const n = outcome.table ? outcome.table.rows.length : outcome.affected
    status = `${outcome.table ? `строк: ${n}` : `${outcome.command}: ${n}`} · ${outcome.ms < 1 ? '<1' : Math.round(outcome.ms)} мс`
  }

  return (
    <div className="pr-console">
      <div className="pr-console-head">
        <span>результат</span>
        {status && <span className={running ? 'pr-spinner' : undefined}>{status}</span>}
      </div>
      <div className="pr-console-body">
        {failure && (
          <div className="pr-msg pr-msg-error">
            <div className="pr-msg-title">Сбой</div>
            <pre>{failure}</pre>
          </div>
        )}

        {outcome && !outcome.ok && (
          <div className="pr-msg pr-msg-error">
            <div className="pr-msg-title">Ошибка Postgres</div>
            <pre>
              {outcome.error}
              {outcome.hint ? `\nподсказка: ${outcome.hint}` : ''}
            </pre>
          </div>
        )}

        {outcome?.ok && outcome.table && <ResultTable table={outcome.table} />}

        {outcome?.ok && !outcome.table && (
          <p className="pr-note">
            Команда выполнилась и откатилась: каждый прогон идёт в своей транзакции, данные задачи не меняются.
          </p>
        )}

        {!outcome && !failure && !running && (
          <p className="pr-empty">
            Здесь появится результат. Запрос выполняет настоящий PostgreSQL прямо в браузере; все изменения данных откатываются
            после каждого прогона.
          </p>
        )}
      </div>
    </div>
  )
}

/** Строк у задач немного, но бесконечный generate_series не должен повесить страницу. */
const MAX_ROWS = 200

function ResultTable({ table }: { table: SqlTable }) {
  if (table.rows.length === 0) return <p className="pr-note">(0 строк)</p>
  return (
    <div className="sql-table-wrap">
      <table className="sql-table">
        <thead>
          <tr>
            {table.columns.map((c, i) => (
              <th key={i}>{c}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {table.rows.slice(0, MAX_ROWS).map((row, i) => (
            <tr key={i}>
              {row.map((cell, j) => (
                <td key={j} className={cell === null ? 'sql-null' : undefined}>
                  {cell === null ? 'NULL' : cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      {table.rows.length > MAX_ROWS && <p className="pr-note">…и ещё {table.rows.length - MAX_ROWS}</p>}
    </div>
  )
}
