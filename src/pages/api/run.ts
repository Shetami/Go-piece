import type { APIRoute } from 'astro'
import { handleError, json, readJson, requireCode, toChunks } from '../../lib/api.ts'
import { compile, hasMain } from '../../lib/playground.ts'
import type { RunResult } from '../../components/practice/protocol.ts'

export const prerender = false

/**
 * POST /api/run — просто запустить код. Никаких ответов и вердиктов: это
 * песочница, и ею пользуются, пока думают над задачей.
 *
 * Отдельно от /api/check намеренно. Проверка знает правильный ответ и потому
 * обязана держать код у себя; прогон не знает ничего и работает с чем угодно,
 * в том числе на странице задачи, которой ещё нет.
 */
export const POST: APIRoute = async ({ request }) => {
  try {
    const body = await readJson<{ code?: unknown }>(request)
    let code = requireCode(body.code)

    /*
     * У задач «реализуй» в заготовке нет main — и не должно быть: проверка
     * запускает её как go test. Но пока человек пишет тело функции, ему нужен
     * компилятор, а песочница без main откажется собирать. Дописываем пустой
     * main и честно говорим об этом в ответе.
     */
    const stubbedMain = !hasMain(code)
    if (stubbedMain) code = `${code.replace(/\s*$/, '')}\n\nfunc main() {}\n`

    const result = await compile(code, true)

    const payload: RunResult = {
      errors: result.Errors ?? '',
      vet: result.VetErrors ?? '',
      output: toChunks(result.Events),
      stubbedMain,
    }
    return json(payload)
  } catch (err) {
    return handleError(err)
  }
}
