import type { GcConfig, GcScenario, MutWorkload } from '../engine/types.ts'
import { resolveConfig } from '../engine/world.ts'
import { gcScenarioById } from '../engine/scenarios.ts'

/**
 * Сценарий ↔ строка для адреса.
 *
 * Хранится только то, что отличается от пресета: `s=mem-limit&seed=3&gogc=400`.
 * Изменённые нагрузки не раскладываются на параметры, а едут одним base64-JSON в `w`.
 */

export interface StandState {
  base: string
  seed: number
  config: GcConfig
  /** null — нагрузки пресета без изменений. */
  workloads: MutWorkload[] | null
}

const CONFIG_KEYS = Object.keys(resolveConfig(undefined)) as (keyof GcConfig)[]

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

export function initialState(base: GcScenario, seed?: number): StandState {
  return {
    base: base.id,
    seed: seed ?? base.seed ?? 1,
    config: resolveConfig(base.config),
    workloads: null,
  }
}

export function toScenario(st: StandState): GcScenario {
  const base = gcScenarioById(st.base)
  if (!base) throw new Error(`нет сценария ${st.base}`)
  return { ...base, config: st.config, workloads: st.workloads ?? base.workloads }
}

export function encodeState(st: StandState): string {
  const base = gcScenarioById(st.base)
  const baseCfg = resolveConfig(base?.config)
  const q = new URLSearchParams()
  q.set('s', st.base)
  q.set('seed', String(st.seed))
  for (const k of CONFIG_KEYS) {
    if (st.config[k] !== baseCfg[k]) q.set(k, String(st.config[k]))
  }
  if (st.workloads) q.set('w', b64encode(JSON.stringify(st.workloads)))
  return q.toString()
}

/** GOGC и GOMEMLIMIT умеют быть выключенными, поэтому «false» разбирается раньше чисел. */
function parseValue(raw: string): unknown {
  if (raw === 'false') return false
  if (raw === 'true') return true
  const n = Number(raw)
  return Number.isFinite(n) ? n : raw
}

export function decodeState(query: string, fallback: GcScenario): StandState {
  const q = new URLSearchParams(query.replace(/^[#?]/, ''))
  const base = gcScenarioById(q.get('s') ?? '') ?? fallback
  const st = initialState(base)
  const seed = Number(q.get('seed'))
  if (Number.isInteger(seed) && seed > 0) st.seed = seed

  const cfg: Record<string, unknown> = { ...st.config }
  for (const k of CONFIG_KEYS) {
    const raw = q.get(k)
    if (raw !== null) cfg[k] = parseValue(raw)
  }
  st.config = cfg as unknown as GcConfig

  const w = q.get('w')
  if (w) {
    try {
      const parsed = JSON.parse(b64decode(w)) as unknown
      if (Array.isArray(parsed)) st.workloads = parsed as MutWorkload[]
    } catch {}
  }
  return st
}
