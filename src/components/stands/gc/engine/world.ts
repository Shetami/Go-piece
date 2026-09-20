import type {
  Cell,
  GcConfig,
  GcEvent,
  GcEventType,
  GcScenario,
  GcWorld,
  Mutator,
} from './types.ts'
import { DEFAULT_GC_CONFIG } from './types.ts'
import type { Rng } from './rng.ts'

/** Одна запланированная горутина: кто и когда рождается. */
export interface SpawnPlanItem {
  tick: number
  workload: number
}

export interface Ctx {
  world: GcWorld
  events: GcEvent[]
  plan: SpawnPlanItem[]
  scenario: GcScenario
}

/** Сколько указателей держит на стеке горутина, если нагрузка не сказала иного. */
const DEFAULT_STACK_SLOTS = 3
/** Сколько исходящих указателей у блока. */
export const SLOTS_PER_CELL = 2
/** Предел глобальных корней: дальше retain затирает самый старый. */
const MAX_GLOBALS = 48
/** Потолок отношения помощи: иначе у самой цели оно уходит в бесконечность. */
const MAX_ASSIST_RATIO = 24

export function resolveConfig(partial: Partial<GcConfig> | undefined): GcConfig {
  return { ...DEFAULT_GC_CONFIG, ...(partial ?? {}) }
}

export function buildSpawnPlan(scenario: GcScenario): SpawnPlanItem[] {
  const plan: SpawnPlanItem[] = []
  scenario.workloads.forEach((w, wi) => {
    for (let i = 0; i < w.count; i++) {
      plan.push({ tick: w.spawnAt === 'staggered' ? i : w.spawnAt, workload: wi })
    }
  })
  plan.sort((a, b) => a.tick - b.tick || a.workload - b.workload)
  return plan
}

export function createWorld(scenario: GcScenario): GcWorld {
  const config = resolveConfig(scenario.config)
  const cells: Cell[] = []
  for (let i = 0; i < config.heapCapacity; i++) {
    cells.push({
      id: i,
      used: false,
      color: 'white',
      slots: new Array<number | null>(SLOTS_PER_CELL).fill(null),
      owner: 0,
      bornTick: 0,
      bornCycle: 0,
      bornBlack: false,
      lost: false,
    })
  }
  return {
    tick: 0,
    config,
    phase: 'off',
    stwLeft: 0,
    cycle: 0,
    cycleFreed: 0,
    cycleAssistTicks: 0,
    cycleGcTicks: 0,
    cycleStart: 0,
    cells,
    muts: [],
    globals: [],
    greyq: [],
    sweepq: [],
    heapMarked: 0,
    scanWorkTarget: 0,
    scanWorkDone: 0,
    goal: config.initialGoal,
    trigger: Math.max(1, Math.round(config.initialGoal * 0.7)),
    assistRatio: 0,
    allocRate: 0,
    barrierOn: false,
    slots: [],
    stats: {
      cycles: 0,
      stwTicks: 0,
      maxPause: 0,
      gcSlotTicks: 0,
      mutatorSlotTicks: 0,
      assistTicks: 0,
      allocated: 0,
      freed: 0,
      lost: 0,
      peakHeap: 0,
      maxOvershoot: 0,
    },
    nextMutId: 1,
    spawnPtr: 0,
    finished: false,
  }
}

export function emit(
  ctx: Ctx,
  type: GcEventType,
  actors: GcEvent['actors'] = {},
  payload: Record<string, unknown> = {},
): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export function getCell(w: GcWorld, id: number): Cell {
  const c = w.cells[id]
  if (!c) throw new Error(`нет блока #${id}`)
  return c
}

export function getMut(w: GcWorld, id: number): Mutator {
  const m = w.muts[id - 1]
  if (!m) throw new Error(`нет горутины G${id}`)
  return m
}

export function heapUsed(w: GcWorld): number {
  let n = 0
  for (const c of w.cells) if (c.used) n++
  return n
}

/** Сколько процессоров рантайм отдаёт фоновой разметке: дробь превращается в дробного маркера. */
export function markWorkerBudget(c: GcConfig): number {
  return c.gomaxprocs * c.gcCpuShare
}

/**
 * Закрасить объект серым.
 *
 * Это единственный способ попасть во фронт разметки: и корни, и барьер записи,
 * и просмотр чужих полей делают именно это. Возвращает true, если цвет изменился.
 */
export function shade(w: GcWorld, id: number | null): boolean {
  if (id === null) return false
  const c = w.cells[id]
  if (!c || !c.used || c.color !== 'white') return false
  c.color = 'grey'
  w.greyq.push(id)
  return true
}

/**
 * Запись указателя в кучу или в глобальную переменную — через барьер.
 *
 * Гибридный барьер Go 1.8:
 *
 *   shade(*slot)                      // удаляемый объект: барьер Юасы
 *   if стек пишущей горутины серый:   // барьер Дейкстры
 *       shade(ptr)
 *   *slot = ptr
 *
 * Первая строка — главная: она ловит именно тот случай, ради которого барьер
 * существует. Записи в стек барьером НЕ закрываются, и это не дыра: спрятать
 * объект на стеке можно, только прочитав указатель откуда-то из кучи, а когда
 * оттуда его уберут, сработает первая строка.
 */
export function writeBarrier(ctx: Ctx, mut: Mutator | null, oldVal: number | null, ptr: number | null, where: string): void {
  const w = ctx.world
  if (w.phase !== 'mark') return

  if (!w.config.writeBarrier) {
    // Барьера нет. Опасна не любая запись, а запись белого указателя в уже
    // просмотренный объект: к нему разметка больше не вернётся.
    if (ptr !== null && w.cells[ptr]?.color === 'white') {
      emit(ctx, 'barrier.missed', { cells: [ptr], mut: mut ? [mut.id] : [] }, { where, hidden: ptr })
    }
    return
  }

  if (shade(w, oldVal)) {
    emit(ctx, 'barrier.shade', { cells: [oldVal!], mut: mut ? [mut.id] : [] }, { rule: 'deletion', where })
  }
  if (ptr !== null && mut !== null && !mut.stackScanned && shade(w, ptr)) {
    emit(ctx, 'barrier.shade', { cells: [ptr], mut: [mut.id] }, { rule: 'dijkstra', where })
  }
}

/** Запись в поле объекта. */
export function writeHeapSlot(ctx: Ctx, mut: Mutator | null, cellId: number, slot: number, ptr: number | null): void {
  const c = getCell(ctx.world, cellId)
  const old = c.slots[slot] ?? null
  if (old === ptr) return
  writeBarrier(ctx, mut, old, ptr, `объект #${cellId}`)
  c.slots[slot] = ptr
}

/** Запись в глобальную переменную. Глобальные корни просматриваются один раз, поэтому барьер обязателен. */
export function writeGlobal(ctx: Ctx, mut: Mutator | null, slot: number, ptr: number | null): void {
  const w = ctx.world
  const old = w.globals[slot] ?? null
  if (old === ptr) return
  writeBarrier(ctx, mut, old, ptr, 'глобальная переменная')
  w.globals[slot] = ptr
}

/** Запись в слот стека — без барьера, как в настоящем рантайме. */
export function writeStack(mut: Mutator, slot: number, ptr: number | null): void {
  mut.stack[slot] = ptr
}

function takeFreeCell(w: GcWorld): Cell | undefined {
  for (const c of w.cells) if (!c.used) return c
  return undefined
}

/**
 * Выделить один блок.
 *
 * Во время разметки блок рождается чёрным: разметка уже прошла корни, и доказать,
 * что новорождённый объект мусор, всё равно нельзя — он доживёт до следующего цикла.
 */
export function allocCell(ctx: Ctx, mut: Mutator): Cell | null {
  const w = ctx.world
  const cell = takeFreeCell(w)
  if (!cell) {
    emit(ctx, 'oom', { mut: [mut.id] }, { capacity: w.config.heapCapacity, goal: w.goal })
    w.finished = true
    w.finishReason = 'oom'
    return null
  }
  const black = w.phase === 'mark' && w.config.allocBlack
  cell.used = true
  cell.color = black ? 'black' : 'white'
  cell.slots = new Array<number | null>(SLOTS_PER_CELL).fill(null)
  cell.owner = mut.workload
  cell.bornTick = w.tick
  cell.bornCycle = w.cycle
  cell.bornBlack = black
  cell.lost = false

  mut.allocated++
  w.stats.allocated++
  if (black) emit(ctx, 'alloc.black', { cells: [cell.id], mut: [mut.id] }, { cycle: w.cycle })
  else emit(ctx, 'alloc', { cells: [cell.id], mut: [mut.id] }, { phase: w.phase })
  return cell
}

/**
 * Освободить блок: память возвращается в общий котёл и тут же может достаться
 * новому объекту. Пометку `lost` блок получает раньше, на разборе итогов
 * разметки, — здесь она только попадает в статистику.
 */
export function freeCell(ctx: Ctx, id: number): void {
  const w = ctx.world
  const c = getCell(w, id)
  if (!c.used) return
  c.used = false
  c.color = 'white'
  c.slots = new Array<number | null>(SLOTS_PER_CELL).fill(null)
  // Пометку `lost` при освобождении НЕ снимаем: пусть на схеме останется шрам —
  // здесь лежал живой объект, а память отдали. Он исчезнет, когда блок займут снова.
  w.stats.freed++
  w.cycleFreed++
}

/**
 * Что на самом деле достижимо из корней.
 *
 * Разметка этого не знает — она строит свой ответ по цветам. Модель считает
 * настоящую достижимость отдельно, чтобы поймать сборщика за руку, если он
 * соберётся освободить живой объект: при исправном барьере это невозможно.
 */
export function reachableSet(w: GcWorld): Set<number> {
  const seen = new Set<number>()
  const stack: number[] = []
  const push = (id: number | null | undefined) => {
    if (id === null || id === undefined) return
    const c = w.cells[id]
    if (!c || !c.used || seen.has(id)) return
    seen.add(id)
    stack.push(id)
  }
  for (const g of w.globals) push(g)
  for (const m of w.muts) {
    if (m.state === 'done') continue
    for (const s of m.stack) push(s)
  }
  while (stack.length > 0) {
    const id = stack.pop()!
    for (const s of getCell(w, id).slots) push(s)
  }
  return seen
}

/** Осталась ли разметке работа прямо сейчас: серые объекты плюс непросмотренные стеки. */
export function markWorkPending(w: GcWorld): number {
  const stacks = w.muts.filter((m) => m.state !== 'done' && !m.stackScanned).length
  return w.greyq.length + stacks
}

/**
 * Пересчитать цель и порог запуска.
 *
 * Цель задаёт GOGC: живое × (1 + GOGC/100). Порог — раньше цели ровно настолько,
 * сколько программа успеет выделить, пока идёт разметка. Пейсер именно это и
 * пытается угадать: стартовать так, чтобы разметка закончилась у самой цели.
 */
export function recomputePacer(ctx: Ctx): void {
  const w = ctx.world
  const c = w.config

  let goal = c.gogc === false ? Number.POSITIVE_INFINITY : Math.round(w.heapMarked * (1 + c.gogc / 100))
  goal = Math.max(goal, w.heapMarked + 2, c.initialGoal)

  let limited = false
  if (c.memLimit !== false && goal > c.memLimit) {
    goal = c.memLimit
    limited = true
  }
  if (!Number.isFinite(goal)) goal = c.heapCapacity

  const scanWork = Math.max(1, w.heapMarked)
  const markRate = Math.max(0.5, markWorkerBudget(c) * c.scanRate)
  const markTicks = scanWork / markRate
  const runway = Math.max(1, Math.round(w.allocRate * markTicks))
  const trigger = Math.min(Math.max(goal - runway, w.heapMarked + 1), Math.max(1, goal - 1))

  w.goal = goal
  w.trigger = trigger

  if (limited) {
    emit(ctx, 'limit.hit', {}, { goal, memLimit: c.memLimit, live: w.heapMarked, gogc: c.gogc })
  }
  emit(ctx, 'pacer.set', {}, {
    live: w.heapMarked,
    goal,
    trigger,
    runway,
    allocRate: Math.round(w.allocRate * 100) / 100,
    markTicks: Math.round(markTicks * 10) / 10,
    limited,
  })
}

/**
 * Отношение помощи: сколько блоков обязан просмотреть тот, кто выделил один блок.
 *
 *   осталось работы разметки / осталось места до цели
 *
 * Считается от ОЦЕНКИ всей работы цикла, а не от длины очереди серых: очередь
 * то пустеет, то наполняется, и по ней помощь включалась бы слишком поздно.
 * Чем ближе куча к цели, тем дороже каждый следующий блок — у самой цели
 * аллокация фактически останавливается, пока разметка не закончится.
 */
export function updateAssistRatio(w: GcWorld): void {
  const left = Math.max(0, w.scanWorkTarget - w.scanWorkDone)
  const room = Math.max(1, w.goal - heapUsed(w))
  w.assistRatio = Math.min(MAX_ASSIST_RATIO, left / room)
}

export function createMutator(ctx: Ctx, workload: number, rng: Rng): Mutator {
  const w = ctx.world
  const wl = ctx.scenario.workloads[workload]
  if (!wl) throw new Error(`нет нагрузки #${workload}`)
  void rng
  const mut: Mutator = {
    id: w.nextMutId++,
    name: wl.name,
    workload,
    stack: new Array<number | null>(wl.stackSlots ?? DEFAULT_STACK_SLOTS).fill(null),
    // Родившаяся во время разметки горутина стартует с пустым стеком: сканировать нечего.
    stackScanned: w.phase === 'mark',
    state: 'runnable',
    phaseIdx: -1,
    phaseLeft: 0,
    repeatsLeft: wl.repeat === 'forever' ? Number.POSITIVE_INFINITY : (wl.repeat ?? 1) - 1,
    assistDebt: 0,
    assistedThisCycle: false,
    createdTick: w.tick,
    cpuTicks: 0,
    assistTicks: 0,
    stwTicks: 0,
    waitTicks: 0,
    allocated: 0,
  }
  w.muts.push(mut)
  return mut
}

export { MAX_GLOBALS, MAX_ASSIST_RATIO, DEFAULT_STACK_SLOTS }
