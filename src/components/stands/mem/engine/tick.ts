import type { Goroutine, HeapObject, Span } from './types.ts'
import { MAX_SMALL, PAGE, SIZE_CLASSES, TINY_SIZE, classFor } from './types.ts'
import type { Ctx } from './world.ts'
import {
  allocPages,
  cacheOf,
  centralOf,
  emit,
  freeSpan,
  getSpan,
  newSpan,
  releasePages,
  usedPages,
} from './world.ts'
import type { Rng } from './rng.ts'

/**
 * Один шаг всей системы.
 *
 * Порядок фиксирован: сначала умирают объекты и возвращается память, потом
 * работают горутины. Так внутри одного тика освободившийся слот успевает
 * достаться новому объекту — как и в жизни.
 */
export function tick(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  if (w.finished) return
  w.tick++
  w.slots = []

  phaseSpawn(ctx)
  phaseReclaim(ctx)
  phaseScavenge(ctx)
  phaseShrinkStacks(ctx)
  phaseRun(ctx, rng)
  phaseAccount(ctx)
}

/* ───────────────────────────── горутины ───────────────────────────── */

function phaseSpawn(ctx: Ctx): void {
  const w = ctx.world
  while (w.spawnPtr < ctx.plan.length) {
    const item = ctx.plan[w.spawnPtr]
    if (!item || item.tick >= w.tick) break
    w.spawnPtr++
    createG(ctx, item.workload)
  }
}

function createG(ctx: Ctx, workload: number): void {
  const w = ctx.world
  const wl = ctx.scenario.workloads[workload]
  if (!wl) throw new Error(`нет нагрузки #${workload}`)
  const g: Goroutine = {
    id: w.nextGid++,
    name: wl.name,
    workload,
    state: 'runnable',
    stackSize: w.config.stackStart,
    stackUsed: 0,
    stackPages: [],
    stackSlot: null,
    copyLeft: 0,
    depth: 0,
    phaseIdx: -1,
    phaseLeft: 0,
    repeatsLeft: wl.repeat === 'forever' ? Number.POSITIVE_INFINITY : (wl.repeat ?? 1) - 1,
    createdTick: w.tick,
    cpuTicks: 0,
    stallTicks: 0,
    allocBytes: 0,
    allocLeft: 0,
    stackGrows: 0,
  }
  w.gs.push(g)
  attachStack(ctx, g, g.stackSize)
  emit(ctx, 'g.start', { g: [g.id] }, { name: wl.name, stack: g.stackSize })
}

/* ──────────────────────────────── стеки ──────────────────────────────── */

/** Выдать горутине память под стек: слот в общем куске или свои страницы. */
function attachStack(ctx: Ctx, g: Goroutine, size: number): boolean {
  const w = ctx.world
  if (size >= PAGE) {
    const pages = allocPages(ctx, { n: size / PAGE, kind: 'stack', owner: g.id, g: g.id })
    if (!pages) return false
    g.stackPages = pages
    g.stackSlot = null
    return true
  }

  let chunk = w.stackChunks.find((c) => c.size === size && c.free.length > 0)
  if (!chunk) {
    const pages = allocPages(ctx, { n: 1, kind: 'stack', owner: -1, g: g.id })
    if (!pages) return false
    const slots = Math.floor(PAGE / size)
    chunk = {
      id: w.nextChunkId++,
      size,
      page: pages[0]!,
      slots,
      free: Array.from({ length: slots }, (_, i) => i),
    }
    w.stackChunks.push(chunk)
  }
  const index = chunk.free.shift()!
  g.stackPages = []
  g.stackSlot = { chunk: chunk.id, index }
  return true
}

/** Вернуть память из-под стека. */
function detachStack(ctx: Ctx, g: Goroutine): void {
  const w = ctx.world
  if (g.stackSlot) {
    const chunk = w.stackChunks.find((c) => c.id === g.stackSlot!.chunk)
    if (chunk) {
      chunk.free.push(g.stackSlot.index)
      if (chunk.free.length === chunk.slots) {
        releasePages(ctx, [chunk.page])
        w.stackChunks = w.stackChunks.filter((c) => c.id !== chunk.id)
      }
    }
    g.stackSlot = null
  }
  if (g.stackPages.length > 0) {
    releasePages(ctx, g.stackPages)
    g.stackPages = []
  }
}

/**
 * Стек кончился.
 *
 * Рантайм не может «дорастить» стек на месте: рядом чужая память. Он выделяет
 * вдвое больший кусок, копирует туда всё содержимое и правит все указатели,
 * которые вели внутрь старого стека. Отсюда и цена — она пропорциональна тому,
 * сколько стека реально занято.
 */
function growStack(ctx: Ctx, g: Goroutine, need: number): boolean {
  const w = ctx.world
  let size = g.stackSize
  while (size < need) size *= 2

  detachStack(ctx, g)
  if (!attachStack(ctx, g, size)) return false

  const copied = g.stackUsed
  const ticks = Math.max(1, Math.ceil(copied / w.config.stackCopyRate))
  const old = g.stackSize
  g.stackSize = size
  g.state = 'growing'
  g.copyLeft = ticks
  g.stackGrows++
  w.stats.stackGrows++
  w.stats.stackCopyBytes += copied

  emit(ctx, 'stack.grow', { g: [g.id] }, {
    from: old,
    to: size,
    copied,
    ticks,
    grows: g.stackGrows,
  })
  return true
}

/**
 * Ужать стек, если горутина давно не заглядывала так глубоко.
 * В рантайме это делает сборщик мусора, когда просматривает стек.
 */
function phaseShrinkStacks(ctx: Ctx): void {
  const w = ctx.world
  if (w.config.shrinkEvery <= 0 || w.tick % w.config.shrinkEvery !== 0) return

  for (const g of w.gs) {
    if (g.state === 'done' || g.state === 'growing') continue
    if (g.stackSize <= w.config.stackStart) continue
    if (g.stackUsed >= g.stackSize / 4) continue
    const half = Math.max(w.config.stackStart, g.stackSize / 2)
    detachStack(ctx, g)
    if (!attachStack(ctx, g, half)) return
    emit(ctx, 'stack.shrink', { g: [g.id] }, { from: g.stackSize, to: half, used: g.stackUsed })
    g.stackSize = half
    w.stats.stackShrinks++
  }
}

/* ─────────────────────────── аллокация ─────────────────────────── */

const align = (x: number, a: number) => Math.ceil(x / a) * a

interface AllocOpts {
  escapes: boolean
  pointers: boolean
  lifetime: number
}

/** Путь, по которому пошла аллокация, — он же подпись на карточке процессора. */
type AllocPath = 'stack' | 'tiny' | 'fast' | 'refill' | 'large' | 'blocked'

function allocate(ctx: Ctx, g: Goroutine, p: number, size: number, o: AllocOpts): AllocPath {
  const w = ctx.world
  w.stats.askedBytes += size
  g.allocBytes += size

  // 1. Объект не убегает из функции — он просто кадр на стеке.
  if (!o.escapes && w.config.escapeAnalysis) {
    const need = g.stackUsed + align(size, 8)
    if (need > g.stackSize) {
      growStack(ctx, g, need)
      return 'blocked'
    }
    g.stackUsed = need
    w.stats.onStack++
    w.stats.slotBytes += size
    emit(ctx, 'alloc.stack', { g: [g.id] }, { size, stackUsed: g.stackUsed, stackSize: g.stackSize })
    return 'stack'
  }

  // 2. Мелочь без указателей складывается в один общий блок на 16 байт.
  if (w.config.tinyAllocator && size < TINY_SIZE && !o.pointers) {
    const c = cacheOf(w, p)
    const off = align(c.tinyOffset, size >= 8 ? 8 : size >= 4 ? 4 : 1)
    if (c.tinySpanObj !== null && off + size <= TINY_SIZE) {
      c.tinyOffset = off + size
      c.tinyObjects++
      w.stats.tiny++
      emit(ctx, 'alloc.tiny', { g: [g.id], p: [p] }, {
        size,
        objects: c.tinyObjects,
        used: c.tinyOffset,
        free: TINY_SIZE - c.tinyOffset,
      })
      return 'tiny'
    }
    // Место в блоке кончилось — берём новый блок обычным путём, классом 16.
    const path = allocSmall(ctx, g, p, TINY_SIZE, TINY_SIZE, o, true)
    if (path !== 'blocked') {
      c.tinyOffset = size
      c.tinyObjects = 1
      c.tinySpanObj = w.nextObjId - 1
      w.stats.tiny++
      w.stats.tinyBlocks++
      emit(ctx, 'alloc.tiny', { g: [g.id], p: [p] }, { size, objects: 1, used: size, free: TINY_SIZE - size, fresh: true })
    }
    return path
  }

  // 3. Большой объект идёт мимо всех кэшей — прямо в кучу, целыми страницами.
  if (size > MAX_SMALL) {
    const npages = Math.ceil(size / PAGE)
    const pages = allocPages(ctx, { n: npages, kind: 'large', owner: w.nextObjId, g: g.id })
    if (!pages) return 'blocked'
    const obj: HeapObject = {
      id: w.nextObjId++,
      size,
      slotSize: npages * PAGE,
      sizeClass: -1,
      span: null,
      pages,
      owner: g.workload,
      bornTick: w.tick,
      diesAt: w.tick + o.lifetime,
      tiny: false,
    }
    w.objects.push(obj)
    w.stats.large++
    w.stats.slotBytes += obj.slotSize
    w.stats.roundAsked += size
    w.stats.roundSlot += obj.slotSize
    emit(ctx, 'alloc.large', { g: [g.id], pages }, {
      size,
      pages: npages,
      slotSize: obj.slotSize,
      waste: obj.slotSize - size,
      heapPages: usedPages(w),
    })
    return 'large'
  }

  // 4. Обычный мелкий объект: класс размеров → кэш процессора → mcentral → куча.
  const cls = classFor(size)!
  return allocSmall(ctx, g, p, size, cls.size, o, false)
}

function allocSmall(
  ctx: Ctx,
  g: Goroutine,
  p: number,
  size: number,
  slotSize: number,
  o: AllocOpts,
  forTiny: boolean,
): AllocPath {
  const w = ctx.world
  const cls = classFor(slotSize)!
  const cache = cacheOf(w, p)

  let path: AllocPath = 'fast'
  let spanId = cache.spans[cls.index]
  let span = spanId === undefined ? undefined : getSpan(w, spanId)

  if (!span || span.used >= span.slots) {
    const refilled = refill(ctx, g, p, cls.index)
    if (refilled === null) return 'blocked'
    span = refilled
    path = 'refill'
  }

  span.used++
  const waste = slotSize - size
  const obj: HeapObject = {
    id: w.nextObjId++,
    size,
    slotSize,
    sizeClass: cls.index,
    span: span.id,
    pages: [],
    owner: g.workload,
    bornTick: w.tick,
    diesAt: w.tick + o.lifetime,
    tiny: forTiny,
  }
  w.objects.push(obj)
  w.stats.slotBytes += slotSize
  if (!forTiny) {
    w.stats.roundAsked += size
    w.stats.roundSlot += slotSize
  }
  if (path === 'fast') w.stats.fast++

  if (!forTiny && waste > 0 && waste / slotSize >= 0.2) {
    emit(ctx, 'class.waste', { g: [g.id] }, {
      size,
      slotSize,
      waste,
      percent: Math.round((waste / slotSize) * 100),
      prevClass: cls.index > 0 ? SIZE_CLASSES[cls.index - 1] : null,
    })
  }
  if (path === 'fast' && !forTiny) {
    emit(ctx, 'alloc.fast', { g: [g.id], p: [p], span: [span.id] }, {
      size,
      slotSize,
      left: span.slots - span.used,
      sizeClass: cls.index,
    })
  }
  return path
}

/**
 * Кэш процессора опустел — идём за новым спаном в общий список этого класса.
 *
 * Здесь кончается быстрый путь без блокировок: mcentral один на весь процесс,
 * и если в тот же тик за тем же классом пришёл другой процессор, придётся ждать.
 */
function refill(ctx: Ctx, g: Goroutine, p: number, cls: number): Span | null {
  const w = ctx.world
  const central = centralOf(w, cls)
  const cache = cacheOf(w, p)

  if (w.config.centralLock && central.lockedAt === w.tick && central.lockedBy !== p) {
    w.stats.contended++
    w.stats.contendedTicks++
    g.stallTicks++
    emit(ctx, 'central.contended', { g: [g.id], p: [p] }, {
      sizeClass: cls,
      objSize: SIZE_CLASSES[cls],
      holder: central.lockedBy,
    })
    return null
  }
  central.lockedAt = w.tick
  central.lockedBy = p

  // Спан, который был у кэша, уезжает обратно в центральный список.
  const oldId = cache.spans[cls]
  if (oldId !== undefined) {
    const old = getSpan(w, oldId)
    old.owner = { kind: 'mcentral' }
    if (old.used >= old.slots) central.full.push(old.id)
    else if (old.used > 0) central.partial.push(old.id)
  }

  let span: Span | null = null
  const fromPartial = central.partial.shift()
  if (fromPartial !== undefined) {
    span = getSpan(w, fromPartial)
  } else {
    emit(ctx, 'central.empty', { g: [g.id], p: [p] }, {
      sizeClass: cls,
      objSize: SIZE_CLASSES[cls],
      full: central.full.length,
    })
    span = newSpan(ctx, cls, g.id)
    if (!span) return null
  }

  span.owner = { kind: 'mcache', p }
  cache.spans[cls] = span.id
  w.stats.refills++
  emit(ctx, 'cache.refill', { g: [g.id], p: [p], span: [span.id] }, {
    sizeClass: cls,
    objSize: SIZE_CLASSES[cls],
    free: span.slots - span.used,
    fromCentral: fromPartial !== undefined,
  })
  return span
}

/* ───────────────────────── возврат памяти ───────────────────────── */

/**
 * Объекты с истёкшим сроком жизни освобождаются.
 * В настоящем рантайме это делает сборщик мусора — см. предыдущую лекцию.
 */
function phaseReclaim(ctx: Ctx): void {
  const w = ctx.world
  const dead = w.objects.filter((o) => o.diesAt <= w.tick)
  if (dead.length === 0) return
  w.objects = w.objects.filter((o) => o.diesAt > w.tick)

  const freedSpans: number[] = []
  for (const obj of dead) {
    if (obj.span !== null) {
      const span = getSpan(w, obj.span)
      span.used--
      if (span.owner.kind === 'mcentral') {
        const central = centralOf(w, span.sizeClass)
        // Спан, в котором снова есть место, переезжает из «занятых» в «частичные»:
        // в рантайме это делает подметальщик.
        if (span.used > 0 && central.full.includes(span.id)) {
          central.full = central.full.filter((x) => x !== span.id)
          central.partial.push(span.id)
        }
        if (span.used <= 0) freedSpans.push(span.id)
      }
    } else if (obj.pages.length > 0) {
      releasePages(ctx, obj.pages)
    }
  }
  emit(ctx, 'obj.free', {}, { count: dead.length, bytes: dead.reduce((s, o) => s + o.slotSize, 0) })

  for (const id of freedSpans) {
    const span = getSpan(w, id)
    if (span.used > 0 || span.owner.kind !== 'mcentral') continue
    const central = centralOf(w, span.sizeClass)
    central.partial = central.partial.filter((x) => x !== id)
    central.full = central.full.filter((x) => x !== id)
    freeSpan(ctx, span)
  }
}

/**
 * Страницы, которые давно никому не нужны, возвращаются операционной системе.
 * Адреса остаются за процессом, но физической памяти за ними больше нет.
 */
function phaseScavenge(ctx: Ctx): void {
  const w = ctx.world
  if (w.config.scavengeAfter <= 0) return
  const ready = w.pages.filter(
    (p) => p.kind === 'free' && p.freeSince !== null && w.tick - p.freeSince >= w.config.scavengeAfter,
  )
  if (ready.length === 0) return
  for (const p of ready) {
    p.kind = 'returned'
    p.freeSince = null
  }
  w.stats.returned += ready.length
  emit(ctx, 'scavenge', { pages: ready.map((p) => p.id) }, {
    pages: ready.length,
    bytes: ready.length * PAGE,
    idle: w.config.scavengeAfter,
  })
}

/* ──────────────────────────── исполнение ──────────────────────────── */

function rotate<T>(items: T[], by: number): T[] {
  if (items.length === 0) return items
  const k = by % items.length
  return [...items.slice(k), ...items.slice(0, k)]
}

function phaseRun(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  const ready = w.gs.filter((g) => g.state !== 'done')
  for (const g of ready) if (g.state !== 'growing') g.state = 'runnable'

  const order = rotate(ready, w.tick)
  let p = 0
  let i = 0
  while (p < w.config.gomaxprocs && i < order.length) {
    const g = order[i++]!
    const path = step(ctx, g, p, rng)
    w.slots.push({ p, g: g.id, path })
    p++
    if (w.finished) break
  }
  while (p < w.config.gomaxprocs) {
    w.slots.push({ p, g: null, path: null })
    p++
  }
}

function step(ctx: Ctx, g: Goroutine, p: number, rng: Rng): string {
  // Стек копируется — горутина в это время не делает ничего.
  if (g.state === 'growing') {
    g.copyLeft--
    g.stallTicks++
    if (g.copyLeft <= 0) g.state = 'running'
    return 'grow'
  }

  g.state = 'running'
  if (g.phaseIdx === -1 && !advance(ctx, g, rng)) return 'done'

  const wl = ctx.scenario.workloads[g.workload]
  const ph = wl?.phases[g.phaseIdx]
  if (!ph) return 'done'

  switch (ph.kind) {
    case 'cpu': {
      g.cpuTicks++
      g.phaseLeft--
      if (g.phaseLeft <= 0) advance(ctx, g, rng)
      return 'cpu'
    }
    case 'alloc': {
      const size = rng.size(ph.size)
      const path = allocate(ctx, g, p, size, {
        escapes: ph.escapes !== false,
        pointers: ph.pointers !== false,
        lifetime: rng.duration(ph.lifetime ?? [8, 20]),
      })
      // Заблокированная аллокация повторится в следующем тике — счётчик не трогаем.
      if (path !== 'blocked') {
        g.allocLeft--
        if (g.allocLeft <= 0) advance(ctx, g, rng)
      }
      return path
    }
    case 'recurse': {
      const need = g.stackUsed + ph.frame
      if (need > g.stackSize) {
        growStack(ctx, g, need)
        return 'grow'
      }
      g.stackUsed = need
      g.depth++
      g.cpuTicks++
      if (g.depth >= ph.depth) {
        // Возврат из рекурсии: кадры уходят разом, память стека остаётся за горутиной.
        g.stackUsed = Math.max(0, g.stackUsed - ph.depth * ph.frame)
        g.depth = 0
        advance(ctx, g, rng)
      }
      return 'recurse'
    }
  }
}

/** Перейти к следующей фазе. false — горутина закончилась. */
function advance(ctx: Ctx, g: Goroutine, rng: Rng): boolean {
  const w = ctx.world
  const wl = ctx.scenario.workloads[g.workload]
  if (!wl) throw new Error(`нет нагрузки #${g.workload}`)

  let next = g.phaseIdx + 1
  if (next >= wl.phases.length) {
    if (g.repeatsLeft > 0) {
      g.repeatsLeft--
      next = 0
      // Функция вернулась: всё, что она держала на стеке, исчезает вместе с кадром.
      g.stackUsed = 0
      g.depth = 0
    } else {
      g.state = 'done'
      g.stackUsed = 0
      detachStack(ctx, g)
      return false
    }
  }
  g.phaseIdx = next
  const ph = wl.phases[next]!
  if (ph.kind === 'cpu') g.phaseLeft = rng.duration(ph.ticks)
  else if (ph.kind === 'alloc') g.allocLeft = ph.count ?? 1
  else g.depth = 0
  void w
  return true
}

function phaseAccount(ctx: Ctx): void {
  const w = ctx.world
  w.stats.peakPages = Math.max(w.stats.peakPages, usedPages(w))
  if (w.finished) return

  const alive = w.gs.filter((g) => g.state !== 'done')
  if (alive.length === 0 && w.spawnPtr >= ctx.plan.length) {
    w.finished = true
    w.finishReason = 'all-done'
  }
}
