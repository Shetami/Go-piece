import type { MapEvent, MapEventType, MapWorld, OpTrace, Probe } from './types.ts'
import type { Rng } from './rng.ts'

/** Всё, что нужно операции: мир, куда писать события и след, и генератор. */
export interface Ctx {
  world: MapWorld
  events: MapEvent[]
  trace: OpTrace
  rng: Rng
  /** Строка сценария текущей команды. */
  line: number
}

export function emit(ctx: Ctx, type: MapEventType, payload: Record<string, unknown> = {}): void {
  ctx.events.push({ tick: ctx.world.tick, type, line: ctx.line, payload })
}

/** Записать проверенную группу (страницу) в след и в счётчики. */
export function recordProbe(ctx: Ctx, p: Probe): Probe {
  ctx.trace.probes.push(p)
  ctx.world.stats.probes++
  return p
}

export function pow2ceil(n: number): number {
  let p = 1
  while (p < n) p *= 2
  return p
}
