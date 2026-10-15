import type { Ctx } from './ctx.ts'
import { emit, recordProbe } from './ctx.ts'
import { hashKey, isNaNKey, keyEqual, tophash } from './hash.ts'
import type { Bucket, LegacyIter, LegacyMap, LSlot, Probe, Yield } from './types.ts'
import {
  GROUP_SLOTS,
  TOP_EMPTY_ONE,
  TOP_EMPTY_REST,
  TOP_EVAC_EMPTY,
  TOP_EVAC_X,
  TOP_EVAC_Y,
} from './types.ts'

/**
 * Карта до Go 1.24 — модель runtime/map.go (hmap и bmap).
 *
 * Корзина — 8 слотов и массив tophash; полная корзина обрастает цепочкой
 * корзин переполнения. Рост — при среднем заполнении 6,5 на корзину или при
 * слишком большом числе корзин переполнения; переезд идёт постепенно,
 * по одной-две корзины на каждую запись (growWork → evacuate).
 */

const emptyLSlot = (): LSlot => ({ top: TOP_EMPTY_REST, key: null, val: null, hash: 0 })
const newPage = (): LSlot[] => Array.from({ length: GROUP_SLOTS }, emptyLSlot)
const newBucket = (): Bucket => ({ pages: [newPage()], evacuated: false })

export const isEmptyTop = (t: number) => t <= TOP_EMPTY_ONE
export const isFullTop = (t: number) => t > TOP_EVAC_EMPTY

/** 6,5 записи на корзину в среднем: count > 8 && count > 13 * 2^B / 2. */
export const overLoadFactor = (count: number, B: number) => count > GROUP_SLOTS && count > 13 * ((1 << B) / 2)
/** Корзин переполнения столько же, сколько основных (B ограничен 15). */
export const tooManyOverflow = (noverflow: number, B: number) => noverflow >= 1 << Math.min(B, 15)

export const growing = (h: LegacyMap) => h.oldbuckets !== null
/** Сколько корзин было в старом массиве, минус один. */
const oldMask = (h: LegacyMap) => (h.sameSizeGrow ? 1 << h.B : 1 << (h.B - 1)) - 1
const mask = (h: LegacyMap) => (1 << h.B) - 1

function allocArray(ctx: Ctx, h: LegacyMap, B: number): number {
  const id = ctx.world.nextId++
  h.arrays[id] = Array.from({ length: 1 << B }, newBucket)
  return id
}

export function createLegacy(ctx: Ctx, hint: number): LegacyMap {
  let B = 0
  while (overLoadFactor(hint, B)) B++
  const h: LegacyMap = {
    count: 0,
    B,
    hash0: ctx.rng.u32(),
    buckets: null,
    oldbuckets: null,
    arrays: {},
    nevacuate: 0,
    noverflow: 0,
    sameSizeGrow: false,
    writing: 0,
  }
  // makemap: при B == 0 корзина выделяется лениво, на первой записи.
  if (B > 0) h.buckets = allocArray(ctx, h, B)
  return h
}

export function legacyHash(ctx: Ctx, h: LegacyMap, key: string): number {
  return isNaNKey(key) ? ctx.rng.u32() : hashKey(key, h.hash0)
}

/** Где искать ключ: в старой корзине, если её ещё не эвакуировали, иначе в новой. */
function lookupBucket(h: LegacyMap, hash: number): { arr: number; bucket: number } {
  if (growing(h)) {
    const ob = hash & oldMask(h)
    if (!h.arrays[h.oldbuckets!]![ob]!.evacuated) return { arr: h.oldbuckets!, bucket: ob }
  }
  return { arr: h.buckets!, bucket: hash & mask(h) }
}

export interface LLoc {
  arr: number
  bucket: number
  page: number
  slot: number
}

/** Проход по цепочке корзины: tophash, потом ключ; emptyRest обрывает поиск. */
function scan(
  ctx: Ctx | null,
  h: LegacyMap,
  arr: number,
  bucket: number,
  key: string,
  hash: number,
  onFree?: (loc: LLoc) => void,
): LLoc | null {
  const top = tophash(hash)
  const b = h.arrays[arr]![bucket]!
  for (let p = 0; p < b.pages.length; p++) {
    const page = b.pages[p]!
    const probe: Probe = { arr, bucket, page: p, match: [], compared: [], stop: false }
    if (ctx) recordProbe(ctx, probe)
    for (let i = 0; i < GROUP_SLOTS; i++) {
      const s = page[i]!
      if (s.top !== top) {
        if (isEmptyTop(s.top)) onFree?.({ arr, bucket, page: p, slot: i })
        if (s.top === TOP_EMPTY_REST) {
          probe.stop = true
          return null
        }
        continue
      }
      probe.match.push(i)
      const eq = keyEqual(s.key, key)
      probe.compared.push({ slot: i, eq })
      if (ctx) {
        ctx.world.stats.compares++
        if (!eq) {
          ctx.world.stats.falseMatches++
          emit(ctx, 'h2.false', { key, other: s.key, tophash: top, bucket, slot: i, legacy: true, nan: isNaNKey(key) })
        }
      }
      if (eq) return { arr, bucket, page: p, slot: i }
    }
  }
  return null
}

export function legacyFind(ctx: Ctx, h: LegacyMap, key: string, hash: number): LLoc | null {
  if (h.buckets === null || h.count === 0) return null
  const { arr, bucket } = lookupBucket(h, hash)
  return scan(ctx, h, arr, bucket, key, hash)
}

function findQuiet(h: LegacyMap, key: string): LSlot | null {
  if (h.buckets === null || isNaNKey(key)) return null
  const hash = hashKey(key, h.hash0)
  const { arr, bucket } = lookupBucket(h, hash)
  const loc = scan(null, h, arr, bucket, key, hash)
  return loc ? h.arrays[arr]![bucket]!.pages[loc.page]![loc.slot]! : null
}

export const legacyHas = (h: LegacyMap, key: string) => findQuiet(h, key) !== null

const slotAt = (h: LegacyMap, l: LLoc) => h.arrays[l.arr]![l.bucket]!.pages[l.page]![l.slot]!

export function legacyPut(ctx: Ctx, h: LegacyMap, key: string, val: number): boolean {
  const hash = legacyHash(ctx, h, key)
  ctx.trace.key = key
  ctx.trace.hash = hash
  if (h.buckets === null) {
    h.buckets = allocArray(ctx, h, h.B)
    emit(ctx, 'small.alloc', { impl: 'legacy' })
  }

  for (let attempt = 0; attempt < 4; attempt++) {
    ctx.trace.probes = []
    const bucket = hash & mask(h)
    if (growing(h)) growWork(ctx, h, bucket)

    let free: LLoc | null = null
    const found = scan(ctx, h, h.buckets, bucket, key, hash, (l) => {
      free ??= l
    })
    if (found) {
      const s = slotAt(h, found)
      const old = s.val
      s.val = val
      ctx.trace.hit = found
      ctx.trace.outcome = `ключ уже был — перезаписали значение ${old} → ${val}`
      emit(ctx, 'put.update', { key, old, val, legacy: true, probes: ctx.trace.probes.length })
      return false
    }

    // Ключа нет. Прежде чем вставлять — не пора ли расти?
    if (!growing(h) && (overLoadFactor(h.count + 1, h.B) || tooManyOverflow(h.noverflow, h.B))) {
      hashGrow(ctx, h)
      continue // again: корзина ключа в новом массиве другая
    }

    let at: LLoc | null = free
    if (!at) {
      const b = h.arrays[h.buckets]![bucket]!
      b.pages.push(newPage())
      h.noverflow++
      ctx.world.stats.overflows++
      at = { arr: h.buckets, bucket, page: b.pages.length - 1, slot: 0 }
      emit(ctx, 'bucket.overflow', { key, bucket, chain: b.pages.length, noverflow: h.noverflow, B: h.B })
    }
    const s = slotAt(h, at)
    s.top = tophash(hash)
    s.key = key
    s.val = val
    s.hash = hash
    h.count++
    ctx.trace.hit = at
    ctx.trace.outcome = `новый ключ — корзина ${bucket}${at.page > 0 ? `, переполнение №${at.page}` : ''}, слот ${at.slot}`
    emit(ctx, 'put.new', {
      key,
      val,
      legacy: true,
      bucket,
      page: at.page,
      slot: at.slot,
      tophash: s.top,
      probes: ctx.trace.probes.length,
      count: h.count,
      B: h.B,
    })
    return true
  }
  throw new Error('legacyPut: рост не помог')
}

function hashGrow(ctx: Ctx, h: LegacyMap) {
  const load = overLoadFactor(h.count + 1, h.B)
  const fromB = h.B
  const noverflow = h.noverflow
  h.oldbuckets = h.buckets
  h.B += load ? 1 : 0
  h.buckets = allocArray(ctx, h, h.B)
  h.sameSizeGrow = !load
  h.nevacuate = 0
  h.noverflow = 0
  ctx.world.stats.grows++
  emit(ctx, 'grow.start', {
    sameSize: !load,
    fromB,
    toB: h.B,
    count: h.count,
    noverflow,
    old: 1 << fromB,
    buckets: 1 << h.B,
  })
}

/** Перед записью в корзину — эвакуировать её старую корзину и ещё одну по порядку. */
function growWork(ctx: Ctx, h: LegacyMap, bucket: number) {
  evacuate(ctx, h, bucket & oldMask(h))
  if (growing(h)) evacuate(ctx, h, h.nevacuate)
}

function evacuate(ctx: Ctx, h: LegacyMap, oldb: number) {
  const old = h.arrays[h.oldbuckets!]!
  const b = old[oldb]!
  if (!b.evacuated) {
    const newbit = oldMask(h) + 1
    const dst = h.arrays[h.buckets!]!
    const x = { bucket: oldb, page: 0, slot: 0, n: 0 }
    const y = { bucket: oldb + newbit, page: 0, slot: 0, n: 0 }
    for (const page of b.pages) {
      for (const s of page) {
        if (!isFullTop(s.top)) {
          s.top = TOP_EVAC_EMPTY
          continue
        }
        const useY = !h.sameSizeGrow && (s.hash & newbit) !== 0
        const d = useY ? y : x
        const db = dst[d.bucket]!
        if (d.slot === GROUP_SLOTS) {
          db.pages.push(newPage())
          h.noverflow++
          d.page++
          d.slot = 0
        }
        Object.assign(db.pages[d.page]![d.slot]!, { top: s.top, key: s.key, val: s.val, hash: s.hash })
        d.slot++
        d.n++
        s.top = useY ? TOP_EVAC_Y : TOP_EVAC_X
      }
    }
    b.evacuated = true
    ctx.world.stats.evacuated++
    ctx.world.stats.moved += x.n + y.n
    emit(ctx, 'evac', {
      old: oldb,
      x: x.bucket,
      y: h.sameSizeGrow ? null : y.bucket,
      toX: x.n,
      toY: y.n,
      sameSize: h.sameSizeGrow,
      nevacuate: h.nevacuate,
      total: old.length,
    })
  }

  if (oldb === h.nevacuate) {
    h.nevacuate++
    while (h.nevacuate < old.length && old[h.nevacuate]!.evacuated) h.nevacuate++
    if (h.nevacuate === old.length) {
      emit(ctx, 'grow.done', { sameSize: h.sameSizeGrow, B: h.B, buckets: 1 << h.B, overflow: h.noverflow })
      h.oldbuckets = null
      h.sameSizeGrow = false
    }
  }
}

export function legacyDelete(ctx: Ctx, h: LegacyMap, key: string): boolean {
  const hash = legacyHash(ctx, h, key)
  ctx.trace.key = key
  ctx.trace.hash = hash
  if (h.buckets === null || h.count === 0) {
    ctx.trace.outcome = 'карта пуста — delete ничего не делает'
    emit(ctx, 'del.miss', { key, legacy: true, probes: 0 })
    return false
  }
  const bucket = hash & mask(h)
  if (growing(h)) growWork(ctx, h, bucket)
  const loc = scan(ctx, h, h.buckets, bucket, key, hash)
  if (!loc) {
    ctx.trace.outcome = 'ключа нет — delete ничего не делает'
    emit(ctx, 'del.miss', { key, legacy: true, probes: ctx.trace.probes.length, nan: isNaNKey(key) })
    return false
  }
  const s = slotAt(h, loc)
  Object.assign(s, { top: TOP_EMPTY_ONE, key: null, val: null, hash: 0 })
  const rest = markEmptyRest(h.arrays[h.buckets]![bucket]!, loc.page, loc.slot)
  h.count--
  ctx.trace.hit = loc
  ctx.trace.outcome = rest > 0 ? `слот освобождён, ${rest} слот(ов) с конца помечены emptyRest` : 'слот освобождён и помечен emptyOne'
  emit(ctx, 'del.hit', { key, legacy: true, bucket, page: loc.page, slot: loc.slot, rest, count: h.count })
  if (h.count === 0) {
    const old = h.hash0
    h.hash0 = ctx.rng.u32()
    emit(ctx, 'seed.reset', { old, seed: h.hash0, why: 'delete', legacy: true })
  }
  return true
}

/**
 * Если за удалённым слотом в цепочке больше ничего нет, он и все пустые слоты
 * перед ним становятся emptyRest — поиск сможет остановиться раньше.
 */
function markEmptyRest(b: Bucket, page: number, slot: number): number {
  const pages = b.pages
  if (slot === GROUP_SLOTS - 1) {
    if (page + 1 < pages.length && pages[page + 1]![0]!.top !== TOP_EMPTY_REST) return 0
  } else if (pages[page]![slot + 1]!.top !== TOP_EMPTY_REST) return 0

  let n = 0
  let p = page
  let i = slot
  for (;;) {
    pages[p]![i]!.top = TOP_EMPTY_REST
    n++
    if (i === 0) {
      if (p === 0) break
      p--
      i = GROUP_SLOTS - 1
    } else i--
    if (pages[p]![i]!.top !== TOP_EMPTY_ONE) break
  }
  return n
}

export function legacyClear(ctx: Ctx, h: LegacyMap) {
  const was = h.count
  const overflow = h.buckets === null ? 0 : h.arrays[h.buckets]!.reduce((n, b) => n + b.pages.length - 1, 0)
  if (h.buckets !== null) h.arrays[h.buckets] = h.arrays[h.buckets]!.map(newBucket)
  h.oldbuckets = null
  h.sameSizeGrow = false
  h.nevacuate = 0
  h.noverflow = 0
  h.count = 0
  const old = h.hash0
  h.hash0 = ctx.rng.u32()
  ctx.trace.outcome = `карта пуста, ${1 << h.B} корзин(ы) остались, цепочки переполнения отброшены`
  emit(ctx, 'clear', { was, buckets: 1 << h.B, overflow, impl: 'legacy' })
  emit(ctx, 'seed.reset', { old, seed: h.hash0, why: 'clear', legacy: true })
}

/** Массивы корзин, на которые не ссылаются ни карта, ни итератор, — мусор. */
export function legacyCollect(ctx: Ctx, h: LegacyMap) {
  const it = ctx.world.iter
  const keep = new Set([h.buckets, h.oldbuckets])
  if (it && !it.done && it.impl === 'legacy') {
    keep.add(it.arr)
    if (it.src) keep.add(it.src.arr)
  }
  for (const id of Object.keys(h.arrays).map(Number)) if (!keep.has(id)) delete h.arrays[id]
}

export function legacyCapacity(h: LegacyMap): number {
  if (h.buckets === null) return 0
  return h.arrays[h.buckets]!.reduce((n, b) => n + b.pages.length * GROUP_SLOTS, 0)
}

/* ───────────────────────── Обход ───────────────────────── */

export function legacyIterStart(ctx: Ctx, h: LegacyMap): LegacyIter {
  const r = ctx.rng.u32()
  return {
    impl: 'legacy',
    arr: h.buckets ?? -1,
    B: h.B,
    startBucket: r & mask(h),
    offset: (r >>> h.B) & (GROUP_SLOTS - 1),
    bucketIdx: 0,
    src: null,
    page: 0,
    i: 0,
  }
}

export type LIterStep =
  | { kind: 'yield'; y: Yield; arr: number; bucket: number; page: number; slot: number; visited: number }
  | { kind: 'skip'; key: string; arr: number; bucket: number; page: number; slot: number }
  | { kind: 'done'; visited: number }

/** Один шаг mapiternext. */
export function legacyIterNext(h: LegacyMap, it: LegacyIter & { visited: number }): LIterStep {
  let visited = 0
  if (it.arr < 0) return { kind: 'done', visited }
  const n = 1 << it.B

  for (;;) {
    if (it.src === null) {
      if (it.bucketIdx >= n) return { kind: 'done', visited }
      const bucket = (it.startBucket + it.bucketIdx) & (n - 1)
      it.bucketIdx++
      // Обход начался посреди роста, а рост ещё идёт: старая корзина может быть не перенесена.
      if (growing(h) && it.B === h.B) {
        const ob = bucket & oldMask(h)
        const oldb = h.arrays[h.oldbuckets!]![ob]!
        if (!oldb.evacuated) it.src = { arr: h.oldbuckets!, bucket: ob, check: h.sameSizeGrow ? null : bucket }
        else it.src = { arr: it.arr, bucket, check: null }
      } else it.src = { arr: it.arr, bucket, check: null }
      it.page = 0
      it.i = 0
    }

    const b = h.arrays[it.src.arr]![it.src.bucket]!
    while (it.page < b.pages.length) {
      const page = b.pages[it.page]!
      while (it.i < GROUP_SLOTS) {
        const off = (it.i + it.offset) & (GROUP_SLOTS - 1)
        it.i++
        visited++
        it.visited++
        const s = page[off]!
        if (isEmptyTop(s.top) || s.top === TOP_EVAC_EMPTY) continue
        // Из старой корзины берём только ключи, которые уедут в текущую новую.
        if (it.src.check !== null && (s.hash & ((1 << it.B) - 1)) !== it.src.check) continue
        const at = { arr: it.src.arr, bucket: it.src.bucket, page: it.page, slot: off }
        if ((s.top !== TOP_EVAC_X && s.top !== TOP_EVAC_Y) || isNaNKey(s.key!)) {
          return { kind: 'yield', y: { key: s.key!, val: s.val!, reread: false }, ...at, visited }
        }
        // Запись уже переехала: «золотые» данные — в новом массиве.
        const cur = findQuiet(h, s.key!)
        if (!cur) return { kind: 'skip', key: s.key!, ...at }
        return { kind: 'yield', y: { key: s.key!, val: cur.val!, reread: true }, ...at, visited }
      }
      it.page++
      it.i = 0
    }
    it.src = null
  }
}
