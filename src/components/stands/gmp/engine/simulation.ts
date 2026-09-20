import type { Scenario, SimEvent, Snapshot, World } from './types.ts'
import type { Ctx, SpawnPlanItem } from './world.ts'
import { buildSpawnPlan, createWorld, resolveConfig } from './world.ts'
import { Rng } from './rng.ts'
import { tick as runTick } from './tick.ts'

/**
 * Симуляция: движок плюс история состояний.
 *
 * История — массив полных снимков, а не дельт. При потолке в тысячу тиков
 * и десятках горутин это сотни килобайт, зато шаг назад, перемотка по
 * таймлайну и сравнение двух прогонов становятся тривиальными.
 */
export class Simulation {
  readonly scenario: Scenario
  readonly seed: number
  readonly plan: SpawnPlanItem[]

  private ctx: Ctx
  private rng: Rng
  private snapshots: Snapshot[] = []

  constructor(scenario: Scenario, seed?: number) {
    this.scenario = scenario
    this.seed = seed ?? scenario.seed ?? 1
    const world = createWorld(scenario)
    this.plan = buildSpawnPlan(scenario, resolveConfig(scenario.config).gomaxprocs)
    this.rng = new Rng(this.seed)
    this.ctx = { world, events: [], plan: this.plan, scenario }
    this.snapshots.push({ world: structuredClone(world), events: [] })
  }

  get world(): World {
    return this.ctx.world
  }

  get tick(): number {
    return this.ctx.world.tick
  }

  get finished(): boolean {
    return this.ctx.world.finished
  }

  /** Все события с начала прогона, по порядку. */
  get events(): SimEvent[] {
    return this.snapshots.flatMap((s) => s.events)
  }

  /** Один шаг. Возвращает снимок или null, если симуляция уже закончилась. */
  step(): Snapshot | null {
    if (this.ctx.world.finished) return null
    this.ctx.events = []
    runTick(this.ctx, this.rng)
    const snap: Snapshot = {
      world: structuredClone(this.ctx.world),
      events: this.ctx.events,
    }
    this.snapshots.push(snap)
    return snap
  }

  /** Прогнать до указанного тика включительно. */
  runTo(tick: number): this {
    while (this.ctx.world.tick < tick && !this.ctx.world.finished) this.step()
    return this
  }

  /** Прогнать до естественного конца или до потолка. */
  runToEnd(maxTicks?: number): this {
    const cap = maxTicks ?? this.scenario.stopAfter ?? 1000
    while (!this.ctx.world.finished && this.ctx.world.tick < cap) this.step()
    if (!this.ctx.world.finished) {
      this.ctx.world.finished = true
      this.ctx.world.finishReason = 'stop-after'
    }
    return this
  }

  /** Снимок на конкретном тике. at(0) — состояние до первого шага. */
  at(tick: number): Snapshot | undefined {
    return this.snapshots[tick]
  }

  get history(): readonly Snapshot[] {
    return this.snapshots
  }

  /** Максимальное число потоков, существовавших одновременно. */
  get peakThreads(): number {
    return this.snapshots.reduce((max, s) => Math.max(max, s.world.ms.length), 0)
  }

  countEvents(type: SimEvent['type']): number {
    return this.events.filter((e) => e.type === type).length
  }

  /** Длины локальных очередей на текущем тике — для проверки балансировки. */
  runqLengths(): number[] {
    return this.ctx.world.ps.map((p) => p.runq.length + (p.runnext === null ? 0 : 1))
  }
}
