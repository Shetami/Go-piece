/**
 * Прогресс в этом браузере — для достижений. Только клиент: всё живёт в localStorage.
 *
 * Новое пишется в один ключ `go-piece:progress`. Старое не переезжает:
 * решённые задачи и результаты тестов по-прежнему лежат в своих ключах
 * (их пишут острова практики и тестов) и читаются отсюда как есть —
 * так прогресс, набранный до появления достижений, засчитывается сам.
 *
 * Каждая запись шлёт PROGRESS_EVENT — по нему трекер пересчитывает достижения
 * и показывает уведомление о полученных.
 */
import type { Catalog, Progress, QuizResult } from './achievements.ts'

const KEY = 'go-piece:progress'
const EARNED_KEY = 'go-piece:achievements'

export const PROGRESS_EVENT = 'go-piece:progress'
/** Записаны новые полученные достижения. */
export const EARNED_EVENT = 'go-piece:achievements'

interface Stored {
  read: Record<string, number>
  stands: Record<string, number>
  terms: Record<string, number>
  /** Задача → когда решена впервые. Только решённые после появления достижений. */
  solves: Record<string, number>
  days: string[]
  flags: Record<string, number>
}

function empty(): Stored {
  return { read: {}, stands: {}, terms: {}, solves: {}, days: [], flags: {} }
}

function readJson<T>(key: string): T | null {
  try {
    const raw = localStorage.getItem(key)
    return raw ? (JSON.parse(raw) as T) : null
  } catch {
    return null
  }
}

function writeJson(key: string, value: unknown): void {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch {
    // Приватный режим — прогресс просто не запомнится.
  }
}

function load(): Stored {
  return { ...empty(), ...readJson<Partial<Stored>>(KEY) }
}

/** YYYY-MM-DD по местному времени: «день» для серий — день человека, а не UTC. */
export function localDay(ts: number): string {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

function update(fn: (s: Stored, now: number) => void): void {
  const s = load()
  const now = Date.now()
  fn(s, now)
  const today = localDay(now)
  if (!s.days.includes(today)) s.days.push(today)
  writeJson(KEY, s)
  window.dispatchEvent(new Event(PROGRESS_EVENT))
}

/** Отметить, что что-то произошло, без отдельной записи — например, пройден тест. */
export function notifyProgress(): void {
  update(() => {})
}

export function recordRead(lectureId: string): void {
  update((s, now) => {
    s.read[lectureId] ??= now
  })
}

export function recordStand(standId: string): void {
  update((s, now) => {
    s.stands[standId] ??= now
  })
}

export function recordTerm(termId: string): void {
  update((s, now) => {
    s.terms[termId] ??= now
  })
}

export function recordSolve(taskId: string): void {
  update((s, now) => {
    s.solves[taskId] ??= now
    if (new Date(now).getHours() < 5) s.flags.night ??= now
  })
}

export function recordFlag(name: string): void {
  update((s, now) => {
    s.flags[name] ??= now
  })
}

/** Собрать прогресс: новый ключ плюс старые ключи задач и тестов по каталогу. */
export function loadProgress(catalog: Catalog): Progress {
  const s = load()

  const solved = new Set(Object.keys(s.solves))
  const quizzes = new Map<string, QuizResult>()
  try {
    for (const t of catalog.tasks) {
      if (localStorage.getItem(`go-piece:practice:${t.id}:solved`) === '1') solved.add(t.id)
    }
    for (const id of catalog.quizzes) {
      const r = readJson<QuizResult>(`go-piece:quiz:${id}`)
      if (r && typeof r.best === 'number' && typeof r.total === 'number') quizzes.set(id, r)
    }
  } catch {
    // Нет доступа к хранилищу — считаем только то, что удалось прочитать.
  }

  const perDay = new Map<string, number>()
  for (const ts of Object.values(s.solves)) {
    const d = localDay(ts)
    perDay.set(d, (perDay.get(d) ?? 0) + 1)
  }

  return {
    read: new Set(Object.keys(s.read)),
    quizzes,
    solved,
    stands: new Set(Object.keys(s.stands)),
    terms: new Set(Object.keys(s.terms)),
    days: s.days,
    bestDay: Math.max(0, ...perDay.values()),
    flags: new Set(Object.keys(s.flags)),
  }
}

/** Полученные достижения: id → когда. */
export function loadEarned(): Record<string, number> {
  return readJson<Record<string, number>>(EARNED_KEY) ?? {}
}

export function saveEarned(earned: Record<string, number>): void {
  writeJson(EARNED_KEY, earned)
  window.dispatchEvent(new Event(EARNED_EVENT))
}
