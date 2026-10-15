import type { Ctx } from './ctx.ts'
import { emit, pow2ceil, recordProbe } from './ctx.ts'
import { dirIndex, h2, hashKey, isNaNKey, keyEqual, probeSeq, splitBit } from './hash.ts'
import type { Group, Probe, Slot, SwissIter, SwissMap, Table, Yield } from './types.ts'
import { CTRL_DELETED, CTRL_EMPTY, GROUP_SLOTS } from './types.ts'

/**
 * Swiss table — модель internal/runtime/maps (Go 1.24+).
 *
 * Функции названы и устроены так же, как в рантайме: putSlotSmall, growToTable,
 * PutSlot с запоминанием первого надгробия, pruneTombstones, rehash → grow или
 * split, installTableSplit, Iter.Next. Где модель отступает — сказано рядом.
 */

const emptySlot = (): Slot => ({ ctrl: CTRL_EMPTY, key: null, val: null, hash: 0 })
export const newGroup = (): Group => ({ slots: Array.from({ length: GROUP_SLOTS }, emptySlot) })

export const isFull = (ctrl: number) => ctrl < 0x80
const slotsWhere = (g: Group, pred: (ctrl: number) => boolean) =>
  g.slots.flatMap((s, i) => (pred(s.ctrl) ? [i] : []))

/** matchH2: в каких слотах контрольный байт равен 7 битам хэша. Одна битовая операция в рантайме. */
export const matchH2 = (g: Group, h: number) => slotsWhere(g, (c) => c === h)
const matchEmpty = (g: Group) => slotsWhere(g, (c) => c === CTRL_EMPTY)
const matchDeleted = (g: Group) => slotsWhere(g, (c) => c === CTRL_DELETED)
const matchEmptyOrDeleted = (g: Group) => slotsWhere(g, (c) => !isFull(c))

/** Запас свежей таблицы: 7/8 ёмкости; таблице из одной группы — всё, кроме одного слота. */
export const maxGrowth = (capacity: number) => (capacity <= GROUP_SLOTS ? capacity - 1 : (capacity * 7) / 8)

/** Надгробий в таблице рантайм не хранит, а выводит из запаса. */
export const tombstones = (t: Table) => maxGrowth(t.capacity) - t.used - t.growthLeft

export function createSwiss(ctx: Ctx, hint: number): SwissMap {
  const m: SwissMap = {
    used: 0,
    seed: ctx.rng.u32(),
    small: null,
    dir: [],
    globalDepth: 0,
    tables: {},
    writing: 0,
    tombstonePossible: false,
    clearSeq: 0,
  }
  if (hint <= GROUP_SLOTS) return m

  // make(map[K]V, hint): сразу все таблицы под hint с учётом заполнения 7/8.
  const max = ctx.world.config.maxTableSlots
  const target = Math.floor((hint * GROUP_SLOTS) / 7)
  const dirSize = pow2ceil(Math.ceil(target / max))
  m.globalDepth = Math.log2(dirSize)
  for (let i = 0; i < dirSize; i++) {
    const t = newTable(ctx, Math.floor(target / dirSize), i, m.globalDepth)
    m.tables[t.id] = t
    m.dir.push(t.id)
  }
  return m
}

function newTable(ctx: Ctx, capacity: number, index: number, localDepth: number): Table {
  const cap = Math.max(GROUP_SLOTS, pow2ceil(capacity))
  return {
    id: ctx.world.nextId++,
    capacity: cap,
    used: 0,
    growthLeft: maxGrowth(cap),
    localDepth,
    index,
    groups: Array.from({ length: cap / GROUP_SLOTS }, newGroup),
    replaced: false,
  }
}

export function swissHash(ctx: Ctx, m: SwissMap, key: string): number {
  // NaN хэшируется случайно: иначе все NaN-ключи сваливались бы в одну группу.
  return isNaNKey(key) ? ctx.rng.u32() : hashKey(key, m.seed)
}

export const tableFor = (m: SwissMap, hash: number): Table => m.tables[m.dir[dirIndex(hash, m.globalDepth)]!]!

/** Сравнить ключи у кандидатов matchH2 — с записью в след и счётчики. */
function compareCandidates(ctx: Ctx, g: Group, probe: Probe, key: string, hash: number): number | null {
  for (const i of probe.match) {
    const eq = keyEqual(g.slots[i]!.key, key)
    probe.compared.push({ slot: i, eq })
    ctx.world.stats.compares++
    if (eq) return i
    ctx.world.stats.falseMatches++
    emit(ctx, 'h2.false', {
      key,
      other: g.slots[i]!.key,
      h2: h2(hash),
      group: probe.group,
      slot: i,
      nan: isNaNKey(key),
    })
  }
  return null
}

export interface Loc {
  table: Table | null
  group: number
  slot: number
}

/** Поиск ключа: matchH2 по группам последовательности проб до группы с empty. */
export function swissFind(ctx: Ctx, m: SwissMap, key: string, hash: number): Loc | null {
  if (m.dir.length === 0) {
    if (!m.small) return null
    const p = recordProbe(ctx, { table: null, group: 0, match: matchH2(m.small, h2(hash)), compared: [], stop: true })
    const i = compareCandidates(ctx, m.small, p, key, hash)
    return i === null ? null : { table: null, group: 0, slot: i }
  }
  const t = tableFor(m, hash)
  for (const offset of probeSeq(hash, t.groups.length - 1)) {
    const g = t.groups[offset]!
    const p = recordProbe(ctx, { table: t.id, group: offset, match: matchH2(g, h2(hash)), compared: [], stop: false })
    const i = compareCandidates(ctx, g, p, key, hash)
    if (i !== null) return { table: t, group: offset, slot: i }
    if (matchEmpty(g).length > 0) {
      p.stop = true
      return null
    }
  }
  return null
}

/** Поиск без следа — для итератора, который перечитывает значения. */
function findQuiet(m: SwissMap, key: string): Slot | null {
  if (isNaNKey(key)) return null
  const hash = hashKey(key, m.seed)
  if (m.dir.length === 0) return m.small?.slots.find((s) => isFull(s.ctrl) && s.key === key) ?? null
  const t = tableFor(m, hash)
  for (const offset of probeSeq(hash, t.groups.length - 1)) {
    const g = t.groups[offset]!
    const s = g.slots.find((x) => x.ctrl === h2(hash) && x.key === key)
    if (s) return s
    if (matchEmpty(g).length > 0) return null
  }
  return null
}

export function swissHas(m: SwissMap, key: string): boolean {
  return findQuiet(m, key) !== null
}

function place(slot: Slot, key: string, val: number, hash: number) {
  slot.ctrl = h2(hash)
  slot.key = key
  slot.val = val
  slot.hash = hash
}

/** m[key] = val. Возвращает true, если ключ новый. */
export function swissPut(ctx: Ctx, m: SwissMap, key: string, val: number): boolean {
  const hash = swissHash(ctx, m, key)
  ctx.trace.key = key
  ctx.trace.hash = hash

  if (m.dir.length === 0) {
    // В рантайме проверка именно такая: даже перезапись существующего ключа
    // в полной маленькой карте сначала превращает её в таблицу.
    if (m.used < GROUP_SLOTS) return putSmall(ctx, m, key, val, hash)
    growToTable(ctx, m)
  }
  return putTable(ctx, m, key, val, hash, 0)
}

function putSmall(ctx: Ctx, m: SwissMap, key: string, val: number, hash: number): boolean {
  if (!m.small) {
    m.small = newGroup()
    emit(ctx, 'small.alloc', { impl: 'swiss' })
  }
  const g = m.small
  const p = recordProbe(ctx, { table: null, group: 0, match: matchH2(g, h2(hash)), compared: [], stop: true })
  const found = compareCandidates(ctx, g, p, key, hash)
  if (found !== null) {
    const old = g.slots[found]!.val
    g.slots[found]!.val = val
    ctx.trace.hit = { table: null, group: 0, slot: found }
    ctx.trace.outcome = `ключ уже был — перезаписали значение ${old} → ${val}`
    emit(ctx, 'put.update', { key, old, val, small: true, probes: 1 })
    return false
  }
  // В маленькой карте нет проб, поэтому нет и надгробий: свободный слот — всегда empty.
  const i = matchEmpty(g)[0]!
  place(g.slots[i]!, key, val, hash)
  m.used++
  ctx.trace.hit = { table: null, group: 0, slot: i }
  ctx.trace.outcome = `новый ключ — в слот ${i} единственной группы`
  emit(ctx, 'put.new', { key, val, h2: h2(hash), small: true, slot: i, used: m.used, probes: 1 })
  return true
}

/** Девятый ключ: маленькая карта становится таблицей на 16 слотов. */
function growToTable(ctx: Ctx, m: SwissMap) {
  const small = m.small!
  const t = newTable(ctx, 2 * GROUP_SLOTS, 0, 0)
  for (const s of small.slots) if (isFull(s.ctrl)) uncheckedPut(t, s)
  m.tables[t.id] = t
  m.dir = [t.id]
  m.globalDepth = 0
  m.small = null
  const it = ctx.world.iter
  if (it && it.impl === 'swiss' && it.small && !it.done) it.smallGroup = small
  ctx.world.stats.moved += t.used
  emit(ctx, 'small.promote', { moved: t.used, capacity: t.capacity, table: t.id, iterHolds: Boolean(it && !it.done && it.impl === 'swiss' && it.small) })
}

/** Вставка без поиска — при росте: ключи заведомо разные, надгробий в новой таблице нет. */
function uncheckedPut(t: Table, s: Slot) {
  for (const offset of probeSeq(s.hash, t.groups.length - 1)) {
    const g = t.groups[offset]!
    const free = matchEmptyOrDeleted(g)
    if (free.length === 0) continue
    place(g.slots[free[0]!]!, s.key!, s.val!, s.hash)
    t.used++
    t.growthLeft--
    return
  }
  throw new Error('uncheckedPut: в таблице нет места')
}

function putTable(ctx: Ctx, m: SwissMap, key: string, val: number, hash: number, attempt: number): boolean {
  if (attempt > 8) throw new Error('putTable: таблица не растёт')
  ctx.trace.probes = []
  const t = tableFor(m, hash)
  let tomb: { group: number; slot: number } | null = null

  for (const offset of probeSeq(hash, t.groups.length - 1)) {
    const g = t.groups[offset]!
    const p = recordProbe(ctx, { table: t.id, group: offset, match: matchH2(g, h2(hash)), compared: [], stop: false })
    const found = compareCandidates(ctx, g, p, key, hash)
    if (found !== null) {
      const old = g.slots[found]!.val
      g.slots[found]!.val = val
      ctx.trace.hit = { table: t.id, group: offset, slot: found }
      ctx.trace.outcome = `ключ уже был — перезаписали значение ${old} → ${val}`
      emit(ctx, 'put.update', { key, old, val, probes: ctx.trace.probes.length })
      return false
    }

    if (matchEmptyOrDeleted(g).length === 0) continue // группа забита — дальше по пробам
    const empty = matchEmpty(g)
    if (empty.length === 0) {
      // Здесь только надгробия: ключ ещё может найтись дальше, а место запомним.
      if (!tomb) tomb = { group: offset, slot: matchDeleted(g)[0]! }
      continue
    }

    // Группа с empty — конец последовательности: ключа в таблице нет.
    p.stop = true
    let target = { group: offset, slot: empty[0]! }
    if (tomb) {
      target = tomb
      t.growthLeft++ // надгробие и так съело запас — ниже он вернётся на место
    }
    if (t.growthLeft > 0) {
      place(t.groups[target.group]!.slots[target.slot]!, key, val, hash)
      t.growthLeft--
      t.used++
      m.used++
      ctx.trace.hit = { table: t.id, group: target.group, slot: target.slot }
      ctx.trace.outcome = tomb
        ? `новый ключ — на место надгробия в группе ${target.group}`
        : `новый ключ — в группу ${target.group}, слот ${target.slot}`
      emit(ctx, 'put.new', {
        key,
        val,
        h2: h2(hash),
        group: target.group,
        slot: target.slot,
        probes: ctx.trace.probes.length,
        growthLeft: t.growthLeft,
        used: m.used,
        attempt,
      })
      if (tomb) emit(ctx, 'tomb.reuse', { key, group: target.group, slot: target.slot, growthLeft: t.growthLeft })
      return true
    }

    // Места нет: сначала прополка надгробий, потом рост.
    if (!pruneTombstones(ctx, t)) rehash(ctx, m, t)
    return putTable(ctx, m, key, val, hash, attempt + 1)
  }
  throw new Error('putTable: последовательность проб кончилась без пустой группы')
}

/**
 * Прополка: надгробия, через которые не «проскакивает» ни один живой ключ,
 * снова становятся empty. Запускается, только если надгробий ≥ 10% ёмкости,
 * и применяется, только если освободит ≥ 10%.
 */
function pruneTombstones(ctx: Ctx, t: Table): boolean {
  const tombs = tombstones(t)
  if (tombs * 10 < t.capacity) return false

  const mask = t.groups.length - 1
  const needed = new Set<number>()
  t.groups.forEach((g, gi) => {
    for (const s of g.slots) {
      if (!isFull(s.ctrl) || isNaNKey(s.key!)) continue // NaN не ищут — его цепочку можно рвать
      for (const offset of probeSeq(s.hash, mask)) {
        if (offset === gi) break
        if (matchEmptyOrDeleted(t.groups[offset]!).length > 0) needed.add(offset)
      }
    }
  })
  let removable = 0
  t.groups.forEach((g, gi) => {
    if (!needed.has(gi)) removable += matchDeleted(g).length
  })

  if (removable * 10 < t.capacity) {
    emit(ctx, 'tomb.prune', { ok: false, tombs, removable, needed: needed.size, capacity: t.capacity, table: t.id })
    return false
  }
  t.groups.forEach((g, gi) => {
    if (needed.has(gi)) return
    for (const i of matchDeleted(g)) {
      g.slots[i]!.ctrl = CTRL_EMPTY
      t.growthLeft++
    }
  })
  ctx.world.stats.prunes++
  emit(ctx, 'tomb.prune', { ok: true, tombs, removable, kept: tombs - removable, needed: needed.size, capacity: t.capacity, table: t.id })
  return t.growthLeft > 0
}

function rehash(ctx: Ctx, m: SwissMap, t: Table) {
  const newCapacity = 2 * t.capacity
  if (newCapacity <= ctx.world.config.maxTableSlots) grow(ctx, m, t, newCapacity)
  else split(ctx, m, t)
}

function retire(t: Table) {
  t.replaced = true
  t.index = -1
}

function grow(ctx: Ctx, m: SwissMap, t: Table, newCapacity: number) {
  const nt = newTable(ctx, newCapacity, t.index, t.localDepth)
  const tombs = tombstones(t)
  for (const g of t.groups) for (const s of g.slots) if (isFull(s.ctrl)) uncheckedPut(nt, s)
  m.tables[nt.id] = nt
  m.dir = m.dir.map((id) => (id === t.id ? nt.id : id))
  retire(t)
  ctx.world.stats.grows++
  ctx.world.stats.moved += nt.used
  emit(ctx, 'table.grow', {
    old: t.id,
    table: nt.id,
    from: t.capacity,
    to: nt.capacity,
    moved: nt.used,
    tombs,
    max: ctx.world.config.maxTableSlots,
  })
}

function split(ctx: Ctx, m: SwissMap, t: Table) {
  const depth = t.localDepth + 1
  const left = newTable(ctx, t.capacity, -1, depth)
  const right = newTable(ctx, t.capacity, -1, depth)
  const bit = splitBit(depth)
  for (const g of t.groups) {
    for (const s of g.slots) if (isFull(s.ctrl)) uncheckedPut((s.hash & bit) === 0 ? left : right, s)
  }

  // installTableSplit: если строк каталога на две половины не хватает — каталог удваивается.
  if (t.localDepth === m.globalDepth) {
    const from = m.dir.length
    m.dir = m.dir.flatMap((id) => [id, id])
    m.globalDepth++
    ctx.world.stats.dirGrows++
    emit(ctx, 'dir.grow', { from, to: m.dir.length, depth: m.globalDepth, table: t.id })
  }
  m.dir = m.dir.map((id, i) => (id !== t.id ? id : ((i >> (m.globalDepth - depth)) & 1) === 0 ? left.id : right.id))
  m.tables[left.id] = left
  m.tables[right.id] = right
  retire(t)
  reindex(m)

  ctx.world.stats.splits++
  ctx.world.stats.moved += left.used + right.used
  emit(ctx, 'table.split', {
    old: t.id,
    left: left.id,
    right: right.id,
    leftN: left.used,
    rightN: right.used,
    capacity: t.capacity,
    depth,
    globalDepth: m.globalDepth,
    bit: 32 - depth,
  })
}

/** index таблицы — первая строка каталога, которая на неё указывает. */
function reindex(m: SwissMap) {
  for (const t of Object.values(m.tables)) if (!t.replaced) t.index = m.dir.indexOf(t.id)
}

export function swissDelete(ctx: Ctx, m: SwissMap, key: string): boolean {
  const hash = swissHash(ctx, m, key)
  ctx.trace.key = key
  ctx.trace.hash = hash
  const loc = swissFind(ctx, m, key, hash)
  if (!loc) {
    ctx.trace.outcome = 'ключа нет — delete ничего не делает'
    emit(ctx, 'del.miss', { key, probes: ctx.trace.probes.length, nan: isNaNKey(key) })
    return false
  }
  ctx.trace.hit = { table: loc.table?.id ?? null, group: loc.group, slot: loc.slot }

  if (!loc.table) {
    const s = m.small!.slots[loc.slot]!
    Object.assign(s, emptySlot())
    m.used--
    ctx.trace.outcome = 'слот освобождён — в маленькой карте надгробий нет'
    emit(ctx, 'del.hit', { key, small: true, slot: loc.slot, used: m.used })
  } else {
    const t = loc.table
    const g = t.groups[loc.group]!
    const hasEmpty = matchEmpty(g).length > 0
    const s = g.slots[loc.slot]!
    s.key = null
    s.val = null
    if (hasEmpty) {
      s.ctrl = CTRL_EMPTY
      t.growthLeft++
      ctx.trace.outcome = `в группе ${loc.group} уже есть empty — слот просто освобождён`
    } else {
      s.ctrl = CTRL_DELETED
      m.tombstonePossible = true
      ctx.trace.outcome = `группа ${loc.group} была полной — на месте ключа надгробие`
    }
    t.used--
    m.used--
    emit(ctx, 'del.hit', { key, group: loc.group, slot: loc.slot, tomb: !hasEmpty, used: m.used, growthLeft: t.growthLeft })
    if (!hasEmpty) emit(ctx, 'tomb.set', { key, group: loc.group, slot: loc.slot, tombs: tombstones(t), capacity: t.capacity })
  }

  if (m.used === 0) {
    const old = m.seed
    m.seed = ctx.rng.u32()
    emit(ctx, 'seed.reset', { old, seed: m.seed, why: 'delete' })
  }
  return true
}

export function swissClear(ctx: Ctx, m: SwissMap) {
  const tables = Object.values(m.tables).filter((t) => !t.replaced)
  const capacity = m.dir.length === 0 ? (m.small ? GROUP_SLOTS : 0) : tables.reduce((n, t) => n + t.capacity, 0)
  const was = m.used
  if (m.small) m.small = newGroup()
  for (const t of tables) {
    t.groups = t.groups.map(newGroup)
    t.used = 0
    t.growthLeft = maxGrowth(t.capacity)
  }
  m.used = 0
  m.clearSeq++
  m.tombstonePossible = false
  const old = m.seed
  m.seed = ctx.rng.u32()
  ctx.trace.outcome = `карта пуста, но ${tables.length || (m.small ? 1 : 0)} ${tables.length ? 'таблиц(ы) остались' : 'группа осталась'} на месте`
  emit(ctx, 'clear', { was, capacity, tables: tables.length, small: m.dir.length === 0, impl: 'swiss' })
  emit(ctx, 'seed.reset', { old, seed: m.seed, why: 'clear' })
}

/** Таблицы, которые заменили, — мусор, если их не держит итератор. */
export function swissCollect(ctx: Ctx, m: SwissMap) {
  const it = ctx.world.iter
  const held = it && !it.done && it.impl === 'swiss' ? it.tab : null
  for (const t of Object.values(m.tables)) if (t.replaced && t.id !== held) delete m.tables[t.id]
}

/* ───────────────────────── Обход ───────────────────────── */

export function swissIterStart(ctx: Ctx, m: SwissMap): SwissIter {
  return {
    impl: 'swiss',
    entryOffset: ctx.rng.u32(),
    dirOffset: ctx.rng.u32(),
    dirIdx: 0,
    entryIdx: 0,
    tab: null,
    globalDepth: m.globalDepth,
    small: m.dir.length === 0,
    smallGroup: null,
  }
}

export type IterStep =
  | { kind: 'yield'; y: Yield; table: number | null; group: number; slot: number; visited: number }
  | { kind: 'skip'; key: string; table: number | null; group: number; slot: number }
  | { kind: 'done'; visited: number }

/**
 * Один шаг Iter.Next. Пустые слоты пролистываются тут же — их число и есть
 * цена обхода, — а вот ключ, который удалили, пока итератор держал старую
 * таблицу, — отдельный шаг, чтобы его было видно.
 */
export function swissIterNext(m: SwissMap, it: SwissIter & { visited: number }): IterStep {
  let visited = 0

  if (it.small) {
    const g = it.smallGroup ?? m.small
    if (!g) return { kind: 'done', visited }
    while (it.entryIdx < GROUP_SLOTS) {
      const k = (it.entryIdx + it.entryOffset) & (GROUP_SLOTS - 1)
      it.entryIdx++
      visited++
      it.visited++
      const s = g.slots[k]!
      if (!isFull(s.ctrl)) continue
      if (!it.smallGroup) return { kind: 'yield', y: { key: s.key!, val: s.val!, reread: false }, table: null, group: 0, slot: k, visited }
      // Карта выросла в таблицу: ключ берём из старой группы, значение — из карты.
      const r = reread(m, s)
      if (!r) return { kind: 'skip', key: s.key!, table: null, group: 0, slot: k }
      return { kind: 'yield', y: r, table: null, group: 0, slot: k, visited }
    }
    return { kind: 'done', visited }
  }

  for (;;) {
    // Каталог мог удвоиться, пока шёл обход: позиция пересчитывается в новые строки.
    if (it.globalDepth !== m.globalDepth) {
      const orders = m.globalDepth - it.globalDepth
      it.dirIdx <<= orders
      it.dirOffset = (it.dirOffset << orders) >>> 0
      it.globalDepth = m.globalDepth
    }
    if (it.dirIdx >= m.dir.length) return { kind: 'done', visited }

    if (it.tab === null) {
      let di = (it.dirIdx + it.dirOffset) & (m.dir.length - 1)
      const t = m.tables[m.dir[di]!]!
      if (t.index !== di) {
        // Случайный старт попал в середину строк одной таблицы — выравниваемся на первую.
        it.dirOffset = (it.dirOffset - (di - t.index)) >>> 0
        di = t.index
      }
      it.tab = t.id
      it.entryIdx = 0
    }

    const t = m.tables[it.tab]!
    while (it.entryIdx < t.capacity) {
      const e = (it.entryIdx + it.entryOffset) & (t.capacity - 1)
      it.entryIdx++
      visited++
      it.visited++
      const group = e >> 3
      const slot = e & 7
      const s = t.groups[group]!.slots[slot]!
      if (!isFull(s.ctrl)) continue
      if (!t.replaced) return { kind: 'yield', y: { key: s.key!, val: s.val!, reread: false }, table: t.id, group, slot, visited }
      const r = reread(m, s)
      if (!r) return { kind: 'skip', key: s.key!, table: t.id, group, slot }
      return { kind: 'yield', y: r, table: t.id, group, slot, visited }
    }

    // Таблица пройдена: перепрыгиваем все строки каталога, которые на неё указывали.
    it.dirIdx += 1 << (m.globalDepth - t.localDepth)
    it.tab = null
  }
}

/** Ключ из старой таблицы — значение из новой. Удалённый ключ пропускается. */
function reread(m: SwissMap, s: Slot): Yield | null {
  if (isNaNKey(s.key!)) return { key: s.key!, val: s.val!, reread: false } // NaN не найти — отдаём как есть
  const cur = findQuiet(m, s.key!)
  return cur ? { key: s.key!, val: cur.val!, reread: true } : null
}

/** Ёмкость карты в слотах — по живым таблицам. */
export function swissCapacity(m: SwissMap): number {
  if (m.dir.length === 0) return m.small ? GROUP_SLOTS : 0
  return Object.values(m.tables)
    .filter((t) => !t.replaced)
    .reduce((n, t) => n + t.capacity, 0)
}

