import type { APIRoute } from 'astro'
import type { DatabaseSync } from 'node:sqlite'
import { BadRequest, handleError, json, readJson } from '../../lib/api.ts'
import { requireUser } from '../../lib/server/auth.ts'
import { getDb, transaction } from '../../lib/server/db.ts'
import { MAX_KEYS, MAX_TOTAL_BYTES, MAX_VALUE_BYTES, isSyncedKey, mergeState, type State } from '../../lib/sync.ts'

export const prerender = false

/**
 * Прогресс аккаунта: решённые задачи, результаты тестов, достижения.
 *
 * Браузер присылает всё, что у него есть, сервер сливает это со своей копией
 * (src/lib/sync.ts) и возвращает итог — им браузер и дополняет localStorage.
 * Поэтому отдельного «скачать» нет: GET — то же самое с пустым состоянием.
 */

function load(db: DatabaseSync, userId: number): State {
  const rows = db.prepare('SELECT key, value FROM user_state WHERE user_id = ?').all(userId) as {
    key: string
    value: string
  }[]
  return Object.fromEntries(rows.map((r) => [r.key, r.value]))
}

function parseState(value: unknown): State {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new BadRequest('поле state — объект')
  const entries = Object.entries(value)
  if (entries.length > MAX_KEYS) throw new BadRequest('слишком много ключей', 413)
  const state: State = {}
  let total = 0
  for (const [key, v] of entries) {
    // Чужие ключи молча пропускаем: в localStorage бывает всякое, а сервер хранит только прогресс.
    if (!isSyncedKey(key) || typeof v !== 'string') continue
    total += v.length
    if (v.length > MAX_VALUE_BYTES || total > MAX_TOTAL_BYTES) throw new BadRequest('прогресс слишком большой', 413)
    state[key] = v
  }
  return state
}

export const GET: APIRoute = async ({ cookies }) => {
  try {
    const user = requireUser(cookies)
    return json({ user, state: load(getDb(), user.id) })
  } catch (err) {
    return handleError(err)
  }
}

export const PUT: APIRoute = async ({ request, cookies }) => {
  try {
    const user = requireUser(cookies)
    const incoming = parseState((await readJson<{ state?: unknown }>(request)).state)

    const state = transaction((db) => {
      const stored = load(db, user.id)
      const merged = mergeState(stored, incoming)

      const upsert = db.prepare(
        `INSERT INTO user_state (user_id, key, value, updated_at) VALUES (?, ?, ?, ?)
         ON CONFLICT (user_id, key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
      )
      const now = Date.now()
      for (const [key, value] of Object.entries(merged)) {
        if (stored[key] !== value) upsert.run(user.id, key, value, now)
      }
      return merged
    })

    return json({ user, state })
  } catch (err) {
    return handleError(err)
  }
}
