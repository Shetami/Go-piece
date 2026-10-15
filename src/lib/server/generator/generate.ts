/**
 * Генерация задачи: Claude пишет, песочница проверяет, Claude исправляет.
 *
 * Один разговор на задачу. Первое сообщение — что нужно (тема, формат,
 * сложность) и пример задачи того же формата с сайта, чтобы было понятно,
 * какого уровня и тона ждём. Ответ — JSON по схеме TaskSchema (structured
 * outputs). Его проверяет validate.ts; если нашлись проблемы, они уходят
 * в тот же разговор следующим сообщением, и так до MAX_ROUNDS раз.
 *
 * Только сервер. Ключ API — ANTHROPIC_API_KEY в окружении.
 */
import Anthropic from '@anthropic-ai/sdk'
import { betaZodOutputFormat } from '@anthropic-ai/sdk/helpers/beta/zod'
import { z } from 'zod'
import { taskKind, taskLevel, taskTopic, type TaskKindId, type TaskLevelId, type TaskTopicId } from '../../../data/practice.ts'
import { validate, type Checked } from './validate.ts'

const MODEL = 'claude-opus-5-5'

/** Сколько раз модель может исправить задачу после замечаний проверки. */
const MAX_ROUNDS = 3

export const TaskSchema = z.object({
  title: z.string().describe('Название задачи по-русски, до 60 символов, без кавычек-ёлочек и без двоеточия в начале'),
  description: z.string().describe('Одно-два предложения по-русски: о чём задача и в чём подвох. Не выдавай ответ'),
  statement: z
    .string()
    .describe('Условие в Markdown по-русски. Без ответа и без подсказки, где баг. Код программы сюда не копируй — он в starter'),
  starter: z.string().describe('Что лежит в редакторе при открытии: полный файл Go, package main'),
  solution: z
    .string()
    .describe('Эталон: полный файл Go. Для output — копия starter; для bug и fix — исправленная программа; для implement — решение'),
  tests: z
    .string()
    .describe('Только для implement: скрытые тесты, package main, функции TestXxx(t *testing.T), без import. Для остальных — пустая строка'),
  bugLine: z
    .number()
    .int()
    .describe('Только для bug: номер строки starter (с единицы), где баг и где solution его исправляет. Для остальных — 0'),
  explanation: z
    .string()
    .describe('Разбор в Markdown по-русски: правильный ответ, почему так, какой механизм языка за этим стоит, как не попасться'),
})

export type GeneratedDraft = z.infer<typeof TaskSchema>

/** Пример задачи того же формата с сайта — эталон тона и уровня. */
export interface ExampleTask {
  title: string
  statement: string
  starter: string
  solution: string
  tests?: string | null
  expect?: string | null
  bugLine?: number
}

export interface GenerateRequest {
  topic: TaskTopicId
  kind: TaskKindId
  level: TaskLevelId
  example: ExampleTask | null
  /** Названия задач, которые уже есть: повторять их не надо. */
  avoid: string[]
}

export type Progress = (message: string) => void

export interface GenerateResult {
  draft: GeneratedDraft
  checked: Checked
}

export class GenerateError extends Error {
  constructor(
    message: string,
    /** Можно ли показать текст человеку как есть. */
    readonly visible = true,
  ) {
    super(message)
    this.name = 'GenerateError'
  }
}

export function generatorConfigured(): boolean {
  return Boolean(process.env.ANTHROPIC_API_KEY || process.env.ANTHROPIC_AUTH_TOKEN)
}

let client: Anthropic | null = null
function anthropic(): Anthropic {
  client ??= new Anthropic()
  return client
}

/**
 * Системный промпт не меняется от запроса к запросу — так он кэшируется.
 * Всё, что зависит от запроса, идёт в первое сообщение.
 */
const SYSTEM = `Ты пишешь задачи для раздела «Практика» сайта Go-piece — тренажёра собеседований по Go. Пользователь сайта открывает задачу, решает её в редакторе на странице, код исполняет песочница go.dev (настоящий компилятор Go, последняя версия).

Задача — как на живом собеседовании: короткая, по делу, с одним настоящим подвохом, который проверяет понимание механизма языка или рантайма, а не внимательность к опечаткам. Хорошая задача учит: после разбора человек понимает что-то про Go, чего не понимал.

Форматы задач:

- output («Что выведет?»). Рабочая программа с func main. Пользователь пишет её вывод, сайт запускает программу и сравнивает stdout. Вывод обязан быть детерминированным: никакой зависимости от планирования горутин, порядка обхода map, времени, случайных чисел, адресов в памяти. solution — точная копия starter.
- bug («Найди баг»). Программа с func main компилируется и работает, но делает не то, что задумано. Пользователь кликает по строке с ошибкой. bugLine — номер этой строки в starter. solution — та же программа с исправлением ровно в этой строке (изменение в одну, максимум две соседние строки). Баг должен проявляться при запуске: starter и solution печатают разное (или starter падает). В условии опиши, что программа должна делать, но не намекай, где ошибка. Не оставляй в коде комментариев, которые выдают баг.
- fix («Почини»). Программа с func main компилируется, но падает, зависает (дедлок) или печатает не то. Пользователь правит код, пока вывод не совпадёт с эталонным. solution — исправленная программа; её вывод и есть эталон, поэтому в условии опиши, что она должна печатать. Вывод solution детерминированный, без паники.
- implement («Реализуй»). Заготовка функции (или типа с методами) с doc-комментарием, тело — заглушка, которая компилируется и возвращает нулевые значения. Без func main ни в starter, ни в solution, ни в tests. tests — скрытые тесты: package main, функции TestXxx(t *testing.T), без import (импорты расставит goimports при склейке с кодом пользователя). Тесты проверяют контракт, граничные случаи и тот подвох, ради которого задача; сообщения t.Errorf/t.Fatalf — по-русски и показывают, на каком входе ошибка. Заглушка из starter обязана тесты не проходить. Для конкурентных задач тесты должны быть детерминированными и укладываться в несколько секунд.

Ограничения песочницы: нет сети и файловой системы, время виртуальное (time.Sleep ничего не стоит, часы начинаются с 2009-11-10), программа должна завершиться за пару секунд, стандартная библиотека — только она. Код — по gofmt, идиоматичный Go, имена по-английски, комментарии по-русски и только там, где они помогают.

Тексты (title, description, statement, explanation) — по-русски, просто и точно, без воды и восклицаний. statement — что дано и что спрашивают; если нужно, что проверяют тесты. explanation — разбор: правильный ответ, механизм, который за ним стоит, типичная ошибка, как её не допустить. Код в Markdown оформляй блоками \`\`\`go.

Уровни сложности: easy — разминка на один механизм; medium — обычный вопрос собеседования, где легко ошибиться; hard — злая задача, где взаимодействуют несколько механизмов или нужен опыт с рантаймом.`

function firstMessage(req: GenerateRequest): string {
  const kind = taskKind(req.kind)
  const topic = taskTopic(req.topic)
  const level = taskLevel(req.level)

  const parts = [
    `Напиши новую задачу.`,
    ``,
    `Формат: ${req.kind} («${kind.title}»)`,
    `Тема: ${topic.title} — ${topic.blurb}`,
    `Сложность: ${req.level} («${level.title}»)`,
  ]

  if (req.example) {
    const e = req.example
    parts.push(
      ``,
      `Пример задачи этого формата с сайта — ориентир по тону, объёму и уровню. Не повторяй её, придумай другую:`,
      `<example>`,
      `title: ${e.title}`,
      `statement:\n${e.statement.trim()}`,
      `starter:\n${e.starter.trim()}`,
    )
    if (req.kind !== 'output') parts.push(`solution:\n${e.solution.trim()}`)
    if (e.tests) parts.push(`tests:\n${e.tests.trim()}`)
    if (e.bugLine) parts.push(`bugLine: ${e.bugLine}`)
    parts.push(`</example>`)
  }

  if (req.avoid.length > 0) {
    parts.push(``, `Такие задачи уже есть — не повторяй их сюжет и подвох:`, ...req.avoid.map((t) => `- ${t}`))
  }
  return parts.join('\n')
}

function feedback(problems: string[]): string {
  return [
    'Автоматическая проверка задачи в песочнице go.dev нашла проблемы:',
    '',
    ...problems.map((p) => `- ${p}`),
    '',
    'Исправь задачу и верни её целиком. Если проблема в самой идее (например, вывод недетерминирован по природе), смени идею.',
  ].join('\n')
}

export async function generateTask(req: GenerateRequest, progress: Progress): Promise<GenerateResult> {
  const messages: Anthropic.Beta.BetaMessageParam[] = [{ role: 'user', content: firstMessage(req) }]

  for (let round = 1; round <= MAX_ROUNDS; round++) {
    progress(round === 1 ? 'Claude придумывает задачу…' : `Claude исправляет задачу (попытка ${round} из ${MAX_ROUNDS})…`)

    const message = await ask(messages)
    // Весь ответ, вместе с блоками размышлений, — обратно в историю: так модель
    // продолжает тот же ход мысли, когда получает замечания.
    messages.push({ role: 'assistant', content: message.content })

    const draft = message.parsed_output
    if (!draft) throw new GenerateError('модель вернула ответ не по схеме', false)

    progress(`Проверяю «${draft.title}» в песочнице go.dev…`)
    const result = await validate(req.kind, draft)
    if (result.ok) return { draft, checked: result.task }

    progress(`Проверка нашла проблемы: ${result.problems.length}`)
    messages.push({ role: 'user', content: feedback(result.problems) })
  }

  throw new GenerateError(`за ${MAX_ROUNDS} попытки не получилось задачи, которая проходит проверку. Попробуйте ещё раз или выберите другую тему`)
}

async function ask(messages: Anthropic.Beta.BetaMessageParam[]) {
  let message
  try {
    const stream = anthropic().beta.messages.stream({
      model: MODEL,
      max_tokens: 32000,
      // Если основная модель откажется (ложное срабатывание классификатора на
      // коде с горутинами и паниками), запрос выполнит запасная.
      betas: ['server-side-fallback-2026-07-01'],
      fallbacks: 'default',
      thinking: { type: 'adaptive' },
      output_config: { effort: 'high', format: betaZodOutputFormat(TaskSchema) },
      system: [{ type: 'text', text: SYSTEM, cache_control: { type: 'ephemeral' } }],
      messages,
    })
    message = await stream.finalMessage()
  } catch (err) {
    if (err instanceof Anthropic.RateLimitError) throw new GenerateError('генератор сейчас перегружен — попробуйте через минуту')
    if (err instanceof Anthropic.AuthenticationError) throw new GenerateError('генерация не настроена: ключ API не принят', false)
    if (err instanceof Anthropic.APIError) throw new GenerateError(`Claude API ответил ошибкой ${err.status}`, false)
    throw err
  }

  if (message.stop_reason === 'refusal') throw new GenerateError('модель отказалась писать эту задачу — попробуйте другую тему или формат')
  if (message.stop_reason === 'max_tokens') throw new GenerateError('задача получилась слишком длинной — попробуйте ещё раз')
  return message
}
