import type { KafkaEvent, KafkaScenario, KafkaWorld, Snapshot } from './types.ts'
import type { Ctx } from './world.ts'
import { createWorld } from './world.ts'
import { Rng } from './rng.ts'
import { tick as runTick } from './tick.ts'

/**
 * Симуляция: движок плюс история состояний.
 * Кластер маленький — три брокера и полсотни сообщений, — поэтому полный снимок
 * на каждом тике стоит дёшево, а шаг назад и перемотка становятся чтением готового.
 */
export class Simulation {
  readonly scenario: KafkaScenario
  readonly seed: number

  private ctx: Ctx
  private rng: Rng
  private snapshots: Snapshot[] = []

  constructor(scenario: KafkaScenario, seed?: number) {
    this.scenario = scenario
    this.seed = seed ?? scenario.seed ?? 1
    const world = createWorld(scenario)
    this.rng = new Rng(this.seed)
    this.ctx = { world, events: [], scenario }
    this.snapshots.push({ world: structuredClone(world), events: [] })
  }

  get world(): KafkaWorld {
    return this.ctx.world
  }

  get tick(): number {
    return this.ctx.world.tick
  }

  get finished(): boolean {
    return this.ctx.world.finished
  }

  get events(): KafkaEvent[] {
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

  countEvents(type: KafkaEvent['type']): number {
    return this.events.filter((e) => e.type === type).length
  }

  /** Сколько записей в среднем уходит одним запросом. */
  get avgBatch(): number {
    const s = this.ctx.world.stats
    return s.requests === 0 ? 0 : s.recsSent / s.requests
  }

  /** Средняя задержка от send() до подтверждения. */
  get avgAck(): number {
    const s = this.ctx.world.stats
    return s.ackCount === 0 ? 0 : s.ackLatency / s.ackCount
  }

  /** Средняя задержка от send() до первой обработки потребителем. */
  get avgE2e(): number {
    const s = this.ctx.world.stats
    return s.e2eCount === 0 ? 0 : s.e2eLatency / s.e2eCount
  }
}
