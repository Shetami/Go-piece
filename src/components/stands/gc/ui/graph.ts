import type { Cell, GcWorld } from '../engine/types.ts'
import type { Color } from '../engine/types.ts'

/**
 * Укладка кучи в граф достижимости.
 *
 * Чистая функция: снимок мира на входе, координаты на выходе. Никакой физики и
 * никакой анимации — положение объекта определяется только его расстоянием от
 * корней и номером блока. Это важнее красоты: при шаге назад и перемотке
 * картинка обязана совпасть с той, что была, иначе следить за объектом глазами
 * невозможно.
 *
 * Что должно быть видно на картинке:
 *  — куда ведут корни и что от них достижимо;
 *  — как серый фронт уходит от корней вглубь;
 *  — недостижимые объекты отдельно: именно их соберут;
 *  — и главное, нарушение инварианта: чёрный объект, указывающий на белый.
 */

export const NODE_R = 9
const ROOT_W = 46
const ROOT_H = 22
const ROW_H = 62
/** Внутри одного уровня строки переносятся, если объектов слишком много. */
const SUBROW_H = 26
const PAD_X = 16
const PAD_Y = 14
export const MIN_GAP = 26
/** Больше этого числа объектов в графе не рисуем: получится каша, а не схема. */
export const MAX_NODES = 120

export interface GraphNode {
  id: number
  x: number
  y: number
  color: Color
  lost: boolean
  bornBlack: boolean
  owner: number
  /** Ссылок наружу и сколько из них ведут в живые блоки. */
  out: number
  /** Расстояние от корней; -1 — недостижим ни от одного корня. */
  depth: number
}

export interface GraphRoot {
  key: string
  label: string
  x: number
  y: number
  w: number
  h: number
  /** Стек горутины уже просмотрен — значит, он «чёрный». */
  scanned: boolean
  mut: number | null
  /** Сколько непустых слотов. */
  count: number
}

export interface GraphEdge {
  key: string
  x1: number
  y1: number
  x2: number
  y2: number
  from: number | null
  to: number
  /**
   * Чёрный объект указывает на белый во время разметки.
   * Так выглядит нарушение трёхцветного инварианта — то самое, ради чего
   * существует барьер записи.
   */
  violation: boolean
}

export interface HeapGraph {
  roots: GraphRoot[]
  nodes: GraphNode[]
  edges: GraphEdge[]
  width: number
  height: number
  /** Где проходит разделитель «достижимое / мусор», null — мусора нет. */
  garbageY: number | null
  /** Занятых блоков всего, показано и спрятано. */
  total: number
  shown: number
  hidden: number
  /** Недостижимых блоков всего (не только показанных). */
  garbage: number
  /** Нарушений инварианта на картинке. */
  violations: number
}

/**
 * Разложить уровень, стараясь поставить каждый узел под его родителями.
 *
 * `desired[i]` — куда узел тянется: под единственного родителя, под середину
 * нескольких или под свой корень. Дальше идёт обычное разведение: слева направо
 * раздвигаем до минимального зазора, справа налево вправляем то, что вылезло за
 * край. Если узлов столько, что в строку они не помещаются даже впритык,
 * возвращаемся к равномерной раскладке с переносом.
 *
 * Смысл ровно один: цепочка без ветвлений обязана идти строго вертикально.
 * Иначе по картинке не видно, что это одна цепочка.
 */
function tidy(desired: number[], width: number, top: number): { points: { x: number; y: number }[]; rows: number } | null {
  const n = desired.length
  if (n === 0) return { points: [], rows: 0 }
  const lo = PAD_X + NODE_R
  const hi = width - PAD_X - NODE_R
  if ((n - 1) * MIN_GAP > hi - lo) return null

  // Убираем обязательный зазор из задачи: после сдвига на i*MIN_GAP условие
  // «не ближе зазора» превращается в «не убывает», а это изотоническая
  // регрессия. Её точное решение даёт pool adjacent violators — и оно
  // минимизирует суммарный квадрат отклонения от желаемых мест.
  //
  // На практике это значит вот что: два потомка одного родителя расходятся
  // симметрично вокруг него, а не сдвигают вправо всех соседей по уровню.
  const shifted = desired.map((v, i) => v - i * MIN_GAP)
  const blocks: { sum: number; count: number; mean: number }[] = []
  for (const v of shifted) {
    let b = { sum: v, count: 1, mean: v }
    while (blocks.length > 0 && blocks[blocks.length - 1]!.mean > b.mean) {
      const prev = blocks.pop()!
      const sum = prev.sum + b.sum
      const count = prev.count + b.count
      b = { sum, count, mean: sum / count }
    }
    blocks.push(b)
  }
  const xs: number[] = []
  for (const b of blocks) for (let i = 0; i < b.count; i++) xs.push(b.mean + xs.length * MIN_GAP)

  // Вправить ряд в поле целиком — сначала сдвигом, потом, если не хватило, зажимом.
  const over = xs[n - 1]! - hi
  const under = lo - xs[0]!
  const shift = over > 0 ? -over : under > 0 ? under : 0
  for (let i = 0; i < n; i++) xs[i] = xs[i]! + shift
  if (xs[n - 1]! > hi) {
    xs[n - 1] = hi
    for (let i = n - 2; i >= 0; i--) xs[i] = Math.min(xs[i]!, xs[i + 1]! - MIN_GAP)
  }
  if (xs[0]! < lo) {
    xs[0] = lo
    for (let i = 1; i < n; i++) xs[i] = Math.max(xs[i]!, xs[i - 1]! + MIN_GAP)
  }
  return { points: xs.map((x) => ({ x, y: top })), rows: 1 }
}

/** Разложить ряд по ширине: ровно, по центру, с переносом на подстроки. */
function place(count: number, width: number, top: number): { points: { x: number; y: number }[]; rows: number } {
  if (count === 0) return { points: [], rows: 0 }
  const usable = width - PAD_X * 2
  const perRow = Math.max(1, Math.min(count, Math.floor(usable / MIN_GAP)))
  const points: { x: number; y: number }[] = []
  for (let i = 0; i < count; i++) {
    const row = Math.floor(i / perRow)
    const inRow = Math.min(perRow, count - row * perRow)
    const idx = i - row * perRow
    // Шаг считается по числу элементов именно этой строки — тогда неполная
    // последняя строка оказывается по центру, а не прижатой влево.
    const step = usable / inRow
    points.push({ x: PAD_X + step * (idx + 0.5), y: top + row * SUBROW_H })
  }
  return { points, rows: Math.ceil(count / perRow) }
}

/**
 * Подрезать отрезок так, чтобы он начинался и кончался на границах кружков,
 * а не в их центрах.
 *
 * Считать «низ одного, верх другого» нельзя: у соседей по горизонтали это даёт
 * диагональ, воткнутую в край кружка. Точка стыковки всегда лежит на прямой
 * между центрами — тогда ребро приходит в узел по-честному, с какой бы стороны
 * оно ни шло.
 */
function clip(
  x1: number,
  y1: number,
  x2: number,
  y2: number,
  r1: number,
  r2: number,
): { x1: number; y1: number; x2: number; y2: number } {
  const dx = x2 - x1
  const dy = y2 - y1
  const len = Math.hypot(dx, dy)
  if (len < 0.001) return { x1, y1, x2, y2 }
  const ux = dx / len
  const uy = dy / len
  return { x1: x1 + ux * r1, y1: y1 + uy * r1, x2: x2 - ux * r2, y2: y2 - uy * r2 }
}

/**
 * Построить граф по снимку мира.
 * `width` — ширина картинки в пользовательских единицах SVG.
 */
export function buildGraph(w: GcWorld, width = 1000): HeapGraph {
  const marking = w.phase === 'mark' || w.phase === 'stw-end'

  const alive = w.muts.filter((m) => m.state !== 'done')
  /** Корни: по одному узлу на держателя — стек горутины или набор глобальных. */
  const holders: { key: string; label: string; scanned: boolean; mut: number | null; targets: number[] }[] =
    alive.map((m) => ({
      key: `g${m.id}`,
      label: `G${m.id}`,
      // «Просмотрен» — состояние внутри цикла разметки. Вне цикла флаг ничего
      // не значит, и красить им корень значило бы показывать прошлогодний снег.
      scanned: marking && m.stackScanned,
      mut: m.id,
      targets: [...new Set(m.stack.filter((s): s is number => s !== null))],
    }))
  const globalTargets = [...new Set(w.globals.filter((g): g is number => g !== null))]
  if (globalTargets.length > 0 || w.globals.length > 0) {
    holders.push({
      key: 'globals',
      label: 'глоб.',
      // Глобальные корни закрашиваются в первую паузу и дальше всегда просмотрены.
      scanned: marking,
      mut: null,
      targets: globalTargets,
    })
  }

  // Расстояние от корней. Обходим в ширину — это ровно то, что делает разметка.
  const depth = new Map<number, number>()
  let front = [...new Set(holders.flatMap((h) => h.targets))].filter((id) => w.cells[id]?.used)
  let d = 0
  while (front.length > 0) {
    const next: number[] = []
    for (const id of front) {
      if (depth.has(id)) continue
      depth.set(id, d)
      for (const s of w.cells[id]!.slots) {
        if (s !== null && w.cells[s]?.used && !depth.has(s)) next.push(s)
      }
    }
    front = next
    d++
  }

  const used = w.cells.filter((c) => c.used)
  const garbageCells = used.filter((c) => !depth.has(c.id))
  const maxDepth = depth.size === 0 ? -1 : Math.max(...depth.values())

  // Если блоков слишком много, режем мусор: он одинаковый, а достижимое — нет.
  const budget = Math.max(0, MAX_NODES - depth.size)
  const shownGarbage = garbageCells.slice(0, budget)
  const hidden = garbageCells.length - shownGarbage.length

  /* ───── координаты ───── */

  const rootLayout = place(holders.length, width, PAD_Y + ROOT_H / 2)
  const rootPos = rootLayout.points
  const rootRows = (Math.max(1, rootLayout.rows) - 1) * SUBROW_H + ROOT_H
  const roots: GraphRoot[] = holders.map((h, i) => ({
    key: h.key,
    label: h.label,
    x: rootPos[i]!.x,
    y: rootPos[i]!.y,
    w: ROOT_W,
    h: ROOT_H,
    scanned: h.scanned,
    mut: h.mut,
    count: h.targets.length,
  }))

  // Кто на кого ссылается — нужно, чтобы класть потомков под их родителями.
  const parentsOf = new Map<number, number[]>()
  for (const c of used) {
    for (const s of c.slots) {
      if (s === null || !w.cells[s]?.used) continue
      const list = parentsOf.get(s)
      if (list) list.push(c.id)
      else parentsOf.set(s, [c.id])
    }
  }
  /** Куда тянется объект первого уровня — под середину всех держащих его корней. */
  const rootXs = new Map<number, number[]>()
  holders.forEach((h, i) => {
    for (const t of h.targets) {
      const list = rootXs.get(t)
      if (list) list.push(rootPos[i]!.x)
      else rootXs.set(t, [rootPos[i]!.x])
    }
  })
  const rootX = new Map<number, number>(
    [...rootXs].map(([id, xs]) => [id, xs.reduce((a, b) => a + b, 0) / xs.length]),
  )

  const pos = new Map<number, { x: number; y: number }>()
  let y = PAD_Y + ROOT_H / 2 + rootRows + ROW_H - ROOT_H
  for (let level = 0; level <= maxDepth; level++) {
    const ids = used.filter((c) => depth.get(c.id) === level).map((c) => c.id)
    if (ids.length === 0) continue
    /**
     * Куда объект тянется по горизонтали: под своих родителей с прошлого уровня.
     * Единственный потомок получает икс родителя один в один — так цепочка и
     * становится вертикальной.
     */
    const anchor = (id: number): number | null => {
      if (level === 0) return rootX.get(id) ?? null
      const xs = (parentsOf.get(id) ?? [])
        .filter((pid) => (depth.get(pid) ?? -1) < level)
        .map((pid) => pos.get(pid)?.x)
        .filter((x): x is number => x !== undefined)
      return xs.length === 0 ? null : xs.reduce((a, b) => a + b, 0) / xs.length
    }
    const want = new Map(ids.map((id) => [id, anchor(id)]))
    // Без родителей (такое бывает у перевешенных указателей) — в конец ряда.
    ids.sort((a, b) => (want.get(a) ?? Number.MAX_SAFE_INTEGER) - (want.get(b) ?? Number.MAX_SAFE_INTEGER) || a - b)

    const laid = tidy(
      ids.map((id) => want.get(id) ?? width / 2),
      width,
      y,
    )
    const { points, rows } = laid ?? place(ids.length, width, y)
    ids.forEach((id, i) => pos.set(id, points[i]!))
    y += ROW_H + (Math.max(1, rows) - 1) * SUBROW_H
  }

  const garbageY = shownGarbage.length > 0 || hidden > 0 ? y - ROW_H / 2 + 8 : null
  if (shownGarbage.length > 0) {
    const ids = shownGarbage.map((c) => c.id).sort((a, b) => a - b)
    const { points, rows } = place(ids.length, width, y + 14)
    ids.forEach((id, i) => pos.set(id, points[i]!))
    y += 14 + ROW_H + (rows - 1) * SUBROW_H
  } else if (hidden > 0) {
    y += 40
  }

  const node = (c: Cell): GraphNode => ({
    id: c.id,
    x: pos.get(c.id)!.x,
    y: pos.get(c.id)!.y,
    color: c.color,
    lost: c.lost,
    bornBlack: c.bornBlack,
    owner: c.owner,
    out: c.slots.filter((s) => s !== null && w.cells[s]?.used).length,
    depth: depth.get(c.id) ?? -1,
  })
  const nodes = used.filter((c) => pos.has(c.id)).map(node)

  /* ───── рёбра ───── */

  const edges: GraphEdge[] = []
  for (const [i, h] of holders.entries()) {
    for (const t of h.targets) {
      const p = pos.get(t)
      if (!p) continue
      edges.push({
        key: `${h.key}-${t}`,
        ...clip(rootPos[i]!.x, rootPos[i]!.y + ROOT_H / 2, p.x, p.y, 0, NODE_R),
        from: null,
        to: t,
        violation: false,
      })
    }
  }
  for (const c of used) {
    const from = pos.get(c.id)
    if (!from) continue
    for (const s of c.slots) {
      if (s === null) continue
      const target = w.cells[s]
      const to = pos.get(s)
      if (!target?.used || !to) continue
      // Инвариант: чёрный не имеет права указывать на белый, пока идёт разметка.
      const violation = marking && c.color === 'black' && target.color === 'white'
      edges.push({
        key: `${c.id}-${s}`,
        ...clip(from.x, from.y, to.x, to.y, NODE_R, NODE_R),
        from: c.id,
        to: s,
        violation,
      })
    }
  }

  // Высота считается по фактически поставленным узлам, а не по числу уровней:
  // иначе под последним рядом остаётся пустая полоса в целый уровень.
  const lowest = nodes.length > 0 ? Math.max(...nodes.map((n) => n.y)) : PAD_Y + ROOT_H
  return {
    roots,
    nodes,
    edges,
    width,
    height: Math.max(lowest + NODE_R + PAD_Y, garbageY !== null ? garbageY + 24 : 0, 120),
    garbageY,
    total: used.length,
    shown: nodes.length,
    hidden,
    garbage: garbageCells.length,
    violations: edges.filter((e) => e.violation).length,
  }
}
