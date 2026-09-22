import type { RowSpec, TxnConfig, TxnScenario, TxnSpec } from '../engine/types.ts'
import { resolveConfig } from '../engine/world.ts'
import { txnScenarioById } from '../engine/scenarios.ts'

/**
 * Сценарий ↔ строка для адреса.
 *
 * Хранится только то, что отличается от пресета: `s=lost-update&isolation=repeatable-read`.
 * Изменённые сессии и строки таблицы не раскладываются на параметры, а едут
 * одним base64-JSON в `t` и `r`.
 */

export interface StandState {
  base: string
  config: TxnConfig
  /** null — сессии и строки пресета без изменений. */
  sessions: TxnSpec[] | null
  rows: RowSpec[] | null
}

const CONFIG_KEYS = Object.keys(resolveConfig(undefined)) as (keyof TxnConfig)[]

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

export function initialState(base: TxnScenario): StandState {
  return { base: base.id, config: resolveConfig(base.config), sessions: null, rows: null }
}

export function toScenario(st: StandState): TxnScenario {
  const base = txnScenarioById(st.base)
  if (!base) throw new Error(`нет сценария ${st.base}`)
  const rows = st.rows ?? base.rows
  // Правило «всё подтверждённое на месте» имеет смысл, только пока его ключи есть в таблице.
  const keys = new Set(rows.map((r) => r.key))
  const invariant = base.invariant && base.invariant.keys.every((k) => keys.has(k)) ? base.invariant : undefined
  return { ...base, config: st.config, sessions: st.sessions ?? base.sessions, rows, invariant }
}

export function encodeState(st: StandState): string {
  const base = txnScenarioById(st.base)
  const baseCfg = resolveConfig(base?.config)
  const q = new URLSearchParams()
  q.set('s', st.base)
  for (const k of CONFIG_KEYS) {
    if (st.config[k] !== baseCfg[k]) q.set(k, String(st.config[k]))
  }
  if (st.sessions) q.set('t', b64encode(JSON.stringify(st.sessions)))
  if (st.rows) q.set('r', b64encode(JSON.stringify(st.rows)))
  return q.toString()
}

/** Тумблеры умеют быть выключенными, поэтому «false» разбирается раньше чисел. */
function parseValue(raw: string): unknown {
  if (raw === 'false') return false
  if (raw === 'true') return true
  const n = Number(raw)
  return Number.isFinite(n) ? n : raw
}

export function decodeState(query: string, fallback: TxnScenario): StandState {
  const q = new URLSearchParams(query.replace(/^[#?]/, ''))
  const base = txnScenarioById(q.get('s') ?? '') ?? fallback
  const st = initialState(base)
  const cfg: Record<string, unknown> = { ...st.config }
  for (const k of CONFIG_KEYS) {
    const raw = q.get(k)
    if (raw !== null) cfg[k] = parseValue(raw)
  }
  st.config = cfg as unknown as TxnConfig
  for (const [param, field] of [['t', 'sessions'], ['r', 'rows']] as const) {
    const raw = q.get(param)
    if (!raw) continue
    try {
      const parsed = JSON.parse(b64decode(raw)) as unknown
      if (Array.isArray(parsed)) (st as unknown as Record<string, unknown>)[field] = parsed
    } catch {}
  }
  return st
}
