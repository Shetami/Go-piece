import type { Duration } from './types.ts'

/**
 * Детерминированный генератор (mulberry32) — такой же, как в стенде планировщика.
 * Вся случайность модели идёт только через него, поэтому один seed всегда даёт
 * посимвольно одинаковую ленту событий: ссылкой на прогон можно делиться.
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

  chance(p: number): boolean {
    return this.next() < p
  }

  duration(d: Duration): number {
    if (typeof d === 'number') return Math.max(1, d)
    const [min, max] = d
    return Math.max(1, this.int(min, max))
  }
}
