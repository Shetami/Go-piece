import type { Frame, LoadEvent, LoadLog, LoadScenario, LoadWorld } from './types.ts'
import type { Ctx } from './world.ts'
import { createWorld, emptySample } from './world.ts'
import { tick as runTick } from './tick.ts'

/**
 * Симуляция: движок плюс история состояний.
 * Снимок мира на каждом тике маленький — очередь, воркеры и счётчики, — поэтому
 * шаг назад и перемотка становятся чтением готового. А то, что только растёт, —
 * точки графиков и ушедшие запросы — лежит один раз в журнале `log`.
 * Случайность есть, но с зерном: прогон всегда один и тот же.
 */
export class Simulation {
  readonly scenario: LoadScenario
  readonly log: LoadLog

  private ctx: Ctx
  private frames: Frame[] = []

  constructor(scenario: LoadScenario) {
    this.scenario = scenario
    const world = createWorld(scenario)
    this.log = { series: [emptySample(world.x)], done: [] }
    this.ctx = { world, events: [], scenario, log: this.log }
    this.frames.push({ world: structuredClone(world), events: [] })
  }

  get world(): LoadWorld {
    return this.ctx.world
  }

  get tick(): number {
    return this.ctx.world.tick
  }

  get finished(): boolean {
    return this.ctx.world.finished
  }

  get events(): LoadEvent[] {
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
    const cap = maxTicks ?? this.scenario.stopAfter
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

  countEvents(type: LoadEvent['type']): number {
    return this.events.filter((e) => e.type === type).length
  }

  firstEvent(type: LoadEvent['type']): LoadEvent | undefined {
    return this.events.find((e) => e.type === type)
  }
}
