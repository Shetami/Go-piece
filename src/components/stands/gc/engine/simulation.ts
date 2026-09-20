import type { GcEvent, GcScenario, GcWorld, Snapshot } from './types.ts'
import type { Ctx, SpawnPlanItem } from './world.ts'
import { buildSpawnPlan, createWorld, heapUsed } from './world.ts'
import { Rng } from './rng.ts'
import { tick as runTick } from './tick.ts'

/**
 * Симуляция: движок плюс история состояний.
 *
 * История — массив полных снимков, а не дельт. Куча в модели фиксированного
 * размера (блоки переиспользуются, как и настоящая память), поэтому снимок
 * дёшев, а шаг назад и перемотка по таймлайну становятся тривиальными.
 */
export class Simulation {
  readonly scenario: GcScenario
  readonly seed: number
  readonly plan: SpawnPlanItem[]

  private ctx: Ctx
  private rng: Rng
  private snapshots: Snapshot[] = []

  constructor(scenario: GcScenario, seed?: number) {
    this.scenario = scenario
    this.seed = seed ?? scenario.seed ?? 1
    const world = createWorld(scenario)
    this.plan = buildSpawnPlan(scenario)
    this.rng = new Rng(this.seed)
    this.ctx = { world, events: [], plan: this.plan, scenario }
    this.snapshots.push({ world: structuredClone(world), events: [] })
  }

  get world(): GcWorld {
    return this.ctx.world
  }

  get tick(): number {
    return this.ctx.world.tick
  }

  get finished(): boolean {
    return this.ctx.world.finished
  }

  get events(): GcEvent[] {
    return this.snapshots.flatMap((s) => s.events)
  }

  step(): Snapshot | null {
    if (this.ctx.world.finished) return null
    this.ctx.events = []
    runTick(this.ctx, this.rng)
    const snap: Snapshot = { world: structuredClone(this.ctx.world), events: this.ctx.events }
    this.snapshots.push(snap)
    return snap
  }

  runTo(tick: number): this {
    while (this.ctx.world.tick < tick && !this.ctx.world.finished) this.step()
    return this
  }

  runToEnd(maxTicks?: number): this {
    const cap = maxTicks ?? this.scenario.stopAfter ?? 600
    while (!this.ctx.world.finished && this.ctx.world.tick < cap) this.step()
    if (!this.ctx.world.finished) {
      this.ctx.world.finished = true
      this.ctx.world.finishReason = 'stop-after'
    }
    return this
  }

  at(tick: number): Snapshot | undefined {
    return this.snapshots[tick]
  }

  get history(): readonly Snapshot[] {
    return this.snapshots
  }

  countEvents(type: GcEvent['type']): number {
    return this.events.filter((e) => e.type === type).length
  }

  /** Текущий размер кучи в блоках. */
  get heap(): number {
    return heapUsed(this.ctx.world)
  }

  /** Завершённых циклов сборки. */
  get cycles(): number {
    return this.ctx.world.stats.cycles
  }

  /** Освобождённых достижимых объектов. При исправном барьере записи — всегда ноль. */
  get lost(): number {
    return this.ctx.world.stats.lost
  }

  /** Доля процессорного времени, ушедшая сборщику: фоновая разметка, помощь, паузы, подметание. */
  get gcCpuShare(): number {
    const { gcSlotTicks, mutatorSlotTicks } = this.ctx.world.stats
    const total = gcSlotTicks + mutatorSlotTicks
    return total === 0 ? 0 : gcSlotTicks / total
  }

  /** Сколько тиков программа занималась своим делом — мера полезной работы. */
  get mutatorProgress(): number {
    return this.ctx.world.muts.reduce((sum, m) => sum + m.cpuTicks, 0)
  }

  /** Самая большая куча за прогон. */
  get peakHeap(): number {
    return this.ctx.world.stats.peakHeap
  }
}
