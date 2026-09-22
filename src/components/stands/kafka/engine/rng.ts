import type { Duration } from './types.ts'

/**
 * Детерминированный генератор (mulberry32) — такой же, как в остальных стендах.
 * Один seed обязан давать посимвольно одинаковую ленту событий.
 *
 * Здесь он решает три вещи: паузы между send(), выбор ключа по весам и время
 * обработки сообщения. Сеть и брокеры в модели детерминированы и без него.
 */
export class Rng {
  private s: number

  constructor(seed: number) {
    this.s = seed >>> 0
  }

  next(): number {
    this.s = (this.s + 0x6d2b79f5) >>> 0
    let t = this.s
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }

  int(min: number, max: number): number {
    if (max <= min) return min
    return min + Math.floor(this.next() * (max - min + 1))
  }

  duration(d: Duration): number {
    if (typeof d === 'number') return Math.max(1, d)
    const [min, max] = d
    return Math.max(1, this.int(min, max))
  }

  /** Индекс по весам. Без весов — равновероятно. */
  weighted(n: number, weights?: number[]): number {
    if (!weights || weights.length !== n) return this.int(0, n - 1)
    const total = weights.reduce((a, b) => a + Math.max(0, b), 0)
    if (total <= 0) return this.int(0, n - 1)
    let x = this.next() * total
    for (let i = 0; i < n; i++) {
      x -= Math.max(0, weights[i]!)
      if (x < 0) return i
    }
    return n - 1
  }
}
