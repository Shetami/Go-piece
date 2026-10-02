/**
 * Каталог для достижений: что вообще есть на сайте. Только сборка — тянет
 * коллекции. На клиент он попадает готовым JSON: пропсом острова /achievements/
 * и файлом /achievements/catalog.json для трекера на остальных страницах.
 */
import { getCollection } from 'astro:content'
import type { Catalog } from './achievements.ts'
import { COURSES } from '../data/courses.ts'
import { taskLang } from '../data/practice.ts'
import { allQuizzes } from './quizzes.ts'

/** Стенд = папка в components/stands с обёрткой <Имя>Stand.astro; ui — общие части, не стенд. */
const standModules = import.meta.glob('../components/stands/*/*Stand.astro')

export async function buildCatalog(): Promise<Catalog> {
  const lectures = (await getCollection('lectures', (e) => !e.data.draft)).sort((a, b) => a.data.order - b.data.order)
  const tasks = (await getCollection('tasks', (t) => !t.data.draft)).sort((a, b) => a.id.localeCompare(b.id))
  const terms = await getCollection('glossary')
  const lectureIds = new Set(lectures.map((l) => l.id))

  return {
    courses: COURSES.map((c) => ({
      id: c.id,
      title: c.title,
      lectures: lectures.filter((l) => l.id.startsWith(`${c.id}/`)).map((l) => l.id),
    })).filter((c) => c.lectures.length > 0),
    quizzes: [...allQuizzes().keys()].filter((id) => lectureIds.has(id)).sort(),
    tasks: tasks.map((t) => ({
      id: t.id,
      topic: t.data.topic,
      lang: taskLang(t.data.topic),
      kind: t.data.kind,
      level: t.data.level,
    })),
    stands: Object.keys(standModules)
      .map((p) => p.match(/stands\/([^/]+)\//)?.[1])
      .filter((s): s is string => s !== undefined)
      .sort(),
    terms: terms.length,
  }
}
