/**
 * PostgreSQL для задач: PGlite — настоящий Postgres, собранный в WebAssembly.
 *
 * Модуль не знает, где он работает: в браузере его запускает воркер
 * (sql.worker.ts), в тестах — Node напрямую. Поэтому здесь нет ни таймаутов, ни
 * сообщений: только «открыть базу со схемой задачи» и «выполнить запрос».
 *
 * Каждый запрос идёт внутри транзакции, которая потом откатывается. Схема и
 * данные задачи поэтому неприкосновенны: DELETE без WHERE или DROP TABLE
 * испортят ровно один прогон, а проверка всегда видит исходные данные.
 */

import { PGlite, types } from '@electric-sql/pglite'
import type { Cell, SqlOutcome } from './result.ts'

/**
 * Значения — строками, как их отдаёт сам Postgres. По умолчанию PGlite
 * превращает их в JS: date в Date со сдвигом часового пояса, numeric — то в
 * строку, то в число, bigint — в число с потерей точности. Для задач нужен
 * ровно тот вид, что в psql, иначе ответ «2024-01-31» не совпадёт ни с чем.
 */
const RAW: Record<number, (x: string) => string> = {}
for (const oid of Object.values(types)) {
  if (typeof oid === 'number') RAW[oid] = (x) => x
}

/** Массивы PGlite разбирает сам и по своим oid — собираем обратно в запись Postgres. */
function toCell(value: unknown): Cell {
  if (value === null || value === undefined) return null
  if (Array.isArray(value)) return `{${value.map((v) => toCell(v) ?? 'NULL').join(',')}}`
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

/**
 * Управлять транзакциями запросу нельзя: он сам уже внутри транзакции, которую
 * откатит прогон. COMMIT посреди запроса зафиксировал бы правки в данных задачи.
 */
const TX_CONTROL = /(^|;)\s*(begin|commit|end|rollback|abort|start\s+transaction|savepoint|release|prepare\s+transaction)\b/i

function stripComments(sql: string): string {
  return sql.replace(/--[^\n]*/g, '').replace(/\/\*[\s\S]*?\*\//g, '')
}

/** Поднять базу и залить в неё схему задачи. Ошибка в схеме — поломка задачи, пусть летит. */
export async function openDb(schema: string): Promise<PGlite> {
  const db = await PGlite.create({ parsers: RAW })
  await db.exec(schema)
  return db
}

/** Выполнить запрос (или несколько через `;`) и откатить всё, что он сделал. */
export async function runQuery(db: PGlite, sql: string): Promise<SqlOutcome> {
  if (stripComments(sql).trim() === '') return { ok: false, error: 'запрос пуст' }
  if (TX_CONTROL.test(stripComments(sql))) {
    return {
      ok: false,
      error: 'управлять транзакциями здесь не нужно',
      hint: 'каждый прогон и так идёт в своей транзакции и откатывается — данные задачи всегда исходные',
    }
  }

  const started = performance.now()
  await db.exec('BEGIN')
  try {
    const results = await db.exec(sql, { rowMode: 'array', parsers: RAW })
    const ms = performance.now() - started
    const last = results.at(-1)
    if (!last) return { ok: false, error: 'запрос пуст' }

    const hasRows = last.fields.length > 0
    return {
      ok: true,
      table: hasRows
        ? {
            columns: last.fields.map((f) => f.name),
            rows: (last.rows as unknown[][]).map((row) => row.map(toCell)),
          }
        : null,
      command: commandOf(sql),
      affected: last.affectedRows ?? 0,
      ms,
    }
  } catch (err) {
    const e = err as { message?: string; position?: string | number; hint?: string }
    const position = e.position === undefined ? undefined : Number(e.position)
    return {
      ok: false,
      error: e.message ?? String(err),
      ...(position && Number.isFinite(position) ? { position } : {}),
      ...(e.hint ? { hint: e.hint } : {}),
    }
  } finally {
    await db.exec('ROLLBACK')
  }
}

/** Первое слово последнего запроса — чтобы сказать «UPDATE: 3 строки», а не молчать. */
function commandOf(sql: string): string {
  const statements = stripComments(sql)
    .split(';')
    .map((s) => s.trim())
    .filter(Boolean)
  return (statements.at(-1)?.split(/\s+/)[0] ?? '').toUpperCase()
}
