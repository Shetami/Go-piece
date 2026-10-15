/**
 * Проверка задачи, которую написала модель, — до того, как её кто-то увидит.
 *
 * Модель пишет задачи уверенно и иногда неправильно: баг не в той строке,
 * вывод зависит от порядка обхода мапы, эталон не проходит собственные тесты.
 * Всё это ловится прогоном в той же песочнице go.dev, что проверяет решения
 * пользователей. Найденные проблемы уходят модели обратно текстом — она их
 * исправляет, — поэтому формулировки здесь написаны для неё.
 *
 * Заодно задача приводится к виду, который ждёт /api/check: код проходит
 * через goimports, для «почини» эталонный вывод получается запуском эталона
 * (а не берётся со слов модели), строка с багом пересчитывается после
 * форматирования.
 */
import type { TaskKindId } from '../../../data/practice.ts'
import { toChunks } from '../../api.ts'
import { compile, format, hasMain, MAX_CODE_BYTES, normalizeOutput } from '../../playground.ts'
import { runTests, streams } from '../checker.ts'

export interface Draft {
  starter: string
  solution: string
  tests: string
  bugLine: number
}

export interface Checked {
  starter: string
  solution: string
  tests: string | null
  expect: string | null
  bugLine: number | null
}

export type Validation = { ok: true; task: Checked } | { ok: false; problems: string[] }

interface Run {
  errors: string
  vet: string
  stdout: string
  stderr: string
}

/**
 * Прогон в песочнице. go.dev кэширует ответ на одинаковый исходник, и два
 * прогона одного кода всегда совпали бы — проверка детерминизма ничего бы не
 * проверяла. Поэтому в конец дописывается уникальный комментарий: он не
 * сдвигает строки и не меняет программу, но обходит кэш.
 */
async function run(code: string): Promise<Run> {
  const compiled = await compile(`${code}\n// go-piece: ${Date.now()}-${Math.random().toString(36).slice(2)}\n`, true)
  const { stdout, stderr } = streams(toChunks(compiled.Events))
  return { errors: compiled.Errors ?? '', vet: compiled.VetErrors ?? '', stdout, stderr }
}

/** Песочница сообщает об убитой по времени программе в stderr. */
function timedOut(r: Run): boolean {
  return /timeout running program|process took too long/i.test(r.stderr)
}

/** goimports: расставить импорты и отформатировать. Ошибка разбора — проблема задачи. */
async function tidy(code: string, name: string, problems: string[]): Promise<string | null> {
  if (new TextEncoder().encode(code).length > MAX_CODE_BYTES) {
    problems.push(`${name} длиннее ${MAX_CODE_BYTES} байт — сократи`)
    return null
  }
  if (!/^package main\b/m.test(code)) {
    problems.push(`${name} должен начинаться с package main`)
    return null
  }
  const f = await format(code, true)
  if (f.Error) {
    problems.push(`${name} не разбирается gofmt: ${f.Error}`)
    return null
  }
  return f.Body
}

/**
 * Где заготовка отличается от эталона: отрезок строк заготовки (с единицы,
 * включительно) после отбрасывания общего начала и общего конца. Пустой
 * отрезок (from > to) — в эталоне строки только добавлены.
 */
export function changedLines(starter: string, solution: string): { from: number; to: number } {
  const a = starter.split('\n')
  const b = solution.split('\n')
  let head = 0
  while (head < a.length && head < b.length && a[head] === b[head]) head++
  let tail = 0
  while (tail < a.length - head && tail < b.length - head && a[a.length - 1 - tail] === b[b.length - 1 - tail]) tail++
  return { from: head + 1, to: a.length - tail }
}

/**
 * Номер строки с багом после форматирования. Модель считала строки в своём
 * тексте; goimports мог добавить импорт и сдвинуть всё вниз. Ищем ту же
 * строку по содержимому.
 */
function remapLine(original: string, formatted: string, line: number): number | null {
  const want = original.split('\n')[line - 1]?.trim()
  if (!want) return null
  const lines = formatted.split('\n')
  if (lines[line - 1]?.trim() === want) return line
  const hits = lines.flatMap((l, i) => (l.trim() === want ? [i + 1] : []))
  return hits.length === 1 ? hits[0]! : null
}

export async function validate(kind: TaskKindId, draft: Draft): Promise<Validation> {
  const problems: string[] = []
  const starter = await tidy(draft.starter, 'starter', problems)
  const solution = kind === 'output' ? starter : await tidy(draft.solution, 'solution', problems)
  if (!starter || !solution) return { ok: false, problems }

  switch (kind) {
    case 'output':
      return validateOutput(starter, problems)
    case 'bug':
      return validateBug(draft, starter, solution, problems)
    case 'fix':
      return validateFix(starter, solution, problems)
    case 'implement':
      return validateImplement(draft.tests, starter, solution, problems)
  }
}

async function validateOutput(starter: string, problems: string[]): Promise<Validation> {
  if (!hasMain(starter)) problems.push('в задаче «что выведет» у программы должна быть func main')
  if (problems.length) return { ok: false, problems }

  // Три прогона: вывод, зависящий от планировщика или порядка мапы, честно угадать нельзя.
  const runs = await Promise.all([run(starter), run(starter), run(starter)])
  const a = runs[0]!
  if (a.errors) return { ok: false, problems: [`программа не компилируется:\n${a.errors}`] }
  if (timedOut(a)) problems.push('программа не уложилась во время песочницы — вероятно, зависает')
  if (normalizeOutput(a.stdout) === '') problems.push('программа ничего не печатает в stdout — отвечать нечего')
  if (runs.some((r) => r.stdout !== a.stdout)) {
    problems.push(
      'вывод отличается между запусками — он зависит от планирования горутин или порядка обхода мапы. ' +
        'Сделай вывод детерминированным: синхронизируй горутины или сортируй ключи',
    )
  }
  if (problems.length) return { ok: false, problems }
  return { ok: true, task: { starter, solution: starter, tests: null, expect: null, bugLine: null } }
}

async function validateBug(draft: Draft, starter: string, solution: string, problems: string[]): Promise<Validation> {
  if (!hasMain(starter) || !hasMain(solution)) problems.push('в задаче «найди баг» и у starter, и у solution должна быть func main')
  if (starter === solution) problems.push('starter и solution совпадают — бага нет')

  const bugLine = remapLine(draft.starter, starter, draft.bugLine)
  if (bugLine === null) {
    problems.push(`bugLine = ${draft.bugLine} не указывает на непустую строку starter (или такая строка в нём не одна)`)
  } else {
    const changed = changedLines(starter, solution)
    const width = changed.to - changed.from + 1
    // Если в эталоне строка только добавлена, баг — в соседней строке заготовки.
    const inside = width > 0 ? bugLine >= changed.from && bugLine <= changed.to : Math.abs(bugLine - changed.from) <= 1
    if (!inside) {
      problems.push(
        `bugLine = ${draft.bugLine}, но starter и solution различаются в строках ${changed.from}–${changed.to} ` +
          '(после goimports). Баг и исправление должны быть в одной строке',
      )
    }
    if (width > 3) problems.push(`исправление занимает ${width} строк — в «найди баг» оно должно быть в одной-двух`)
  }
  if (problems.length) return { ok: false, problems }

  const [broken, fixed, again] = await Promise.all([run(starter), run(solution), run(solution)])
  if (broken.errors) problems.push(`starter не компилируется — в «найди баг» код должен собираться:\n${broken.errors}`)
  if (fixed.errors) problems.push(`solution не компилируется:\n${fixed.errors}`)
  if (problems.length) return { ok: false, problems }

  if (fixed.stderr) problems.push(`solution падает:\n${fixed.stderr.slice(0, 500)}`)
  if (fixed.stdout !== again.stdout) problems.push('вывод solution отличается между запусками — сделай его детерминированным')
  if (broken.stdout === fixed.stdout && broken.stderr === fixed.stderr) {
    problems.push('starter и solution печатают одно и то же — баг никак не проявляется при запуске. Пусть он меняет вывод')
  }
  if (problems.length) return { ok: false, problems }
  return { ok: true, task: { starter, solution, tests: null, expect: null, bugLine } }
}

async function validateFix(starter: string, solution: string, problems: string[]): Promise<Validation> {
  if (!hasMain(starter) || !hasMain(solution)) problems.push('в задаче «почини» и у starter, и у solution должна быть func main')
  if (starter === solution) problems.push('starter и solution совпадают — чинить нечего')
  if (problems.length) return { ok: false, problems }

  const [broken, fixed, again] = await Promise.all([run(starter), run(solution), run(solution)])
  if (broken.errors) problems.push(`starter не компилируется. В «почини» код собирается, но ведёт себя неправильно:\n${broken.errors}`)
  if (fixed.errors) problems.push(`solution не компилируется:\n${fixed.errors}`)
  if (problems.length) return { ok: false, problems }

  if (fixed.stderr) problems.push(`solution падает или зависает:\n${fixed.stderr.slice(0, 500)}`)
  if (normalizeOutput(fixed.stdout) === '') problems.push('solution ничего не печатает — сверять не с чем')
  if (fixed.stdout !== again.stdout) problems.push('вывод solution отличается между запусками — сделай его детерминированным')
  if (fixed.vet) problems.push(`go vet ругается на solution:\n${fixed.vet}`)
  if (!broken.stderr && normalizeOutput(broken.stdout) === normalizeOutput(fixed.stdout)) {
    problems.push('starter уже печатает то же, что solution, — чинить нечего. Пусть сломанная версия печатает другое или падает')
  }
  if (problems.length) return { ok: false, problems }
  return { ok: true, task: { starter, solution, tests: null, expect: fixed.stdout, bugLine: null } }
}

async function validateImplement(rawTests: string, starter: string, solution: string, problems: string[]): Promise<Validation> {
  for (const [name, code] of [
    ['starter', starter],
    ['solution', solution],
    ['tests', rawTests],
  ] as const) {
    if (hasMain(code)) problems.push(`в ${name} не должно быть func main — проверка запускает код как go test`)
  }
  if (!/func Test\w+\(t \*testing\.T\)/.test(rawTests)) problems.push('в tests нет ни одной функции TestXxx(t *testing.T)')
  if (/^import\b/m.test(rawTests)) problems.push('в tests не должно быть import — импорты расставит goimports после склейки с кодом')
  if (problems.length) return { ok: false, problems }

  const tests = rawTests.trim()
  const [good, stub] = await Promise.all([runTests(solution, tests), runTests(starter, tests)])
  if (!good.verdict.pass) {
    problems.push(`solution не проходит tests — ${good.verdict.title}:\n${(good.verdict.detail ?? '').slice(0, 1500)}`)
  }
  if (!stub.built) {
    problems.push(`starter не собирается вместе с tests — сигнатуры в заготовке и в тестах расходятся:\n${(stub.verdict.detail ?? '').slice(0, 1000)}`)
  } else if (stub.verdict.pass) {
    problems.push('заготовка starter уже проходит tests — тесты ничего не проверяют. Добавь проверки, на которых заглушка падает')
  }
  if (problems.length) return { ok: false, problems }
  return { ok: true, task: { starter, solution, tests, expect: null, bugLine: null } }
}
