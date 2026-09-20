import type { MemEvent, MemScenario, MemWorld, Snapshot } from './types.ts'
import type { Ctx, SpawnPlanItem } from './world.ts'
import { buildSpawnPlan, createWorld, mappedPages, usedPages } from './world.ts'
import { Rng } from './rng.ts'
import { tick as runTick } from './tick.ts'

/**
 * Симуляция: движок плюс история состояний.
 * Куча в модели фиксированного размера, объекты переиспользуют слоты, поэтому
 * снимок дёшев, а шаг назад и перемотка — это чтение готового снимка.
 */
export class Simulation {
  readonly scenario: MemScenario
  readonly seed: number
  readonly plan: SpawnPlanItem[]

  private ctx: Ctx
  private rng: Rng
  private snapshots: Snapshot[] = []

  constructor(scenario: MemScenario, seed?: number) {
    this.scenario = scenario
    this.seed = seed ?? scenario.seed ?? 1
    const world = createWorld(scenario)
    this.plan = buildSpawnPlan(scenario)
    this.rng = new Rng(this.seed)
    this.ctx = { world, events: [], plan: this.plan, scenario }
    this.snapshots.push({ world: structuredClone(world), events: [] })
  }

  get world(): MemWorld {
    return this.ctx.world
  }

  get tick(): number {
    return this.ctx.world.tick
  }

  get finished(): boolean {
    return this.ctx.world.finished
  }

  get events(): MemEvent[] {
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
    const cap = maxTicks ?? this.scenario.stopAfter ?? 400
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

  countEvents(type: MemEvent['type']): number {
    return this.events.filter((e) => e.type === type).length
  }

  /** Сколько аллокаций обошлись кэшем процессора — без единой блокировки. */
  get fastShare(): number {
    const s = this.ctx.world.stats
    const total = s.fast + s.tiny + s.refills + s.large + s.onStack
    return total === 0 ? 0 : (s.fast + s.tiny + s.onStack) / total
  }

  /** Доля памяти, потерянной на округлении до класса размеров. */
  get waste(): number {
    const s = this.ctx.world.stats
    return s.roundSlot === 0 ? 0 : (s.roundSlot - s.roundAsked) / s.roundSlot
  }

  /** Занятые страницы кучи прямо сейчас. */
  get heapPages(): number {
    return usedPages(this.ctx.world)
  }

  /** Сколько страниц рантайм держит у операционной системы. */
  get rss(): number {
    return mappedPages(this.ctx.world)
  }
}
