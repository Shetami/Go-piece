import type { G, M, MState, SimEvent, World } from '../engine/types.ts'

/**
 * Схема мира на одном тике. Чистая проекция: получает снимок, ничего не хранит.
 */

const WL_COLORS = 8
const RUNQ_SHOWN = 24

const M_STATE_LABEL: Record<MState, string> = {
  idle: 'спит',
  spinning: 'ищет',
  running: 'исполняет',
  syscall: 'в ядре',
  blocked: 'в ядре, без P',
}

interface Highlight {
  g: Set<number>
  m: Set<number>
  p: Set<number>
}

export function highlightOf(e: SimEvent | null): Highlight {
  return {
    g: new Set(e?.actors.g ?? []),
    m: new Set(e?.actors.m ?? []),
    p: new Set(e?.actors.p ?? []),
  }
}

function GChip({ g, hl, small }: { g: G; hl: Highlight; small?: boolean }) {
  const color = `var(--wl-${g.workload % WL_COLORS})`
  const title = `G${g.id} · ${g.name} · ${g.state}${g.waitReason ? ` (${g.waitReason})` : ''}\nна CPU ${g.runTicks}, в очереди ${g.runqTicks}, ждала ${g.waitTicks}`
  return (
    <span
      title={title}
      className={`gchip ${small ? 'gchip-sm' : ''} ${hl.g.has(g.id) ? 'is-hl' : ''} state-${g.state}`}
      style={{ '--c': color } as React.CSSProperties}
    >
      G{g.id}
    </span>
  )
}

function MChip({ m, hl }: { m: M; hl: Highlight }) {
  return (
    <span
      className={`mchip m-${m.state} ${hl.m.has(m.id) ? 'is-hl' : ''}`}
      title={`M${m.id}: ${M_STATE_LABEL[m.state]}`}
    >
      M{m.id}
      <small>{M_STATE_LABEL[m.state]}</small>
    </span>
  )
}

function Queue({ ids, w, hl, empty, limit = RUNQ_SHOWN }: {
  ids: number[]
  w: World
  hl: Highlight
  empty: string
  limit?: number
}) {
  if (ids.length === 0) return <span className="board-empty">{empty}</span>
  const shown = ids.slice(0, limit)
  return (
    <span className="queue">
      {shown.map((id) => {
        const g = w.gs[id - 1]
        return g ? <GChip key={id} g={g} hl={hl} small /> : null
      })}
      {ids.length > limit && <span className="board-more">+{ids.length - limit}</span>}
    </span>
  )
}

export function Board({ world: w, highlight: hl }: { world: World; highlight: Highlight }) {
  const mById = (id: number | null) => (id === null ? undefined : w.ms[id - 1])
  const gById = (id: number | null) => (id === null ? undefined : w.gs[id - 1])

  const kernelMs = w.ms.filter((m) => m.state === 'blocked')
  const idleMs = w.ms.filter((m) => m.state === 'idle')
  const done = w.gs.filter((g) => g.state === 'dead').length

  const chanWaits = Object.entries(w.chans).filter(([, c]) => c.sendq.length + c.recvq.length > 0)
  const mutexes = Object.entries(w.mutexes).filter(([, mu]) => mu.holder !== null || mu.waitq.length > 0)

  return (
    <div className="board">
      <div className="board-ps">
        {w.ps.map((p) => {
          const m = mById(p.m)
          const running = m ? gById(m.g) : undefined
          const load = p.runq.length
          return (
            <section key={p.id} className={`pcard p-${p.state} ${hl.p.has(p.id) ? 'is-hl' : ''}`}>
              <header>
                <strong>P{p.id}</strong>
                <span className="pcard-state">
                  {p.state === 'idle' ? 'свободен' : p.state === 'syscall' ? 'ждёт syscall' : 'занят'}
                </span>
              </header>
              <div className="pcard-row">
                <span className="board-label">поток</span>
                {m ? <MChip m={m} hl={hl} /> : <span className="board-empty">—</span>}
              </div>
              <div className="pcard-row">
                <span className="board-label">исполняет</span>
                {running ? (
                  <span className="pcard-running">
                    <GChip g={running} hl={hl} />
                    {Number.isFinite(running.quantumUsed) && running.state === 'running' && (
                      <span
                        className="quantum"
                        title={`квант: ${running.quantumUsed} из ${w.config.quantum}`}
                      >
                        <span style={{ width: `${Math.min(100, (running.quantumUsed / w.config.quantum) * 100)}%` }} />
                      </span>
                    )}
                  </span>
                ) : (
                  <span className="board-empty">—</span>
                )}
              </div>
              <div className="pcard-row">
                <span className="board-label">runnext</span>
                {p.runnext !== null && gById(p.runnext) ? (
                  <GChip g={gById(p.runnext)!} hl={hl} small />
                ) : (
                  <span className="board-slot" />
                )}
              </div>
              <div className="pcard-q">
                <span className="board-label">
                  очередь {load}/{w.config.runqCapacity}
                </span>
                <Queue ids={p.runq} w={w} hl={hl} empty="пусто" />
              </div>
            </section>
          )
        })}
      </div>

      <div className="board-zones">
        <div className="zone">
          <span className="board-label">глобальная очередь · {w.globrunq.length}</span>
          <Queue ids={w.globrunq} w={w} hl={hl} empty="пусто" limit={40} />
        </div>

        <div className="zone">
          <span className="board-label">в ядре без P · {kernelMs.length}</span>
          {kernelMs.length === 0 ? (
            <span className="board-empty">никого</span>
          ) : (
            <span className="queue">
              {kernelMs.map((m) => (
                <span key={m.id} className="pair">
                  <MChip m={m} hl={hl} />
                  {gById(m.g) && <GChip g={gById(m.g)!} hl={hl} small />}
                </span>
              ))}
            </span>
          )}
        </div>

        <div className="zone">
          <span className="board-label">netpoller · {w.netpoll.length}</span>
          <Queue ids={w.netpoll.map((e) => e.g)} w={w} hl={hl} empty="пусто" />
        </div>

        {w.timers.length > 0 && (
          <div className="zone">
            <span className="board-label">таймеры · {w.timers.length}</span>
            <Queue ids={w.timers.map((e) => e.g)} w={w} hl={hl} empty="" />
          </div>
        )}

        {chanWaits.map(([name, c]) => (
          <div className="zone" key={`chan-${name}`}>
            <span className="board-label">канал «{name}»</span>
            <span className="queue">
              {c.sendq.length > 0 && <span className="board-label">send:</span>}
              <Queue ids={c.sendq} w={w} hl={hl} empty="" />
              {c.recvq.length > 0 && <span className="board-label">recv:</span>}
              <Queue ids={c.recvq} w={w} hl={hl} empty="" />
            </span>
          </div>
        ))}

        {mutexes.map(([name, mu]) => (
          <div className="zone" key={`mu-${name}`}>
            <span className="board-label">мьютекс «{name}»</span>
            <span className="queue">
              {mu.holder !== null && gById(mu.holder) && (
                <>
                  <span className="board-label">держит</span>
                  <GChip g={gById(mu.holder)!} hl={hl} small />
                </>
              )}
              {mu.waitq.length > 0 && <span className="board-label">ждут</span>}
              <Queue ids={mu.waitq} w={w} hl={hl} empty="" />
            </span>
          </div>
        ))}

        <div className="zone">
          <span className="board-label">спящие потоки · {idleMs.length}</span>
          {idleMs.length === 0 ? (
            <span className="board-empty">нет</span>
          ) : (
            <span className="queue">
              {idleMs.slice(0, 16).map((m) => (
                <MChip key={m.id} m={m} hl={hl} />
              ))}
              {idleMs.length > 16 && <span className="board-more">+{idleMs.length - 16}</span>}
            </span>
          )}
        </div>

        <div className="zone zone-stats">
          <span>
            потоков <b>{w.ms.length}</b> при GOMAXPROCS <b>{w.config.gomaxprocs}</b>
          </span>
          <span>
            горутин живо <b>{w.gs.length - done}</b>, завершено <b>{done}</b>
          </span>
        </div>
      </div>
    </div>
  )
}
