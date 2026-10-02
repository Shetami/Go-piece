/**
 * Прогресс аккаунта — что из localStorage уезжает на сервер и как две копии
 * сливаются в одну. Чистый модуль: его импортируют и браузер, и серверный
 * маршрут /api/progress, и тесты.
 *
 * Состояние — это просто ключи localStorage со строковыми значениями, как они
 * лежат в браузере. Острова продолжают читать и писать localStorage напрямую,
 * а синхронизация переносит ключи туда и обратно. Поэтому слияние обязано
 * быть коммутативным и идемпотентным: прогресс только копится — решённая задача
 * не «разрешается», лучший результат теста не падает, достижение не пропадает.
 */

export type State = Record<string, string>

/** Ключи, которые принадлежат аккаунту. Черновики кода и тема остаются в браузере. */
const SYNCED = /^go-piece:(progress|achievements|quiz:[^:]{1,200}|practice:[^:]{1,200}:solved)$/

export function isSyncedKey(key: string): boolean {
  return SYNCED.test(key)
}

/** Больше этого на одно значение не бывает: прогресс — это идентификаторы и даты. */
export const MAX_VALUE_BYTES = 256 * 1024
/** Всё состояние целиком: на порядки больше реального. */
export const MAX_TOTAL_BYTES = 2 * 1024 * 1024
/** И ключей столько не бывает: задач, тестов и служебных ключей — сотни. */
export const MAX_KEYS = 5000

function parse(raw: string): unknown {
  try {
    return JSON.parse(raw)
  } catch {
    return undefined
  }
}

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
}

/**
 * Ключи по алфавиту. Слияние выдаёт одну и ту же строку для одного и того же
 * содержимого — иначе копии, различающиеся лишь порядком ключей, гоняли бы
 * синхронизацию туда-обратно.
 */
function sorted<T>(obj: Record<string, T>): Record<string, T> {
  return Object.fromEntries(Object.entries(obj).sort(([x], [y]) => (x < y ? -1 : x > y ? 1 : 0)))
}

/** Отметки «когда впервые»: объединение, при совпадении — более раннее время. */
function mergeFirstSeen(a: unknown, b: unknown): Record<string, number> {
  const out: Record<string, number> = {}
  for (const src of [a, b]) {
    if (!isObject(src)) continue
    for (const [k, v] of Object.entries(src)) {
      if (typeof v !== 'number') continue
      out[k] = k in out ? Math.min(out[k]!, v) : v
    }
  }
  return sorted(out)
}

function mergeDays(a: unknown, b: unknown): string[] {
  const days = new Set<string>()
  for (const src of [a, b]) {
    if (Array.isArray(src)) for (const d of src) if (typeof d === 'string') days.add(d)
  }
  return [...days].sort()
}

/** go-piece:progress — см. Stored в src/lib/progress.ts. */
function mergeProgress(a: Record<string, unknown>, b: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = { ...a, ...b }
  for (const field of ['read', 'stands', 'terms', 'solves', 'flags']) {
    out[field] = mergeFirstSeen(a[field], b[field])
  }
  out.days = mergeDays(a.days, b.days)
  return sorted(out)
}

/** go-piece:quiz:<id> — остаётся лучший результат. */
function mergeQuiz(a: Record<string, unknown>, b: Record<string, unknown>): Record<string, unknown> {
  const num = (v: unknown) => (typeof v === 'number' ? v : -1)
  // При равном результате — тот, где вопросов больше: порядок аргументов не должен влиять.
  const diff = num(a.best) - num(b.best) || num(a.total) - num(b.total)
  return diff > 0 ? a : diff < 0 ? b : JSON.stringify(a) < JSON.stringify(b) ? a : b
}

/**
 * Слить два значения одного ключа. Если что-то не разбирается — побеждает
 * то, что разбирается; если не разбирается ни одно — `b`.
 */
export function mergeValue(key: string, a: string | undefined, b: string | undefined): string | undefined {
  if (a === undefined) return b
  if (b === undefined || a === b) return a

  if (key.endsWith(':solved')) return a === '1' || b === '1' ? '1' : b

  const pa = parse(a)
  const pb = parse(b)
  if (!isObject(pb)) return isObject(pa) ? a : b
  if (!isObject(pa)) return b

  if (key === 'go-piece:progress') return JSON.stringify(mergeProgress(pa, pb))
  if (key === 'go-piece:achievements') return JSON.stringify(mergeFirstSeen(pa, pb))
  if (key.startsWith('go-piece:quiz:')) return JSON.stringify(mergeQuiz(pa, pb))
  return b
}

/** Слить два состояния целиком. */
export function mergeState(a: State, b: State): State {
  const out: State = {}
  for (const key of new Set([...Object.keys(a), ...Object.keys(b)])) {
    const v = mergeValue(key, a[key], b[key])
    if (v !== undefined) out[key] = v
  }
  return out
}
