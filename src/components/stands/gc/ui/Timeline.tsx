import { useEffect, useMemo, useRef } from 'react'
import type { GcEvent, GcEventType, Snapshot } from '../engine/types.ts'
import { EVENT_IMPORTANCE } from '../engine/types.ts'
import { explain } from '../explain/events.ts'

/**
 * Лента и перемотка. Устроены так же, как в стенде планировщика, но знают
 * про события сборщика: свой список важности и свой разбор.
 */

/** Полоса перемотки с засечками ключевых событий. */
export function Scrubber({ history, cursor, last, watchFor, onSeek }: {
  history: readonly Snapshot[]
  cursor: number
  last: number
  watchFor: GcEventType[]
  onSeek: (t: number) => void
}) {
  const marks = useMemo(() => {
    const out: { tick: number; type: GcEventType }[] = []
    for (const s of history) {
      const e = s.events.find((x) => EVENT_IMPORTANCE[x.type] === 'key' || watchFor.includes(x.type))
      if (e) out.push({ tick: s.world.tick, type: e.type })
    }
    return out
  }, [history, watchFor])

  return (
    <div className="scrubber">
      <div className="scrubber-marks" aria-hidden="true">
        {marks.map((m) => (
          <span
            key={m.tick}
            className={`mark mark-${m.type.replace('.', '-')}`}
            style={{ left: `${(m.tick / Math.max(1, last)) * 100}%` }}
          />
        ))}
      </div>
      <input
        type="range"
        min={0}
        max={last}
        value={cursor}
        onChange={(e) => onSeek(Number(e.target.value))}
        aria-label="Перемотка по тикам"
      />
    </div>
  )
}

/** Лента событий: новые сверху, клик выбирает событие и прыгает на его тик. */
export function EventFeed({ events, selected, onSelect, showMinor }: {
  events: GcEvent[]
  selected: GcEvent | null
  onSelect: (e: GcEvent) => void
  showMinor: boolean
}) {
  const shown = useMemo(() => {
    const indexed = events.map((e, i) => ({ e, key: i }))
    const filtered = showMinor ? indexed : indexed.filter(({ e }) => EVENT_IMPORTANCE[e.type] !== 'low')
    return filtered.slice(-150).reverse()
  }, [events, showMinor])

  const listRef = useRef<HTMLOListElement>(null)
  useEffect(() => {
    listRef.current?.scrollTo({ top: 0 })
  }, [events.length])

  if (shown.length === 0) {
    return <p className="feed-empty">Событий пока нет — нажмите ▶ или «шаг».</p>
  }

  return (
    <ol className="feed" ref={listRef}>
      {shown.map(({ e, key }) => (
        <li key={key}>
          <button
            type="button"
            className={`feed-item imp-${EVENT_IMPORTANCE[e.type]} ${selected === e ? 'is-sel' : ''}`}
            onClick={() => onSelect(e)}
          >
            <span className="feed-tick">{e.tick}</span>
            <span className={`feed-dot dot-${e.type.split('.')[0]}`} />
            <span className="feed-title">{explain(e).title}</span>
          </button>
        </li>
      ))}
    </ol>
  )
}

export interface TermInfo {
  title: string
  short: string
}

export function EventDetail({ event, terms, onClose }: {
  event: GcEvent
  terms: Record<string, TermInfo>
  onClose: () => void
}) {
  const x = explain(event)
  return (
    <div className="detail" role="region" aria-label="Разбор события">
      <header>
        <span className="detail-tick">тик {event.tick}</span>
        <h4>{x.title}</h4>
        <button type="button" className="btn-ghost" onClick={onClose} aria-label="Закрыть разбор">
          ✕
        </button>
      </header>
      {x.body.map((para, i) => (
        <p key={i}>{para}</p>
      ))}
      {x.runtime && (
        <p className="detail-runtime">
          <span>В рантайме:</span> <code>{x.runtime}</code>
        </p>
      )}
      {x.model && (
        <p className="detail-model">
          <span>Упрощение модели:</span> {x.model}
        </p>
      )}
      {x.terms.length > 0 && (
        <p className="detail-terms">
          {x.terms.map((id) => (
            <a key={id} href={`/glossary/${id}/`} title={terms[id]?.short}>
              {terms[id]?.title ?? id}
            </a>
          ))}
        </p>
      )}
    </div>
  )
}
