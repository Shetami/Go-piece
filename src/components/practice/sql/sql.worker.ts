/**
 * Воркер с Postgres. Отдельный поток нужен по двум причинам: загрузка WASM
 * не должна замораживать страницу, а бесконечный рекурсивный CTE должен
 * убиваться. Отменить запрос внутри PGlite нечем — statement_timeout в нём не
 * срабатывает, — поэтому зависший воркер просто уничтожают снаружи (session.ts).
 */

import type { PGlite } from '@electric-sql/pglite'
import { openDb, runQuery } from './engine.ts'
import type { WorkerReply, WorkerRequest } from './session.ts'

let db: Promise<PGlite> | null = null

function reply(message: WorkerReply) {
  self.postMessage(message)
}

self.onmessage = async (event: MessageEvent<WorkerRequest>) => {
  const req = event.data
  try {
    if (req.type === 'init') {
      db = openDb(req.schema)
      await db
      reply({ id: req.id, type: 'ready' })
      return
    }
    if (!db) throw new Error('база не поднята')
    reply({ id: req.id, type: 'result', outcome: await runQuery(await db, req.sql) })
  } catch (err) {
    reply({ id: req.id, type: 'fatal', error: err instanceof Error ? err.message : String(err) })
  }
}
