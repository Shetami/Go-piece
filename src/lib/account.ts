/**
 * Аккаунт в браузере: вход, выход и синхронизация прогресса с сервером.
 * Только клиент.
 *
 * Острова по-прежнему читают и пишут localStorage — это быстрый локальный
 * кэш, и без входа сайт работает как раньше. Когда человек вошёл, этот модуль
 * отправляет прогресс на сервер (/api/progress) и дописывает в localStorage
 * то, что пришло с других устройств. Какие ключи уезжают и как сливаются —
 * src/lib/sync.ts.
 *
 * Чей прогресс сейчас в localStorage, помнит ключ OWNER_KEY:
 *   - нет владельца (гость) и человек входит — гостевой прогресс переходит
 *     в аккаунт: это тот же человек, просто он ещё не вошёл;
 *   - владелец другой — локальный прогресс сначала стирается, чтобы чужие
 *     решения не попали в аккаунт;
 *   - выход стирает прогресс из браузера: следующий за этим компьютером
 *     начинает с нуля, а прогресс ждёт на сервере.
 */
import { isSyncedKey, mergeValue, type State } from './sync.ts'
import { EARNED_EVENT, PROGRESS_EVENT } from './progress.ts'

export interface Account {
  id: number
  email: string
}

const OWNER_KEY = 'go-piece:owner'
/** Сменился вошедший пользователь (или он вышел). */
export const ACCOUNT_EVENT = 'go-piece:account'

/** Пауза перед отправкой: решение задачи даёт несколько записей подряд. */
const PUSH_DELAY = 800

export function cachedAccount(): Account | null {
  try {
    const raw = localStorage.getItem(OWNER_KEY)
    const a = raw ? (JSON.parse(raw) as Account) : null
    return a && typeof a.id === 'number' && typeof a.email === 'string' ? a : null
  } catch {
    return null
  }
}

function setOwner(a: Account | null): void {
  try {
    if (a) localStorage.setItem(OWNER_KEY, JSON.stringify({ id: a.id, email: a.email }))
    else localStorage.removeItem(OWNER_KEY)
  } catch {
    // Без хранилища синхронизировать всё равно нечего.
  }
  window.dispatchEvent(new Event(ACCOUNT_EVENT))
}

function localState(): State {
  const state: State = {}
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key && isSyncedKey(key)) state[key] = localStorage.getItem(key) ?? ''
    }
  } catch {
    // Нет доступа — отправим пустое.
  }
  return state
}

/** Стереть из браузера всё, что принадлежит аккаунту, вместе с черновиками задач. */
function clearLocal(): void {
  try {
    const keys: string[] = []
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key?.startsWith('go-piece:') && key !== OWNER_KEY) keys.push(key)
    }
    for (const key of keys) localStorage.removeItem(key)
  } catch {
    // Нечего стирать.
  }
}

/**
 * Дописать пришедшее с сервера. Сливаем с тем, что лежит в localStorage
 * сейчас, а не с отправленным: пока шёл запрос, могла решиться ещё задача.
 */
function apply(server: State): boolean {
  let changed = false
  try {
    for (const [key, value] of Object.entries(server)) {
      if (!isSyncedKey(key)) continue
      const local = localStorage.getItem(key) ?? undefined
      const merged = mergeValue(key, local, value)
      if (merged !== undefined && merged !== local) {
        localStorage.setItem(key, merged)
        changed = true
      }
    }
  } catch {
    // Хранилище переполнено или закрыто — останется как есть.
  }
  return changed
}

async function request<T>(method: string, url: string, body?: unknown, keepalive = false): Promise<T> {
  const res = await fetch(url, {
    method,
    headers: body === undefined ? undefined : { 'content-type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
    credentials: 'same-origin',
    keepalive,
  })
  const data = (await res.json().catch(() => ({}))) as T & { error?: string }
  if (!res.ok) throw new AccountError(data.error ?? `ошибка сервера (${res.status})`, res.status)
  return data
}

export class AccountError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message)
    this.name = 'AccountError'
  }
}

/** Синхронизации идут строго по очереди. */
let queue: Promise<void> = Promise.resolve()

/** Отправить локальный прогресс и забрать слитый. Без входа — ничего не делает. */
export function sync(keepalive = false): Promise<void> {
  const run = async () => {
    const owner = cachedAccount()
    if (!owner) return
    try {
      const res = await request<{ user: Account; state: State }>('PUT', '/api/progress', { state: localState() }, keepalive)
      if (apply(res.state)) window.dispatchEvent(new Event(PROGRESS_EVENT))
    } catch (err) {
      // Сессия кончилась или её стёрли — считаем, что вышли, но прогресс в браузере не трогаем.
      if (err instanceof AccountError && err.status === 401) setOwner(null)
      else console.warn('go-piece: не удалось синхронизировать прогресс', err)
    }
  }
  queue = queue.then(run, run)
  return queue
}

let timer: number | undefined

function schedulePush(): void {
  if (!cachedAccount()) return
  window.clearTimeout(timer)
  timer = window.setTimeout(() => {
    timer = undefined
    void sync()
  }, PUSH_DELAY)
}

/** Человек вошёл (или зарегистрировался) — разобраться, чей прогресс в браузере, и слить. */
async function adopt(user: Account): Promise<void> {
  const prev = cachedAccount()
  if (prev && prev.id !== user.id) clearLocal()
  setOwner(user)
  await sync()
  // Достижения и отметки на странице пересчитаются по свежему прогрессу.
  window.dispatchEvent(new Event(PROGRESS_EVENT))
}

export async function register(email: string, password: string): Promise<Account> {
  const { user } = await request<{ user: Account }>('POST', '/api/auth/register', { email, password })
  await adopt(user)
  return user
}

export async function login(email: string, password: string): Promise<Account> {
  const { user } = await request<{ user: Account }>('POST', '/api/auth/login', { email, password })
  await adopt(user)
  return user
}

export async function logout(): Promise<void> {
  window.clearTimeout(timer)
  await sync()
  await request('POST', '/api/auth/logout', {}).catch(() => {})
  clearLocal()
  setOwner(null)
}

/**
 * Кука сессии httpOnly, из скрипта её не видно. Если в браузере нет отметки
 * о владельце (например, localStorage очистили), спрашиваем сервер.
 */
export async function refreshAccount(): Promise<Account | null> {
  try {
    const { user } = await request<{ user: Account | null }>('GET', '/api/auth/me')
    if (user) await adopt(user)
    else if (cachedAccount()) setOwner(null)
    return user
  } catch {
    return cachedAccount()
  }
}

/** Подключить синхронизацию на странице: вызывается один раз из BaseLayout. */
export function startSync(): void {
  window.addEventListener(PROGRESS_EVENT, schedulePush)
  window.addEventListener(EARNED_EVENT, schedulePush)
  // Уходим со страницы, не дождавшись паузы, — отправляем сразу, запрос переживёт выгрузку.
  window.addEventListener('pagehide', () => {
    if (timer === undefined) return
    window.clearTimeout(timer)
    timer = undefined
    void sync(true)
  })
  // Подтянуть прогресс с других устройств.
  void sync()
}
