import type { Duration } from './types.ts'

/**
 * Детерминированный генератор (mulberry32).
 *
 * Вся случайность модели идёт только через него: выбор жертвы при краже,
 * длительности фаз, разброс времени системных вызовов. Один и тот же seed
 * обязан давать посимвольно одинаковую ленту событий — на этом держатся
 * шаринг ссылок на стенд и воспроизводимость тестов.
 */
export class Rng {
  private s: number

  constructor(seed: number) {
    this.s = seed >>> 0
  }

  /** [0, 1) */
  next(): number {
    this.s = (this.s + 0x6d2b79f5) >>> 0
    let t = this.s
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }

  /** Целое из [min, max] включительно. */
  int(min: number, max: number): number {
    if (max <= min) return min
    return min + Math.floor(this.next() * (max - min + 1))
  }

  pick<T>(items: readonly T[]): T | undefined {
    if (items.length === 0) return undefined
    return items[Math.floor(this.next() * items.length)]
  }

  /** Разворачивает Duration в конкретное число тиков. */
  duration(d: Duration): number {
    if (d === 'forever') return Number.POSITIVE_INFINITY
    if (typeof d === 'number') return Math.max(1, d)
    const [min, max] = d
    return Math.max(1, this.int(min, max))
  }
}
