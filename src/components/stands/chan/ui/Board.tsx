import type { Chan, ChanEvent, ChanWorld, Goroutine, Waiter } from '../engine/types.ts'

/**
 * Схема мира на одном тике. Чистая проекция: получает снимок, ничего не хранит.
 *
 * Главный элемент — карточка канала: очередь отправителей слева, кольцевой буфер
 * посередине, очередь получателей справа. Всё поведение канала видно в одной строке.
 */

const WL_COLORS = 8

export interface Highlight {
  g: Set<number>
  chans: Set<number>
}

export function highlightOf(e: ChanEvent | null): Highlight {
  return {
    g: new Set(e?.actors.g ?? []),
    chans: new Set(e?.actors.chan ?? []),
  }
}

/** Подпись операции на карточке процессора. */
const OP_LABEL: Record<string, string> = {
  cpu: 'считает',
  'send.direct': 'отдала в руки',
  'send.buffer': 'в буфер',
  'send.block': 'уснула на отправке',
  'recv.direct': 'взяла из рук',
  'recv.buffer': 'из буфера',
  'recv.block': 'уснула на приёме',
  'recv.closed': 'нулевое значение',
  'select.default': 'default',
  'select.block': 'уснула в select',
  close: 'закрыла канал',
  nil: 'nil-канал',
  panic: 'паника',
  done: 'закончила',
}

/** Ступени, которые проходит операция с каналом, — от самой дешёвой к самой дорогой. */
const STAGES = [
  {
    key: 'handoff',
    label: 'встречный ждёт',
    hint: 'В очереди уже кто-то стоит — значение копируется напрямую из стека в стек, мимо буфера',
  },
  {
    key: 'buffer',
    label: 'место в буфере',
    hint: 'Встречного нет, но в кольце есть свободная ячейка — отправитель не блокируется',
  },
  {
    key: 'park',
    label: 'парковка',
    hint: 'Ни встречного, ни места: горутина встаёт в очередь канала и засыпает',
  },
] as const

function OpStrip({ events }: { events: ChanEvent[] }) {
  const active = new Set<string>()
  for (const e of events) {
    if (e.type === 'send.direct' || e.type === 'recv.direct' || e.type === 'recv.wake') active.add('handoff')
    if (e.type === 'send.buffer' || e.type === 'recv.buffer' || e.type === 'buf.full') active.add('buffer')
    if (e.type === 'send.block' || e.type === 'recv.block' || e.type === 'select.block') active.add('park')
  }
  const closed = events.some((e) => e.type === 'chan.close')
  const panic = events.some((e) => e.type === 'send.closed')

  return (
    <div className="chan-path" role="group" aria-label="Путь операции с каналом">
      {STAGES.map((s, i) => (
        <span key={s.key} className={`chan-stage ${active.has(s.key) ? 'is-now' : ''}`} title={s.hint}>
          {s.label}
          {i < STAGES.length - 1 && <i aria-hidden="true">→</i>}
        </span>
      ))}
      {closed && <span className="chan-stage-note">канал закрыт — проснулись все</span>}
      {panic && <span className="chan-stage-panic">паника: отправка в закрытый канал</span>}
    </div>
  )
}

function WaitChip({ w, g, hl, dir }: { w: Waiter; g: Goroutine; hl: Highlight; dir: 'send' | 'recv' }) {
  return (
    <span
      className={`chan-waiter ${hl.g.has(w.g) ? 'is-hl' : ''} ${w.fromSelect ? 'is-select' : ''}`}
      style={{ '--c': `var(--wl-${g.workload % WL_COLORS})` } as React.CSSProperties}
      title={
        `G${w.g} · ${g.name}\n` +
        `${dir === 'send' ? 'ждёт, кому отдать значение' : 'ждёт значения'} с тика ${w.since}` +
        (w.fromSelect ? '\nпришла из select — стоит сразу в нескольких очередях' : '')
      }
    >
      G{w.g}
      {w.fromSelect && <b title="Этот sudog заведён оператором select">sel</b>}
    </span>
  )
}

function ChannelCard({ w, c, hl }: { w: ChanWorld; c: Chan; hl: Highlight }) {
  const locked = c.lockedAt === w.tick
  const kind = c.isNil ? 'не создан через make' : c.cap === 0 ? 'без буфера' : `буфер ${c.qcount}/${c.cap}`

  return (
    <section
      className={`chan-card ${hl.chans.has(c.id) ? 'is-hl' : ''} ${locked ? 'is-locked' : ''} ${c.closed ? 'is-closed' : ''} ${c.isNil ? 'is-nil' : ''}`}
    >
      <header>
        <strong>{c.name}</strong>
        <span className="chan-kind">{kind}</span>
        {c.closed && <span className="chan-badge" title="Канал закрыт: приём отдаёт нулевые значения, отправка паникует">закрыт</span>}
        {c.isNil && <span className="chan-badge chan-badge-nil" title="Канал не создан через make: любая операция блокируется навсегда">nil</span>}
      </header>

      <div className="chan-row">
        <span
          className="chan-q chan-q-send"
          title={`Очередь отправителей (sendq): ${c.sendq.length}. Строго FIFO — значение заберут у того, кто уснул раньше всех`}
        >
          {c.sendq.length === 0 ? (
            <i className="chan-q-empty">—</i>
          ) : (
            c.sendq.map((x) => <WaitChip key={x.g} w={x} g={w.gs[x.g - 1]!} hl={hl} dir="send" />)
          )}
          <em aria-hidden="true">→</em>
        </span>

        {c.isNil ? (
          <span className="chan-void" title="У nil-канала нет ни буфера, ни очередей — ждать здесь можно только вечно">
            ∅
          </span>
        ) : c.cap === 0 ? (
          <span className="chan-meet" title="Небуферизованный канал ничего не хранит: это точка встречи, а не ящик">
            точка встречи
          </span>
        ) : (
          <span className="chan-buf" role="img" aria-label={`Буфер: ${c.qcount} из ${c.cap}`}>
            {c.buf.map((owner, i) => {
              const g = owner === null ? null : w.gs[owner - 1]
              return (
                <span
                  key={i}
                  className={`chan-cell ${owner === null ? 'is-empty' : 'is-full'} ${i === c.sendx ? 'is-sendx' : ''} ${i === c.recvx ? 'is-recvx' : ''}`}
                  style={
                    g ? ({ '--c': `var(--wl-${g.workload % WL_COLORS})` } as React.CSSProperties) : undefined
                  }
                  title={
                    (owner === null ? `ячейка ${i}: свободна` : `ячейка ${i}: значение от G${owner}`) +
                    (i === c.recvx ? '\n← отсюда читает получатель (recvx)' : '') +
                    (i === c.sendx ? '\n→ сюда пишет отправитель (sendx)' : '')
                  }
                />
              )
            })}
          </span>
        )}

        <span
          className="chan-q chan-q-recv"
          title={`Очередь получателей (recvq): ${c.recvq.length}. Одновременно непустой может быть только одна из двух очередей`}
        >
          <em aria-hidden="true">→</em>
          {c.recvq.length === 0 ? (
            <i className="chan-q-empty">—</i>
          ) : (
            c.recvq.map((x) => <WaitChip key={x.g} w={x} g={w.gs[x.g - 1]!} hl={hl} dir="recv" />)
          )}
        </span>
      </div>

      <footer className="chan-stats">
        <span title="Всего значений прошло через канал">передач <b>{c.stats.sent}</b></span>
        <span title="Из них мимо буфера — прямо из стека в стек">из рук <b>{c.stats.direct}</b></span>
        <span title="И через кольцевой буфер">через буфер <b>{c.stats.buffered}</b></span>
        <span title="Сколько раз здесь засыпали отправители и получатели">
          парковок <b>{c.stats.blockedSends + c.stats.blockedRecvs}</b>
        </span>
      </footer>
    </section>
  )
}

function PSlot({ w, p, hl }: { w: ChanWorld; p: number; hl: Highlight }) {
  const slot = w.slots.find((s) => s.p === p)
  const g = slot?.g ? w.gs[slot.g - 1] : undefined
  const chan = slot?.chan !== null && slot?.chan !== undefined ? w.chans[slot.chan] : undefined
  const op = slot?.op ?? null

  return (
    <span
      className={`chan-slot ${op ? `op-${op.split('.')[0]}` : 'is-idle'} ${g && hl.g.has(g.id) ? 'is-hl' : ''}`}
      style={g ? ({ '--c': `var(--wl-${g.workload % WL_COLORS})` } as React.CSSProperties) : undefined}
      title={g ? `P${p}: G${g.id} «${g.name}» — ${OP_LABEL[op ?? ''] ?? op}` : `P${p} простаивает: готовых горутин нет`}
    >
      <b>P{p}</b>
      {g ? (
        <>
          <span className="chan-slot-g">G{g.id}</span>
          <i>
            {OP_LABEL[op ?? ''] ?? op}
            {chan ? ` · ${chan.name}` : ''}
          </i>
        </>
      ) : (
        <i className="chan-slot-idle">простаивает</i>
      )}
    </span>
  )
}

function GoroutineChip({ g, w, hl }: { g: Goroutine; w: ChanWorld; hl: Highlight }) {
  const waitOn = g.wait ? g.wait.chans.map((id) => w.chans[id]?.name).filter(Boolean).join(', ') : ''
  const waited = g.wait ? w.tick - g.wait.since : 0
  const kind = g.wait?.kind
  return (
    <span
      className={`chan-g state-${g.state} ${hl.g.has(g.id) ? 'is-hl' : ''}`}
      style={{ '--c': `var(--wl-${g.workload % WL_COLORS})` } as React.CSSProperties}
      title={
        `G${g.id} · ${g.name}\n` +
        (g.state === 'waiting'
          ? `спит ${waited} тиков: ${kind === 'send' ? 'отправка в' : kind === 'recv' ? 'приём из' : 'select на'} ${waitOn}\n`
          : '') +
        `отправила ${g.sent}, приняла ${g.received}, парковок ${g.blocks}, всего ждала ${g.waitTicks} тиков`
      }
    >
      G{g.id}
      {g.state === 'waiting' && (
        <b>
          {kind === 'send' ? '→' : kind === 'recv' ? '←' : 'sel'} {waitOn}
        </b>
      )}
      {g.state === 'waiting' && <small>{waited}т</small>}
    </span>
  )
}

export function Board({ world: w, events, highlight: hl }: {
  world: ChanWorld
  events: ChanEvent[]
  highlight: Highlight
}) {
  const alive = w.gs.filter((g) => g.state !== 'done')
  const parked = alive.filter((g) => g.state === 'waiting')
  const s = w.stats

  return (
    <div className="board chan-board">
      <OpStrip events={events} />

      <div className="chan-cards">
        {w.chans.map((c) => (
          <ChannelCard key={c.id} w={w} c={c} hl={hl} />
        ))}
      </div>

      <div className="chan-slots">
        {w.slots.map((slot) => (
          <PSlot key={slot.p} w={w} p={slot.p} hl={hl} />
        ))}
      </div>

      <div className="board-zones">
        <div className="zone">
          <span className="board-label">горутины · живых {alive.length}, спят {parked.length}</span>
          <span className="queue">
            {alive.slice(0, 18).map((g) => (
              <GoroutineChip key={g.id} g={g} w={w} hl={hl} />
            ))}
            {alive.length > 18 && <span className="board-more">+{alive.length - 18}</span>}
          </span>
        </div>

        <div className="zone zone-stats">
          <span title="Все значения, дошедшие до получателя или до буфера">
            передач <b>{s.transfers}</b>
          </span>
          <span title="Из них прошли мимо буфера — прямо из стека в стек">
            из рук в руки <b>{s.transfers === 0 ? 0 : Math.round((s.direct / s.transfers) * 100)}%</b>
          </span>
          <span title="Сколько раз горутина засыпала на канале">
            парковок <b>{s.parks}</b>
          </span>
          <span title="Среднее время, проведённое в парковке">
            ждали в среднем <b>{s.parks === 0 ? 0 : Math.round((s.waitTicks / s.parks) * 10) / 10}</b> т.
          </span>
        </div>
      </div>
    </div>
  )
}
