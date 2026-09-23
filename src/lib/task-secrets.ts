/**
 * Закрытая часть задачи: скрытые тесты и эталонный вывод.
 *
 * Это ответы. Импортировать модуль можно только из src/pages/api/* — они
 * собираются в серверную функцию и на клиент не уезжают. Ни одна страница и ни
 * один React-остров не должны его трогать; тест practice.test.ts это стережёт.
 */

const checks = import.meta.glob('../content/tasks/**/check.go', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

const expects = import.meta.glob('../content/tasks/**/expect.txt', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

function byTaskId(files: Record<string, string>, name: string): Map<string, string> {
  const out = new Map<string, string>()
  for (const [path, source] of Object.entries(files)) {
    const m = path.match(new RegExp(`content/tasks/(.+)/${name.replace('.', '\\.')}$`))
    if (m?.[1]) out.set(m[1], source)
  }
  return out
}

const CHECKS = byTaskId(checks, 'check.go')
const EXPECTS = byTaskId(expects, 'expect.txt')

/** Скрытые тесты задачи `implement` — дописываются к коду пользователя. */
export function taskCheck(id: string): string | null {
  return CHECKS.get(id) ?? null
}

/** Эталонный вывод задачи `fix` — с ним сверяется то, что напечатал пользователь. */
export function taskExpect(id: string): string | null {
  return EXPECTS.get(id) ?? null
}
