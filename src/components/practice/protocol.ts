/**
 * Что остров говорит серверу и что слышит в ответ.
 *
 * Модуль общий для React-острова и маршрутов в src/pages/api, поэтому здесь
 * только типы и чистые функции — ничего, что нельзя показывать пользователю.
 */

import type { TaskKindId } from '../../data/practice.ts'

/** Кусок вывода программы; `delay` — наносекунды с предыдущего куска. */
export interface OutputChunk {
  kind: 'stdout' | 'stderr'
  text: string
  delay: number
}

/** Результат прогона: собралось ли, что напечатало, что сказал vet. */
export interface RunResult {
  /** Ошибки компиляции. Непустая строка — программа не запускалась вовсе. */
  errors: string
  /** Замечания go vet: собралось, но, скорее всего, не то, что хотели. */
  vet: string
  output: OutputChunk[]
  /**
   * К коду пользователя дописали пустой `func main`, чтобы было что компилировать.
   * Так устроены задачи «реализуй»: там main нет и быть не должно.
   */
  stubbedMain: boolean
}

/** Вердикт проверки. `pass` — засчитано; `title` короткий, `detail` можно показать моноширинным. */
export interface Verdict {
  pass: boolean
  title: string
  detail?: string
}

export interface CheckResult extends RunResult {
  verdict: Verdict
}

export interface ApiError {
  error: string
}

/** Тело запроса на проверку. Поля зависят от типа задачи. */
export interface CheckRequest {
  task: string
  kind: TaskKindId
  /** Код из редактора — нужен для `fix` и `implement`. */
  code?: string
  /** Ответ пользователя для `output`. */
  answer?: string
  /** Выбранная строка (с единицы) для `bug`. */
  line?: number
}

/** Ответ сервера — либо результат, либо ошибка. Разводим по наличию поля. */
export function isApiError(value: unknown): value is ApiError {
  return typeof value === 'object' && value !== null && typeof (value as ApiError).error === 'string'
}

/** Весь вывод одним текстом — как его видно в терминале. */
export function outputText(chunks: OutputChunk[]): string {
  return chunks.map((c) => c.text).join('')
}

/**
 * Сравнение выводов. Хвостовые пробелы и лишние пустые строки в конце
 * не должны решать судьбу ответа — всё остальное считается значимым,
 * включая порядок строк и регистр.
 */
export function sameOutput(a: string, b: string): boolean {
  return normalize(a) === normalize(b)
}

export function normalize(text: string): string {
  return text
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map((line) => line.replace(/[ \t]+$/, ''))
    .join('\n')
    .replace(/\n+$/, '')
}
