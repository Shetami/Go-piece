import type { APIRoute } from 'astro'
import { getCollection } from 'astro:content'
import { BadRequest, handleError, json, readJson, requireCode, requireString, toChunks } from '../../lib/api.ts'
import { compile, format, hasMain, type CompileResult } from '../../lib/playground.ts'
import { taskCode } from '../../lib/task-files.ts'
import { taskCheck, taskExpect } from '../../lib/task-secrets.ts'
import { taskLang } from '../../data/practice.ts'
import { normalize, sameOutput, type CheckRequest, type CheckResult, type OutputChunk, type Verdict } from '../../components/practice/protocol.ts'

export const prerender = false

/**
 * POST /api/check — проверить решение.
 *
 * Здесь живут ответы, поэтому маршрут сам достаёт всё, что нужно для проверки:
 * эталонный вывод, скрытые тесты, номер сломанной строки. Клиент присылает
 * только то, что сделал человек, и не может повлиять на критерий.
 *
 * Механика у каждого типа задачи своя:
 *   output    — сервер запускает НЕТРОНУТУЮ заготовку и сверяет её вывод с ответом;
 *               правки в редакторе на вердикт не влияют, иначе ответ можно подогнать;
 *   bug       — сверяется номер строки, в песочницу не ходим вовсе;
 *   fix       — запускается код пользователя, вывод сверяется с expect.txt;
 *   implement — к коду дописываются скрытые тесты, и всё это идёт как go test.
 */
export const POST: APIRoute = async ({ request }) => {
  try {
    const body = await readJson<CheckRequest>(request)
    const id = requireString(body.task, 'task')

    const tasks = await getCollection('tasks')
    const task = tasks.find((t) => t.id === id)
    if (!task) throw new BadRequest(`нет такой задачи: ${id}`, 404)

    // SQL-задачи проверяются в браузере, где живёт их Postgres. Сюда приходят
    // только «найди баг» — номеру строки всё равно, на каком он языке.
    if (taskLang(task.data.topic) === 'sql' && task.data.kind !== 'bug') {
      throw new BadRequest('SQL-задачи этого типа проверяются в браузере')
    }

    switch (task.data.kind) {
      case 'output':
        return json(await checkOutput(id, body.answer))
      case 'bug':
        return json(checkBug(task.data.bugLine!, body.line))
      case 'fix':
        return json(await checkFix(id, requireCode(body.code)))
      case 'implement':
        return json(await checkImplement(id, requireCode(body.code)))
    }
  } catch (err) {
    return handleError(err)
  }
}

/** Пустой прогон — для вердиктов, которым песочница не нужна. */
const NO_RUN = { errors: '', vet: '', output: [] as OutputChunk[], stubbedMain: false }

function streams(output: OutputChunk[]) {
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
async function checkOutput(id: string, answer: unknown): Promise<CheckResult> {
  if (typeof answer !== 'string') throw new BadRequest('поле answer должно быть строкой')

  const compiled = await compile(taskCode(id).starter, true)
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
async function checkFix(id: string, code: string): Promise<CheckResult> {
  const expected = taskExpect(id)
  if (expected === null) throw new BadRequest(`задача ${id}: нет expect.txt`, 500)

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
async function checkImplement(id: string, code: string): Promise<CheckResult> {
  const tests = taskCheck(id)
  if (tests === null) throw new BadRequest(`задача ${id}: нет check.go`, 500)

  if (hasMain(code)) {
    throw new BadRequest('уберите func main — проверка запускает код как go test, и main ей мешает')
  }

  const merged = `${code.replace(/\s*$/, '')}\n\n${tests.replace(/^package\s+\w+\s*/m, '')}`
  const formatted = await format(merged, true)
  if (formatted.Error) {
    return asResult({ Errors: '', Events: null }, { pass: false, title: 'Код не разбирается', detail: formatted.Error })
  }

  const compiled = await compile(formatted.Body, true)
  if (compiled.Errors) {
    return asResult(compiled, { pass: false, title: 'Не компилируется', detail: compiled.Errors })
  }

  const chunks = toChunks(compiled.Events)
  const { stdout, stderr } = streams(chunks)
  const pass = /^PASS$/m.test(stdout) && !/^(FAIL|---\s+FAIL)/m.test(stdout)

  // Тесты видеть нельзя, а вот их жалобы — можно и нужно: без них проверка бесполезна.
  const failures = stdout
    .split('\n')
    .filter((line) => /^(---\s+FAIL|\s+\S+\.go:\d+:|FAIL)/.test(line))
    .join('\n')

  return asResult(compiled, {
    pass,
    title: pass ? 'Тесты прошли' : stderr ? 'Программа упала' : 'Тесты не прошли',
    detail: pass ? undefined : stderr.split('\n').slice(0, 5).join('\n') || failures || stdout,
  })
}
