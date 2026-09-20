import type { Goroutine, MemEvent, MemWorld, Page, Span } from '../engine/types.ts'
import { PAGE, SIZE_CLASSES } from '../engine/types.ts'

/**
 * Схема мира на одном тике. Чистая проекция: получает снимок, ничего не хранит.
 *
 * Главных элементов два: путь аллокации сверху — по нему видно, докуда дошёл
 * запрос памяти в этом тике, — и сетка страниц, где видно, чем занята куча.
 */

const WL_COLORS = 8

export interface Highlight {
  g: Set<number>
  pages: Set<number>
  spans: Set<number>
}

export function highlightOf(e: MemEvent | null): Highlight {
  return {
    g: new Set(e?.actors.g ?? []),
    pages: new Set(e?.actors.pages ?? []),
    spans: new Set(e?.actors.span ?? []),
  }
}

const bytes = (b: number) => (b >= 1024 ? `${Math.round((b / 1024) * 10) / 10} КБ` : `${b} Б`)

/** Ступени пути аллокации — от самой дешёвой к самой дорогой. */
const STAGES = [
  { key: 'stack', label: 'стек', hint: 'Объект не убегает из функции — он просто кадр на стеке, куча его не видит' },
  { key: 'tiny', label: 'блок 16 Б', hint: 'Мелочь без указателей складывается вплотную в общий блок' },
  { key: 'fast', label: 'кэш P', hint: 'Свободный слот в спане своего процессора — без единой блокировки' },
  { key: 'refill', label: 'mcentral', hint: 'Кэш пуст: за новым спаном идём в общий список под блокировкой' },
  { key: 'heap', label: 'mheap', hint: 'Свободных спанов нет — куча нарезает новый из страниц' },
  { key: 'os', label: 'ОС', hint: 'Своей памяти не хватило — системный вызов за новыми страницами' },
] as const

function PathStrip({ w, events }: { w: MemWorld; events: MemEvent[] }) {
  const active = new Set<string>()
  for (const s of w.slots) {
    if (s.path === 'stack') active.add('stack')
    if (s.path === 'tiny') active.add('tiny')
    if (s.path === 'fast') active.add('fast')
    if (s.path === 'refill') active.add('refill')
    if (s.path === 'large') active.add('heap')
  }
  for (const e of events) {
    if (e.type === 'cache.refill' || e.type === 'central.contended') active.add('refill')
    if (e.type === 'heap.grow' || e.type === 'central.empty' || e.type === 'alloc.large') active.add('heap')
    if (e.type === 'os.map') active.add('os')
    if (e.type === 'alloc.tiny') active.add('tiny')
    if (e.type === 'alloc.stack') active.add('stack')
  }
  const blocked = events.some((e) => e.type === 'central.contended')

  return (
    <div className="mem-path" role="group" aria-label="Путь аллокации">
      {STAGES.map((s, i) => (
        <span key={s.key} className={`mem-stage ${active.has(s.key) ? 'is-now' : ''}`} title={s.hint}>
          {s.label}
          {i < STAGES.length - 1 && <i aria-hidden="true">→</i>}
        </span>
      ))}
      {blocked && <span className="mem-stage-wait">кто-то ждёт блокировку</span>}
    </div>
  )
}

function HeapGrid({ w, hl, spanByPage }: { w: MemWorld; hl: Highlight; spanByPage: Map<number, Span> }) {
  return (
    <div className="mem-heap" role="img" aria-label="Страницы кучи">
      {w.pages.map((page: Page) => {
        const span = spanByPage.get(page.id)
        const cls =
          page.kind === 'span'
            ? 'is-span'
            : page.kind === 'large'
              ? 'is-large'
              : page.kind === 'stack'
                ? 'is-stack'
                : page.kind === 'returned'
                  ? 'is-returned'
                  : 'is-free'
        const title =
          page.kind === 'span' && span
            ? `страница ${page.id}: спан класса ${bytes(span.objSize)} — занято ${span.used} из ${span.slots}`
            : page.kind === 'large'
              ? `страница ${page.id}: часть большого объекта`
              : page.kind === 'stack'
                ? `страница ${page.id}: стеки горутин`
                : page.kind === 'returned'
                  ? `страница ${page.id}: возвращена операционной системе`
                  : `страница ${page.id}: свободна, но память ещё за процессом`
        const fill = span ? Math.round((span.used / Math.max(1, span.slots)) * 100) : 100
        return (
          <span
            key={page.id}
            className={`mem-page ${cls} ${hl.pages.has(page.id) ? 'is-hl' : ''}`}
            style={{ '--c': span ? `var(--wl-${span.sizeClass % WL_COLORS})` : undefined } as React.CSSProperties}
            title={title}
          >
            {page.kind === 'span' && <i style={{ height: `${fill}%` }} aria-hidden="true" />}
          </span>
        )
      })}
    </div>
  )
}

function CacheCard({ w, p, hl }: { w: MemWorld; p: number; hl: Highlight }) {
  const cache = w.mcaches[p]!
  const slot = w.slots.find((s) => s.p === p)
  const g = slot?.g ? w.gs[slot.g - 1] : undefined
  const spans = Object.entries(cache.spans)
    .map(([cls, id]) => ({ cls: Number(cls), span: w.spans[id - 1] }))
    .filter((x) => x.span && x.span.owner.kind === 'mcache' && x.span.owner.p === p)

  return (
    <section className={`mem-cache ${slot?.path ? `path-${slot.path}` : ''}`}>
      <header>
        <strong>P{p}</strong>
        <span>{g ? `G${g.id} ${g.name}` : 'простаивает'}</span>
      </header>
      <div className="mem-cache-spans">
        {spans.length === 0 ? (
          <span className="board-empty">кэш пуст</span>
        ) : (
          spans.map(({ cls, span }) => (
            <span
              key={cls}
              className={`mem-span ${hl.spans.has(span!.id) ? 'is-hl' : ''}`}
              style={{ '--c': `var(--wl-${cls % WL_COLORS})` } as React.CSSProperties}
              title={`класс ${bytes(SIZE_CLASSES[cls]!)}: занято ${span!.used} из ${span!.slots}`}
            >
              {bytes(SIZE_CLASSES[cls]!)}
              <i style={{ width: `${Math.round((span!.used / Math.max(1, span!.slots)) * 100)}%` }} />
            </span>
          ))
        )}
      </div>
      <div className="mem-tiny" title="Общий блок на 16 байт: сколько в нём уже лежит мелких объектов">
        <span className="board-label">блок 16 Б</span>
        <span className="mem-tiny-bar">
          <i style={{ width: `${Math.round((cache.tinyOffset / 16) * 100)}%` }} />
        </span>
        <b>{cache.tinyObjects > 0 ? `${cache.tinyObjects} об.` : '—'}</b>
      </div>
    </section>
  )
}

function GoroutineChip({ g, hl }: { g: Goroutine; hl: Highlight }) {
  const fill = Math.min(100, Math.round((g.stackUsed / g.stackSize) * 100))
  return (
    <span
      className={`mem-g state-${g.state} ${hl.g.has(g.id) ? 'is-hl' : ''}`}
      style={{ '--c': `var(--wl-${g.workload % WL_COLORS})` } as React.CSSProperties}
      title={
        `G${g.id} · ${g.name}\n` +
        `стек ${bytes(g.stackUsed)} из ${bytes(g.stackSize)}, ростов ${g.stackGrows}\n` +
        `выделила ${bytes(g.allocBytes)}, простояла ${g.stallTicks} тиков`
      }
    >
      G{g.id}
      <span className="mem-stack-bar">
        <i style={{ width: `${fill}%` }} />
      </span>
      <small>{bytes(g.stackSize)}</small>
      {g.state === 'growing' && <b title="Стек переезжает — горутина в это время не работает">переезд</b>}
    </span>
  )
}

export function Board({ world: w, events, highlight: hl }: { world: MemWorld; events: MemEvent[]; highlight: Highlight }) {
  const spanByPage = new Map<number, Span>()
  for (const s of w.spans) {
    if (s.owner.kind === 'free') continue
    for (const pg of s.pages) spanByPage.set(pg, s)
  }

  const used = w.pages.filter((p) => p.kind === 'span' || p.kind === 'large' || p.kind === 'stack').length
  const free = w.pages.filter((p) => p.kind === 'free').length
  const alive = w.gs.filter((g) => g.state !== 'done')

  /** Классы, которые сейчас хоть кем-то используются, — их и показываем в общем списке. */
  const centrals = w.mcentrals
    .map((c, i) => ({ c, i }))
    .filter(({ c }) => c.partial.length > 0 || c.full.length > 0)

  return (
    <div className="board mem-board">
      <PathStrip w={w} events={events} />

      <HeapGrid w={w} hl={hl} spanByPage={spanByPage} />

      <div className="mem-legend">
        <span className="lg lg-span">спан (цвет — класс размера, заливка — заполненность)</span>
        <span className="lg lg-large">большой объект</span>
        <span className="lg lg-stack">стеки горутин</span>
        <span className="lg lg-free">свободна, память за процессом</span>
        <span className="lg lg-returned">возвращена ОС</span>
      </div>

      <div className="mem-caches">
        {w.mcaches.map((c) => (
          <CacheCard key={c.p} w={w} p={c.p} hl={hl} />
        ))}
      </div>

      <div className="board-zones">
        <div className="zone">
          <span className="board-label">общий список спанов · классов в работе {centrals.length}</span>
          {centrals.length === 0 ? (
            <span className="board-empty">пока пусто</span>
          ) : (
            <span className="mem-centrals">
              {centrals.slice(0, 10).map(({ c, i }) => (
                <span
                  key={i}
                  className={`mem-central ${c.lockedAt === w.tick ? 'is-locked' : ''}`}
                  style={{ '--c': `var(--wl-${i % WL_COLORS})` } as React.CSSProperties}
                  title={`класс ${bytes(SIZE_CLASSES[i]!)}: ${c.partial.length} спанов со свободными слотами, ${c.full.length} занятых целиком${c.lockedAt === w.tick ? '\nблокировка занята в этом тике' : ''}`}
                >
                  {bytes(SIZE_CLASSES[i]!)}
                  <b>
                    {c.partial.length}/{c.partial.length + c.full.length}
                  </b>
                </span>
              ))}
            </span>
          )}
        </div>

        <div className="zone">
          <span className="board-label">горутины · {alive.length}</span>
          <span className="queue">
            {alive.slice(0, 16).map((g) => (
              <GoroutineChip key={g.id} g={g} hl={hl} />
            ))}
            {alive.length > 16 && <span className="board-more">+{alive.length - 16}</span>}
          </span>
        </div>

        <div className="zone zone-stats">
          <span title="Страницы под спаны, большие объекты и стеки">
            занято <b>{used}</b> стр. · {bytes(used * PAGE)}
          </span>
          <span title="Свободные страницы, за которыми ещё стоит настоящая память">
            свободно <b>{free}</b>
          </span>
          <span title="Живых объектов в куче прямо сейчас">
            объектов <b>{w.objects.length}</b>
          </span>
          <span title="Спанов, которые сейчас кому-то принадлежат">
            спанов <b>{w.spans.filter((s) => s.owner.kind !== 'free').length}</b>
          </span>
        </div>
      </div>
    </div>
  )
}
