import { MIN_TOPHASH } from './types.ts'

/** Ширина хэша — как uintptr на 64-битной платформе. */
export const HASH_BITS = 64

const u64 = (n: bigint) => BigInt.asUintN(64, n)

/**
 * Хэш ключа с солью карты, 64 бита. В рантайме это memhash на AES-раундах;
 * здесь — FNV-1a 64 с финализатором murmur3 (fmix64). Для модели важно одно:
 * биты перемешаны хорошо, и смена соли меняет хэш каждого ключа целиком.
 */
export function hashKey(key: string, seed: bigint): bigint {
  let h = 0xcbf29ce484222325n ^ seed
  for (let i = 0; i < key.length; i++) {
    h ^= BigInt(key.charCodeAt(i))
    h = u64(h * 0x100000001b3n)
  }
  h ^= seed >> 32n
  h ^= h >> 33n
  h = u64(h * 0xff51afd7ed558ccdn)
  h ^= h >> 33n
  h = u64(h * 0xc4ceb9fe1a85ec53n)
  h ^= h >> 33n
  return h
}

/** NaN != NaN: такой ключ нельзя найти, и рантайм хэширует его случайным числом. */
export const isNaNKey = (key: string) => key === 'NaN'

/** Равенство ключей так, как его видит карта: NaN не равен ничему, даже себе. */
export const keyEqual = (a: string | null, b: string) => a === b && !isNaNKey(b)

/** Swiss table: младшие 7 бит — в контрольный байт. */
export const h2 = (hash: bigint) => Number(hash & 0x7fn)
/** Swiss table: остальное — выбор группы, с которой начинаются пробы. */
export const h1 = (hash: bigint) => hash >> 7n

/** Младшие биты хэша как номер: группа, корзина. mask — 2^k-1. */
export const lowBits = (hash: bigint, mask: number) => Number(hash & BigInt(mask))

/** Строка каталога: старшие globalDepth бит хэша. */
export function dirIndex(hash: bigint, globalDepth: number): number {
  return globalDepth === 0 ? 0 : Number(hash >> BigInt(HASH_BITS - globalDepth))
}

/** Бит, по которому таблица глубины depth-1 делится на две глубины depth. */
export const splitBit = (depth: number) => 1n << BigInt(HASH_BITS - depth)

/** Старая карта: старшие 8 бит хэша; значения меньше 5 заняты под состояния слота. */
export function tophash(hash: bigint): number {
  const top = Number(hash >> BigInt(HASH_BITS - 8))
  return top < MIN_TOPHASH ? top + MIN_TOPHASH : top
}

/** Последовательность проб: h, h+1, h+3, h+6, … по модулю числа групп. */
export function* probeSeq(hash: bigint, mask: number): Generator<number> {
  let offset = lowBits(h1(hash), mask)
  for (let i = 1; i <= mask + 1; i++) {
    yield offset
    offset = (offset + i) & mask
  }
}

export const hex2 = (n: number) => n.toString(16).padStart(2, '0')
/** Соль старой карты: hash0 там uint32. */
export const hex8 = (n: bigint) => u64(n).toString(16).padStart(8, '0')
export const hex16 = (n: bigint) => u64(n).toString(16).padStart(16, '0')
export const bin = (n: bigint | number, width: number) => u64(BigInt(n)).toString(2).padStart(width, '0').slice(-width)
