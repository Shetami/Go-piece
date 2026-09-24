/**
 * Результат SQL-запроса и всё, что с ним делает проверка: сравнение с эталоном,
 * разбор ответа «что вернёт запрос», текстовый вид для вердикта.
 *
 * Модуль чистый — ни PGlite, ни DOM. Его импортируют и остров, и воркер, и
 * тесты, поэтому значения здесь уже строки ровно в том виде, в каком их печатает
 * psql: `t` и `f` для булевых, `2.50` для numeric, `2024-01-31` для дат. NULL —
 * это `null`, а не пустая строка: пустая строка в SQL — полноправное значение.
 */

export type Cell = string | null

export interface SqlTable {
  columns: string[]
  rows: Cell[][]
}

/** Чем кончился прогон: таблица последнего запроса или ошибка Postgres. */
export type SqlOutcome =
  | {
      ok: true
      /** Строки последнего запроса. `null` — последняя команда строк не возвращает (UPDATE без RETURNING и т. п.). */
      table: SqlTable | null
      /** Тег последней команды: SELECT, UPDATE… */
      command: string
      affected: number
      ms: number
    }
  | {
      ok: false
      error: string
      /** Позиция ошибки в тексте запроса, с единицы, в символах — так её считает Postgres. */
      position?: number
      hint?: string
    }

const NUMBER = /^[+-]?(\d+\.?\d*|\.\d+)(e[+-]?\d+)?$/i

/**
 * Каноничный вид ячейки для сравнения. Разница между `2.5` и `2.50`,
 * `true` и `t` — это разница в записи, а не в ответе: numeric хранит масштаб,
 * и человек не обязан его угадывать. Всё остальное сравнивается как есть.
 */
export function canonCell(cell: Cell): string {
  if (cell === null) return '\u0000NULL'
  const v = cell.trim()
  if (NUMBER.test(v)) {
    const n = Number(v)
    if (Number.isFinite(n)) return String(Number(n.toFixed(6)))
  }
  if (v === 'true') return 't'
  if (v === 'false') return 'f'
  return v
}

function rowKey(row: Cell[]): string {
  return row.map(canonCell).join('\u0001')
}

/** Строка таблицы так, как её показал бы psql: ячейки через « | », NULL словом. */
export function formatRow(row: Cell[]): string {
  return row.map((c) => (c === null ? 'NULL' : c)).join(' | ')
}

export function formatRows(rows: Cell[][]): string {
  return rows.length === 0 ? '(0 строк)' : rows.map(formatRow).join('\n')
}

export interface Comparison {
  pass: boolean
  /** Что не так — коротко и без выдачи эталона целиком. */
  detail?: string
}

function rowsWord(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return `${n} строка`
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return `${n} строки`
  return `${n} строк`
}

/**
 * Совпадает ли результат с эталоном.
 *
 * Имена колонок не сравниваются: `count(*)` и `AS total` — один и тот же ответ.
 * Число колонок и их порядок — сравниваются, их условие задачи называет явно.
 * Порядок строк важен, только если задача его требует (`ordered`): без ORDER BY
 * Postgres волен отдавать строки как угодно, и требовать порядка было бы нечестно.
 */
export function compareTables(expected: SqlTable, got: SqlTable, ordered: boolean): Comparison {
  if (got.columns.length !== expected.columns.length) {
    return {
      pass: false,
      detail: `колонок ${got.columns.length}, а нужно ${expected.columns.length}. Сверьтесь с условием: какие колонки и в каком порядке.`,
    }
  }

  const want = expected.rows.map(rowKey)
  const have = got.rows.map(rowKey)

  if (ordered) {
    if (want.length === have.length && want.every((k, i) => k === have[i])) return { pass: true }
  } else {
    const a = [...want].sort()
    const b = [...have].sort()
    if (a.length === b.length && a.every((k, i) => k === b[i])) return { pass: true }
  }

  // Что именно не так. Эталон целиком не показываем — только расхождение.
  const counts = new Map<string, number>()
  for (const k of want) counts.set(k, (counts.get(k) ?? 0) + 1)
  const extra: Cell[][] = []
  got.rows.forEach((row, i) => {
    const k = have[i]!
    const left = counts.get(k) ?? 0
    if (left > 0) counts.set(k, left - 1)
    else extra.push(row)
  })
  const missing = [...counts.values()].reduce((s, n) => s + n, 0)

  const lines: string[] = []
  if (have.length !== want.length) lines.push(`получилось ${rowsWord(have.length)}, а должно быть ${want.length}.`)
  if (extra.length > 0) {
    const shown = extra.slice(0, 3).map((r) => `  ${formatRow(r)}`)
    lines.push(`лишние или неверные строки (${extra.length}):\n${shown.join('\n')}${extra.length > 3 ? '\n  …' : ''}`)
  }
  if (missing > 0 && extra.length === 0) lines.push(`не хватает строк: ${missing}.`)
  if (lines.length === 0) lines.push('строки те же, но порядок не тот. Проверьте ORDER BY.')
  return { pass: false, detail: lines.join('\n') }
}

/**
 * Ответ на «что вернёт запрос» — текст, который человек набрал руками.
 *
 * Строка ответа — строка результата, ячейки через `|`. Прощаем то, что psql
 * печатает сам: строку заголовка (если совпадает с именами колонок), черту из
 * минусов и плюсов, подпись «(N rows)». Пустой ответ, «пусто» и «(0 строк)»
 * означают, что строк нет.
 */
export function parseAnswer(text: string, columns: string[]): Cell[][] {
  const lines = text
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l !== '')
    .filter((l) => !/^[-+|\s]+$/.test(l))
    .filter((l) => !/^\(\d+\s+(rows?|строк[аи]?)\)$/i.test(l))

  if (lines.length === 1 && /^(пусто|ничего|нет строк|\(?0 строк\)?|empty)$/i.test(lines[0]!)) return []

  const rows = lines.map((l) =>
    l
      .replace(/^\|/, '')
      .replace(/\|$/, '')
      .split('|')
      .map((c) => c.trim())
      .map((c): Cell => (/^null$/i.test(c) ? null : c)),
  )

  const header = rows[0]
  if (header && columns.length > 0 && header.length === columns.length && header.every((c, i) => c === columns[i])) {
    rows.shift()
  }
  return rows
}
