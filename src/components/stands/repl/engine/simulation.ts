import type { Frame, ReplEvent, ReplScenario, ReplWorld } from './types.ts'
import type { Ctx } from './world.ts'
import { createWorld } from './world.ts'
import { tick as runTick } from './tick.ts'

/**
 * Симуляция: движок плюс история состояний.
 * Мир маленький — несколько узлов и клиентов, — поэтому полный
 * снимок на каждом тике стоит дёшево, а шаг назад и перемотка становятся
 * чтением готового. Случайности в этом стенде нет: расписание операций задано
 * сценарием, и прогон всегда один и тот же.
 */
export class Simulation {
  readonly scenario: ReplScenario

  private ctx: Ctx
  private frames: Frame[] = []

  constructor(scenario: ReplScenario) {
    this.scenario = scenario
    const world = createWorld(scenario)
    this.ctx = { world, events: [], scenario }
    this.frames.push({ world: structuredClone(world), events: [] })
  }

  get world(): ReplWorld {
    return this.ctx.world
  }

  get tick(): number {
    return this.ctx.world.tick
  }

  get finished(): boolean {
    return this.ctx.world.finished
  }

  get events(): ReplEvent[] {
    return this.frames.flatMap((f) => f.events)
  }

  step(): Frame | null {
    if (this.ctx.world.finished) return null
    this.ctx.events = []
    runTick(this.ctx)
    const frame: Frame = { world: structuredClone(this.ctx.world), events: this.ctx.events }
    this.frames.push(frame)
    return frame
  }

  runTo(tick: number): this {
    while (this.ctx.world.tick < tick && !this.ctx.world.finished) this.step()
    return this
  }

  runToEnd(maxTicks?: number): this {
    const cap = maxTicks ?? this.scenario.stopAfter ?? 200
    while (!this.ctx.world.finished && this.ctx.world.tick < cap) this.step()
    if (!this.ctx.world.finished) {
      this.ctx.world.finished = true
      this.ctx.world.finishReason = 'stop-after'
    }
    return this
  }

  at(tick: number): Frame | undefined {
    return this.frames[tick]
  }

  get history(): readonly Frame[] {
    return this.frames
  }

  countEvents(type: ReplEvent['type']): number {
    return this.events.filter((e) => e.type === type).length
  }

  firstEvent(type: ReplEvent['type']): ReplEvent | undefined {
    return this.events.find((e) => e.type === type)
  }
}
