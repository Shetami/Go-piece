/**
 * Регистрация, вход и сессии. Почта без подтверждения, пароль хранится как
 * scrypt-хэш с солью, сессия — случайный токен в httpOnly-куке.
 *
 * Только сервер.
 */
import { createHash, randomBytes, scrypt, timingSafeEqual } from 'node:crypto'
import { promisify } from 'node:util'
import type { AstroCookies } from 'astro'
import { BadRequest } from '../api.ts'
import { getDb } from './db.ts'

const scryptAsync = promisify(scrypt) as (pw: string, salt: Buffer, len: number) => Promise<Buffer>

export const SESSION_COOKIE = 'go_piece_session'
const SESSION_DAYS = 30
const KEY_LEN = 64

export const MIN_PASSWORD = 8
const MAX_PASSWORD = 200

export interface User {
  id: number
  email: string
}

export function normalizeEmail(value: unknown): string {
  if (typeof value !== 'string') throw new BadRequest('укажите почту')
  const email = value.trim().toLowerCase()
  // Без подтверждения проверять глубже смысла нет: важно лишь, чтобы это было похоже на почту.
  if (email.length > 254 || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) throw new BadRequest('почта выглядит неправильно')
  return email
}

export function requirePassword(value: unknown): string {
  if (typeof value !== 'string' || value.length < MIN_PASSWORD) {
    throw new BadRequest(`пароль — не короче ${MIN_PASSWORD} символов`)
  }
  if (value.length > MAX_PASSWORD) throw new BadRequest(`пароль — не длиннее ${MAX_PASSWORD} символов`)
  return value
}

/** Формат: scrypt$<соль base64>$<хэш base64>. */
async function hashPassword(password: string): Promise<string> {
  const salt = randomBytes(16)
  const hash = await scryptAsync(password, salt, KEY_LEN)
  return `scrypt$${salt.toString('base64')}$${hash.toString('base64')}`
}

async function verifyPassword(password: string, stored: string): Promise<boolean> {
  const [scheme, salt, hash] = stored.split('$')
  if (scheme !== 'scrypt' || !salt || !hash) return false
  const expected = Buffer.from(hash, 'base64')
  const actual = await scryptAsync(password, Buffer.from(salt, 'base64'), expected.length)
  return timingSafeEqual(actual, expected)
}

/**
 * Хэш, с которым сравнивается пароль, когда такой почты нет: вход по
 * несуществующей почте идёт столько же времени, сколько по существующей.
 */
let dummyHash: Promise<string> | null = null

export async function register(email: string, password: string): Promise<User> {
  const db = getDb()
  if (db.prepare('SELECT 1 FROM users WHERE email = ?').get(email)) {
    throw new BadRequest('эта почта уже зарегистрирована', 409)
  }
  const hash = await hashPassword(password)
  try {
    const row = db
      .prepare('INSERT INTO users (email, password_hash, created_at) VALUES (?, ?, ?) RETURNING id')
      .get(email, hash, Date.now()) as { id: number }
    return { id: row.id, email }
  } catch (err) {
    // Две регистрации одной почты одновременно: вторая упирается в UNIQUE.
    if (String(err).includes('UNIQUE')) throw new BadRequest('эта почта уже зарегистрирована', 409)
    throw err
  }
}

export async function login(email: string, password: string): Promise<User> {
  const row = getDb().prepare('SELECT id, password_hash FROM users WHERE email = ?').get(email) as
    | { id: number; password_hash: string }
    | undefined
  dummyHash ??= hashPassword(randomBytes(16).toString('hex'))
  const ok = await verifyPassword(password, row?.password_hash ?? (await dummyHash))
  if (!row || !ok) throw new BadRequest('неверная почта или пароль', 401)
  return { id: row.id, email }
}

function tokenHash(token: string): string {
  return createHash('sha256').update(token).digest('hex')
}

export function startSession(user: User, cookies: AstroCookies, url: URL): void {
  const token = randomBytes(32).toString('base64url')
  const expires = Date.now() + SESSION_DAYS * 24 * 60 * 60 * 1000
  const db = getDb()
  db.prepare('DELETE FROM sessions WHERE expires_at < ?').run(Date.now())
  db.prepare('INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)').run(
    tokenHash(token),
    user.id,
    expires,
  )
  cookies.set(SESSION_COOKIE, token, {
    path: '/',
    httpOnly: true,
    sameSite: 'lax',
    // В Docker сайт может открываться по http — тогда Secure-куку браузер не сохранит.
    secure: url.protocol === 'https:',
    expires: new Date(expires),
  })
}

export function endSession(cookies: AstroCookies): void {
  const token = cookies.get(SESSION_COOKIE)?.value
  if (token) getDb().prepare('DELETE FROM sessions WHERE token_hash = ?').run(tokenHash(token))
  cookies.delete(SESSION_COOKIE, { path: '/' })
}

/** Пользователь по куке сессии или null. */
export function currentUser(cookies: AstroCookies): User | null {
  const token = cookies.get(SESSION_COOKIE)?.value
  if (!token) return null
  const row = getDb()
    .prepare(
      `SELECT u.id, u.email FROM sessions s JOIN users u ON u.id = s.user_id
       WHERE s.token_hash = ? AND s.expires_at > ?`,
    )
    .get(tokenHash(token), Date.now()) as { id: number; email: string } | undefined
  return row ? { id: row.id, email: row.email } : null
}

export function requireUser(cookies: AstroCookies): User {
  const user = currentUser(cookies)
  if (!user) throw new BadRequest('нужно войти', 401)
  return user
}
