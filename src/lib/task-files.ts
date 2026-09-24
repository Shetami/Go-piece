/**
 * Открытая часть задачи: заготовка и эталонное решение.
 *
 * Файлы подтягиваются на сборке, а не читаются с диска в рантайме, — иначе бы
 * это не пережило деплой на Vercel, где рядом с функцией лежит только бандл.
 *
 * Скрытые проверки сюда сознательно не попадают: они в task-secrets.ts, который
 * импортируют только серверные маршруты. Разделение не косметическое — стоит
 * острову утащить этот модуль на клиент, и открытым станет всё, что в нём есть.
 */

const starters = import.meta.glob('../content/tasks/**/starter.{go,sql}', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

const solutions = import.meta.glob('../content/tasks/**/solution.{go,sql}', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

/** Схема и данные SQL-задачи: заливаются в Postgres перед первым запросом. */
const schemas = import.meta.glob('../content/tasks/**/schema.sql', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

/** `../content/tasks/slices/append-aliasing/starter.go` → `slices/append-aliasing`. */
function taskIdOf(path: string, file: string): string | null {
  const m = path.match(new RegExp(`content/tasks/(.+)/${file.replace('.', '\\.')}$`))
  return m?.[1] ?? null
}

function byTaskId(files: Record<string, string>, name: string): Map<string, string> {
  const out = new Map<string, string>()
  for (const [path, source] of Object.entries(files)) {
    const id = taskIdOf(path, name)
    if (id) out.set(id, source)
  }
  return out
}

const STARTERS = new Map([...byTaskId(starters, 'starter.go'), ...byTaskId(starters, 'starter.sql')])
const SOLUTIONS = new Map([...byTaskId(solutions, 'solution.go'), ...byTaskId(solutions, 'solution.sql')])
const SCHEMAS = byTaskId(schemas, 'schema.sql')

export interface TaskCode {
  starter: string
  solution: string
}

/**
 * Код задачи. Отсутствие файла — ошибка сборки, а не пустой редактор:
 * задача без заготовки просто не работает, и узнать об этом лучше сразу.
 */
export function taskCode(id: string): TaskCode {
  const starter = STARTERS.get(id)
  if (starter === undefined) throw new Error(`задача ${id}: нет starter.go или starter.sql`)
  const solution = SOLUTIONS.get(id)
  if (solution === undefined) throw new Error(`задача ${id}: нет solution.go или solution.sql`)
  return { starter, solution }
}

/** Схема SQL-задачи. Для SQL её отсутствие — такая же ошибка сборки, как отсутствие заготовки. */
export function taskSchema(id: string): string {
  const schema = SCHEMAS.get(id)
  if (schema === undefined) throw new Error(`задача ${id}: нет schema.sql`)
  return schema
}

/** Все задачи, у которых есть код, — для проверки целостности в тестах. */
export function taskIdsWithCode(): string[] {
  return [...STARTERS.keys()].sort()
}
