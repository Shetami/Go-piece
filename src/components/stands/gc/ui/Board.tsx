import type { Cell, GcEvent, GcPhase, GcWorld, Mutator, SlotKind, SlotView } from '../engine/types.ts'
import { HeapGraph } from './HeapGraph.tsx'

/**
 * Схема мира на одном тике. Чистая проекция: получает снимок, ничего не хранит.
 *
 * У кучи два вида, и они отвечают на разные вопросы.
 *
 * «Сетка» — сколько занято: каждый квадрат это блок памяти, видно рост кучи,
 * работу подметальщика и дыры от освобождённых блоков.
 *
 * «Граф» — до чего можно дойти от корней. Это и есть определение живого объекта,
 * поэтому здесь видно то, чего в сетке не увидеть: фронт разметки, уходящий от
 * корней вглубь, недостижимые объекты отдельной группой и нарушение
 * трёхцветного инварианта — ссылку из чёрного в белое.
 */

export type HeapView = 'grid' | 'graph'

const WL_COLORS = 8

export interface Highlight {
  cells: Set<number>
  mut: Set<number>
}

export function highlightOf(e: GcEvent | null): Highlight {
  return { cells: new Set(e?.actors.cells ?? []), mut: new Set(e?.actors.mut ?? []) }
}

const PHASES: { key: GcPhase; label: string; hint: string }[] = [
  { key: 'off', label: 'вне цикла', hint: 'Сборки нет: программа выделяет память, куча растёт к порогу' },
  { key: 'stw-start', label: 'пауза', hint: 'Мир остановлен: включается барьер записи, закрашиваются глобальные корни' },
  { key: 'mark', label: 'разметка', hint: 'Идёт одновременно с программой: маркеры разбирают очередь серых' },
  { key: 'stw-end', label: 'пауза', hint: 'Мир остановлен: разметка закрывается, белое признаётся мусором' },
  { key: 'sweep', label: 'подметание', hint: 'Белые блоки освобождаются — без остановки мира' },
]

const SLOT_LABEL: Record<SlotKind, string> = {
  dedicated: 'маркер',
  fractional: 'дробный маркер',
  'idle-worker': 'простой → маркер',
  sweeper: 'подметальщик',
  mutator: 'программа',
  assist: 'помощь разметке',
  stw: 'пауза',
  free: 'свободен',
}

function PhaseStrip({ w }: { w: GcWorld }) {
  return (
    <div className="gc-phases" role="group" aria-label="Фаза цикла сборки">
      {PHASES.map((p) => (
        <span
          key={p.key}
          className={`gc-phase gc-phase-${p.key} ${w.phase === p.key ? 'is-now' : ''}`}
          title={p.hint}
        >
          {p.label}
        </span>
      ))}
      <span className="gc-phase-cycle">
        {w.cycle === 0 ? 'сборки ещё не было' : `цикл ${w.cycle}`}
        {w.barrierOn && <b title="Пока идёт разметка, каждая запись указателя проходит через барьер"> барьер вкл</b>}
      </span>
    </div>
  )
}

function CellSquare({ c, w, hl }: { c: Cell; w: GcWorld; hl: Highlight }) {
  const marking = w.phase !== 'off'
  const cls = c.lost ? 'is-lost' : !c.used ? 'is-free' : marking ? `is-${c.color}` : 'is-used'
  const title = c.lost
    ? `блок #${c.id}: сборщик решил, что он мусор, хотя до него можно дойти от корней.\n` +
      `${c.used ? 'Сейчас его освободят.' : 'Память уже отдана — программа будет писать по чужому адресу.'}`
    : !c.used
      ? `блок #${c.id}: свободен`
      : `блок #${c.id} · ${c.slots.filter((s) => s !== null).length} ссылок наружу\n` +
        `цвет: ${c.color}${c.bornBlack ? '\nвыделен чёрным во время разметки' : ''}`
  return (
    <span
      className={`gc-cell ${cls} ${hl.cells.has(c.id) ? 'is-hl' : ''}`}
      style={{ '--c': `var(--wl-${c.owner % WL_COLORS})` } as React.CSSProperties}
      title={title}
    />
  )
}

function PacerBar({ w }: { w: GcWorld }) {
  const used = w.cells.filter((c) => c.used).length
  const limit = w.config.memLimit === false ? null : w.config.memLimit
  const goal = Number.isFinite(w.goal) ? w.goal : w.config.heapCapacity
  const scale = Math.max(goal, used, w.trigger, limit ?? 0) * 1.08
  const pct = (v: number) => `${Math.min(100, (v / scale) * 100)}%`

  return (
    <div className="gc-pacer">
      <div className="gc-pacer-bar">
        <span className="gc-pacer-live" style={{ width: pct(w.heapMarked) }} title={`живое по прошлому циклу: ${w.heapMarked}`} />
        <span className="gc-pacer-heap" style={{ width: pct(used) }} title={`куча сейчас: ${used}`} />
        <span className="gc-pacer-mark gc-pacer-trigger" style={{ left: pct(w.trigger) }} title={`порог запуска: ${w.trigger}`} />
        <span className="gc-pacer-mark gc-pacer-goal" style={{ left: pct(goal) }} title={`цель: ${goal}`} />
        {limit !== null && (
          <span className="gc-pacer-mark gc-pacer-limit" style={{ left: pct(limit) }} title={`GOMEMLIMIT: ${limit}`} />
        )}
      </div>
      <div className="gc-pacer-legend">
        <span>
          куча <b>{used}</b>
        </span>
        <span>
          живое <b>{w.heapMarked}</b>
        </span>
        <span className="is-trigger">
          порог <b>{w.trigger}</b>
        </span>
        <span className="is-goal">
          цель <b>{Number.isFinite(w.goal) ? w.goal : '∞'}</b>
        </span>
        {limit !== null && (
          <span className="is-limit">
            лимит <b>{limit}</b>
          </span>
        )}
        {used > goal && <span className="is-over">перелёт на {used - goal}</span>}
      </div>
    </div>
  )
}

function MutChip({ m, w, hl }: { m: Mutator; w: GcWorld; hl: Highlight }) {
  const debt = m.assistDebt >= 1 ? Math.round(m.assistDebt) : 0
  return (
    <span
      className={`gc-mut state-${m.state} ${hl.mut.has(m.id) ? 'is-hl' : ''}`}
      style={{ '--c': `var(--wl-${m.workload % WL_COLORS})` } as React.CSSProperties}
      title={
        `G${m.id} · ${m.name} · ${m.state}\n` +
        `своей работы ${m.cpuTicks} тиков, помощи ${m.assistTicks}, в паузах ${m.stwTicks}, ждала процессор ${m.waitTicks}\n` +
        `выделила ${m.allocated} блоков, корней на стеке ${m.stack.filter((s) => s !== null).length}` +
        `${w.phase === 'mark' ? `\nстек ${m.stackScanned ? 'просмотрен (чёрный)' : 'ещё не просмотрен (серый)'}` : ''}`
      }
    >
      G{m.id}
      {debt > 0 && <small title="Долг помощи в разметке">долг {debt}</small>}
      {w.phase === 'mark' && !m.stackScanned && <i title="Стек ещё не просмотрен">?</i>}
    </span>
  )
}

export function Board({ world: w, highlight: hl, view, onView }: {
  world: GcWorld
  highlight: Highlight
  view: HeapView
  onView: (v: HeapView) => void
}) {
  // На нулевом тике процессоры ещё ничего не делали — показываем их пустыми,
  // чтобы ряд карточек не появлялся из ниоткуда после первого шага.
  const slots: SlotView[] =
    w.slots.length > 0
      ? w.slots
      : Array.from({ length: w.config.gomaxprocs }, (_, p) => ({ p, kind: 'free' as SlotKind }))
  const used = w.cells.filter((c) => c.used).length
  const live = w.cells.filter((c) => c.used && c.color === 'black').length
  const alive = w.muts.filter((m) => m.state !== 'done')
  const scanPct = w.scanWorkTarget === 0 ? 0 : Math.min(100, (w.scanWorkDone / w.scanWorkTarget) * 100)

  return (
    <div className="board gc-board">
      <PhaseStrip w={w} />

      <div className="gc-view">
        <span className="board-label">куча · {used} из {w.config.heapCapacity}</span>
        <span className="gc-view-tabs" role="group" aria-label="Вид кучи">
          <button
            type="button"
            className={view === 'graph' ? 'is-now' : ''}
            onClick={() => onView('graph')}
            title="Граф достижимости: кто на кого ссылается и до чего можно дойти от корней"
          >
            граф
          </button>
          <button
            type="button"
            className={view === 'grid' ? 'is-now' : ''}
            onClick={() => onView('grid')}
            title="Сетка блоков: сколько памяти занято и как работает подметальщик"
          >
            сетка
          </button>
        </span>
      </div>

      {view === 'graph' ? (
        <HeapGraph world={w} highlight={hl} />
      ) : (
        <div className="gc-heap" role="img" aria-label={`Куча: ${used} блоков занято из ${w.config.heapCapacity}`}>
          {w.cells.map((c) => (
            <CellSquare key={c.id} c={c} w={w} hl={hl} />
          ))}
        </div>
      )}

      <div className="gc-legend">
        <span className="lg lg-white">белый — разметка не нашла</span>
        <span className="lg lg-grey">серый — нашла, но не просмотрела</span>
        <span className="lg lg-black">чёрный — просмотрен, живой</span>
        <span className="lg lg-used">вне цикла — просто занятый блок</span>
        {view === 'graph' ? (
          <>
            <span className="lg lg-root">корень — стек горутины или глобальные</span>
            <span className="lg lg-violation">ссылка из чёрного в белое — инвариант нарушен</span>
          </>
        ) : (
          <span className="lg lg-free">пусто — свободен</span>
        )}
        <span className="lg lg-lost">красный — освобождён живым</span>
      </div>

      <PacerBar w={w} />

      <div className="gc-slots">
        {slots.map((s) => {
          const m = s.mut === undefined ? undefined : w.muts[s.mut - 1]
          return (
            <span key={s.p} className={`gc-slot gc-slot-${s.kind}`} title={SLOT_LABEL[s.kind]}>
              <b>P{s.p}</b>
              <span>
                {s.kind === 'mutator' || s.kind === 'assist'
                  ? `G${s.mut} ${s.kind === 'assist' ? 'помогает' : (m?.name ?? '')}`
                  : SLOT_LABEL[s.kind]}
              </span>
              {s.work !== undefined && s.work > 0 && <i>{s.work}</i>}
            </span>
          )
        })}
      </div>

      <div className="board-zones">
        <div className="zone">
          <span className="board-label">горутины · {alive.length}</span>
          <span className="queue">
            {alive.slice(0, 24).map((m) => (
              <MutChip key={m.id} m={m} w={w} hl={hl} />
            ))}
            {alive.length > 24 && <span className="board-more">+{alive.length - 24}</span>}
          </span>
        </div>

        <div className="zone">
          <span className="board-label">фронт разметки</span>
          <span className="gc-front">
            <span title="Объекты, которые найдены, но ещё не просмотрены">
              серых <b>{w.greyq.length}</b>
            </span>
            <span title="Стеки горутин, которые ещё предстоит просмотреть">
              стеков <b>{alive.filter((m) => !m.stackScanned).length}</b>
            </span>
            <span title="Оценка работы цикла и сколько из неё сделано">
              просмотрено <b>{w.scanWorkDone}</b>
              {w.scanWorkTarget > 0 && ` из ~${w.scanWorkTarget}`}
            </span>
            <span className="gc-progress" aria-hidden="true">
              <span style={{ width: `${scanPct}%` }} />
            </span>
          </span>
        </div>

        <div className="zone">
          <span className="board-label">подметание · {w.sweepq.length}</span>
          <span className="gc-front">
            <span title="Блоки, признанные мусором и ждущие освобождения">
              ждут <b>{w.sweepq.length}</b>
            </span>
            <span title="Сколько блоков сборщик освободил за весь прогон">
              освобождено всего <b>{w.stats.freed}</b>
            </span>
          </span>
        </div>

        <div className="zone zone-stats">
          <span>
            глобальных корней <b>{w.globals.filter((g) => g !== null).length}</b>
          </span>
          <span>
            чёрных сейчас <b>{live}</b>
          </span>
          <span title="Сколько блоков разметки обязан оплатить тот, кто выделит один блок">
            ставка помощи <b>×{w.assistRatio.toFixed(1)}</b>
          </span>
          <span title="Скользящее среднее: блоков за тик">
            скорость аллокации <b>{w.allocRate.toFixed(1)}</b>
          </span>
        </div>
      </div>
    </div>
  )
}
