/**
 * Остров страницы задачи: редактор, кнопки, вывод и вердикт.
 *
 * Механика зависит от типа задачи, и вся разница собрана здесь, в одном месте:
 *
 *   output    — редактор заперт, отвечают текстом в поле «что выведет»;
 *   bug       — редактор заперт, отвечают кликом по строке;
 *   fix       — редактор открыт, ответ — сам код;
 *   implement — то же, но проверяют скрытые тесты.
 *
 * Запертый редактор — не защита (код всё равно на странице), а подсказка о
 * жанре: в задаче «что выведет» править код бессмысленно, вердикт всё равно
 * выносится по нетронутой заготовке с сервера. Замок снимается кнопкой, когда
 * человек уже ответил и хочет поковыряться.
 */

import { useCallback, useEffect, useRef, useState } from 'react'
import type { TaskKindId } from '../../../data/practice.ts'
import { checkTask, formatCode, runCode } from '../client.ts'
import type { CheckResult, RunResult, Verdict } from '../protocol.ts'
import Console from './Console.tsx'
import Editor, { type EditorHandle } from './Editor.tsx'
import { goDiagnostics } from './editor.ts'
import { Text } from '@codemirror/state'

interface Props {
  taskId: string
  kind: TaskKindId
  /** Заголовок над полем ответа — из TASK_KINDS. */
  prompt: string
  starter: string
  /** Эталон. Для output и bug он совпадает с заготовкой и кнопкой не показывается. */
  solution: string
}

const EMPTY_RUN: RunResult = { errors: '', vet: '', output: [], stubbedMain: false }

/** Черновик и отметка «решено» переживают перезагрузку — но только в этом браузере. */
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
    // Приватный режим или запрет на хранилище — черновик просто не переживёт перезагрузку.
  }
}

export default function TaskRunner({ taskId, kind, prompt, starter, solution }: Props) {
  const picksLine = kind === 'bug'
  const editsCode = kind === 'fix' || kind === 'implement'

  const [code, setCode] = useState(starter)
  const [answer, setAnswer] = useState('')
  const [pickedLine, setPickedLine] = useState(0)
  const [unlocked, setUnlocked] = useState(editsCode)

  const [run, setRun] = useState<RunResult>(EMPTY_RUN)
  const [verdict, setVerdict] = useState<Verdict | null>(null)
  const [failure, setFailure] = useState<string>()
  const [busy, setBusy] = useState<null | 'run' | 'check' | 'fmt'>(null)
  const [solved, setSolved] = useState(false)

  const editor = useRef<EditorHandle>(null)

  // Черновик и «решено» читаем после гидратации: на сервере localStorage нет.
  useEffect(() => {
    const draft = load(storageKey(taskId, 'code'))
    if (draft !== null && draft !== starter) {
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

  /** Решено — распахиваем разбор: он живёт в MDX за пределами острова. */
  const markSolved = useCallback(() => {
    setSolved(true)
    save(storageKey(taskId, 'solved'), '1')
    document.querySelector<HTMLDetailsElement>('.reveal')?.setAttribute('open', '')
  }, [taskId])

  /** Перенести жалобы компилятора и vet в подчёркивания редактора. */
  const showGoErrors = useCallback((result: RunResult) => {
    const doc = Text.of(code.split('\n'))
    editor.current?.setDiagnostics([
      ...goDiagnostics(result.errors, doc, 'error'),
      ...goDiagnostics(result.vet, doc, 'warning'),
    ])
  }, [code])

  const doRun = useCallback(async () => {
    setBusy('run')
    setFailure(undefined)
    setVerdict(null)
    try {
      const result = await runCode(code)
      setRun(result)
      showGoErrors(result)
    } catch (err) {
      setFailure(err instanceof Error ? err.message : String(err))
      setRun(EMPTY_RUN)
    } finally {
      setBusy(null)
    }
  }, [code, showGoErrors])

  const doFormat = useCallback(async () => {
    if (!unlocked) return
    setBusy('fmt')
    setFailure(undefined)
    try {
      const result = await formatCode(code)
      if ('syntaxError' in result) {
        editor.current?.setDiagnostics(goDiagnostics(result.syntaxError, Text.of(code.split('\n')), 'error'))
      } else {
        setCode(result.code)
        editor.current?.setDiagnostics([])
      }
    } catch (err) {
      setFailure(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(null)
    }
  }, [code, unlocked])

  const doCheck = useCallback(async () => {
    setBusy('check')
    setFailure(undefined)
    try {
      const result: CheckResult = await checkTask({
        task: taskId,
        kind,
        ...(editsCode ? { code } : {}),
        ...(kind === 'output' ? { answer } : {}),
        ...(picksLine ? { line: pickedLine } : {}),
      })
      setRun(result)
      setVerdict(result.verdict)
      showGoErrors(result)
      if (result.verdict.pass) markSolved()
    } catch (err) {
      setFailure(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(null)
    }
  }, [answer, code, editsCode, kind, markSolved, pickedLine, picksLine, showGoErrors, taskId])

  const reset = useCallback(() => {
    setCode(starter)
    setPickedLine(0)
    setVerdict(null)
    setRun(EMPTY_RUN)
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

  // Можно ли сейчас нажать «Проверить» и что мешает, если нельзя.
  const blocker =
    kind === 'output' && answer.trim() === ''
      ? 'напишите ожидаемый вывод'
      : picksLine && pickedLine === 0
        ? 'отметьте строку слева от номеров'
        : picksLine && dirty
          ? 'код изменён, номера строк уехали — нажмите «Сбросить»'
          : null

  return (
    <div className="pr-runner not-prose">
      <div className="pr-toolbar">
        <button type="button" className="pr-btn pr-btn-main" onClick={doRun} disabled={busy !== null}>
          {busy === 'run' ? 'Запускаю…' : 'Запустить'}
        </button>

        <button type="button" className="pr-btn" onClick={doCheck} disabled={busy !== null || blocker !== null} title={blocker ?? undefined}>
          {busy === 'check' ? 'Проверяю…' : 'Проверить'}
        </button>

        {unlocked && (
          <button type="button" className="pr-btn" onClick={doFormat} disabled={busy !== null} title="gofmt + недостающие импорты (⌘S)">
            {busy === 'fmt' ? 'Форматирую…' : 'Формат'}
          </button>
        )}

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
            <span>main.go</span>
            {!unlocked && <span className="pr-lock">только чтение</span>}
            {picksLine && <span className="pr-hint">кликните в колонку слева от номера строки</span>}
          </div>
          <Editor
            ref={editor}
            code={code}
            onChange={setCode}
            onRun={doRun}
            onFormat={doFormat}
            onPick={picksLine ? setPickedLine : undefined}
            pickedLine={picksLine ? pickedLine : undefined}
            readOnly={!unlocked}
          />
          <div className="pr-keys">
            <kbd>⌘</kbd>+<kbd>↵</kbd> запустить{unlocked && <> · <kbd>⌘</kbd>+<kbd>S</kbd> формат · <kbd>Tab</kbd> отступ</>}
          </div>
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
                rows={6}
                spellCheck={false}
                placeholder={'строка за строкой,\nровно так, как напечатает программа'}
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

          <Console
            errors={run.errors}
            vet={run.vet}
            output={run.output}
            stubbedMain={run.stubbedMain}
            running={busy === 'run' || busy === 'check'}
            failure={failure}
          />
        </div>
      </div>
    </div>
  )
}
