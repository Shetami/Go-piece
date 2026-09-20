import type { Cell, MutPhase, Mutator, SlotKind, SlotView } from './types.ts'
import type { Ctx } from './world.ts'
import {
  MAX_GLOBALS,
  allocCell,
  createMutator,
  emit,
  freeCell,
  getCell,
  heapUsed,
  markWorkerBudget,
  reachableSet,
  recomputePacer,
  markWorkPending,
  shade,
  updateAssistRatio,
  writeGlobal,
  writeHeapSlot,
  writeStack,
} from './world.ts'
import type { Rng } from './rng.ts'

/**
 * Один шаг всей системы.
 *
 * Порядок важен. Сначала двигается фаза цикла — по тому, что успело случиться
 * в прошлом тике, — потом пейсер решает, не пора ли начинать новый цикл, и
 * только после этого процессоры делятся между сборщиком и программой. Так
 * снимок остаётся честным: если в нём написано «пауза», значит в этот тик
 * действительно никто не работал.
 */
export function tick(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  if (w.finished) return
  w.tick++
  w.slots = []
  const allocBefore = w.stats.allocated

  phaseSpawn(ctx, rng)
  phaseAdvance(ctx)
  phaseTrigger(ctx)
  phaseSlots(ctx, rng)
  phaseAccount(ctx, w.stats.allocated - allocBefore)
}

function phaseSpawn(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  while (w.spawnPtr < ctx.plan.length) {
    const item = ctx.plan[w.spawnPtr]
    if (!item || item.tick >= w.tick) break
    w.spawnPtr++
    createMutator(ctx, item.workload, rng)
  }
}

/* ────────────────────────────── пейсер ────────────────────────────── */

function phaseTrigger(ctx: Ctx): void {
  const w = ctx.world
  if (w.phase !== 'off') return
  if (w.config.gogc === false && w.config.memLimit === false) return
  const used = heapUsed(w)
  if (used < w.trigger) return
  startCycle(ctx, used)
}

/**
 * Начало цикла: sweep termination.
 *
 * Мир останавливается, включается барьер записи, и под паузой закрашиваются
 * глобальные корни. Стеки горутин под паузу не попадают — их просмотрят уже
 * во время разметки, по одному.
 */
function startCycle(ctx: Ctx, used: number): void {
  const w = ctx.world
  w.cycle++
  w.cycleStart = w.tick
  w.cycleFreed = 0
  w.cycleAssistTicks = 0
  w.cycleGcTicks = 0
  w.phase = 'stw-start'
  w.stwLeft = w.config.stwTicks
  w.barrierOn = true
  // Оценка работы цикла: всё живое прошлого цикла плюс по стеку на горутину.
  // От неё пейсер считает, сколько помощи брать с каждого выделенного блока.
  w.scanWorkTarget = Math.max(1, w.heapMarked + w.muts.filter((m) => m.state !== 'done').length)
  w.scanWorkDone = 0

  emit(ctx, 'gc.trigger', {}, {
    heap: used,
    trigger: w.trigger,
    goal: w.goal,
    cycle: w.cycle,
    live: w.heapMarked,
  })
  emit(ctx, 'stw.enter', {}, { kind: 'sweep-termination', ticks: w.config.stwTicks, cycle: w.cycle })

  for (const m of w.muts) {
    m.stackScanned = m.state === 'done'
    m.assistedThisCycle = false
    m.assistDebt = 0
  }

  const shaded: number[] = []
  for (const g of w.globals) {
    if (shade(w, g) && g !== null) shaded.push(g)
  }
  emit(ctx, 'mark.roots', { cells: shaded.slice(0, 12) }, {
    globals: w.globals.filter((g) => g !== null).length,
    shaded: shaded.length,
    stacks: w.muts.filter((m) => m.state !== 'done').length,
  })
}

/* ──────────────────────────── процессоры ──────────────────────────── */

function rotate<T>(items: T[], by: number): T[] {
  if (items.length === 0) return items
  const k = by % items.length
  return [...items.slice(k), ...items.slice(0, k)]
}

/**
 * Раздача процессоров на этот тик.
 *
 * Порядок раздачи и есть «25% CPU сборщику»: выделенные маркеры забирают свои
 * слоты первыми, программа получает то, что осталось, а совсем пустые слоты
 * дорабатывают маркерами вхолостую.
 */
function phaseSlots(ctx: Ctx, rng: Rng): void {
  const w = ctx.world
  const c = w.config
  const slots: SlotView[] = []

  if (w.phase === 'stw-start' || w.phase === 'stw-end') {
    for (let p = 0; p < c.gomaxprocs; p++) slots.push({ p, kind: 'stw' })
    w.stats.gcSlotTicks += c.gomaxprocs
    w.cycleGcTicks += c.gomaxprocs
    w.stats.stwTicks++
    for (const m of w.muts) {
      if (m.state === 'done') continue
      m.state = 'stopped'
      m.stwTicks++
    }
    w.slots = slots
    return
  }

  let p = 0

  if (w.phase === 'mark') {
    const budget = markWorkerBudget(c)
    const dedicated = Math.min(c.gomaxprocs, Math.floor(budget))
    const frac = budget - dedicated
    // Дробный маркер работает не постоянно, а долю тиков: ровно ту, что осталась от доли CPU.
    const fractionalNow = frac > 0.0001 && Math.floor(w.tick * frac) > Math.floor((w.tick - 1) * frac)
    for (let i = 0; i < dedicated && p < c.gomaxprocs; i++, p++) {
      slots.push({ p, kind: 'dedicated', work: markWork(ctx, c.scanRate, 'dedicated', null) })
    }
    if (fractionalNow && p < c.gomaxprocs) {
      slots.push({ p, kind: 'fractional', work: markWork(ctx, c.scanRate, 'fractional', null) })
      p++
    }
  }

  if (w.phase === 'sweep' && w.sweepq.length > 0 && p < c.gomaxprocs) {
    slots.push({ p, kind: 'sweeper', work: sweepWork(ctx, c.sweepRate) })
    p++
  }

  w.stats.gcSlotTicks += p
  w.cycleGcTicks += p

  const ready = w.muts.filter((m) => m.state !== 'done')
  for (const m of ready) m.state = 'runnable'

  const order = rotate(ready, w.tick)
  let i = 0
  while (p < c.gomaxprocs && i < order.length) {
    const m = order[i++]!
    const kind = runMutator(ctx, m, rng)
    slots.push({ p, kind, mut: m.id })
    if (kind === 'assist') {
      w.stats.gcSlotTicks++
      w.cycleGcTicks++
    } else {
      w.stats.mutatorSlotTicks++
    }
    p++
    if (w.finished) break
  }

  while (p < c.gomaxprocs) {
    if (w.phase === 'mark' && c.idleWorkers) {
      const work = markWork(ctx, c.scanRate, 'idle', null)
      if (work > 0) {
        slots.push({ p, kind: 'idle-worker', work })
        w.stats.gcSlotTicks++
        w.cycleGcTicks++
        p++
        continue
      }
    }
    slots.push({ p, kind: 'free' })
    p++
  }

  for (const m of ready) if (m.state === 'runnable') m.waitTicks++
  w.slots = slots
}

/* ───────────────────────────── разметка ───────────────────────────── */

/**
 * Единица работы разметки: просмотреть серый объект и закрасить его детей.
 *
 * Стеки горутин идут первыми — пока стек не просмотрен, он серый, и горутине
 * приходится закрашивать всё, что она куда-либо записывает.
 */
function markWork(ctx: Ctx, budget: number, by: string, mut: Mutator | null): number {
  const w = ctx.world
  if (w.phase !== 'mark') return 0

  let done = 0
  const scanned: number[] = []
  let greyed = 0

  while (done < budget) {
    const pending = w.muts.find((m) => m.state !== 'done' && !m.stackScanned)
    if (pending) {
      let shadedHere = 0
      for (const s of pending.stack) if (shade(w, s)) shadedHere++
      pending.stackScanned = true
      greyed += shadedHere
      done++
      w.scanWorkDone++
      emit(ctx, 'mark.stack', { mut: [pending.id] }, {
        by,
        slots: pending.stack.filter((s) => s !== null).length,
        shaded: shadedHere,
      })
      continue
    }

    const id = w.greyq.shift()
    if (id === undefined) break
    const cell = getCell(w, id)
    if (!cell.used) continue
    for (const s of cell.slots) if (shade(w, s)) greyed++
    cell.color = 'black'
    scanned.push(id)
    done++
    w.scanWorkDone++
  }

  if (scanned.length > 0) {
    emit(ctx, 'mark.scan', { cells: scanned, mut: mut ? [mut.id] : [] }, {
      by,
      count: scanned.length,
      greyed,
      greyLeft: w.greyq.length,
    })
  }
  return done
}

function sweepWork(ctx: Ctx, budget: number): number {
  const w = ctx.world
  const freed: number[] = []
  while (freed.length < budget) {
    const id = w.sweepq.shift()
    if (id === undefined) break
    freeCell(ctx, id)
    freed.push(id)
  }
  if (freed.length > 0) {
    emit(ctx, 'sweep.freed', { cells: freed.slice(0, 12) }, { count: freed.length, left: w.sweepq.length })
  }
  return freed.length
}

/* ───────────────────────────── мутаторы ───────────────────────────── */

function runMutator(ctx: Ctx, mut: Mutator, rng: Rng): SlotKind {
  const w = ctx.world

  if (w.phase === 'mark' && w.config.markAssist && mut.assistDebt >= 1) {
    const done = markWork(ctx, w.config.scanRate, 'assist', mut)
    // Работы не нашлось — долг списывается: помогать больше нечему.
    mut.assistDebt = done === 0 ? 0 : Math.max(0, mut.assistDebt - done)
    mut.assistTicks++
    mut.state = 'assist'
    w.stats.assistTicks++
    w.cycleAssistTicks++
    if (!mut.assistedThisCycle) {
      mut.assistedThisCycle = true
      emit(ctx, 'assist.begin', { mut: [mut.id] }, {
        debt: Math.round(mut.assistDebt * 10) / 10,
        ratio: Math.round(w.assistRatio * 100) / 100,
        heap: heapUsed(w),
        goal: w.goal,
      })
    } else {
      emit(ctx, 'assist.work', { mut: [mut.id] }, { done, debtLeft: Math.round(mut.assistDebt * 10) / 10 })
    }
    return 'assist'
  }

  mut.state = 'running'
  mut.cpuTicks++
  stepPhase(ctx, mut, rng)
  return 'mutator'
}

function currentPhase(ctx: Ctx, mut: Mutator): MutPhase | undefined {
  const wl = ctx.scenario.workloads[mut.workload]
  return wl?.phases[mut.phaseIdx]
}

function stepPhase(ctx: Ctx, mut: Mutator, rng: Rng): void {
  if (mut.phaseIdx === -1 && !advance(ctx, mut, rng)) return
  const ph = currentPhase(ctx, mut)
  if (!ph) return

  switch (ph.kind) {
    case 'cpu': {
      mut.phaseLeft--
      if (mut.phaseLeft <= 0) advance(ctx, mut, rng)
      return
    }
    case 'alloc': {
      doAlloc(ctx, mut, ph, rng)
      advance(ctx, mut, rng)
      return
    }
    case 'drop': {
      doDrop(ctx, mut, rng)
      advance(ctx, mut, rng)
      return
    }
    case 'move': {
      doMove(ctx, mut)
      advance(ctx, mut, rng)
      return
    }
  }
}

/** Перейти к следующей фазе. false — горутина закончилась. */
function advance(ctx: Ctx, mut: Mutator, rng: Rng): boolean {
  const wl = ctx.scenario.workloads[mut.workload]
  if (!wl) throw new Error(`нет нагрузки #${mut.workload}`)

  let next = mut.phaseIdx + 1
  if (next >= wl.phases.length) {
    if (mut.repeatsLeft > 0) {
      mut.repeatsLeft--
      next = 0
    } else {
      mut.state = 'done'
      // Горутина ушла — её стек вместе с корнями исчез, всё, что держалось только
      // на нём, станет мусором в ближайшем цикле.
      mut.stack.fill(null)
      return false
    }
  }
  mut.phaseIdx = next
  const ph = wl.phases[next]
  if (ph?.kind === 'cpu') mut.phaseLeft = rng.duration(ph.ticks)
  else mut.phaseLeft = 1
  return true
}

/**
 * Выделить структуру: голова ложится в слот стека, остальные блоки — цепочкой за ней.
 * Прошлый житель слота теряет последнюю ссылку и становится мусором.
 */
function doAlloc(ctx: Ctx, mut: Mutator, ph: Extract<MutPhase, { kind: 'alloc' }>, rng: Rng): void {
  const w = ctx.world

  // Аллокация подметает за собой: в рантайме это proportional sweeping.
  if (w.sweepq.length > 0) sweepWork(ctx, 1)

  let head: Cell | null = null
  let prev: Cell | null = null
  for (let i = 0; i < ph.blocks; i++) {
    const cell = allocCell(ctx, mut)
    if (!cell) return
    if (prev) writeHeapSlot(ctx, mut, prev.id, 0, cell.id)
    else head = cell
    prev = cell
  }
  if (!head) return

  writeStack(mut, rng.int(0, mut.stack.length - 1), head.id)

  if (ph.retain) {
    const slot = w.globals.length < MAX_GLOBALS ? w.globals.length : rng.int(0, MAX_GLOBALS - 1)
    if (slot >= w.globals.length) w.globals.push(null)
    writeGlobal(ctx, mut, slot, head.id)
  }

  if (w.phase === 'mark' && w.config.markAssist) {
    mut.assistDebt += ph.blocks * w.assistRatio
  }
}

function doDrop(ctx: Ctx, mut: Mutator, rng: Rng): void {
  void ctx
  const filled = mut.stack.map((s, i) => (s === null ? -1 : i)).filter((i) => i >= 0)
  if (filled.length === 0) return
  writeStack(mut, filled[rng.int(0, filled.length - 1)]!, null)
}

/**
 * Спрятать объект за уже просмотренным.
 *
 * Это та самая опасная последовательность, ради которой существует барьер
 * записи: белый объект переезжает под чёрный, а единственная другая ссылка на
 * него исчезает. С барьером он станет серым и выживет, без барьера — останется
 * белым, и подметальщик освободит живую память.
 */
function doMove(ctx: Ctx, mut: Mutator): void {
  const w = ctx.world

  const dst = w.cells.find((c) => c.used && c.color === 'black' && c.slots.some((s) => s === null))
  if (!dst) return

  let donor: { cell: Cell; slot: number; victim: number } | undefined
  for (const c of w.cells) {
    if (!c.used || c.color === 'black' || c.id === dst.id) continue
    for (let i = 0; i < c.slots.length; i++) {
      const v = c.slots[i]
      if (v === null || v === undefined || v === dst.id) continue
      const target = w.cells[v]
      if (target?.used && target.color === 'white') {
        donor = { cell: c, slot: i, victim: v }
        break
      }
    }
    if (donor) break
  }
  if (!donor) return

  const free = dst.slots.indexOf(null)
  writeHeapSlot(ctx, mut, dst.id, free, donor.victim)
  writeHeapSlot(ctx, mut, donor.cell.id, donor.slot, null)
}

/* ───────────────────────────── фазы цикла ─────────────────────────── */

function phaseAdvance(ctx: Ctx): void {
  const w = ctx.world

  switch (w.phase) {
    case 'stw-start': {
      w.stwLeft--
      if (w.stwLeft > 0) return
      w.phase = 'mark'
      w.stats.maxPause = Math.max(w.stats.maxPause, w.config.stwTicks)
      updateAssistRatio(w)
      emit(ctx, 'stw.exit', {}, { kind: 'sweep-termination', ticks: w.config.stwTicks, cycle: w.cycle })
      return
    }
    case 'mark': {
      updateAssistRatio(w)
      if (markWorkPending(w) > 0) return
      emit(ctx, 'mark.drained', {}, {
        cycle: w.cycle,
        ticks: w.tick - w.cycleStart,
        black: w.cells.filter((c) => c.used && c.color !== 'white').length,
      })
      w.phase = 'stw-end'
      w.stwLeft = w.config.stwTicks
      emit(ctx, 'stw.enter', {}, { kind: 'mark-termination', ticks: w.config.stwTicks, cycle: w.cycle })
      return
    }
    case 'stw-end': {
      w.stwLeft--
      if (w.stwLeft > 0) return
      w.stats.maxPause = Math.max(w.stats.maxPause, w.config.stwTicks)
      emit(ctx, 'stw.exit', {}, { kind: 'mark-termination', ticks: w.config.stwTicks, cycle: w.cycle })
      finishMark(ctx)
      return
    }
    case 'sweep': {
      if (w.sweepq.length > 0) return
      finishCycle(ctx)
      return
    }
    default:
      return
  }
}

/** Разметка закончилась: белое — мусор. Здесь же модель проверяет сборщика на честность. */
function finishMark(ctx: Ctx): void {
  const w = ctx.world
  w.barrierOn = false

  const used = heapUsed(w)
  const white = w.cells.filter((c) => c.used && c.color === 'white')
  const reachable = reachableSet(w)
  const lost = white.filter((c) => reachable.has(c.id))

  for (const c of lost) c.lost = true
  w.stats.lost += lost.length
  if (lost.length > 0) {
    emit(ctx, 'heap.lost', { cells: lost.map((c) => c.id).slice(0, 16) }, {
      count: lost.length,
      writeBarrier: w.config.writeBarrier,
      allocBlack: w.config.allocBlack,
      cycle: w.cycle,
    })
  }

  w.heapMarked = used - white.length
  w.sweepq = white.map((c) => c.id)
  w.phase = 'sweep'
  emit(ctx, 'sweep.begin', {}, {
    garbage: white.length,
    live: w.heapMarked,
    heap: used,
    cycle: w.cycle,
  })
}

function finishCycle(ctx: Ctx): void {
  const w = ctx.world
  // Новый цикл начинается с чистого листа: в рантайме для этого просто меняются
  // местами битовые карты разметки, перекрашивать объекты не нужно.
  for (const c of w.cells) if (c.used) c.color = 'white'

  w.phase = 'off'
  w.stats.cycles++
  const heapAfter = heapUsed(w)

  emit(ctx, 'sweep.done', {}, { freed: w.cycleFreed, heap: heapAfter, cycle: w.cycle })
  emit(ctx, 'cycle.done', {}, {
    cycle: w.cycle,
    ticks: w.tick - w.cycleStart,
    live: w.heapMarked,
    freed: w.cycleFreed,
    heap: heapAfter,
    goal: w.goal,
    stw: w.config.stwTicks * 2,
    assistTicks: w.cycleAssistTicks,
    gcTicks: w.cycleGcTicks,
    cpuShare: Math.round((w.cycleGcTicks / Math.max(1, (w.tick - w.cycleStart) * w.config.gomaxprocs)) * 100),
  })
  recomputePacer(ctx)
}

/* ──────────────────────────── учёт и финиш ────────────────────────── */

function phaseAccount(ctx: Ctx, allocatedThisTick: number): void {
  const w = ctx.world

  // Скользящее среднее скорости аллокации — то, по чему пейсер угадывает разбег.
  w.allocRate = w.allocRate * 0.85 + allocatedThisTick * 0.15

  const used = heapUsed(w)
  w.stats.peakHeap = Math.max(w.stats.peakHeap, used)
  if (Number.isFinite(w.goal) && used > w.goal) {
    w.stats.maxOvershoot = Math.max(w.stats.maxOvershoot, used - w.goal)
  }

  if (w.finished) return

  const alive = w.muts.filter((m) => m.state !== 'done')
  const pending = w.spawnPtr < ctx.plan.length
  if (alive.length === 0 && !pending && w.phase === 'off') {
    w.finished = true
    w.finishReason = 'all-done'
  }
}
