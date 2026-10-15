import type { APIRoute } from 'astro'
import { getCollection } from 'astro:content'
import { BadRequest, handleError, json, readJson } from '../../lib/api.ts'
import { currentUser } from '../../lib/server/auth.ts'
import {
  dailyLimit,
  listGenerated,
  recentTitles,
  recordAttempt,
  saveGenerated,
  usedToday,
} from '../../lib/server/generated.ts'
import { GenerateError, generateTask, generatorConfigured, type ExampleTask } from '../../lib/server/generator/generate.ts'
import { taskCode } from '../../lib/task-files.ts'
import { taskCheck } from '../../lib/task-secrets.ts'
import {
  TASK_KIND_IDS,
  TASK_LEVEL_IDS,
  TASK_TOPICS,
  taskLang,
  type TaskKindId,
  type TaskLevelId,
  type TaskTopicId,
} from '../../data/practice.ts'

export const prerender = false

/**
 * /api/generate — задачи, которые Claude пишет по запросу.
 *
 * GET  — мои сгенерированные задачи и сколько генераций осталось на сутки.
 * POST — сгенерировать новую: { topic, kind, level }. Генерация занимает
 *        от полуминуты до пары минут, поэтому ответ — поток NDJSON:
 *        { type: 'progress', message } … и в конце { type: 'done', id, url }
 *        или { type: 'error', message }.
 *
 * Только для вошедших: каждая генерация стоит денег. Только Go: SQL-задачи
 * проверяются в браузере, и проверить сгенерированную на сервере нечем.
 */

export const GET: APIRoute = async ({ cookies }) => {
  try {
    const user = currentUser(cookies)
    if (!user) return json({ user: null, configured: generatorConfigured(), limit: dailyLimit(), used: 0, tasks: [] })
    return json({
      user,
      configured: generatorConfigured(),
      limit: dailyLimit(),
      used: usedToday(user.id),
      tasks: listGenerated(user.id),
    })
  } catch (err) {
    return handleError(err)
  }
}

/** Кто сейчас генерирует. Одна генерация на человека за раз. */
const running = new Set<number>()

export const POST: APIRoute = async ({ request, cookies }) => {
  let userId: number
  let params: { topic: TaskTopicId; kind: TaskKindId; level: TaskLevelId }
  try {
    const user = currentUser(cookies)
    if (!user) throw new BadRequest('чтобы генерировать задачи, войдите', 401)
    if (!generatorConfigured()) throw new BadRequest('генерация задач на этом сервере не настроена', 503)
    params = readParams(await readJson(request))
    if (running.has(user.id)) throw new BadRequest('предыдущая задача ещё генерируется', 409)
    if (usedToday(user.id) >= dailyLimit()) {
      throw new BadRequest(`лимит — ${dailyLimit()} генераций в сутки. Попробуйте позже`, 429)
    }
    userId = user.id
  } catch (err) {
    return handleError(err)
  }

  running.add(userId)
  recordAttempt(userId)

  const encoder = new TextEncoder()
  const body = new ReadableStream<Uint8Array>({
    async start(controller) {
      const send = (event: object) => {
        try {
          controller.enqueue(encoder.encode(`${JSON.stringify(event)}\n`))
        } catch {
          // Человек закрыл вкладку — генерацию доводим до конца, задача останется в его списке.
        }
      }
      try {
        const avoid = [...(await contentTitles(params.topic, params.kind)), ...recentTitles(params.topic, params.kind)]
        const { draft, checked } = await generateTask(
          { ...params, example: await pickExample(params.topic, params.kind), avoid },
          (message) => send({ type: 'progress', message }),
        )
        const code = saveGenerated(userId, {
          ...params,
          title: draft.title.trim(),
          description: draft.description.trim(),
          statement: draft.statement.trim(),
          explanation: draft.explanation.trim(),
          ...checked,
        })
        send({ type: 'done', id: `gen/${code}`, url: `/practice/gen/${code}/` })
      } catch (err) {
        if (!(err instanceof GenerateError) || !err.visible) console.error('generate:', err)
        const message = err instanceof GenerateError && err.visible ? err.message : 'не получилось сгенерировать задачу'
        send({ type: 'error', message })
      } finally {
        running.delete(userId)
        try {
          controller.close()
        } catch {
          // Уже закрыт.
        }
      }
    },
  })

  return new Response(body, {
    headers: {
      'content-type': 'application/x-ndjson; charset=utf-8',
      'cache-control': 'no-store',
      // Иначе прокси вроде nginx копит поток и отдаёт его разом в конце.
      'x-accel-buffering': 'no',
    },
  })
}

function readParams(body: Record<string, unknown>): { topic: TaskTopicId; kind: TaskKindId; level: TaskLevelId } {
  const topic = TASK_TOPICS.find((t) => t.id === body.topic)
  if (!topic) throw new BadRequest('неизвестная тема')
  if (taskLang(topic.id) !== 'go') throw new BadRequest('генерировать можно только задачи на Go')
  if (!TASK_KIND_IDS.includes(body.kind as TaskKindId)) throw new BadRequest('неизвестный формат')
  if (!TASK_LEVEL_IDS.includes(body.level as TaskLevelId)) throw new BadRequest('неизвестная сложность')
  return { topic: topic.id, kind: body.kind as TaskKindId, level: body.level as TaskLevelId }
}

async function goTasks(kind: TaskKindId) {
  return getCollection('tasks', (t) => !t.data.draft && t.data.kind === kind && taskLang(t.data.topic) === 'go')
}

async function contentTitles(topic: TaskTopicId, kind: TaskKindId): Promise<string[]> {
  return (await goTasks(kind)).filter((t) => t.data.topic === topic).map((t) => t.data.title)
}

/**
 * Пример для модели — случайная задача того же формата, лучше той же темы.
 * Случайная, чтобы генерации не тянулись к одному и тому же образцу.
 */
async function pickExample(topic: TaskTopicId, kind: TaskKindId): Promise<ExampleTask | null> {
  const all = await goTasks(kind)
  const sameTopic = all.filter((t) => t.data.topic === topic)
  const pool = sameTopic.length > 0 ? sameTopic : all
  const entry = pool[Math.floor(Math.random() * pool.length)]
  if (!entry) return null
  const { starter, solution } = taskCode(entry.id)
  return {
    title: entry.data.title,
    // Условие — всё, что до разбора.
    statement: (entry.body ?? '').split('<Reveal>')[0] ?? '',
    starter,
    solution,
    tests: kind === 'implement' ? taskCheck(entry.id) : null,
    bugLine: entry.data.bugLine,
  }
}
