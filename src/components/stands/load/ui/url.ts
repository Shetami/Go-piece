import type { LoadConfig, LoadScenario, Phase } from '../engine/types.ts'
import { resolveConfig } from '../engine/world.ts'
import { loadScenarioById } from '../engine/scenarios.ts'

/**
 * Сценарий ↔ строка для адреса. Хранится только то, что отличается от пресета:
 * `s=retry-storm&cancelOnTimeout=true`. Изменённые фазы нагрузки едут в `p`
 * как `40:2,60:1` — с тика 40 поток ×2, с тика 60 обычный.
 */

export interface StandState {
  base: string
  config: LoadConfig
  phases: Phase[] | null
}

const CONFIG_KEYS = Object.keys(resolveConfig(undefined)) as (keyof LoadConfig)[]

export function initialState(base: LoadScenario): StandState {
  return { base: base.id, config: resolveConfig(base.config), phases: null }
}

export function toScenario(st: StandState): LoadScenario {
  const base = loadScenarioById(st.base)
  if (!base) throw new Error(`нет сценария ${st.base}`)
  return { ...base, config: st.config, phases: st.phases ?? base.phases }
}

export function encodeState(st: StandState): string {
  const base = loadScenarioById(st.base)
  const baseCfg = resolveConfig(base?.config)
  const q = new URLSearchParams()
  q.set('s', st.base)
  for (const k of CONFIG_KEYS) if (st.config[k] !== baseCfg[k]) q.set(k, String(st.config[k]))
  if (st.phases) q.set('p', st.phases.map((p) => `${p.at}:${p.x}`).join(','))
  return q.toString()
}

function parseValue(raw: string): unknown {
  if (raw === 'false') return false
  if (raw === 'true') return true
  const n = Number(raw)
  return Number.isFinite(n) ? n : raw
}

export function decodeState(query: string, fallback: LoadScenario): StandState {
  const q = new URLSearchParams(query.replace(/^[#?]/, ''))
  const base = loadScenarioById(q.get('s') ?? '') ?? fallback
  const st = initialState(base)
  const cfg: Record<string, unknown> = { ...st.config }
  for (const k of CONFIG_KEYS) {
    const raw = q.get(k)
    if (raw !== null) cfg[k] = parseValue(raw)
  }
  st.config = cfg as unknown as LoadConfig
  const p = q.get('p')
  if (p !== null) {
    st.phases = p
      .split(',')
      .map((s) => s.split(':').map(Number))
      .filter(([at, x]) => Number.isFinite(at) && Number.isFinite(x) && x! >= 0)
      .map(([at, x]) => ({ at: at!, x: x! }))
  }
  return st
}
