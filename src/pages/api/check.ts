import type { APIRoute } from 'astro'
import { getCollection } from 'astro:content'
import { BadRequest, handleError, json, readJson, requireString } from '../../lib/api.ts'
import { taskCode } from '../../lib/task-files.ts'
import { taskCheck, taskExpect } from '../../lib/task-secrets.ts'
import { taskLang } from '../../data/practice.ts'
import { checkSubmission, type TaskSecrets } from '../../lib/server/checker.ts'
import { GENERATED_PREFIX, generatedSecrets } from '../../lib/server/generated.ts'
import type { CheckRequest } from '../../components/practice/protocol.ts'

export const prerender = false

/**
 * POST /api/check — проверить решение.
 *
 * Здесь живут ответы, поэтому маршрут сам достаёт всё, что нужно для проверки:
 * эталонный вывод, скрытые тесты, номер сломанной строки. Клиент присылает
 * только то, что сделал человек, и не может повлиять на критерий. Сама
 * механика проверки — в src/lib/server/checker.ts.
 *
 * Задачи бывают двух происхождений: из src/content/tasks (id `<тема>/<задача>`)
 * и сгенерированные на сайте (id `gen/<код>`), которые лежат в базе.
 */
export const POST: APIRoute = async ({ request }) => {
  try {
    const body = await readJson<CheckRequest>(request)
    const id = requireString(body.task, 'task')
    const secrets = id.startsWith(GENERATED_PREFIX) ? generatedSecrets(id) : await contentSecrets(id)
    return json(await checkSubmission(secrets, body))
  } catch (err) {
    return handleError(err)
  }
}

async function contentSecrets(id: string): Promise<TaskSecrets> {
  const tasks = await getCollection('tasks')
  const task = tasks.find((t) => t.id === id)
  if (!task) throw new BadRequest(`нет такой задачи: ${id}`, 404)

  // SQL-задачи проверяются в браузере, где живёт их Postgres. Сюда приходят
  // только «найди баг» — номеру строки всё равно, на каком он языке.
  if (taskLang(task.data.topic) === 'sql' && task.data.kind !== 'bug') {
    throw new BadRequest('SQL-задачи этого типа проверяются в браузере')
  }

  return {
    kind: task.data.kind,
    starter: taskCode(id).starter,
    bugLine: task.data.bugLine,
    check: task.data.kind === 'implement' ? taskCheck(id) : null,
    expect: task.data.kind === 'fix' ? taskExpect(id) : null,
  }
}
