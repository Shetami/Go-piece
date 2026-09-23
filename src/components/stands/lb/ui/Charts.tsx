import { useEffect, useMemo, useRef, useState } from 'react'
import type { LbLog } from '../engine/types.ts'
import { percentile } from '../engine/world.ts'

/**
 * Графики прогона: общая ось тиков, данные нарисованы до курсора.
 * Масштаб по вертикали берётся по всему прогону, чтобы при проигрывании
 * ось не прыгала. У каждой реплики свой цвет из категориальной палитры —
 * в том же порядке, что на карточках реплик. На остальных графиках синий —
 * ответы и обычная задержка, оранжевый — хвост, красный — ошибки.
 */

export type SeriesColor = 'work' | 'wait' | 'in' | 'waste' | `r${number}`

export interface Series {
  label: string
  color: SeriesColor
  values: (number | null)[]
  /** Закрасить область под линией; для stacked — слой стопки. */
  area?: boolean
}

interface RefLine {
  y: number
  label: string
}

/** Окно для перцентилей задержки, тиков. */
export const LAT_WINDOW = 20
/** Окно для потоков, тиков. */
export const FLOW_WINDOW = 10

export interface ChartData {
  /** Незавершённые запросы по репликам; null — реплики ещё нет. */
  load: (number | null)[][]
  p50: (number | null)[]
  p99: (number | null)[]
  arrived: number[]
  ok: number[]
  errors: number[]
  anyErrors: boolean
}

/** Всё, что рисуют графики, считается один раз на прогон. */
export function chartData(log: LbLog): ChartData {
  const n = log.series.length
  const reps = Math.max(...log.series.map((s) => s.load.length))
  const byTick: number[][] = Array.from({ length: n }, () => [])
  for (const c of log.done) if (c.outcome === 'ok') byTick[c.t]?.push(c.lat)
  const p50: (number | null)[] = []
  const p99: (number | null)[] = []
  for (let t = 0; t < n; t++) {
    const win = byTick.slice(Math.max(0, t - LAT_WINDOW + 1), t + 1).flat()
    p50.push(win.length ? percentile(win, 50) : null)
    p99.push(win.length ? percentile(win, 99) : null)
  }
  const sum = (pick: (i: number) => number) =>
    Array.from({ length: n }, (_, t) => {
      let s = 0
      for (let i = Math.max(1, t - FLOW_WINDOW + 1); i <= t; i++) s += pick(i)
      return s
    })
  return {
    load: Array.from({ length: reps }, (_, r) => log.series.map((s) => s.load[r] ?? null)),
    p50,
    p99,
    arrived: sum((i) => log.series[i]!.arrived),
    ok: sum((i) => log.series[i]!.ok),
    errors: sum((i) => log.series[i]!.errors + log.series[i]!.timeouts),
    anyErrors: log.series.some((s) => s.errors + s.timeouts > 0),
  }
}

function useWidth<T extends HTMLElement>(): [React.RefObject<T | null>, number] {
  const ref = useRef<T>(null)
  const [w, setW] = useState(640)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const ro = new ResizeObserver(([e]) => e && setW(Math.max(260, Math.round(e.contentRect.width))))
    ro.observe(el)
    return () => ro.disconnect()
  }, [])
  return [ref, w]
}

/** Круглая верхняя граница оси: 4, 5, 10, 20, 25, 50… */
function niceMax(v: number): number {
  if (v <= 4) return 4
  const p = 10 ** Math.floor(Math.log10(v))
  for (const m of [1, 2, 2.5, 5, 10]) if (m * p >= v) return m * p
  return 10 * p
}

const H = 118
const M = { l: 34, r: 104, t: 8, b: 18 }

export function Chart({ title, note, series, stacked = false, last, cursor, refs = [], onSeek, fmt = String }: {
  title: string
  note?: string
  series: Series[]
  stacked?: boolean
  last: number
  cursor: number
  refs?: RefLine[]
  onSeek: (t: number) => void
  fmt?: (v: number) => string
}) {
  const [box, width] = useWidth<HTMLDivElement>()
  const [hover, setHover] = useState<number | null>(null)

  // В стопке каждый слой лежит на сумме предыдущих.
  const tops = useMemo(() => {
    if (!stacked) return series.map((s) => s.values)
    const acc: number[] = Array.from({ length: last + 1 }, () => 0)
    return series.map((s) =>
      s.values.map((v, t) => {
        acc[t]! += v ?? 0
        return acc[t]!
      }),
    )
  }, [series, stacked, last])

  const yMax = niceMax(Math.max(1, ...tops.flatMap((v) => v.map((x) => x ?? 0)), ...refs.map((r) => r.y)))
  const iw = width - M.l - M.r
  const ih = H - M.t - M.b
  const x = (t: number) => M.l + (t / Math.max(1, last)) * iw
  const y = (v: number) => M.t + ih - (v / yMax) * ih

  const linePath = (vals: (number | null)[]) => {
    let d = ''
    let pen = false
    for (let t = 0; t <= cursor; t++) {
      const v = vals[t]
      if (v === null || v === undefined) {
        pen = false
        continue
      }
      d += `${pen ? 'L' : 'M'}${x(t).toFixed(1)} ${y(v).toFixed(1)}`
      pen = true
    }
    return d
  }
  const areaPath = (top: (number | null)[], bottom: (number | null)[] | null) => {
    if (cursor < 1) return ''
    let d = `M${x(0)} ${y(bottom?.[0] ?? 0)}`
    for (let t = 0; t <= cursor; t++) d += `L${x(t).toFixed(1)} ${y(top[t] ?? 0).toFixed(1)}`
    for (let t = cursor; t >= 0; t--) d += `L${x(t).toFixed(1)} ${y(bottom?.[t] ?? 0).toFixed(1)}`
    return `${d}Z`
  }

  // Подписи у концов линий. Если две подписи налезают, вторая уходит: её несут легенда и подсказка.
  const endLabels = useMemo(() => {
    const out: { i: number; y: number; text: string }[] = []
    series.forEach((s, i) => {
      const v = s.values[cursor]
      if (v === null || v === undefined) return
      out.push({ i, y: y(tops[i]![cursor] ?? v), text: `${s.label} ${fmt(v)}` })
    })
    out.sort((a, b) => a.y - b.y)
    return out.filter((l, k) => k === 0 || l.y - out[k - 1]!.y >= 12)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [series, tops, cursor, yMax, width])

  const tickAt = (clientX: number, el: SVGSVGElement) => {
    const r = el.getBoundingClientRect()
    const t = Math.round(((clientX - r.left - M.l) / iw) * last)
    return Math.max(0, Math.min(last, t))
  }
  const hv = hover !== null && hover <= cursor ? hover : null

  return (
    <div className="ld-chart" ref={box}>
      <div className="ld-chart-head">
        <span className="ld-chart-title">{title}</span>
        {note && <span className="ld-chart-note">{note}</span>}
        {series.length > 1 && (
          <span className="ld-legend">
            {series.map((s) => (
              <span key={s.label}>
                <i className={`ld-key ${s.area ? 'is-area' : ''} ld-c-${s.color}`} />
                {s.label}
              </span>
            ))}
          </span>
        )}
      </div>
      <svg
        width={width}
        height={H}
        viewBox={`0 0 ${width} ${H}`}
        role="img"
        aria-label={`${title}: ${series.map((s) => `${s.label} ${s.values[cursor] ?? '—'}`).join(', ')} на тике ${cursor}`}
        onPointerMove={(e) => setHover(tickAt(e.clientX, e.currentTarget))}
        onPointerLeave={() => setHover(null)}
        onClick={(e) => onSeek(tickAt(e.clientX, e.currentTarget))}
      >
        {[0, yMax / 2, yMax].map((v) => (
          <g key={v}>
            <line x1={M.l} x2={M.l + iw} y1={y(v)} y2={y(v)} className="ld-grid" />
            <text x={M.l - 6} y={y(v) + 3.5} textAnchor="end" className="ld-axis">
              {fmt(v)}
            </text>
          </g>
        ))}
        {[0, Math.round(last / 2), last].map((t) => (
          <text key={t} x={x(t)} y={H - 4} textAnchor={t === 0 ? 'start' : t === last ? 'end' : 'middle'} className="ld-axis">
            {t}
          </text>
        ))}
        {refs.map((r) => (
          <g key={r.label}>
            <line x1={M.l} x2={M.l + iw} y1={y(r.y)} y2={y(r.y)} className="ld-ref" />
            <text x={M.l + 4} y={y(r.y) - 3} className="ld-axis">
              {r.label}
            </text>
          </g>
        ))}
        {series.map((s, i) =>
          s.area ? <path key={`a${i}`} d={areaPath(tops[i]!, stacked && i > 0 ? tops[i - 1]! : null)} className={`ld-area ld-c-${s.color}`} /> : null,
        )}
        {series.map((s, i) => (
          <path key={`l${i}`} d={linePath(tops[i]!)} className={`ld-line ld-c-${s.color}`} />
        ))}
        <line x1={x(cursor)} x2={x(cursor)} y1={M.t} y2={M.t + ih} className="ld-now" />
        {endLabels.map((l) => (
          <g key={l.i}>
            <line x1={x(cursor) + 4} x2={x(cursor) + 12} y1={l.y} y2={l.y} className={`ld-line ld-c-${series[l.i]!.color}`} />
            <text x={x(cursor) + 15} y={l.y + 3.5} className="ld-end">
              {l.text}
            </text>
          </g>
        ))}
        {hv !== null && <line x1={x(hv)} x2={x(hv)} y1={M.t} y2={M.t + ih} className="ld-cross" />}
      </svg>
      {hv !== null && (
        <div className="ld-tip" style={{ left: Math.min(x(hv) + 10, width - 150) }}>
          <span className="ld-tip-tick">тик {hv}</span>
          {series.map((s) => (
            <span key={s.label} className="ld-tip-row">
              <i className={`ld-key ld-c-${s.color}`} />
              <b>{s.values[hv] === null || s.values[hv] === undefined ? '—' : fmt(s.values[hv]!)}</b> {s.label}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
