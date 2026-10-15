import type { MapConfig, MapScenario } from '../engine/types.ts'
import { resolveConfig } from '../engine/world.ts'
import { mapScenarioById } from '../engine/scenarios.ts'

/**
 * Сценарий ↔ строка для адреса.
 *
 * Хранится только то, что отличается от пресета: `s=growth&seed=3&maxTableSlots=16`.
 * Изменённый текст сценария едет одним base64 в `x`.
 */

export interface StandState {
  base: string
  seed: number
  config: MapConfig
  /** null — сценарий пресета без изменений. */
  script: string | null
}

const CONFIG_KEYS = Object.keys(resolveConfig(undefined)) as (keyof MapConfig)[]

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

export function initialState(base: MapScenario, seed?: number): StandState {
  return { base: base.id, seed: seed ?? base.seed ?? 1, config: resolveConfig(base.config), script: null }
}

export function toScenario(st: StandState): MapScenario {
  const base = mapScenarioById(st.base)
  if (!base) throw new Error(`нет сценария ${st.base}`)
  return { ...base, config: st.config, script: st.script ?? base.script }
}

export function encodeState(st: StandState): string {
  const baseCfg = resolveConfig(mapScenarioById(st.base)?.config)
  const q = new URLSearchParams()
  q.set('s', st.base)
  q.set('seed', String(st.seed))
  for (const k of CONFIG_KEYS) if (st.config[k] !== baseCfg[k]) q.set(k, String(st.config[k]))
  if (st.script !== null) q.set('x', b64encode(st.script))
  return q.toString()
}

/** Тумблеры умеют быть выключенными, поэтому «false» разбирается раньше чисел. */
function parseValue(raw: string): unknown {
  if (raw === 'false') return false
  if (raw === 'true') return true
  const n = Number(raw)
  return Number.isFinite(n) ? n : raw
}

export function decodeState(query: string, fallback: MapScenario): StandState {
  const q = new URLSearchParams(query.replace(/^[#?]/, ''))
  const base = mapScenarioById(q.get('s') ?? '') ?? fallback
  const st = initialState(base)
  const seed = Number(q.get('seed'))
  if (Number.isInteger(seed) && seed > 0) st.seed = seed

  const cfg: Record<string, unknown> = { ...st.config }
  for (const k of CONFIG_KEYS) {
    const raw = q.get(k)
    if (raw !== null) cfg[k] = parseValue(raw)
  }
  st.config = cfg as unknown as MapConfig
  if (st.config.impl !== 'swiss' && st.config.impl !== 'legacy') st.config.impl = 'swiss'

  const x = q.get('x')
  if (x) {
    try {
      st.script = b64decode(x)
    } catch {}
  }
  return st
}
