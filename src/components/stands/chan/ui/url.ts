import type { ChanConfig, ChanScenario, ChanSpec, ChanWorkload } from '../engine/types.ts'
import { resolveConfig } from '../engine/world.ts'
import { chanScenarioById } from '../engine/scenarios.ts'

/**
 * Сценарий ↔ строка для адреса.
 *
 * Хранится только то, что отличается от пресета: `s=buffer&seed=3&gomaxprocs=8`.
 * Изменённые каналы и нагрузки не раскладываются на параметры, а едут одним
 * base64-JSON в `c` и `w`.
 */

export interface StandState {
  base: string
  seed: number
  config: ChanConfig
  /** null — каналы и нагрузки пресета без изменений. */
  chans: ChanSpec[] | null
  workloads: ChanWorkload[] | null
}

const CONFIG_KEYS = Object.keys(resolveConfig(undefined)) as (keyof ChanConfig)[]

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

export function initialState(base: ChanScenario, seed?: number): StandState {
  return {
    base: base.id,
    seed: seed ?? base.seed ?? 1,
    config: resolveConfig(base.config),
    chans: null,
    workloads: null,
  }
}

export function toScenario(st: StandState): ChanScenario {
  const base = chanScenarioById(st.base)
  if (!base) throw new Error(`нет сценария ${st.base}`)
  return {
    ...base,
    config: st.config,
    chans: st.chans ?? base.chans,
    workloads: st.workloads ?? base.workloads,
  }
}

export function encodeState(st: StandState): string {
  const base = chanScenarioById(st.base)
  const baseCfg = resolveConfig(base?.config)
  const q = new URLSearchParams()
  q.set('s', st.base)
  q.set('seed', String(st.seed))
  for (const k of CONFIG_KEYS) {
    if (st.config[k] !== baseCfg[k]) q.set(k, String(st.config[k]))
  }
  if (st.chans) q.set('c', b64encode(JSON.stringify(st.chans)))
  if (st.workloads) q.set('w', b64encode(JSON.stringify(st.workloads)))
  return q.toString()
}

/** Тумблеры умеют быть выключенными, поэтому «false» разбирается раньше чисел. */
function parseValue(raw: string): unknown {
  if (raw === 'false') return false
  if (raw === 'true') return true
  const n = Number(raw)
  return Number.isFinite(n) ? n : raw
}

export function decodeState(query: string, fallback: ChanScenario): StandState {
  const q = new URLSearchParams(query.replace(/^[#?]/, ''))
  const base = chanScenarioById(q.get('s') ?? '') ?? fallback
  const st = initialState(base)
  const seed = Number(q.get('seed'))
  if (Number.isInteger(seed) && seed > 0) st.seed = seed

  const cfg: Record<string, unknown> = { ...st.config }
  for (const k of CONFIG_KEYS) {
    const raw = q.get(k)
    if (raw !== null) cfg[k] = parseValue(raw)
  }
  st.config = cfg as unknown as ChanConfig

  const c = q.get('c')
  if (c) {
    try {
      const parsed = JSON.parse(b64decode(c)) as unknown
      if (Array.isArray(parsed)) st.chans = parsed as ChanSpec[]
    } catch {}
  }
  const w = q.get('w')
  if (w) {
    try {
      const parsed = JSON.parse(b64decode(w)) as unknown
      if (Array.isArray(parsed)) st.workloads = parsed as ChanWorkload[]
    } catch {}
  }
  return st
}
