import type {
  Goroutine,
  Mcache,
  Mcentral,
  MemConfig,
  MemEvent,
  MemEventType,
  MemScenario,
  MemWorld,
  Page,
  Span,
} from './types.ts'
import { DEFAULT_MEM_CONFIG, PAGE, SIZE_CLASSES } from './types.ts'

/** Одна запланированная горутина: кто и когда рождается. */
export interface SpawnPlanItem {
  tick: number
  workload: number
}

export interface Ctx {
  world: MemWorld
  events: MemEvent[]
  plan: SpawnPlanItem[]
  scenario: MemScenario
}

export function resolveConfig(partial: Partial<MemConfig> | undefined): MemConfig {
  return { ...DEFAULT_MEM_CONFIG, ...(partial ?? {}) }
}

export function buildSpawnPlan(scenario: MemScenario): SpawnPlanItem[] {
  const plan: SpawnPlanItem[] = []
  scenario.workloads.forEach((w, wi) => {
    for (let i = 0; i < w.count; i++) {
      plan.push({ tick: w.spawnAt === 'staggered' ? i : w.spawnAt, workload: wi })
    }
  })
  plan.sort((a, b) => a.tick - b.tick || a.workload - b.workload)
  return plan
}

export function createWorld(scenario: MemScenario): MemWorld {
  const config = resolveConfig(scenario.config)
  const pages: Page[] = []
  for (let i = 0; i < config.heapPages; i++) {
    // В начале вся арена — это адреса, за которыми ещё нет настоящей памяти.
    pages.push({ id: i, kind: 'returned', owner: null, freeSince: null })
  }
  const mcaches: Mcache[] = []
  for (let p = 0; p < config.gomaxprocs; p++) {
    mcaches.push({ p, spans: {}, tinyOffset: 0, tinyObjects: 0, tinySpanObj: null })
  }
  const mcentrals: Mcentral[] = SIZE_CLASSES.map((_, i) => ({
    sizeClass: i,
    partial: [],
    full: [],
    lockedAt: -1,
    lockedBy: null,
  }))

  return {
    tick: 0,
    config,
    pages,
    spans: [],
    stackChunks: [],
    objects: [],
    gs: [],
    mcaches,
    mcentrals,
    slots: [],
    stats: {
      onStack: 0,
      tiny: 0,
      tinyBlocks: 0,
      fast: 0,
      refills: 0,
      large: 0,
      askedBytes: 0,
      slotBytes: 0,
      roundAsked: 0,
      roundSlot: 0,
      contended: 0,
      contendedTicks: 0,
      mapped: 0,
      returned: 0,
      peakPages: 0,
      stackGrows: 0,
      stackCopyBytes: 0,
      stackShrinks: 0,
      spansCreated: 0,
      spansFreed: 0,
    },
    nextGid: 1,
    nextSpanId: 1,
    nextChunkId: 1,
    nextObjId: 1,
    spawnPtr: 0,
    finished: false,
  }
}

export function emit(
  ctx: Ctx,
  type: MemEventType,
  actors: MemEvent['actors'] = {},
  payload: Record<string, unknown> = {},
): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export function getG(w: MemWorld, id: number): Goroutine {
  const g = w.gs[id - 1]
  if (!g) throw new Error(`нет горутины G${id}`)
  return g
}

export function getSpan(w: MemWorld, id: number): Span {
  const s = w.spans[id - 1]
  if (!s) throw new Error(`нет спана #${id}`)
  return s
}

/** Сколько страниц сейчас занято под что угодно. */
export function usedPages(w: MemWorld): number {
  return w.pages.filter((p) => p.kind !== 'free' && p.kind !== 'returned').length
}

/** Сколько памяти рантайм держит у операционной системы: занятое плюс свободное, но не возвращённое. */
export function mappedPages(w: MemWorld): number {
  return w.pages.filter((p) => p.kind !== 'returned').length
}

/**
 * Найти `n` подряд идущих страниц.
 *
 * Куча раздаётся не байтами, а страницами, и большому объекту нужен непрерывный
 * кусок: свободных страниц может быть вдоволь, но если они разбросаны, взять их
 * нельзя. Это внешняя фрагментация.
 */
function findRun(w: MemWorld, n: number): number[] | null {
  let run: number[] = []
  for (const p of w.pages) {
    if (p.kind === 'free' || p.kind === 'returned') {
      run.push(p.id)
      if (run.length === n) return run
    } else {
      run = []
    }
  }
  return null
}

export interface PageRequest {
  n: number
  kind: 'span' | 'large' | 'stack'
  owner: number
  /** Для события: кто просил память. */
  g?: number
}

/**
 * Выдать страницы под спан, большой объект или стек.
 *
 * Если в найденном куске есть страницы, которые рантайм уже вернул операционной
 * системе, за них придётся снова сходить в ядро — это событие os.map.
 */
export function allocPages(ctx: Ctx, req: PageRequest): number[] | null {
  const w = ctx.world
  const run = findRun(w, req.n)
  if (!run) {
    const free = w.pages.filter((p) => p.kind === 'free' || p.kind === 'returned').length
    if (free >= req.n) {
      emit(ctx, 'frag.external', { g: req.g ? [req.g] : [] }, {
        need: req.n,
        free,
        reason: 'нет подряд идущих страниц',
      })
    }
    emit(ctx, 'oom', { g: req.g ? [req.g] : [] }, { need: req.n, free, heapPages: w.config.heapPages })
    w.finished = true
    w.finishReason = 'oom'
    return null
  }

  const remapped = run.filter((id) => w.pages[id]!.kind === 'returned')
  if (remapped.length > 0) {
    // Свободные страницы были, но не подряд — пришлось занимать новые.
    // Это внешняя фрагментация: память есть, а взять её нельзя.
    const scattered = w.pages.filter((p) => p.kind === 'free' && !run.includes(p.id)).length
    if (scattered >= req.n) {
      emit(ctx, 'frag.external', { g: req.g ? [req.g] : [], pages: run }, {
        need: req.n,
        scattered,
        kind: req.kind,
      })
    }
  }
  for (const id of run) {
    const page = w.pages[id]!
    page.kind = req.kind
    page.owner = req.owner
    page.freeSince = null
  }
  if (remapped.length > 0) {
    w.stats.mapped += remapped.length
    emit(ctx, 'os.map', { pages: remapped, g: req.g ? [req.g] : [] }, {
      pages: remapped.length,
      bytes: remapped.length * PAGE,
      total: mappedPages(w),
      why: req.kind,
    })
  }
  w.stats.peakPages = Math.max(w.stats.peakPages, usedPages(w))
  return run
}

export function releasePages(ctx: Ctx, ids: number[]): void {
  const w = ctx.world
  for (const id of ids) {
    const page = w.pages[id]!
    page.kind = 'free'
    page.owner = null
    page.freeSince = w.tick
  }
}

/* ──────────────────────────── спаны и классы ──────────────────────────── */

/** Новый спан класса `cls`: страницы из кучи, нарезанные на объекты одного размера. */
export function newSpan(ctx: Ctx, cls: number, g: number | undefined): Span | null {
  const w = ctx.world
  const objSize = SIZE_CLASSES[cls]!
  const npages = Math.max(w.config.spanPages, Math.ceil(objSize / PAGE))
  const pages = allocPages(ctx, { n: npages, kind: 'span', owner: w.nextSpanId, g })
  if (!pages) return null

  const span: Span = {
    id: w.nextSpanId++,
    sizeClass: cls,
    objSize,
    pages,
    slots: Math.floor((npages * PAGE) / objSize),
    used: 0,
    owner: { kind: 'mcentral' },
    createdTick: w.tick,
  }
  w.spans.push(span)
  w.stats.spansCreated++
  emit(ctx, 'heap.grow', { span: [span.id], pages, g: g ? [g] : [] }, {
    sizeClass: cls,
    objSize,
    pages: npages,
    slots: span.slots,
    heapPages: usedPages(w),
  })
  return span
}

/** Спан опустел — страницы возвращаются в кучу и могут достаться другому классу. */
export function freeSpan(ctx: Ctx, span: Span): void {
  const w = ctx.world
  span.owner = { kind: 'free' }
  releasePages(ctx, span.pages)
  w.stats.spansFreed++
  emit(ctx, 'span.free', { span: [span.id], pages: span.pages }, {
    sizeClass: span.sizeClass,
    objSize: span.objSize,
    pages: span.pages.length,
    livedTicks: w.tick - span.createdTick,
  })
}

export function centralOf(w: MemWorld, cls: number): Mcentral {
  const c = w.mcentrals[cls]
  if (!c) throw new Error(`нет mcentral для класса ${cls}`)
  return c
}

export function cacheOf(w: MemWorld, p: number): Mcache {
  const c = w.mcaches[p]
  if (!c) throw new Error(`нет mcache для P${p}`)
  return c
}
