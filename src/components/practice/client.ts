/**
 * Обращения острова к серверным маршрутам. Здесь же — единственное место,
 * где сетевая ошибка превращается в текст, который не стыдно показать.
 */

import { isApiError, type CheckRequest, type CheckResult, type RunResult } from './protocol.ts'

async function post<T>(url: string, body: unknown): Promise<T> {
  let res: Response
  try {
    res = await fetch(url, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(body),
    })
  } catch {
    throw new Error('нет связи с сервером')
  }

  let data: unknown
  try {
    data = await res.json()
  } catch {
    throw new Error(`сервер ответил ${res.status} без внятного тела`)
  }

  if (!res.ok) {
    throw new Error(isApiError(data) ? data.error : `сервер ответил ${res.status}`)
  }
  return data as T
}

export function runCode(code: string): Promise<RunResult> {
  return post<RunResult>('/api/run', { code })
}

export function checkTask(request: CheckRequest): Promise<CheckResult> {
  return post<CheckResult>('/api/check', request)
}

/**
 * Форматирование. Синтаксическая ошибка — не сбой запроса: gofmt честно
 * говорит, где не разобрал, и это сообщение стоит показать в редакторе.
 */
export async function formatCode(code: string): Promise<{ code: string } | { syntaxError: string }> {
  try {
    return await post<{ code: string }>('/api/fmt', { code })
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    if (/\.go:\d+/.test(message)) return { syntaxError: message }
    throw err
  }
}
