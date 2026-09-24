/**
 * Сессия с Postgres в воркере — со стороны страницы.
 *
 * База поднимается лениво, при первом запросе: PGlite весит несколько мегабайт,
 * и тянуть их ради того, кто только читает условие, незачем. Запросы идут
 * строго по одному. Если запрос не уложился в отведённое время, воркер
 * уничтожается вместе с базой, а следующий запрос поднимет её заново — данные
 * задачи всё равно берутся из схемы, терять нечего.
 */

import type { SqlOutcome } from './result.ts'

export type WorkerRequest = { id: number; type: 'init'; schema: string } | { id: number; type: 'run'; sql: string }

export type WorkerReply =
  | { id: number; type: 'ready' }
  | { id: number; type: 'result'; outcome: SqlOutcome }
  | { id: number; type: 'fatal'; error: string }

/** Задачи крошечные: всё, что дольше, — почти наверняка бесконечная рекурсия. */
const RUN_TIMEOUT_MS = 5000

export class SqlSession {
  private worker: Worker | null = null
  private ready: Promise<void> | null = null
  private seq = 0
  private queue: Promise<unknown> = Promise.resolve()
  private pending = new Map<number, (reply: WorkerReply) => void>()

  constructor(private readonly schema: string) {}

  /** Выполнить запрос. Ошибки Postgres — часть результата; бросает только поломка самой среды. */
  run(sql: string): Promise<SqlOutcome> {
    const job = this.queue.then(() => this.runNow(sql))
    this.queue = job.catch(() => undefined)
    return job
  }

  /** Поднять базу заранее — например, когда человек начал печатать. */
  warmUp(): void {
    this.start().catch(() => undefined)
  }

  dispose(): void {
    this.worker?.terminate()
    this.worker = null
    this.ready = null
    for (const resolve of this.pending.values()) resolve({ id: -1, type: 'fatal', error: 'сессия закрыта' })
    this.pending.clear()
  }

  private start(): Promise<void> {
    if (this.ready) return this.ready
    const worker = new Worker(new URL('./sql.worker.ts', import.meta.url), { type: 'module' })
    worker.onmessage = (event: MessageEvent<WorkerReply>) => {
      const resolve = this.pending.get(event.data.id)
      this.pending.delete(event.data.id)
      resolve?.(event.data)
    }
    worker.onerror = (event) => {
      for (const resolve of this.pending.values()) resolve({ id: -1, type: 'fatal', error: event.message || 'воркер упал' })
      this.pending.clear()
    }
    this.worker = worker
    this.ready = this.send({ type: 'init', schema: this.schema }).then((reply) => {
      if (reply.type === 'fatal') {
        this.dispose()
        throw new Error(`не удалось поднять базу: ${reply.error}`)
      }
    })
    return this.ready
  }

  private send(req: { type: 'init'; schema: string } | { type: 'run'; sql: string }): Promise<WorkerReply> {
    const id = ++this.seq
    return new Promise((resolve) => {
      this.pending.set(id, resolve)
      this.worker!.postMessage({ ...req, id } satisfies WorkerRequest)
    })
  }

  private async runNow(sql: string): Promise<SqlOutcome> {
    await this.start()

    let timer: ReturnType<typeof setTimeout> | undefined
    const timeout = new Promise<'timeout'>((resolve) => {
      timer = setTimeout(() => resolve('timeout'), RUN_TIMEOUT_MS)
    })
    const reply = await Promise.race([this.send({ type: 'run', sql }), timeout])
    clearTimeout(timer)

    if (reply === 'timeout') {
      this.dispose()
      return {
        ok: false,
        error: `запрос не уложился в ${RUN_TIMEOUT_MS / 1000} с и был остановлен`,
        hint: 'на таких маленьких данных это почти всегда бесконечная рекурсия: у рекурсивной части CTE нет условия остановки',
      }
    }
    if (reply.type === 'fatal') throw new Error(reply.error)
    if (reply.type !== 'result') throw new Error('воркер ответил не тем')
    return reply.outcome
  }
}
