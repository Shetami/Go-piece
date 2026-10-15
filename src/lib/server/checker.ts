/**
 * Проверка решения задачи на Go — без привязки к тому, откуда задача взялась.
 *
 * Задачи из src/content/tasks и задачи, сгенерированные на сайте, проверяются
 * одинаково; разница только в том, где лежат их скрытые части. Маршрут
 * /api/check достаёт их и передаёт сюда, генератор пользуется теми же
 * функциями, чтобы убедиться, что новая задача вообще решается.
 *
 * Механика у каждого типа своя:
 *   output    — запускается НЕТРОНУТАЯ заготовка, её вывод сверяется с ответом;
 *   bug       — сверяется номер строки, в песочницу не ходим вовсе;
 *   fix       — запускается код пользователя, вывод сверяется с эталонным;
 *   implement — к коду дописываются скрытые тесты, и всё это идёт как go test.
 *
 * Только сервер: здесь в руках ответы.
 */
import type { TaskKindId } from '../../data/practice.ts'
import { normalize, sameOutput, type CheckResult, type OutputChunk, type Verdict } from '../../components/practice/protocol.ts'
import { BadRequest, requireCode, toChunks } from '../api.ts'
import { compile, format, hasMain, type CompileResult } from '../playground.ts'

/** Всё, что нужно для проверки. Поля, не нужные этому типу задачи, не заполняются. */
export interface TaskSecrets {
  kind: TaskKindId
  /** Заготовка — по ней выносится вердикт в «что выведет». */
  starter: string
  bugLine?: number
  /** Скрытые тесты задачи «реализуй». */
  check?: string | null
  /** Эталонный вывод задачи «почини». */
  expect?: string | null
}

export interface Submission {
  code?: unknown
  answer?: unknown
  line?: unknown
}

export async function checkSubmission(task: TaskSecrets, body: Submission): Promise<CheckResult> {
  switch (task.kind) {
    case 'output':
      return checkOutput(task.starter, body.answer)
    case 'bug':
      return checkBug(task.bugLine!, body.line)
    case 'fix':
      if (task.expect == null) throw new BadRequest('у задачи нет эталонного вывода', 500)
      return checkFix(task.expect, requireCode(body.code))
    case 'implement':
      if (task.check == null) throw new BadRequest('у задачи нет скрытых тестов', 500)
      return checkImplement(task.check, requireCode(body.code))
  }
}

/** Пустой прогон — для вердиктов, которым песочница не нужна. */
const NO_RUN = { errors: '', vet: '', output: [] as OutputChunk[], stubbedMain: false }

export function streams(output: OutputChunk[]) {
  const pick = (kind: 'stdout' | 'stderr') =>
    output
      .filter((c) => c.kind === kind)
      .map((c) => c.text)
      .join('')
  return { stdout: pick('stdout'), stderr: pick('stderr') }
}

function asResult(compiled: CompileResult, verdict: Verdict, stubbedMain = false): CheckResult {
  return {
    errors: compiled.Errors ?? '',
    vet: compiled.VetErrors ?? '',
    output: toChunks(compiled.Events),
    stubbedMain,
    verdict,
  }
}

/**
 * «Что выведет программа». Запускаем заготовку с сервера, а не то, что сейчас
 * в редакторе: человек мог поставить туда что угодно, в том числе ровно тот
 * вывод, который написал в ответе.
 *
 * Сверяем только stdout. Если программа вдобавок упала, дамп горутин в ответе
 * требовать бессмысленно — покажем его, но спрашивать не будем.
 */
async function checkOutput(starter: string, answer: unknown): Promise<CheckResult> {
  if (typeof answer !== 'string') throw new BadRequest('поле answer должно быть строкой')

  const compiled = await compile(starter, true)
  if (compiled.Errors) {
    // Заготовка задачи не собирается — это наша поломка, а не ошибка отвечающего.
    return asResult(compiled, { pass: false, title: 'Заготовка задачи не компилируется — это баг сайта', detail: compiled.Errors })
  }

  const { stdout } = streams(toChunks(compiled.Events))
  const pass = sameOutput(answer, stdout)

  return asResult(compiled, {
    pass,
    title: pass ? 'Верно' : 'Не сходится',
    detail: pass ? undefined : `вы написали:\n${normalize(answer) || '(пусто)'}\n\nпрограмма напечатала:\n${normalize(stdout) || '(пусто)'}`,
  })
}

/** «Найди баг»: сверяем номер строки. Единственная проверка, которой не нужен Go. */
function checkBug(bugLine: number, line: unknown): CheckResult {
  if (typeof line !== 'number' || !Number.isInteger(line) || line < 1) {
    throw new BadRequest('нужно выбрать строку')
  }
  const pass = line === bugLine
  return {
    ...NO_RUN,
    verdict: {
      pass,
      title: pass ? 'Да, дело в этой строке' : 'Не здесь',
      detail: pass ? undefined : `строка ${line} — не та. Запустите код и посмотрите, что он печатает на самом деле.`,
    },
  }
}

/** «Почини»: запускаем то, что написал человек, и сверяем вывод с эталонным. */
async function checkFix(expected: string, code: string): Promise<CheckResult> {
  const compiled = await compile(code, true)
  if (compiled.Errors) {
    return asResult(compiled, { pass: false, title: 'Не компилируется', detail: compiled.Errors })
  }

  const { stdout, stderr } = streams(toChunks(compiled.Events))
  if (stderr) {
    // Собралось, но упало: паника или дедлок. Это отдельный разговор, не «вывод не совпал».
    return asResult(compiled, { pass: false, title: 'Программа упала', detail: stderr.split('\n').slice(0, 3).join('\n') })
  }

  const pass = sameOutput(stdout, expected)
  return asResult(compiled, {
    pass,
    title: pass ? 'Готово' : 'Вывод не тот',
    detail: pass ? undefined : `ожидалось:\n${normalize(expected)}\n\nполучилось:\n${normalize(stdout) || '(пусто)'}`,
  })
}

/**
 * «Реализуй»: к коду дописываются скрытые тесты.
 *
 * Песочница включает режим go test только для однофайловой отправки, поэтому
 * тесты именно дописываются текстом, без своего блока импортов, а недостающие
 * импорты потом проставляет goimports. Побочный плюс: сообщение об ошибке
 * синтаксиса приходит от gofmt ещё до компиляции.
 */
async function checkImplement(tests: string, code: string): Promise<CheckResult> {
  if (hasMain(code)) {
    throw new BadRequest('уберите func main — проверка запускает код как go test, и main ей мешает')
  }
  const run = await runTests(code, tests)
  return asResult(run.compiled, run.verdict)
}

export interface TestRun {
  compiled: CompileResult
  /** Собралось ли вместе с тестами — отдельно от того, прошли ли тесты. */
  built: boolean
  verdict: Verdict
}

/** Склеить код со скрытыми тестами и прогнать как go test. */
export async function runTests(code: string, tests: string): Promise<TestRun> {
  const merged = `${code.replace(/\s*$/, '')}\n\n${tests.replace(/^package\s+\w+\s*/m, '')}`
  const formatted = await format(merged, true)
  if (formatted.Error) {
    const compiled = { Errors: '', Events: null }
    return { compiled, built: false, verdict: { pass: false, title: 'Код не разбирается', detail: formatted.Error } }
  }

  const compiled = await compile(formatted.Body, true)
  if (compiled.Errors) {
    return { compiled, built: false, verdict: { pass: false, title: 'Не компилируется', detail: compiled.Errors } }
  }

  const { stdout, stderr } = streams(toChunks(compiled.Events))
  const pass = /^PASS$/m.test(stdout) && !/^(FAIL|---\s+FAIL)/m.test(stdout)

  // Тесты видеть нельзя, а вот их жалобы — можно и нужно: без них проверка бесполезна.
  const failures = stdout
    .split('\n')
    .filter((line) => /^(---\s+FAIL|\s+\S+\.go:\d+:|FAIL)/.test(line))
    .join('\n')

  return {
    compiled,
    built: true,
    verdict: {
      pass,
      title: pass ? 'Тесты прошли' : stderr ? 'Программа упала' : 'Тесты не прошли',
      detail: pass ? undefined : stderr.split('\n').slice(0, 5).join('\n') || failures || stdout,
    },
  }
}
