import type { Acks, Group, KafkaEvent, KafkaEventType, KafkaScenario, KafkaWorld, Msg, Partition, Rec } from '../engine/types.ts'

/**
 * Схема кластера на одном тике. Чистая проекция: получает снимок, ничего не хранит.
 *
 * Читается сверху вниз, как путь сообщения: продюсеры со своими пакетами,
 * то, что сейчас летит по сети, логи реплик по партициям и брокерам, группы
 * потребителей. Любое сообщение кликабельно — клик показывает его путь.
 */

const WL_COLORS = 8
/** Сколько последних записей лога видно в ячейке реплики. */
const LOG_TAIL = 10

export interface Highlight {
  recs: Set<number>
  brokers: Set<number>
  partitions: Set<number>
  producers: Set<number>
  consumers: Set<number>
}

export function highlightOf(e: KafkaEvent | null): Highlight {
  return {
    recs: new Set(e?.actors.rec ?? []),
    brokers: new Set(e?.actors.broker ?? []),
    partitions: new Set(e?.actors.partition ?? []),
    producers: new Set(e?.actors.producer ?? []),
    consumers: new Set(e?.actors.consumer ?? []),
  }
}

/** Цвет сообщения: по ключу, если он есть, — так видно, что один ключ живёт в одной партиции. */
export function recColor(rec: Rec, sc: KafkaScenario): string {
  const keys = sc.producers[rec.producer]?.keys ?? []
  const i = rec.key === null ? rec.producer : Math.max(0, keys.indexOf(rec.key))
  return `var(--wl-${i % WL_COLORS})`
}

export function RecChip({ rec, sc, focus, hl, onFocus, variant, suffix }: {
  rec: Rec
  sc: KafkaScenario
  focus: number | null
  hl: Highlight
  onFocus: (id: number) => void
  variant?: 'pending' | 'dup'
  suffix?: string
}) {
  const bad = rec.state === 'lost' || rec.state === 'failed'
  return (
    <button
      type="button"
      className={[
        'kf-rec',
        focus === rec.id && 'is-focus',
        hl.recs.has(rec.id) && 'is-hl',
        variant === 'pending' && 'is-pending',
        variant === 'dup' && 'is-dup',
        bad && 'is-bad',
      ]
        .filter(Boolean)
        .join(' ')}
      style={{ '--c': recColor(rec, sc) } as React.CSSProperties}
      title={
        `m${rec.id}${rec.key === null ? '' : ` · ключ «${rec.key}»`} · p${rec.partition}` +
        (rec.offset === null ? '' : ` · оффсет ${rec.offset}`) +
        `\nсостояние: ${STATE_LABEL[rec.state]}` +
        (variant === 'pending' ? '\nвыше HW — ещё не закоммичено, потребители его не видят' : '') +
        (variant === 'dup' ? '\nдубль: повтор пакета без идемпотентности' : '') +
        '\nклик — показать путь'
      }
      onClick={() => onFocus(rec.id)}
    >
      m{rec.id}
      {suffix && <small>{suffix}</small>}
    </button>
  )
}

const STATE_LABEL: Record<Rec['state'], string> = {
  batched: 'в пакете у продюсера',
  'in-flight': 'в пути к брокеру',
  acked: 'подтверждено',
  failed: 'не доставлено',
  lost: 'потеряно',
}

/* ─────────────────────────── полоса пути ─────────────────────────── */

const STAGE_OF: Partial<Record<KafkaEventType, string>> = {
  'rec.send': 'send',
  'batch.seal': 'seal',
  'produce.request': 'request',
  'produce.append': 'append',
  'produce.duplicate': 'append',
  'produce.dedup': 'append',
  'repl.fetch': 'replicate',
  'hw.advance': 'commit',
  'produce.ack': 'ack',
  'consumer.fetch': 'fetch',
  'consumer.process': 'process',
  'offset.commit': 'offset',
}

const ALERTS: Partial<Record<KafkaEventType, string>> = {
  'produce.error': 'ошибка записи — повтор',
  'produce.failed': 'не доставлено',
  'produce.duplicate': 'дубль в логе',
  'resp.lost': 'ответ потерян',
  'rec.lost': 'потеря подтверждённого',
  'log.truncate': 'лог обрезан',
  'leader.elect': 'смена лидера',
  'broker.down': 'брокер упал',
  'isr.shrink': 'ISR сжался',
  'group.rebalance': 'ребалансировка',
  'rec.reprocess': 'повторная обработка',
  'rec.skipped': 'пропуск',
}

function stagesFor(acks: Acks): { key: string; label: string; hint: string }[] {
  const s = {
    send: { key: 'send', label: 'send()', hint: 'Приложение отдало сообщение продюсеру' },
    seal: { key: 'seal', label: 'пакет', hint: 'Пакет закрыт: набрал batch.size или прождал linger.ms' },
    request: { key: 'request', label: 'запрос', hint: 'ProduceRequest летит лидеру партиции' },
    append: { key: 'append', label: 'лог лидера', hint: 'Лидер дописал пакет в конец лога и назначил оффсеты' },
    replicate: { key: 'replicate', label: 'реплики', hint: 'Фолловеры скопировали записи запросами Fetch' },
    commit: { key: 'commit', label: 'HW', hint: 'High watermark перешёл записи: они закоммичены и видны потребителям' },
    ack: { key: 'ack', label: 'ack', hint: 'Продюсер получил подтверждение' },
    fetch: { key: 'fetch', label: 'poll', hint: 'Потребитель прочитал закоммиченные записи' },
    process: { key: 'process', label: 'обработка', hint: 'Код приложения обработал сообщение' },
    offset: { key: 'offset', label: 'коммит', hint: 'Группа закоммитила оффсет' },
  }
  const head = [s.send, s.seal, s.request]
  const tail = [s.fetch, s.process, s.offset]
  if (acks === 'all') return [...head, s.append, s.replicate, s.commit, s.ack, ...tail]
  if (acks === 1) return [...head, s.append, s.ack, s.replicate, s.commit, ...tail]
  return [...head, s.ack, s.append, s.replicate, s.commit, ...tail]
}

function PathStrip({ events, acks }: { events: KafkaEvent[]; acks: Acks }) {
  const active = new Set<string>()
  const alerts = new Set<string>()
  for (const e of events) {
    const st = STAGE_OF[e.type]
    if (st) active.add(st)
    const a = ALERTS[e.type]
    if (a && !(e.type === 'group.rebalance' && e.payload.first)) alerts.add(a)
  }
  return (
    <div className="chan-path kf-path" role="group" aria-label="Остановки пути сообщения — подсвечены те, что пройдены в этом тике">
      {stagesFor(acks).map((s, i, all) => (
        <span key={s.key} className={`chan-stage ${active.has(s.key) ? 'is-now' : ''}`} title={s.hint}>
          {s.label}
          {i < all.length - 1 && <i aria-hidden="true">→</i>}
        </span>
      ))}
      {[...alerts].map((a) => (
        <span key={a} className="chan-stage-panic">
          {a}
        </span>
      ))}
    </div>
  )
}

/* ─────────────────────────── продюсеры ─────────────────────────── */

function Producers({ w, sc, hl, focus, onFocus }: ZoneProps) {
  return (
    <section className="kf-zone">
      <span className="board-label">продюсеры · {w.config.acks === 'all' ? 'acks=all' : `acks=${w.config.acks}`} · linger {w.config.lingerTicks} · пакет до {w.config.batchSize}</span>
      <div className="kf-producers">
        {w.producers.map((p) => {
          const spec = sc.producers[p.spec]!
          const open = Object.values(p.open).map((id) => w.batches[id - 1]!)
          const ready = p.ready.map((id) => w.batches[id - 1]!)
          return (
            <div key={p.id} className={`kf-card ${hl.producers.has(p.id) ? 'is-hl' : ''}`}>
              <header>
                <strong>{p.name}</strong>
                <span className="kf-muted">
                  send() {p.created}/{spec.messages}
                </span>
              </header>
              <div className="kf-batches" title="Аккумулятор: пакеты по партициям, ещё не отправленные">
                {open.length + ready.length === 0 && <i className="kf-empty">аккумулятор пуст</i>}
                {[...ready, ...open].map((b) => (
                  <span key={b.id} className={`kf-batch ${b.state === 'open' ? 'is-open' : 'is-ready'} ${b.attempts > 0 ? 'is-retry' : ''}`} title={
                    `пакет #${b.id} в p${b.partition}\n` +
                    (b.state === 'open'
                      ? `открыт ${w.tick - b.createdTick} т.: ждёт попутчиков до linger.ms или batch.size`
                      : b.attempts > 0
                        ? `ждёт повтора (попыток ${b.attempts})`
                        : 'полон, ждёт отправки')
                  }>
                    <b>p{b.partition}</b>
                    {b.recs.map((id) => (
                      <RecChip key={id} rec={w.recs[id - 1]!} sc={sc} focus={focus} hl={hl} onFocus={onFocus} />
                    ))}
                    <small>{b.state === 'open' ? `${w.tick - b.createdTick}т` : b.attempts > 0 ? 'повтор' : 'готов'}</small>
                  </span>
                ))}
              </div>
              <div className="kf-inflight" title="Запросы в пути: отправлены, ответа ещё нет">
                {p.inFlight.map((f) => {
                  const b = w.batches[f.batch - 1]!
                  return (
                    <span key={f.req} className="kf-req">
                      #{b.id} → B{f.broker} <small>{w.tick - f.sentTick}т</small>
                    </span>
                  )
                })}
              </div>
              <footer className="chan-stats">
                <span title="Запросов записи отправлено">запросов <b>{p.stats.requests}</b></span>
                <span title="Сообщений с подтверждением">ack <b>{p.stats.acked}</b></span>
                <span title="Повторов после ошибок и таймаутов">повторов <b>{p.stats.retries}</b></span>
                {p.stats.failed > 0 && (
                  <span className="metric-bad" title="Истёк delivery.timeout.ms">не доставлено <b>{p.stats.failed}</b></span>
                )}
              </footer>
            </div>
          )
        })}
      </div>
    </section>
  )
}

/* ─────────────────────────── сеть ─────────────────────────── */

function msgLabel(m: Msg, w: KafkaWorld): { from: string; to: string; what: string; recs: number[]; kind: string } | null {
  const P = (id: number) => w.producers[id]?.name ?? `P${id}`
  const C = (id: number) => w.consumers[id]?.name ?? `C${id}`
  switch (m.kind) {
    case 'produce':
      return { from: P(m.producer), to: `B${m.broker}`, what: `запись p${m.partition}`, recs: w.batches[m.batch - 1]!.recs, kind: 'produce' }
    case 'produce-resp':
      return { from: `B${m.broker}`, to: P(m.producer), what: m.error ? `ошибка ${m.error}` : `ok p${m.partition} @${m.baseOffset}`, recs: [], kind: m.error ? 'error' : 'ack' }
    case 'repl-resp':
      return m.entries.length === 0 ? null : { from: `B${m.from}`, to: `B${m.broker}`, what: `копия p${m.partition}`, recs: m.entries.map((e) => e.rec), kind: 'repl' }
    case 'fetch-resp':
      return m.entries.length === 0 ? null : { from: `B${m.broker}`, to: C(m.consumer), what: `данные p${m.partition}`, recs: m.entries.map((e) => e.rec), kind: 'fetch' }
    case 'commit':
      return {
        from: C(m.consumer),
        to: 'координатор',
        what: `коммит ${Object.entries(m.offsets).map(([p, o]) => `p${p}→${o}`).join(' ')}`,
        recs: [],
        kind: 'commit',
      }
    default:
      return null
  }
}

function Network({ w, sc, hl, focus, onFocus }: ZoneProps) {
  const shown = w.net.map((m) => ({ m, l: msgLabel(m, w) })).filter((x) => x.l !== null)
  const hidden = w.net.length - shown.length
  return (
    <section className="kf-zone">
      <span className="board-label">
        в сети · {w.net.length}
        {hidden > 0 && <span className="kf-muted"> (из них {hidden} пустых Fetch — реплики и потребители спрашивают, нет ли нового)</span>}
      </span>
      <div className="kf-net">
        {shown.length === 0 && <i className="kf-empty">ничего полезного не летит</i>}
        {shown.map(({ m, l }) => (
          <span key={m.id} className={`kf-msg kf-msg-${l!.kind}`}>
            <b>{l!.from}</b>→<b>{l!.to}</b> {l!.what}
            {l!.recs.map((id) => (
              <RecChip key={id} rec={w.recs[id - 1]!} sc={sc} focus={focus} hl={hl} onFocus={onFocus} />
            ))}
            <small>{m.deliverAt - w.tick}т</small>
          </span>
        ))}
      </div>
    </section>
  )
}

/* ─────────────────────────── кластер ─────────────────────────── */

function ReplicaCell({ w, sc, p, b, hl, focus, onFocus }: ZoneProps & { p: Partition; b: number }) {
  const rep = p.logs[b]
  const broker = w.brokers[b]!
  if (!rep) return <div className="kf-replica is-none" aria-hidden="true" />
  const isLeader = p.leader === b
  const inIsr = p.isr.includes(b)
  const hw = isLeader ? p.hw : rep.hw
  const log = rep.log
  const from = Math.max(0, log.length - LOG_TAIL)
  const role = isLeader ? 'лидер' : inIsr ? 'в ISR' : 'вне ISR'
  return (
    <div
      className={[
        'kf-replica',
        isLeader && 'is-leader',
        !inIsr && 'is-out',
        !broker.alive && 'is-dead',
        hl.brokers.has(b) && hl.partitions.has(p.id) && 'is-hl',
      ]
        .filter(Boolean)
        .join(' ')}
      title={
        `p${p.id} на B${b}: ${role}${broker.alive ? '' : ', брокер мёртв'}\n` +
        `LEO ${log.length} — сюда ляжет следующая запись\n` +
        `HW ${hw}${isLeader ? ' — потребители видят записи ниже него' : ' — как его знает эта реплика'}`
      }
    >
      <span className={`kf-role ${isLeader ? 'is-leader' : inIsr ? 'is-isr' : 'is-out'}`}>{role}</span>
      <span className="kf-log">
        {from > 0 && <i className="kf-log-more">…{from}</i>}
        {log.slice(from).map((e, i) => {
          const off = from + i
          const rec = w.recs[e.rec - 1]!
          // Та же запись раньше в логе — значит, эта копия записана повтором.
          const dup = log.findIndex((x) => x.rec === e.rec) < off
          return (
            <span key={off} className="kf-entry">
              {off === hw && off > from && <span className="kf-hw" title={`HW = ${hw}`} />}
              <RecChip rec={rec} sc={sc} focus={focus} hl={hl} onFocus={onFocus} variant={dup ? 'dup' : off >= hw ? 'pending' : undefined} />
            </span>
          )
        })}
        {log.length === 0 && <i className="kf-empty">пусто</i>}
      </span>
      <span className="kf-leo">
        HW {hw} · LEO {log.length}
      </span>
    </div>
  )
}

function Cluster(props: ZoneProps) {
  const { w, hl } = props
  return (
    <section className="kf-zone">
      <span className="board-label">кластер · партиции × брокеры · фактор репликации {w.config.replicationFactor}, min.insync {w.config.minInsyncReplicas}</span>
      <div className="kf-grid" style={{ '--brokers': w.brokers.length } as React.CSSProperties}>
        <span className="kf-corner" />
        {w.brokers.map((b) => {
          const leads = w.partitions.filter((p) => p.leader === b.id).length
          return (
            <span
              key={b.id}
              className={`kf-broker ${b.alive ? '' : 'is-dead'} ${b.slow > 1 ? 'is-slow' : ''} ${hl.brokers.has(b.id) ? 'is-hl' : ''}`}
              title={`B${b.id}: ${b.alive ? 'жив' : 'мёртв'}${b.slow > 1 ? `, реплицирует в ${b.slow} раз медленнее` : ''}, лидер ${leads} партиций`}
            >
              <b>B{b.id}</b>
              {!b.alive && <em>упал{b.detectAt !== null && b.detectAt > w.tick ? ` · заметят через ${b.detectAt - w.tick}т` : ''}</em>}
              {b.alive && b.slow > 1 && <em>медленный ×{b.slow}</em>}
            </span>
          )
        })}
        {w.partitions.map((p) => (
          <PartitionRow key={p.id} {...props} p={p} />
        ))}
      </div>
    </section>
  )
}

function PartitionRow(props: ZoneProps & { p: Partition }) {
  const { w, p, hl } = props
  return (
    <>
      <span className={`kf-pname ${hl.partitions.has(p.id) ? 'is-hl' : ''}`}>
        <b>p{p.id}</b>
        <small>
          {p.leader === null ? 'нет лидера' : `эпоха ${p.epoch}`}
          <br />
          ISR {p.isr.map((b) => `B${b}`).join(',')}
        </small>
      </span>
      {w.brokers.map((b) => (
        <ReplicaCell key={b.id} {...props} b={b.id} />
      ))}
    </>
  )
}

/* ─────────────────────────── группы ─────────────────────────── */

function lag(w: KafkaWorld, g: Group, p: number): number {
  return Math.max(0, w.partitions[p]!.hw - (g.committed[p] ?? 0))
}

function Groups({ w, sc, hl, focus, onFocus }: ZoneProps) {
  return (
    <section className="kf-zone">
      <span className="board-label">группы потребителей</span>
      <div className="kf-groups">
        {w.groups.map((g) => {
          const members = w.consumers.filter((c) => c.group === g.name)
          const total = w.partitions.reduce((n, p) => n + lag(w, g, p.id), 0)
          return (
            <div key={g.name} className="kf-card">
              <header>
                <strong>{g.name}</strong>
                <span className="kf-muted">поколение {g.generation}</span>
                {g.rebalanceUntil !== null && (
                  <span className="chan-badge" title="Ребалансировка: партиции никто не читает">
                    ребалансировка · {g.rebalanceUntil - w.tick}т
                  </span>
                )}
                <span className={`kf-lag ${total > 8 ? 'is-high' : ''}`} title="Лаг: HW минус закоммиченный оффсет, сумма по партициям">
                  лаг {total}
                </span>
              </header>
              <div className="kf-members">
                {members.map((c) => (
                  <div key={c.id} className={`kf-member state-${c.state} ${hl.consumers.has(c.id) ? 'is-hl' : ''}`}>
                    <span className="kf-member-name">
                      <b>{c.name}</b>
                      <small>
                        {c.state === 'dead'
                          ? 'упал'
                          : c.state === 'waiting'
                            ? 'ждёт партиций'
                            : c.assigned.length === 0
                              ? 'без партиций'
                              : c.assigned.map((x) => `p${x}`).join(', ')}
                      </small>
                    </span>
                    {c.processing && (
                      <span className="kf-proc" title="Обрабатывается прямо сейчас">
                        ⚙ <RecChip rec={w.recs[c.processing.rec - 1]!} sc={sc} focus={focus} hl={hl} onFocus={onFocus} suffix={`${c.processing.left}т`} />
                      </span>
                    )}
                    {c.buffer.length > 0 && (
                      <span className="kf-buf" title="Отдано poll'ом, ждёт обработки">
                        {c.buffer.map((x) => (
                          <RecChip key={`${x.partition}-${x.offset}`} rec={w.recs[x.rec - 1]!} sc={sc} focus={focus} hl={hl} onFocus={onFocus} />
                        ))}
                      </span>
                    )}
                    <small className="kf-muted">обработал {c.processedCount}</small>
                  </div>
                ))}
              </div>
              <div className="kf-offsets">
                {w.partitions.map((p) => {
                  const owner = members.find((c) => c.state === 'active' && c.assigned.includes(p.id))
                  const l = lag(w, g, p.id)
                  return (
                    <span key={p.id} className={`kf-off ${l > 8 ? 'is-high' : ''}`} title={`p${p.id}: закоммичено ${g.committed[p.id] ?? 0}, HW ${p.hw}, лаг ${l}${owner ? `, читает ${owner.name}` : ', никто не читает'}`}>
                      p{p.id} <b>{g.committed[p.id] ?? 0}</b>/{p.hw}
                      <i style={{ width: `${Math.min(100, l * 8)}%` }} />
                    </span>
                  )
                })}
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}

/* ─────────────────────────── вся схема ─────────────────────────── */

interface ZoneProps {
  w: KafkaWorld
  sc: KafkaScenario
  hl: Highlight
  focus: number | null
  onFocus: (id: number) => void
}

export function Board({ world, events, scenario, highlight, focus, onFocus }: {
  world: KafkaWorld
  events: KafkaEvent[]
  scenario: KafkaScenario
  highlight: Highlight
  focus: number | null
  onFocus: (id: number) => void
}) {
  const props: ZoneProps = { w: world, sc: scenario, hl: highlight, focus, onFocus }
  return (
    <div className="board kf-board">
      <PathStrip events={events} acks={world.config.acks} />
      <div className="kf-top">
        <Producers {...props} />
        <Network {...props} />
      </div>
      <Cluster {...props} />
      <Groups {...props} />
    </div>
  )
}
