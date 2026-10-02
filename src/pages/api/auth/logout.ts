import type { APIRoute } from 'astro'
import { handleError, json } from '../../../lib/api.ts'
import { endSession } from '../../../lib/server/auth.ts'

export const prerender = false

/** POST /api/auth/logout — закончить сессию в этом браузере. */
export const POST: APIRoute = async ({ cookies }) => {
  try {
    endSession(cookies)
    return json({ ok: true })
  } catch (err) {
    return handleError(err)
  }
}
