import { defineCollection, reference } from 'astro:content'
import { glob } from 'astro/loaders'
import { z } from 'astro/zod'
import { TASK_KIND_IDS, TASK_LEVEL_IDS, TASK_TOPIC_IDS } from './data/practice.ts'

/**
 * Лекции. id — путь без расширения: `go-runtime/scheduler`.
 * Первый сегмент пути — курс, он же раздел навигации.
 */
const lectures = defineCollection({
  loader: glob({ pattern: '**/*.mdx', base: './src/content/lectures' }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    /** Порядок внутри курса. */
    order: z.number(),
    tags: z.array(z.string()).default([]),
    /** Сколько минут читать — пишется руками, честнее автоподсчёта для текста со стендами. */
    minutes: z.number().optional(),
    draft: z.boolean().default(false),
    updated: z.coerce.date().optional(),
  }),
})

/**
 * Справочник. Один термин — один файл, id = имя файла.
 * `<Term id="...">` в лекциях ссылается именно на этот id.
 */
const glossary = defineCollection({
  loader: glob({ pattern: '**/*.mdx', base: './src/content/glossary' }),
  schema: z.object({
    title: z.string(),
    /** Определение в одну строку — уходит во всплывающую подсказку. */
    short: z.string().max(220),
    /** Другие названия и написания — для поиска и для «см. также». */
    aliases: z.array(z.string()).default([]),
    category: z.enum(['сущность', 'механизм', 'настройка', 'состояние', 'инструмент']),
    related: z.array(reference('glossary')).default([]),
    /** Где смотреть в исходниках Go. */
    source: z.string().optional(),
  }),
})

/**
 * Задачи для практики. Одна задача — папка:
 *
 *   src/content/tasks/<тема>/<задача>/
 *     index.mdx     условие, а после `<Reveal>` — разбор
 *     starter.go    что лежит в редакторе при открытии
 *     solution.go   эталон; его же показывает кнопка «решение»
 *     check.go      скрытые тесты (только kind: implement) — на клиент не уезжают
 *     expect.txt    эталонный вывод (только kind: fix) — тоже остаётся на сервере
 *
 * id задачи — `<тема>/<задача>`, без хвостового `index`.
 * Код держим отдельными .go файлами, а не строками во фронтматтере: так их
 * правит настоящий редактор, проверяет настоящий gofmt и видит `go vet`.
 */
const tasks = defineCollection({
  loader: glob({
    pattern: '**/index.mdx',
    base: './src/content/tasks',
    generateId: ({ entry }) => entry.replace(/\/index\.mdx$/, ''),
  }),
  schema: z
    .object({
      title: z.string(),
      description: z.string(),
      /** В каком виде спрашивают: что выведет, найди баг, почини, реализуй. */
      kind: z.enum(TASK_KIND_IDS),
      /** О чём спрашивают. */
      topic: z.enum(TASK_TOPIC_IDS),
      level: z.enum(TASK_LEVEL_IDS).default('medium'),
      /** Порядок внутри темы. */
      order: z.number().default(0),
      /**
       * Строка со сломанным местом, считая с единицы, — только для kind: bug.
       * Считается по starter.go. Это и есть правильный ответ, поэтому
       * на страницу оно не попадает: сверяет сервер.
       */
      bugLine: z.number().int().positive().optional(),
      /** Термины справочника, которые стоит перечитать рядом с задачей. */
      related: z.array(reference('glossary')).default([]),
      /** Лекция, из которой эта задача растёт, если такая есть: `go-runtime/channels`. */
      lecture: z.string().optional(),
      draft: z.boolean().default(false),
    })
    .superRefine((data, ctx) => {
      // Задаче «найди баг» без правильной строки нечего проверять — ловим на сборке.
      if (data.kind === 'bug' && data.bugLine === undefined) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'для kind: bug нужен bugLine', path: ['bugLine'] })
      }
      if (data.kind !== 'bug' && data.bugLine !== undefined) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'bugLine имеет смысл только для kind: bug', path: ['bugLine'] })
      }
    }),
})

export const collections = { lectures, glossary, tasks }
