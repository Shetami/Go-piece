/**
 * Общая обвязка серверных маршрутов: разбор тела, ответы, перевод результата
 * песочницы в наш протокол. Всё, что здесь есть, исполняется только на сервере.
 */

import type { OutputChunk } from '../components/practice/protocol.ts'
import { MAX_CODE_BYTES, PlaygroundError, type PlaygroundEvent } from './playground.ts'

export function json(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      'content-type': 'application/json; charset=utf-8',
      // Ответы зависят от тела запроса и ничего не кэшируют по адресу.
      'cache-control': 'no-store',
    },
  })
}

export function fail(message: string, status = 400): Response {
  return json({ error: message }, status)
}

/** Ошибка, на которую маршрут отвечает понятным текстом, а не пятисоткой. */
export class BadRequest extends Error {
  constructor(
    message: string,
    readonly status = 400,
  ) {
    super(message)
    this.name = 'BadRequest'
  }
}

export async function readJson<T>(request: Request): Promise<T> {
  let parsed: unknown
  try {
    parsed = await request.json()
  } catch {
    throw new BadRequest('ожидался JSON')
  }
  if (typeof parsed !== 'object' || parsed === null) throw new BadRequest('ожидался объект')
  return parsed as T
}

/**
 * Проверка кода перед отправкой в песочницу. Поле ввода открыто миру, а за
 * каждым запросом стоит чужой сервис, у которого есть лимиты, — поэтому
 * длину режем у себя, а не узнаём о ней из 429.
 */
export function requireCode(value: unknown, field = 'code'): string {
  if (typeof value !== 'string') throw new BadRequest(`поле ${field} должно быть строкой`)
  if (value.trim() === '') throw new BadRequest('код пустой')
  if (new TextEncoder().encode(value).length > MAX_CODE_BYTES) {
    throw new BadRequest(`код длиннее ${MAX_CODE_BYTES} байт — для задачи здесь этого точно много`)
  }
  return value
}

export function requireString(value: unknown, field: string): string {
  if (typeof value !== 'string' || value === '') throw new BadRequest(`поле ${field} обязательно`)
  return value
}

/** События песочницы → куски вывода нашего протокола. */
export function toChunks(events: PlaygroundEvent[] | null): OutputChunk[] {
  return (events ?? []).map((e) => ({
    kind: e.Kind === 'stderr' ? 'stderr' : 'stdout',
    text: e.Message,
    delay: e.Delay,
  }))
}

/** Один обработчик ошибок на все маршруты: свои — с текстом, чужие — без подробностей. */
export function handleError(err: unknown): Response {
  if (err instanceof BadRequest) return fail(err.message, err.status)
  if (err instanceof PlaygroundError) return fail(err.message, err.status)
  console.error('practice api:', err)
  return fail('не получилось выполнить запрос', 500)
}
