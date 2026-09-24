/**
 * Тесты «Проверь себя». Модули подтягиваются на сборке, как и код задач практики:
 * `../data/quizzes/go-runtime/scheduler.ts` → тест лекции `go-runtime/scheduler`.
 */
import type { Quiz } from '../components/quiz/types.ts'

const modules = import.meta.glob('../data/quizzes/**/*.ts', { import: 'default', eager: true }) as Record<string, Quiz>

const QUIZZES = new Map<string, Quiz>()
for (const [path, quiz] of Object.entries(modules)) {
  const id = path.match(/data\/quizzes\/(.+)\.ts$/)?.[1]
  if (id) QUIZZES.set(id, quiz)
}

/** Тест лекции или undefined, если его пока нет. */
export function quizFor(lectureId: string): Quiz | undefined {
  return QUIZZES.get(lectureId)
}

export function allQuizzes(): Map<string, Quiz> {
  return QUIZZES
}

/** Ключ localStorage с результатом теста — общий для острова и страницы-сводки. */
export function quizStorageKey(lectureId: string): string {
  return `go-piece:quiz:${lectureId}`
}
