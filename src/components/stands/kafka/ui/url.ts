import type { ConsumerSpec, Fault, KafkaConfig, KafkaScenario, ProducerSpec } from '../engine/types.ts'
import { resolveConfig } from '../engine/world.ts'
import { kafkaScenarioById } from '../engine/scenarios.ts'

/**
 * Сценарий ↔ строка для адреса.
 *
 * Хранится только то, что отличается от пресета: `s=acks&seed=3&acks=all`.
 * Изменённые продюсеры, потребители и сбои едут одним base64-JSON в `p`, `c` и `f`.
 */

export interface StandState {
  base: string
  seed: number
  config: KafkaConfig
  /** null — как в пресете. */
  producers: ProducerSpec[] | null
  consumers: ConsumerSpec[] | null
  faults: Fault[] | null
}

const CONFIG_KEYS = Object.keys(resolveConfig(undefined)) as (keyof KafkaConfig)[]

function b64encode(s: string): string {
  const bytes = new TextEncoder().encode(s)
  let bin = ''
  for (const b of bytes) bin += String.fromCharCode(b)
  return btoa(bin).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

function b64decode(s: string): string {
  const bin = atob(s.replace(/-/g, '+').replace(/_/g, '/'))
  return new TextDecoder().decode(Uint8Array.from(bin, (c) => c.charCodeAt(0)))
}

export function initialState(base: KafkaScenario, seed?: number): StandState {
  return {
    base: base.id,
    seed: seed ?? base.seed ?? 1,
    config: resolveConfig(base.config),
    producers: null,
    consumers: null,
    faults: null,
  }
}

export function toScenario(st: StandState): KafkaScenario {
  const base = kafkaScenarioById(st.base)
  if (!base) throw new Error(`нет сценария ${st.base}`)
  return {
    ...base,
    config: resolveConfig(st.config),
    producers: st.producers ?? base.producers,
    consumers: st.consumers ?? base.consumers,
    faults: st.faults ?? base.faults,
  }
}

export function encodeState(st: StandState): string {
  const base = kafkaScenarioById(st.base)
  const baseCfg = resolveConfig(base?.config)
  const q = new URLSearchParams()
  q.set('s', st.base)
  q.set('seed', String(st.seed))
  for (const k of CONFIG_KEYS) {
    if (st.config[k] !== baseCfg[k]) q.set(k, String(st.config[k]))
  }
  if (st.producers) q.set('p', b64encode(JSON.stringify(st.producers)))
  if (st.consumers) q.set('c', b64encode(JSON.stringify(st.consumers)))
  if (st.faults) q.set('f', b64encode(JSON.stringify(st.faults)))
  return q.toString()
}

/** Тумблеры умеют быть выключенными, поэтому «false» разбирается раньше чисел. */
function parseValue(raw: string): unknown {
  if (raw === 'false') return false
  if (raw === 'true') return true
  const n = Number(raw)
  return Number.isFinite(n) ? n : raw
}

function parseList<T>(raw: string | null): T[] | null {
  if (!raw) return null
  try {
    const parsed = JSON.parse(b64decode(raw)) as unknown
    return Array.isArray(parsed) ? (parsed as T[]) : null
  } catch {
    return null
  }
}

export function decodeState(query: string, fallback: KafkaScenario): StandState {
  const q = new URLSearchParams(query.replace(/^[#?]/, ''))
  const base = kafkaScenarioById(q.get('s') ?? '') ?? fallback
  const st = initialState(base)
  const seed = Number(q.get('seed'))
  if (Number.isInteger(seed) && seed > 0) st.seed = seed

  const cfg: Record<string, unknown> = { ...st.config }
  for (const k of CONFIG_KEYS) {
    const raw = q.get(k)
    if (raw !== null) cfg[k] = parseValue(raw)
  }
  st.config = resolveConfig(cfg as unknown as KafkaConfig)
  st.producers = parseList<ProducerSpec>(q.get('p'))
  st.consumers = parseList<ConsumerSpec>(q.get('c'))
  st.faults = parseList<Fault>(q.get('f'))
  return st
}
