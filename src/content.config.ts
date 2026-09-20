import { defineCollection, reference } from 'astro:content'
import { glob } from 'astro/loaders'
import { z } from 'astro/zod'

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

export const collections = { lectures, glossary }
