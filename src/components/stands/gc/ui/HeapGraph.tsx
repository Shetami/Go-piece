import { useMemo } from 'react'
import type { GcWorld } from '../engine/types.ts'
import { NODE_R, buildGraph } from './graph.ts'
import type { Highlight } from './Board.tsx'

/**
 * Куча как граф достижимости.
 *
 * Сетка блоков отвечает на вопрос «сколько занято». Этот вид отвечает на другой,
 * ради которого сборщик и существует: «до чего можно дойти от корней». Здесь
 * видно, как серый фронт уходит вглубь, почему недостижимое — мусор независимо
 * от того, сколько на него внутренних ссылок, и как выглядит нарушение
 * трёхцветного инварианта.
 */

const WL_COLORS = 8
const VIEW_W = 1000

const COLOR_RU: Record<string, string> = { white: 'белый', grey: 'серый', black: 'чёрный' }

export function HeapGraph({ world: w, highlight: hl }: { world: GcWorld; highlight: Highlight }) {
  const g = useMemo(() => buildGraph(w, VIEW_W), [w])
  const marking = w.phase === 'mark' || w.phase === 'stw-end'

  return (
    <div className="gc-graph">
      <svg
        viewBox={`0 0 ${VIEW_W} ${g.height}`}
        className="gc-graph-svg"
        role="img"
        aria-label={`Граф достижимости: ${g.total} занятых блоков, из них ${g.garbage} недостижимых`}
      >
        {g.nodes.length === 0 && g.roots.length === 0 && (
          <text className="gc-graph-empty" x={VIEW_W / 2} y={54} textAnchor="middle">
            куча пуста — программа ещё ничего не выделила
          </text>
        )}

        {g.garbageY !== null && (
          <g className="gc-graph-divider">
            <line x1={8} y1={g.garbageY} x2={VIEW_W - 8} y2={g.garbageY} />
            <text x={VIEW_W - 12} y={g.garbageY - 6} textAnchor="end">
              ниже — недостижимое: {g.garbage} {g.garbage === 1 ? 'блок' : g.garbage < 5 ? 'блока' : 'блоков'}
              {g.hidden > 0 && ` (показано ${g.garbage - g.hidden})`}
            </text>
          </g>
        )}

        {/* Рёбра рисуются первыми — узлы должны лежать поверх них. */}
        <g className="gc-graph-edges">
          {g.edges.map((e) => {
            const mid = (e.y1 + e.y2) / 2
            const lit = e.from !== null && (hl.cells.has(e.from) || hl.cells.has(e.to))
            return (
              <path
                key={e.key}
                className={`gc-edge ${e.violation ? 'is-violation' : ''} ${e.from === null ? 'is-root' : ''} ${lit ? 'is-hl' : ''}`}
                d={`M ${e.x1} ${e.y1} C ${e.x1} ${mid}, ${e.x2} ${mid}, ${e.x2} ${e.y2}`}
              >
                {e.violation && (
                  <title>
                    Чёрный блок #{e.from} указывает на белый #{e.to}. Инвариант нарушен: разметка сюда уже не
                    вернётся, и живой объект попадёт под подметание.
                  </title>
                )}
              </path>
            )
          })}
        </g>

        <g className="gc-graph-roots">
          {g.roots.map((r) => (
            <g key={r.key} className={`gc-root ${r.scanned ? 'is-scanned' : ''} ${r.mut !== null && hl.mut.has(r.mut) ? 'is-hl' : ''}`}>
              <rect x={r.x - r.w / 2} y={r.y - r.h / 2} width={r.w} height={r.h} rx={5} />
              <text x={r.x} y={r.y + 4} textAnchor="middle">
                {r.label}
              </text>
              <title>
                {r.mut === null
                  ? `Глобальные переменные: ${r.count} непустых указателей. Закрашиваются в первую паузу цикла.`
                  : `Стек G${r.mut}: ${r.count} непустых указателей.\n` +
                    (marking
                      ? r.scanned
                        ? 'Стек уже просмотрен — он «чёрный», барьер Дейкстры для него больше не нужен.'
                        : 'Стек ещё не просмотрен — разметка до него не дошла.'
                      : 'Цикл не идёт.')}
              </title>
            </g>
          ))}
        </g>

        <g className="gc-graph-nodes">
          {g.nodes.map((n) => (
            <g
              key={n.id}
              className={`gc-node ${n.lost ? 'is-lost' : marking ? `is-${n.color}` : 'is-used'} ${
                n.depth < 0 ? 'is-garbage' : ''
              } ${hl.cells.has(n.id) ? 'is-hl' : ''}`}
              style={{ '--c': `var(--wl-${n.owner % WL_COLORS})` } as React.CSSProperties}
            >
              <circle cx={n.x} cy={n.y} r={NODE_R} />
              {n.bornBlack && <circle className="gc-node-mark" cx={n.x + NODE_R - 2} cy={n.y - NODE_R + 2} r={2.5} />}
              <title>
                {`блок #${n.id} · ссылок наружу ${n.out}\n` +
                  (n.depth < 0
                    ? 'недостижим от корней — это и есть мусор'
                    : `достижим от корней, ${n.depth === 0 ? 'прямо со стека или из глобальной' : `через ${n.depth} ${n.depth === 1 ? 'ссылку' : 'ссылки'}`}`) +
                  (marking ? `\nцвет: ${COLOR_RU[n.color]}` : '') +
                  (n.bornBlack ? '\nвыделен чёрным во время разметки' : '') +
                  (n.lost ? '\nсборщик освободил его живым — это баг, которого не бывает при исправном барьере' : '')}
              </title>
            </g>
          ))}
        </g>
      </svg>

      <div className="gc-graph-note">
        {g.violations > 0 ? (
          <span className="gc-graph-alarm">
            инвариант нарушен: {g.violations}{' '}
            {g.violations === 1 ? 'ссылка ведёт' : 'ссылок ведут'} из чёрного в белое — эти объекты разметка уже не
            найдёт
          </span>
        ) : marking ? (
          <span>инвариант цел: ни одна ссылка не ведёт из чёрного в белое</span>
        ) : (
          <span>цикл не идёт — цвет означает просто «занято»</span>
        )}
        {g.hidden > 0 && <span className="gc-graph-hidden">скрыто {g.hidden} недостижимых блоков — их слишком много для схемы</span>}
      </div>
    </div>
  )
}
