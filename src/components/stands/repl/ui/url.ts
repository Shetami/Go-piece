import type { ClientOp, ClientSpec, Fault, ReplConfig, ReplScenario } from '../engine/types.ts'
import { resolveConfig } from '../engine/world.ts'
import { replScenarioById } from '../engine/scenarios.ts'

/**
 * Сценарий ↔ строка для адреса. Хранится только то, что отличается от пресета:
 * `s=failover&standbys=any-1`. Изменённые клиенты и сбои едут base64-JSON в `c` и `f`.
 */

export interface StandState {
  base: string
  config: ReplConfig
  clients: ClientSpec[] | null
  faults: Fault[] | null
}

const CONFIG_KEYS = Object.keys(resolveConfig(undefined)) as (keyof ReplConfig)[]

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

export function initialState(base: ReplScenario): StandState {
  return { base: base.id, config: resolveConfig(base.config), clients: null, faults: null }
}

export function toScenario(st: StandState): ReplScenario {
  const base = replScenarioById(st.base)
  if (!base) throw new Error(`нет сценария ${st.base}`)
  const n = st.config.replicas
  // Реплики, которых больше нет, из чтения выпадают: чтение идёт на любую, отчёт — на первую.
  const fix = (o: ClientOp): ClientOp => {
    if (o.kind === 'read' && typeof o.from === 'number' && o.from > n) return { ...o, from: 'any' }
    if (o.kind === 'query' && o.from > n) return { ...o, from: 1 }
    return o
  }
  const clients: ClientSpec[] = (st.clients ?? base.clients).map((c) => ({ ...c, ops: c.ops.map(fix) }))
  const faults = (st.faults ?? base.faults).filter((f) => f.kind === 'primary-down' || f.replica <= n)
  return { ...base, config: st.config, clients, faults }
}

export function encodeState(st: StandState): string {
  const base = replScenarioById(st.base)
  const baseCfg = resolveConfig(base?.config)
  const q = new URLSearchParams()
  q.set('s', st.base)
  for (const k of CONFIG_KEYS) if (st.config[k] !== baseCfg[k]) q.set(k, String(st.config[k]))
  if (st.clients) q.set('c', b64encode(JSON.stringify(st.clients)))
  if (st.faults) q.set('f', b64encode(JSON.stringify(st.faults)))
  return q.toString()
}

function parseValue(raw: string): unknown {
  if (raw === 'false') return false
  if (raw === 'true') return true
  const n = Number(raw)
  return Number.isFinite(n) ? n : raw
}

export function decodeState(query: string, fallback: ReplScenario): StandState {
  const q = new URLSearchParams(query.replace(/^[#?]/, ''))
  const base = replScenarioById(q.get('s') ?? '') ?? fallback
  const st = initialState(base)
  const cfg: Record<string, unknown> = { ...st.config }
  for (const k of CONFIG_KEYS) {
    const raw = q.get(k)
    if (raw !== null) cfg[k] = parseValue(raw)
  }
  st.config = cfg as unknown as ReplConfig
  for (const [param, field] of [['c', 'clients'], ['f', 'faults']] as const) {
    const raw = q.get(param)
    if (!raw) continue
    try {
      const parsed = JSON.parse(b64decode(raw)) as unknown
      if (Array.isArray(parsed)) (st as unknown as Record<string, unknown>)[field] = parsed
    } catch {}
  }
  return st
}
