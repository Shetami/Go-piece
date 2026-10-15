/**
 * Сгенерированные задачи в базе: запись, чтение, лимиты.
 *
 * id такой задачи — `gen/<код>`, где код — случайные 10 символов. Префикс
 * отличает её от задач из src/content/tasks везде, где ходит id: в /api/check,
 * в ключах прогресса, в адресе страницы (/practice/gen/<код>/).
 *
 * Задача доступна всем, у кого есть ссылка: ею можно поделиться. А вот
 * генерировать можно только после входа — каждая генерация стоит денег.
 *
 * Только сервер.
 */
import { randomBytes } from 'node:crypto'
import type { TaskKindId, TaskLevelId, TaskTopicId } from '../../data/practice.ts'
import { BadRequest } from '../api.ts'
import type { TaskSecrets } from './checker.ts'
import { getDb } from './db.ts'

export const GENERATED_PREFIX = 'gen/'

/** Сколько задач один человек может сгенерировать за сутки. */
export function dailyLimit(): number {
  const n = Number(process.env.GENERATE_DAILY_LIMIT)
  return Number.isInteger(n) && n > 0 ? n : 10
}

const DAY_MS = 24 * 60 * 60 * 1000

export interface NewGeneratedTask {
  topic: TaskTopicId
  kind: TaskKindId
  level: TaskLevelId
  title: string
  description: string
  statement: string
  explanation: string
  starter: string
  solution: string
  tests: string | null
  expect: string | null
  bugLine: number | null
}

/** То, что можно показать на странице. Скрытых частей здесь нет. */
export interface GeneratedTaskView {
  id: string
  topic: TaskTopicId
  kind: TaskKindId
  level: TaskLevelId
  title: string
  description: string
  statement: string
  explanation: string
  starter: string
  solution: string
  createdAt: number
}

export interface GeneratedTaskSummary {
  id: string
  topic: TaskTopicId
  kind: TaskKindId
  level: TaskLevelId
  title: string
  createdAt: number
}

/** Код из адреса или id → код. Чужие символы отсекаем до запроса в базу. */
function codeOf(idOrCode: string): string | null {
  const code = idOrCode.startsWith(GENERATED_PREFIX) ? idOrCode.slice(GENERATED_PREFIX.length) : idOrCode
  return /^[a-z0-9]{10}$/.test(code) ? code : null
}

function newCode(): string {
  // base36 из случайных байт: только строчные буквы и цифры, удобно в адресе.
  return [...randomBytes(10)].map((b) => (b % 36).toString(36)).join('')
}

export function saveGenerated(userId: number, task: NewGeneratedTask): string {
  const code = newCode()
  getDb()
    .prepare(
      `INSERT INTO generated_tasks
        (id, user_id, topic, kind, level, title, description, statement, explanation, starter, solution, tests, expect, bug_line, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    )
    .run(
      code,
      userId,
      task.topic,
      task.kind,
      task.level,
      task.title,
      task.description,
      task.statement,
      task.explanation,
      task.starter,
      task.solution,
      task.tests,
      task.expect,
      task.bugLine,
      Date.now(),
    )
  return code
}

interface Row {
  id: string
  topic: TaskTopicId
  kind: TaskKindId
  level: TaskLevelId
  title: string
  description: string
  statement: string
  explanation: string
  starter: string
  solution: string
  tests: string | null
  expect: string | null
  bug_line: number | null
  created_at: number
}

function row(idOrCode: string): Row | null {
  const code = codeOf(idOrCode)
  if (!code) return null
  return (getDb().prepare('SELECT * FROM generated_tasks WHERE id = ?').get(code) as Row | undefined) ?? null
}

export function generatedView(code: string): GeneratedTaskView | null {
  const r = row(code)
  if (!r) return null
  return {
    id: `${GENERATED_PREFIX}${r.id}`,
    topic: r.topic,
    kind: r.kind,
    level: r.level,
    title: r.title,
    description: r.description,
    statement: r.statement,
    explanation: r.explanation,
    starter: r.starter,
    // В «что выведет» и «найди баг» эталон показывать нельзя: он и есть ответ
    // (а для output — та же заготовка). Так же поступают задачи из контента.
    solution: r.kind === 'fix' || r.kind === 'implement' ? r.solution : r.starter,
    createdAt: r.created_at,
  }
}

/** Скрытые части — для /api/check. */
export function generatedSecrets(id: string): TaskSecrets {
  const r = row(id)
  if (!r) throw new BadRequest(`нет такой задачи: ${id}`, 404)
  return {
    kind: r.kind,
    starter: r.starter,
    bugLine: r.bug_line ?? undefined,
    check: r.tests,
    expect: r.expect,
  }
}

export function listGenerated(userId: number, limit = 50): GeneratedTaskSummary[] {
  const rows = getDb()
    .prepare(
      'SELECT id, topic, kind, level, title, created_at FROM generated_tasks WHERE user_id = ? ORDER BY created_at DESC LIMIT ?',
    )
    .all(userId, limit) as unknown as Pick<Row, 'id' | 'topic' | 'kind' | 'level' | 'title' | 'created_at'>[]
  return rows.map((r) => ({
    id: `${GENERATED_PREFIX}${r.id}`,
    topic: r.topic,
    kind: r.kind,
    level: r.level,
    title: r.title,
    createdAt: r.created_at,
  }))
}

/** Названия последних задач по теме — чтобы генератор не выдавал одно и то же. */
export function recentTitles(topic: TaskTopicId, kind: TaskKindId, limit = 30): string[] {
  const rows = getDb()
    .prepare('SELECT title FROM generated_tasks WHERE topic = ? AND kind = ? ORDER BY created_at DESC LIMIT ?')
    .all(topic, kind, limit) as { title: string }[]
  return rows.map((r) => r.title)
}

/**
 * Сколько генераций человек уже потратил за последние сутки. Считаются попытки,
 * а не сохранённые задачи: неудачная генерация стоит столько же.
 */
export function usedToday(userId: number): number {
  const r = getDb()
    .prepare('SELECT COUNT(*) AS n FROM generation_attempts WHERE user_id = ? AND created_at > ?')
    .get(userId, Date.now() - DAY_MS) as { n: number }
  return r.n
}

export function recordAttempt(userId: number): void {
  getDb().prepare('INSERT INTO generation_attempts (user_id, created_at) VALUES (?, ?)').run(userId, Date.now())
}
