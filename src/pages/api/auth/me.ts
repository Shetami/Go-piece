import type { APIRoute } from 'astro'
import { handleError, json } from '../../../lib/api.ts'
import { currentUser } from '../../../lib/server/auth.ts'

export const prerender = false

/** GET /api/auth/me — кто вошёл в этом браузере: { user } или { user: null }. */
export const GET: APIRoute = async ({ cookies }) => {
  try {
    return json({ user: currentUser(cookies) })
  } catch (err) {
    return handleError(err)
  }
}
