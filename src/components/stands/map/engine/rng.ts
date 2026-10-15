/**
 * Детерминированный генератор (mulberry32) — такой же, как в остальных стендах.
 * Один seed обязан давать посимвольно одинаковую ленту событий.
 *
 * Здесь он отвечает за всё, что в рантайме случайно: соль хэша у каждой карты
 * и при её сбросе, стартовую точку обхода и хэш NaN.
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

  /** Случайные 32 бита. */
  u32(): number {
    return Math.floor(this.next() * 4294967296) >>> 0
  }
}
