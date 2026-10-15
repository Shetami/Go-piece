/**
 * Остров страницы «Сгенерировать задачу»: выбор темы, формата и сложности,
 * ход генерации и список своих задач.
 *
 * Генерация идёт долго, поэтому /api/generate отвечает потоком NDJSON — по
 * строке на шаг. Остров показывает шаги по мере прихода, а в конце уводит на
 * страницу готовой задачи.
 */
import { useEffect, useState } from 'react'
import {
  TASK_KINDS,
  TASK_LEVELS,
  TASK_TOPICS,
  type TaskKindId,
  type TaskLevelId,
  type TaskTopicId,
} from '../../../data/practice.ts'

interface Summary {
  id: string
  topic: TaskTopicId
  kind: TaskKindId
  level: TaskLevelId
  title: string
  createdAt: number
}

interface State {
  user: { id: number; email: string } | null
  configured: boolean
  limit: number
  used: number
  tasks: Summary[]
}

type Event = { type: 'progress'; message: string } | { type: 'done'; id: string; url: string } | { type: 'error'; message: string }

const GO_TOPICS = TASK_TOPICS.filter((t) => t.lang === 'go')

const titleOf = <T extends { id: string; title: string }>(list: readonly T[], id: string) => list.find((x) => x.id === id)?.title ?? id

export default function GenerateForm() {
  const [state, setState] = useState<State | null>(null)
  const [loadError, setLoadError] = useState<string>()

  const [topic, setTopic] = useState<TaskTopicId>('channels')
  const [kind, setKind] = useState<TaskKindId>('output')
  const [level, setLevel] = useState<TaskLevelId>('medium')

  const [busy, setBusy] = useState(false)
  const [log, setLog] = useState<string[]>([])
  const [error, setError] = useState<string>()

  useEffect(() => {
    fetch('/api/generate', { credentials: 'same-origin' })
      .then((r) => r.json() as Promise<State>)
      .then(setState)
      .catch(() => setLoadError('не получилось связаться с сервером'))
  }, [])

  async function generate() {
    setBusy(true)
    setError(undefined)
    setLog([])
    try {
      const res = await fetch('/api/generate', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({ topic, kind, level }),
      })
      if (!res.ok || !res.body) {
        const data = (await res.json().catch(() => ({}))) as { error?: string }
        throw new Error(data.error ?? `сервер ответил ${res.status}`)
      }
      setState((s) => (s ? { ...s, used: s.used + 1 } : s))

      // Поток NDJSON: строки могут прийти кусками, поэтому копим хвост.
      const reader = res.body.pipeThrough(new TextDecoderStream()).getReader()
      let buffer = ''
      for (;;) {
        const { value, done } = await reader.read()
        if (done) break
        buffer += value
        const lines = buffer.split('\n')
        buffer = lines.pop() ?? ''
        for (const line of lines) {
          if (!line.trim()) continue
          const event = JSON.parse(line) as Event
          if (event.type === 'progress') setLog((l) => [...l, event.message])
          if (event.type === 'error') throw new Error(event.message)
          if (event.type === 'done') {
            setLog((l) => [...l, 'Готово — открываю задачу'])
            window.location.href = event.url
            return
          }
        }
      }
      throw new Error('соединение оборвалось до конца генерации — задача, если успела, появится в списке ниже')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setBusy(false)
    }
  }

  if (loadError) return <p className="auth-error">{loadError}</p>
  if (!state) return <p className="text-muted">Загрузка…</p>

  if (!state.configured) {
    return (
      <p className="rounded-lg border border-line bg-surface px-4 py-3 text-muted max-w-2xl">
        Генерация задач на этом сервере не настроена: нужен ключ Claude API в переменной <code>ANTHROPIC_API_KEY</code>.
      </p>
    )
  }

  if (!state.user) {
    return (
      <p className="rounded-lg border border-line bg-surface px-4 py-3 max-w-2xl">
        Генерировать задачи можно после входа: каждая генерация обращается к Claude и стоит денег, поэтому на неё есть
        суточный лимит. <a href="/login/" className="text-accent">Войти</a> или{' '}
        <a href="/register/" className="text-accent">зарегистрироваться</a>.
      </p>
    )
  }

  const left = Math.max(0, state.limit - state.used)

  return (
    <div className="max-w-4xl">
      <Choice label="Тема" options={GO_TOPICS} value={topic} onChange={setTopic} disabled={busy} />
      <Choice label="Формат" options={TASK_KINDS} value={kind} onChange={setKind} disabled={busy} />
      <Choice label="Сложность" options={TASK_LEVELS} value={level} onChange={setLevel} disabled={busy} />

      <div className="flex flex-wrap items-center gap-4 mt-6">
        <button type="button" className="pr-btn pr-btn-main" onClick={generate} disabled={busy || left === 0}>
          {busy ? 'Генерирую…' : 'Сгенерировать задачу'}
        </button>
        <span className="text-sm text-muted">
          {left > 0 ? `осталось на сутки: ${left} из ${state.limit}` : `лимит на сутки исчерпан (${state.limit})`}
        </span>
      </div>

      {(log.length > 0 || busy) && (
        <ol className="mt-6 space-y-1 text-sm font-mono">
          {log.map((line, i) => (
            <li key={i} className={i === log.length - 1 && busy ? 'text-ink' : 'text-muted'}>
              {line}
            </li>
          ))}
        </ol>
      )}
      {busy && <p className="text-xs text-muted mt-3">Обычно это занимает от 30 секунд до двух минут: задача пишется и проверяется в песочнице go.dev.</p>}
      {error && <p className="auth-error mt-4">{error}</p>}

      {state.tasks.length > 0 && (
        <section className="mt-12">
          <h2 className="text-xl font-bold mb-4">Мои задачи</h2>
          <ul className="space-y-2">
            {state.tasks.map((t) => (
              <li key={t.id}>
                <a href={`/practice/${t.id}/`} className="font-semibold text-accent">
                  {t.title}
                </a>
                <span className="text-sm text-muted">
                  {' '}
                  — {titleOf(TASK_KINDS, t.kind)}, {titleOf(TASK_TOPICS, t.topic)}, {titleOf(TASK_LEVELS, t.level).toLowerCase()}
                </span>
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  )
}

function Choice<T extends string>(props: {
  label: string
  options: readonly { id: T; title: string }[]
  value: T
  onChange: (v: T) => void
  disabled: boolean
}) {
  return (
    <div className="flex flex-wrap items-center gap-2 mb-3 text-sm">
      <span className="text-muted w-24 shrink-0">{props.label}:</span>
      {props.options.map((o) => (
        <button
          key={o.id}
          type="button"
          className={`pr-filter${o.id === props.value ? ' is-on' : ''}`}
          onClick={() => props.onChange(o.id)}
          disabled={props.disabled}
        >
          {o.title}
        </button>
      ))}
    </div>
  )
}
