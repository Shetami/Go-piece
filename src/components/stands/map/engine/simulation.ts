import type { Ctx } from './ctx.ts'
import { Rng } from './rng.ts'
import { parseScript, type ParseError } from './script.ts'
import type { MapEvent, MapScenario, MapWorld, Snapshot } from './types.ts'
import { createWorld, mapCapacity, mapLen, tick as runTick, type Runner } from './world.ts'

/**
 * Симуляция: движок плюс история состояний.
 * Карта в стенде — сотня-другая слотов, поэтому полный снимок на каждом тике
 * стоит дёшево, а шаг назад и перемотка становятся чтением готового.
 */
export class Simulation {
  readonly scenario: MapScenario
  readonly seed: number
  readonly errors: ParseError[]

  private ctx: Ctx
  private run: Runner
  private snapshots: Snapshot[] = []

  constructor(scenario: MapScenario, seed?: number) {
    this.scenario = scenario
    this.seed = seed ?? scenario.seed ?? 1
    const { ops, errors } = parseScript(scenario.script)
    this.errors = errors
    const world = createWorld(scenario)
    this.ctx = { world, events: [], trace: { code: '', probes: [], outcome: '' }, rng: new Rng(this.seed), line: 0 }
    this.run = { queue: [{ kind: 'make', line: 0 }, ...ops] }
    this.snapshots.push({ world: structuredClone(world), events: [], trace: null })
  }

  get world(): MapWorld {
    return this.ctx.world
  }

  get tick(): number {
    return this.ctx.world.tick
  }

  get finished(): boolean {
    return this.ctx.world.finished
  }

  get events(): MapEvent[] {
    return this.snapshots.flatMap((s) => s.events)
  }

  step(): Snapshot | null {
    const w = this.ctx.world
    if (w.finished) return null
    const before = w.tick
    runTick(this.ctx, this.run)
    if (w.tick === before) return null
    const snap: Snapshot = { world: structuredClone(w), events: this.ctx.events, trace: this.ctx.trace }
    this.snapshots.push(snap)
    return snap
  }

  runToEnd(maxTicks = 600): this {
    while (!this.ctx.world.finished && this.ctx.world.tick < maxTicks) this.step()
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

  countEvents(type: MapEvent['type']): number {
    return this.events.filter((e) => e.type === type).length
  }

  get len(): number {
    return mapLen(this.ctx.world)
  }

  get capacity(): number {
    return mapCapacity(this.ctx.world)
  }
}
