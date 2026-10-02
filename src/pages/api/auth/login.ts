import type { APIRoute } from 'astro'
import { BadRequest, handleError, json, readJson } from '../../../lib/api.ts'
import { login, normalizeEmail, startSession } from '../../../lib/server/auth.ts'

export const prerender = false

/** POST /api/auth/login — войти по почте и паролю. */
export const POST: APIRoute = async ({ request, cookies, url }) => {
  try {
    const body = await readJson<{ email?: unknown; password?: unknown }>(request)
    if (typeof body.password !== 'string' || body.password === '') throw new BadRequest('введите пароль')
    const user = await login(normalizeEmail(body.email), body.password)
    startSession(user, cookies, url)
    return json({ user })
  } catch (err) {
    return handleError(err)
  }
}
