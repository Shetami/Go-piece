import type { Fault, CacheConfig, CacheScenario, Phase } from '../engine/types.ts'
import { resolveConfig } from '../engine/world.ts'
import { cacheScenarioById } from '../engine/scenarios.ts'

/**
 * Сценарий ↔ строка для адреса. Хранится только то, что отличается от пресета:
 * `s=ttl-sync&coalesce=true`. Фазы нагрузки едут в `p` как `40:2,60:1`,
 * изменённые сбои — base64-JSON в `f`.
 */

export interface StandState {
  base: string
  config: CacheConfig
  phases: Phase[] | null
  faults: Fault[] | null
}

const CONFIG_KEYS = Object.keys(resolveConfig(undefined)) as (keyof CacheConfig)[]

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

export function initialState(base: CacheScenario): StandState {
  return { base: base.id, config: resolveConfig(base.config), phases: null, faults: null }
}

export function toScenario(st: StandState): CacheScenario {
  const base = cacheScenarioById(st.base)
  if (!base) throw new Error(`нет сценария ${st.base}`)
  return { ...base, config: st.config, phases: st.phases ?? base.phases, faults: st.faults ?? base.faults }
}

export function encodeState(st: StandState): string {
  const base = cacheScenarioById(st.base)
  const baseCfg = resolveConfig(base?.config)
  const q = new URLSearchParams()
  q.set('s', st.base)
  for (const k of CONFIG_KEYS) if (st.config[k] !== baseCfg[k]) q.set(k, String(st.config[k]))
  if (st.phases) q.set('p', st.phases.map((p) => `${p.at}:${p.x}`).join(','))
  if (st.faults) q.set('f', b64encode(JSON.stringify(st.faults)))
  return q.toString()
}

function parseValue(raw: string): unknown {
  if (raw === 'false') return false
  if (raw === 'true') return true
  const n = Number(raw)
  return Number.isFinite(n) ? n : raw
}

export function decodeState(query: string, fallback: CacheScenario): StandState {
  const q = new URLSearchParams(query.replace(/^[#?]/, ''))
  const base = cacheScenarioById(q.get('s') ?? '') ?? fallback
  const st = initialState(base)
  const cfg: Record<string, unknown> = { ...st.config }
  for (const k of CONFIG_KEYS) {
    const raw = q.get(k)
    if (raw !== null) cfg[k] = parseValue(raw)
  }
  st.config = cfg as unknown as CacheConfig
  const p = q.get('p')
  if (p !== null) {
    st.phases = p
      .split(',')
      .map((s) => s.split(':').map(Number))
      .filter(([at, x]) => Number.isFinite(at) && Number.isFinite(x) && x! >= 0)
      .map(([at, x]) => ({ at: at!, x: x! }))
  }
  const f = q.get('f')
  if (f) {
    try {
      const parsed = JSON.parse(b64decode(f)) as unknown
      if (Array.isArray(parsed)) st.faults = parsed as Fault[]
    } catch {}
  }
  return st
}
