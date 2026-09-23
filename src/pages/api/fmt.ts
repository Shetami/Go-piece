import type { APIRoute } from 'astro'
import { BadRequest, handleError, json, readJson, requireCode } from '../../lib/api.ts'
import { format } from '../../lib/playground.ts'

export const prerender = false

/**
 * POST /api/fmt — привести код к виду gofmt и дописать недостающие импорты.
 *
 * Форматирует настоящий gofmt на go.dev, а не наша догадка о нём: в Go
 * форматирование каноническое, и «почти gofmt» здесь хуже, чем ничего.
 * Неразобравшийся код — не ошибка сервера: так и отвечаем, с текстом от gofmt,
 * чтобы редактор подчеркнул место.
 */
export const POST: APIRoute = async ({ request }) => {
  try {
    const body = await readJson<{ code?: unknown }>(request)
    const code = requireCode(body.code)

    const result = await format(code, true)
    if (result.Error) return json({ error: result.Error, kind: 'syntax' }, 422)
    if (!result.Body) throw new BadRequest('gofmt вернул пустой ответ', 502)

    return json({ code: result.Body })
  } catch (err) {
    return handleError(err)
  }
}
