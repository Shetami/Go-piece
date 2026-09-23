/**
 * Клиент Go Playground — единственное место в проекте, которое ходит в сеть.
 *
 * Код на Go компилирует и исполняет go.dev: свой песочницы у нас нет и не будет,
 * компилировать чужой код на своём сервере — плохая идея. Отсюда два следствия,
 * о которых стоит помнить, читая всё остальное:
 *
 *  - Время в песочнице ненастоящее. Часы стартуют с 2009-11-10 23:00 UTC и
 *    прыгают вперёд, как только все горутины уснули. Поэтому `time.Sleep`
 *    ничего не стоит, а у события есть `Delay` — на сколько наносекунд позже
 *    предыдущего оно случилось. Вывод можно проиграть с настоящими паузами.
 *  - Тестовый режим включается только для ОДНОГО файла. Многофайловая отправка
 *    (txtar) компилируется, но всегда запускается как программа и требует main.
 *    Поэтому скрытые проверки дописываются к коду пользователя текстом,
 *    а недостающие импорты потом добавляет goimports.
 */

const UPSTREAM = 'https://go.dev/_'

/** Сколько ждём go.dev. Сама песочница режет исполнение примерно на пяти секундах. */
const TIMEOUT_MS = 20_000

/** Предел на размер исходника. Задачи здесь небольшие, а поле ввода — чужое. */
export const MAX_CODE_BYTES = 20_000

/** Кусок вывода программы. `Delay` — наносекунды с предыдущего куска. */
export interface PlaygroundEvent {
  Message: string
  Kind: 'stdout' | 'stderr'
  Delay: number
}

/**
 * Ответ компилятора. Тут важно различать два вида неудачи:
 * `Errors` — программа не собралась, `Events` с `Kind: 'stderr'` — собралась
 * и упала на исполнении (паника, дедлок). Для обучения это разные истории.
 */
export interface CompileResult {
  Errors: string
  Events: PlaygroundEvent[] | null
  VetErrors?: string
}

export interface FormatResult {
  Body: string
  Error: string
}

export class PlaygroundError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message)
    this.name = 'PlaygroundError'
  }
}

async function post<T>(path: string, form: Record<string, string>): Promise<T> {
  const body = new URLSearchParams(form)
  let res: Response
  try {
    res = await fetch(`${UPSTREAM}/${path}`, {
      method: 'POST',
      headers: {
        'content-type': 'application/x-www-form-urlencoded; charset=utf-8',
        'user-agent': 'go-piece (https://go-piece.vercel.app)',
      },
      body,
      signal: AbortSignal.timeout(TIMEOUT_MS),
    })
  } catch (err) {
    const timedOut = err instanceof Error && err.name === 'TimeoutError'
    throw new PlaygroundError(
      timedOut ? 'go.dev не ответил вовремя' : 'не получилось достучаться до go.dev',
      504,
    )
  }

  if (!res.ok) {
    // 429 с go.dev прилетает при частых запросах — так и скажем, а не «ошибка сервера».
    const text = res.status === 429 ? 'go.dev просит подождать: слишком часто' : `go.dev ответил ${res.status}`
    throw new PlaygroundError(text, res.status === 429 ? 429 : 502)
  }

  return (await res.json()) as T
}

/** Собрать и запустить. `vet` ловит то, что компилятор пропускает, — например `Printf` с неверным глаголом. */
export function compile(code: string, vet = true): Promise<CompileResult> {
  return post<CompileResult>('compile', { version: '2', body: code, withVet: String(vet) })
}

/**
 * gofmt. С `imports: true` это goimports — он ещё и дописывает недостающие
 * импорты. На этом держится склейка кода пользователя со скрытыми проверками:
 * проверкам нужен `testing`, но своего блока импортов у них нет.
 */
export function format(code: string, imports = true): Promise<FormatResult> {
  return post<FormatResult>('fmt', { body: code, imports: String(imports) })
}

/** Есть ли в исходнике `func main`. Без него песочница не запустит программу. */
export function hasMain(code: string): boolean {
  return /^func\s+main\s*\(\s*\)/m.test(stripComments(code))
}

/** Грубое вычищение комментариев и строк — только чтобы `hasMain` не ловил `func main` в тексте. */
function stripComments(code: string): string {
  return code
    .replace(/\/\*[\s\S]*?\*\//g, ' ')
    .replace(/\/\/[^\n]*/g, ' ')
    .replace(/"(?:[^"\\\n]|\\.)*"/g, '""')
    .replace(/`[^`]*`/g, '``')
}

/** Склеить вывод программы в один текст — как его видно в терминале. */
export function joinOutput(events: PlaygroundEvent[] | null): string {
  return (events ?? []).map((e) => e.Message).join('')
}

/**
 * Нормализация вывода перед сравнением. Пробелы в конце строк и лишние пустые
 * строки в конце — не то, на чём стоит заваливать решение.
 */
export function normalizeOutput(text: string): string {
  return text
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map((line) => line.replace(/[ \t]+$/, ''))
    .join('\n')
    .replace(/\n+$/, '')
}
