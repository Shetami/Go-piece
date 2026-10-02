import type { APIRoute } from 'astro'
import { buildCatalog } from '../../lib/catalog.ts'

/** Каталог для трекера достижений: собирается заранее, как и весь сайт. */
export const prerender = true

export const GET: APIRoute = async () =>
  new Response(JSON.stringify(await buildCatalog()), { headers: { 'Content-Type': 'application/json' } })
