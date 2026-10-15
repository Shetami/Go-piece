import { MIN_TOPHASH } from './types.ts'

/**
 * Хэш ключа с солью карты. В рантайме это memhash на AES-раундах и 64 бита;
 * здесь — FNV-1a с финализатором murmur3 и 32 бита. Для модели важно одно:
 * биты перемешаны хорошо, и смена соли меняет хэш каждого ключа целиком.
 */
export function hashKey(key: string, seed: number): number {
  let h = (0x811c9dc5 ^ seed) >>> 0
  for (let i = 0; i < key.length; i++) {
    h ^= key.charCodeAt(i)
    h = Math.imul(h, 0x01000193)
  }
  h ^= seed >>> 16
  h ^= h >>> 16
  h = Math.imul(h, 0x85ebca6b)
  h ^= h >>> 13
  h = Math.imul(h, 0xc2b2ae35)
  h ^= h >>> 16
  return h >>> 0
}

/** NaN != NaN: такой ключ нельзя найти, и рантайм хэширует его случайным числом. */
export const isNaNKey = (key: string) => key === 'NaN'

/** Равенство ключей так, как его видит карта: NaN не равен ничему, даже себе. */
export const keyEqual = (a: string | null, b: string) => a === b && !isNaNKey(b)

/** Swiss table: младшие 7 бит — в контрольный байт. */
export const h2 = (hash: number) => hash & 0x7f
/** Swiss table: остальное — выбор группы, с которой начинаются пробы. */
export const h1 = (hash: number) => hash >>> 7

/** Строка каталога: старшие globalDepth бит хэша. */
export function dirIndex(hash: number, globalDepth: number): number {
  // В JS сдвиг на 32 — это сдвиг на 0, поэтому нулевую глубину разбираем отдельно.
  return globalDepth === 0 ? 0 : hash >>> (32 - globalDepth)
}

/** Бит, по которому таблица глубины depth-1 делится на две глубины depth. */
export const splitBit = (depth: number) => (1 << (32 - depth)) >>> 0

/** Старая карта: старшие 8 бит хэша; значения меньше 5 заняты под состояния слота. */
export function tophash(hash: number): number {
  const top = hash >>> 24
  return top < MIN_TOPHASH ? top + MIN_TOPHASH : top
}

/** Последовательность проб: h, h+1, h+3, h+6, … по модулю числа групп. */
export function* probeSeq(hash: number, mask: number): Generator<number> {
  let offset = h1(hash) & mask
  for (let i = 1; i <= mask + 1; i++) {
    yield offset
    offset = (offset + i) & mask
  }
}

export const hex2 = (n: number) => n.toString(16).padStart(2, '0')
export const hex8 = (n: number) => (n >>> 0).toString(16).padStart(8, '0')
export const bin = (n: number, width: number) => (n >>> 0).toString(2).padStart(width, '0').slice(-width)
