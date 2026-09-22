import type { ChanEvent, ChanScenario, ChanWorld, Snapshot } from './types.ts'
import type { Ctx, SpawnPlanItem } from './world.ts'
import { buildSpawnPlan, createWorld } from './world.ts'
import { Rng } from './rng.ts'
import { tick as runTick } from './tick.ts'

/**
 * Симуляция: движок плюс история состояний.
 * Мир маленький — десяток горутин и пара каналов, — поэтому полный снимок на
 * каждом тике стоит дёшево, а шаг назад и перемотка становятся чтением готового.
 */
export class Simulation {
  readonly scenario: ChanScenario
  readonly seed: number
  readonly plan: SpawnPlanItem[]

  private ctx: Ctx
  private rng: Rng
  private snapshots: Snapshot[] = []

  constructor(scenario: ChanScenario, seed?: number) {
    this.scenario = scenario
    this.seed = seed ?? scenario.seed ?? 1
    const world = createWorld(scenario)
    this.plan = buildSpawnPlan(scenario)
    this.rng = new Rng(this.seed)
    this.ctx = { world, events: [], plan: this.plan, scenario }
    this.snapshots.push({ world: structuredClone(world), events: [] })
  }

  get world(): ChanWorld {
    return this.ctx.world
  }

  get tick(): number {
    return this.ctx.world.tick
  }

  get finished(): boolean {
    return this.ctx.world.finished
  }

  get events(): ChanEvent[] {
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
    const cap = maxTicks ?? this.scenario.stopAfter ?? 300
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

  countEvents(type: ChanEvent['type']): number {
    return this.events.filter((e) => e.type === type).length
  }

  /** Доля передач, прошедших мимо буфера — из рук в руки. */
  get directShare(): number {
    const s = this.ctx.world.stats
    return s.transfers === 0 ? 0 : s.direct / s.transfers
  }

  /** Сколько тиков в среднем горутина проводит в парковке. */
  get avgWait(): number {
    const s = this.ctx.world.stats
    return s.parks === 0 ? 0 : s.waitTicks / s.parks
  }

  /** Горутины, которые сейчас спят на каналах. */
  get parked(): number {
    return this.ctx.world.gs.filter((g) => g.state === 'waiting').length
  }
}
