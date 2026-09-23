import { useMemo } from 'react'
import type { Order, SagaEvent, SagaWorld } from '../engine/types.ts'
import type { Simulation } from '../engine/simulation.ts'
import { paidNotShipped } from '../engine/world.ts'
import { Chart, FLOW_WINDOW, chartData } from './Charts.tsx'

/**
 * Три сервиса и брокер на одном тике. Чистая проекция: получает кадр, ничего
 * не хранит. Сверху — графики прогона до курсора, под ними — заказы по
 * состояниям, содержимое outbox и сообщения в пути.
 */

export interface Highlight {
  order: Set<number>
  msg: Set<number>
}

export function highlightOf(e: SagaEvent | null): Highlight {
  return { order: new Set(e?.actors.order ?? []), msg: new Set(e?.actors.msg ?? []) }
}

const SHOWN = 60

const STATE: Record<Order['state'], { label: string; cls: string }> = {
  created: { label: 'создан', cls: 'is-created' },
  paid: { label: 'оплачен', cls: 'is-paid' },
  shipped: { label: 'доставлен', cls: 'is-shipped' },
  cancelled: { label: 'отменён', cls: 'is-cancelled' },
}

function Orders({ w, hl }: { w: SagaWorld; hl: Highlight }) {
  const shown = w.orders.slice(-SHOWN)
  const counts = (['created', 'paid', 'shipped', 'cancelled'] as const).map((s) => ({ s, n: w.orders.filter((o) => o.state === s).length }))
  return (
    <section className="sg-panel" aria-label="Заказы">
      <header>
        <strong>Заказы</strong>
        <span className="kf-muted">
          {counts.map(({ s, n }) => `${STATE[s].label}: ${n}`).join(' · ')}
        </span>
      </header>
      <div className="sg-orders">
        {w.orders.length === 0 && <span className="kf-muted">заказов пока нет</span>}
        {shown.map((o) => {
          const lost = o.state === 'created' && o.published && !w.broker.some((m) => m.order === o.id) && !w.work.some((x) => x.msg.order === o.id)
          return (
            <span
              key={o.id}
              className={`sg-order ${STATE[o.state].cls} ${o.charges > 1 ? 'is-double' : ''} ${lost ? 'is-lost' : ''} ${hl.order.has(o.id) ? 'is-hl' : ''}`}
              title={`№${o.id}: ${STATE[o.state].label}${o.charges > 0 ? `, списаний ${o.charges}` : ''}${o.refunded ? ', деньги возвращены' : ''}${lost ? ' — событие потеряно' : ''}`}
            >
              {o.charges > 1 ? '₽₽' : ''}
            </span>
          )
        })}
        {w.orders.length > SHOWN && <span className="ld-more">показаны последние {SHOWN}</span>}
      </div>
    </section>
  )
}

function Pipeline({ w, hl }: { w: SagaWorld; hl: Highlight }) {
  const pending = w.outbox.filter((r) => r.sentAt === null)
  return (
    <section className={`sg-panel ${w.brokerUp ? '' : 'is-bad'}`} aria-label="Outbox и брокер">
      <header>
        <strong>{w.config.publish === 'outbox' ? 'Outbox и брокер' : 'Брокер'}</strong>
        <span className={`lb-badge ${w.brokerUp ? '' : 'is-bad'}`}>{w.brokerUp ? 'брокер доступен' : 'брокер недоступен'}</span>
      </header>
      {w.config.publish === 'outbox' && (
        <div className="ld-chips" aria-label="Строки outbox">
          <span className="kf-muted">в outbox {pending.length}:</span>
          {pending.slice(0, 24).map((r) => (
            <span key={r.id} className={`ld-chip ${hl.order.has(r.order) ? 'is-hl' : ''}`} title={`заказ №${r.order}: ${r.kind}, записана на тике ${r.writtenAt}`} />
          ))}
          {pending.length > 24 && <span className="ld-more">+{pending.length - 24}</span>}
        </div>
      )}
      <div className="ld-chips" aria-label="Сообщения в брокере">
        <span className="kf-muted">в брокере {w.broker.length}:</span>
        {w.broker.slice(0, 24).map((m) => (
          <span
            key={m.id}
            className={`ld-chip ${m.duplicate ? 'is-zombie' : 'is-joined'} ${hl.msg.has(m.id) || hl.order.has(m.order) ? 'is-hl' : ''}`}
            title={`заказ №${m.order}: ${m.kind}${m.duplicate ? ' — дубль' : ''}, дойдёт на тике ${m.arriveAt}`}
          />
        ))}
        {w.broker.length > 24 && <span className="ld-more">+{w.broker.length - 24}</span>}
      </div>
      <div className="sg-consumers">
        {(['pay', 'ship'] as const).map((c) => (
          <span key={c} className="sg-consumer">
            {c === 'pay' ? 'платежи' : 'доставка'}: обрабатывают <b>{w.work.filter((x) => x.consumer === c).length}</b> из {w.config.workers}
          </span>
        ))}
        {w.retryQ.length > 0 && <span className="sg-consumer">ждут повтора: <b>{w.retryQ.length}</b></span>}
      </div>
    </section>
  )
}

export function Board({ sim, cursor, highlight: hl, onSeek }: { sim: Simulation; cursor: number; highlight: Highlight; onSeek: (t: number) => void }) {
  const w = sim.history[cursor]!.world
  const last = sim.history.length - 1
  const d = useMemo(() => chartData(sim.log), [sim])
  const c = w.config

  return (
    <div className="board ld-board">
      <div className="ld-charts">
        <Chart
          title="Заказы"
          note={`за последние ${FLOW_WINDOW} тиков`}
          series={[
            { label: 'создано', color: 'wait', values: d.created },
            { label: 'доставлено', color: 'work', values: d.shipped },
            ...(d.anyCancelled ? [{ label: 'отменено', color: 'waste' as const, values: d.cancelled }] : []),
          ]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
        <Chart
          title="Между сервисами"
          note="заказов без конечного состояния и событий в пути"
          series={[
            { label: 'заказов в пути', color: 'wait', values: d.inFlight },
            { label: 'событий в брокере', color: 'in', values: d.inBroker },
            ...(d.anyOutbox ? [{ label: 'строк в outbox', color: 'waste' as const, values: d.outbox }] : []),
          ]}
          last={last}
          cursor={cursor}
          onSeek={onSeek}
        />
      </div>

      <p className="lb-summary">
        публикация <b>{c.publish === 'outbox' ? 'через outbox' : 'две записи подряд'}</b> · потребители{' '}
        <b>{c.idempotent ? 'идемпотентны' : 'не проверяют повторы'}</b> · компенсация{' '}
        <b>{c.compensate ? 'включена' : 'выключена'}</b> · оплачено, но не доставлено сейчас: <b>{paidNotShipped(w)}</b>
      </p>
      <div className="rs-view">
        <Orders w={w} hl={hl} />
        <Pipeline w={w} hl={hl} />
      </div>
      <p className="rp-legend" aria-hidden="true">
        <span className="sg-order is-created" /> создан <span className="sg-order is-paid" /> оплачен
        <span className="sg-order is-shipped" /> доставлен <span className="sg-order is-cancelled" /> отменён
        <span className="sg-order is-lost" /> событие потеряно <span className="sg-order is-double">₽₽</span> двойное списание
      </p>
    </div>
  )
}
