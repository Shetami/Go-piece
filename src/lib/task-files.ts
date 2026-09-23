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

const starters = import.meta.glob('../content/tasks/**/starter.go', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

const solutions = import.meta.glob('../content/tasks/**/solution.go', {
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

const STARTERS = byTaskId(starters, 'starter.go')
const SOLUTIONS = byTaskId(solutions, 'solution.go')

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
  if (starter === undefined) throw new Error(`задача ${id}: нет starter.go`)
  const solution = SOLUTIONS.get(id)
  if (solution === undefined) throw new Error(`задача ${id}: нет solution.go`)
  return { starter, solution }
}

/** Все задачи, у которых есть код, — для проверки целостности в тестах. */
export function taskIdsWithCode(): string[] {
  return [...STARTERS.keys()].sort()
}
