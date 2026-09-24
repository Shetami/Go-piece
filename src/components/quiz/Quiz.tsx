/**
 * Остров «Проверь себя» в конце лекции.
 *
 * Каждый вопрос проверяется отдельно: выбрал — «Проверить» — видишь, прав ли,
 * и разбор. Когда отвечены все, показывается счёт, а лучший результат
 * запоминается в localStorage — его же читает страница-сводка /quiz/.
 *
 * Варианты ответов и шаги в вопросах на порядок перемешиваются — иначе
 * правильный ответ стоял бы там, куда его поставил автор, и это быстро
 * замечают. Перемешивание детерминировано (сид — текст вопроса), чтобы
 * разметка с сервера совпала с гидратацией на клиенте.
 */

import { Fragment, useEffect, useMemo, useState, type ReactNode } from 'react'
import type { Question } from './types.ts'

interface Props {
  lectureId: string
  questions: Question[]
  /** Совпадает с quizStorageKey из src/lib/quizzes.ts — модуль тянет import.meta.glob, на клиент его не везём. */
  storageKey: string
}

interface Stored {
  best: number
  total: number
}

function loadStored(key: string): Stored | null {
  try {
    const raw = localStorage.getItem(key)
    return raw ? (JSON.parse(raw) as Stored) : null
  } catch {
    return null
  }
}

function saveStored(key: string, value: Stored): void {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch {
    // Приватный режим — результат просто не запомнится.
  }
}

/** `код` и **жирный** — больше разметки в вопросах нет. */
export function Inline({ text }: { text: string }) {
  const parts: ReactNode[] = []
  const re = /`([^`]+)`|\*\*([^*]+)\*\*/g
  let last = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(text))) {
    if (m.index > last) parts.push(text.slice(last, m.index))
    parts.push(m[1] !== undefined ? <code key={m.index}>{m[1]}</code> : <b key={m.index}>{m[2]}</b>)
    last = re.lastIndex
  }
  if (last < text.length) parts.push(text.slice(last))
  return <>{parts.map((p, i) => <Fragment key={i}>{p}</Fragment>)}</>
}

/**
 * Детерминированная перестановка индексов по сиду. С avoidIdentity гарантированно
 * не совпадает с исходным порядком — для вопросов «расставьте по порядку».
 */
function seededPerm(seed: string, n: number, avoidIdentity: boolean): number[] {
  let h = 2166136261
  for (let i = 0; i < seed.length; i++) h = Math.imul(h ^ seed.charCodeAt(i), 16777619)
  const rnd = () => {
    h = Math.imul(h ^ (h >>> 15), 2246822507)
    h = Math.imul(h ^ (h >>> 13), 3266489909)
    h ^= h >>> 16
    return (h >>> 0) / 4294967296
  }
  const idx = Array.from({ length: n }, (_, i) => i)
  for (let i = n - 1; i > 0; i--) {
    const j = Math.floor(rnd() * (i + 1))
    ;[idx[i], idx[j]] = [idx[j]!, idx[i]!]
  }
  if (avoidIdentity && idx.every((v, i) => v === i)) idx.push(idx.shift()!)
  return idx
}

type Answer = number | number[] | boolean | null

function initialAnswer(q: Question): Answer {
  switch (q.kind) {
    case 'single':
      return null
    case 'multi':
      return []
    case 'truefalse':
      return null
    case 'order':
      return seededPerm(q.q, q.items.length, true)
  }
}

function isCorrect(q: Question, a: Answer): boolean {
  switch (q.kind) {
    case 'single':
      return a === q.answer
    case 'multi': {
      const picked = [...(a as number[])].sort()
      const right = [...q.answer].sort()
      return picked.length === right.length && picked.every((v, i) => v === right[i])
    }
    case 'truefalse':
      return a === q.answer
    case 'order':
      return (a as number[]).every((v, i) => v === i)
  }
}

function isAnswered(q: Question, a: Answer): boolean {
  if (q.kind === 'multi') return (a as number[]).length > 0
  if (q.kind === 'order') return true
  return a !== null
}

const KIND_HINT: Record<Question['kind'], string> = {
  single: 'Один правильный ответ',
  multi: 'Отметьте все правильные',
  truefalse: 'Верно или нет?',
  order: 'Расставьте по порядку',
}

function QuestionCard({
  n,
  q,
  answer,
  checked,
  onAnswer,
  onCheck,
}: {
  n: number
  q: Question
  answer: Answer
  checked: boolean
  onAnswer: (a: Answer) => void
  onCheck: () => void
}) {
  const right = checked && isCorrect(q, answer)

  let body: ReactNode = null
  if (q.kind === 'single' || q.kind === 'multi') {
    const multi = q.kind === 'multi'
    const picked = multi ? (answer as number[]) : [answer as number]
    const correct = multi ? q.answer : [q.answer]
    body = (
      <div className="quiz-options">
        {seededPerm(q.q, q.options.length, false).map((i) => {
          const opt = q.options[i]!
          const on = picked.includes(i)
          const cls = [
            'quiz-option',
            on && 'is-picked',
            checked && correct.includes(i) && 'is-right',
            checked && on && !correct.includes(i) && 'is-wrong',
          ]
            .filter(Boolean)
            .join(' ')
          return (
            <button
              key={i}
              type="button"
              className={cls}
              disabled={checked}
              aria-pressed={on}
              onClick={() => {
                if (!multi) onAnswer(i)
                else onAnswer(on ? picked.filter((x) => x !== i) : [...picked, i])
              }}
            >
              <span className="quiz-mark" aria-hidden>
                {multi ? (on ? '☑' : '☐') : on ? '●' : '○'}
              </span>
              <span>
                <Inline text={opt} />
              </span>
            </button>
          )
        })}
      </div>
    )
  } else if (q.kind === 'truefalse') {
    body = (
      <div className="quiz-tf">
        {[true, false].map((v) => {
          const on = answer === v
          const cls = [
            'quiz-option',
            on && 'is-picked',
            checked && v === q.answer && 'is-right',
            checked && on && v !== q.answer && 'is-wrong',
          ]
            .filter(Boolean)
            .join(' ')
          return (
            <button key={String(v)} type="button" className={cls} disabled={checked} onClick={() => onAnswer(v)}>
              {v ? 'Верно' : 'Неверно'}
            </button>
          )
        })}
      </div>
    )
  } else {
    const order = answer as number[]
    const move = (from: number, to: number) => {
      const next = [...order]
      ;[next[from], next[to]] = [next[to]!, next[from]!]
      onAnswer(next)
    }
    body = (
      <ol className="quiz-order">
        {order.map((item, pos) => (
          <li
            key={item}
            className={['quiz-order-item', checked && (item === pos ? 'is-right' : 'is-wrong')].filter(Boolean).join(' ')}
          >
            <span className="quiz-order-n">{pos + 1}</span>
            <span className="quiz-order-text">
              <Inline text={q.items[item]!} />
            </span>
            {!checked && (
              <span className="quiz-order-btns">
                <button type="button" className="btn-ghost" disabled={pos === 0} onClick={() => move(pos, pos - 1)} aria-label="Выше">
                  ↑
                </button>
                <button
                  type="button"
                  className="btn-ghost"
                  disabled={pos === order.length - 1}
                  onClick={() => move(pos, pos + 1)}
                  aria-label="Ниже"
                >
                  ↓
                </button>
              </span>
            )}
          </li>
        ))}
      </ol>
    )
  }

  return (
    <li className="quiz-q">
      <div className="quiz-kicker">
        {n}. {KIND_HINT[q.kind]}
      </div>
      <p className="quiz-text">
        <Inline text={q.q} />
      </p>
      {body}
      {!checked ? (
        <button type="button" className="btn quiz-check" disabled={!isAnswered(q, answer)} onClick={onCheck}>
          Проверить
        </button>
      ) : (
        <div className={`quiz-verdict ${right ? 'is-right' : 'is-wrong'}`}>
          <b>{right ? 'Верно.' : 'Не совсем.'}</b>{' '}
          {q.kind === 'order' && !right && (
            <span className="quiz-right-order">
              Правильный порядок:{' '}
              {q.items.map((it, i) => (
                <Fragment key={i}>
                  {i > 0 && ' → '}
                  <Inline text={it} />
                </Fragment>
              ))}
              .{' '}
            </span>
          )}
          <Inline text={q.explain} />
        </div>
      )}
    </li>
  )
}

export default function Quiz({ lectureId, questions, storageKey }: Props) {
  const fresh = useMemo(() => questions.map(initialAnswer), [questions])
  const [answers, setAnswers] = useState<Answer[]>(fresh)
  const [checked, setChecked] = useState<boolean[]>(() => questions.map(() => false))
  const [stored, setStored] = useState<Stored | null>(null)

  // localStorage читаем после гидратации: на сервере его нет.
  useEffect(() => setStored(loadStored(storageKey)), [storageKey])

  const done = checked.every(Boolean)
  const score = questions.reduce((s, q, i) => s + (checked[i] && isCorrect(q, answers[i]!) ? 1 : 0), 0)

  useEffect(() => {
    if (!done) return
    const prev = loadStored(storageKey)
    const next = { best: Math.max(prev?.best ?? 0, score), total: questions.length }
    saveStored(storageKey, next)
    setStored(next)
  }, [done, score, storageKey, questions.length])

  const reset = () => {
    setAnswers(fresh)
    setChecked(questions.map(() => false))
  }

  return (
    <div className="quiz" data-lecture={lectureId}>
      <div className="quiz-head">
        <span>
          {questions.length} вопросов по материалу лекции
        </span>
        {stored && (
          <span className="quiz-best">
            лучший результат: {stored.best} из {stored.total}
          </span>
        )}
      </div>
      <ol className="quiz-list">
        {questions.map((q, i) => (
          <QuestionCard
            key={i}
            n={i + 1}
            q={q}
            answer={answers[i]!}
            checked={checked[i]!}
            onAnswer={(a) => setAnswers((prev) => prev.map((x, j) => (j === i ? a : x)))}
            onCheck={() => setChecked((prev) => prev.map((x, j) => (j === i ? true : x)))}
          />
        ))}
      </ol>
      <div className="quiz-foot">
        {done ? (
          <>
            <span className="quiz-score">
              Результат: <b>{score}</b> из {questions.length}
              {score === questions.length && ' — всё верно'}
            </span>
            <button type="button" className="btn-ghost" onClick={reset}>
              пройти заново
            </button>
          </>
        ) : (
          <span className="text-muted">
            Отвечено {checked.filter(Boolean).length} из {questions.length}
          </span>
        )}
      </div>
    </div>
  )
}
