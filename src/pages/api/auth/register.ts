import type { APIRoute } from 'astro'
import { handleError, json, readJson } from '../../../lib/api.ts'
import { normalizeEmail, register, requirePassword, startSession } from '../../../lib/server/auth.ts'

export const prerender = false

/** POST /api/auth/register — завести аккаунт и сразу войти. Почта не подтверждается. */
export const POST: APIRoute = async ({ request, cookies, url }) => {
  try {
    const body = await readJson<{ email?: unknown; password?: unknown }>(request)
    const user = await register(normalizeEmail(body.email), requirePassword(body.password))
    startSession(user, cookies, url)
    return json({ user })
  } catch (err) {
    return handleError(err)
  }
}
