import { bin, dirIndex, h1, h2, hex2, hex8, tophash } from '../engine/hash.ts'
import { maxGrowth, tombstones } from '../engine/swiss.ts'
import type { Bucket, Group, Iter, LegacyMap, LSlot, MapWorld, OpTrace, Probe, Slot, SwissMap, Table } from '../engine/types.ts'
import {
  CTRL_DELETED,
  CTRL_EMPTY,
  TOP_EMPTY_ONE,
  TOP_EMPTY_REST,
  TOP_EVAC_EMPTY,
  TOP_EVAC_X,
  TOP_EVAC_Y,
} from '../engine/types.ts'

/**
 * Схема карты на одном тике. Чистая проекция: получает снимок и след операции,
 * ничего не хранит.
 *
 * Сверху — операция и хэш её ключа по битам: видно, какие биты выбирают таблицу,
 * какие — группу, какие лежат в контрольном байте. Ниже — сама карта: каталог и
 * таблицы Swiss table или массив корзин старой реализации. Группы, которые
 * проверила операция, пронумерованы в порядке проб.
 */

const short = (k: string | null) => (k === null ? '' : k.length > 6 ? `${k.slice(0, 5)}…` : k)

/** Сколько раз и под каким номером операция заходила в группу (страницу). */
function probeIndex(trace: OpTrace | null, pred: (p: Probe) => boolean): { n: number; p: Probe }[] {
  if (!trace) return []
  return trace.probes.flatMap((p, i) => (pred(p) ? [{ n: i + 1, p }] : []))
}

/* ───────────────────────── Операция и хэш ───────────────────────── */

function HashBits({ hash, segments }: { hash: number; segments: { from: number; to: number; cls: string; label: string }[] }) {
  const bits = bin(hash, 32)
  const cls = (i: number) => segments.find((s) => i >= s.from && i < s.to)?.cls ?? ''
  return (
    <span className="map-bits" aria-label={`хэш ${hex8(hash)}`}>
      {bits.split('').map((b, i) => (
        <span key={i} className={cls(i)}>
          {b}
        </span>
      ))}
    </span>
  )
}

function OpCard({ w, trace }: { w: MapWorld; trace: OpTrace | null }) {
  if (!trace) {
    return (
      <div className="map-op">
        <code className="map-op-code">// карта ещё не создана</code>
      </div>
    )
  }
  const hash = trace.hash
  let segments: { from: number; to: number; cls: string; label: string }[] = []
  let legend: React.ReactNode = null

  if (hash !== undefined && w.swiss) {
    const m = w.swiss
    const gd = m.dir.length === 0 ? 0 : m.globalDepth
    const t = m.dir.length === 0 ? null : m.tables[m.dir[dirIndex(hash, gd)]!]
    const gbits = t ? Math.log2(t.groups.length) : 0
    segments = [
      { from: 0, to: gd, cls: 'bit-dir', label: 'каталог' },
      { from: 25 - gbits, to: 25, cls: 'bit-h1', label: 'группа' },
      { from: 25, to: 32, cls: 'bit-h2', label: 'H2' },
    ]
    legend = (
      <span className="map-legend">
        {gd > 0 && (
          <span className="bit-dir">
            строка каталога <b>{dirIndex(hash, gd)}</b>
          </span>
        )}
        {t && (
          <span className="bit-h1" title="H1 — всё, кроме 7 младших бит. Его младшие биты выбирают стартовую группу">
            H1 → группа <b>{h1(hash) & (t.groups.length - 1)}</b>
          </span>
        )}
        <span className="bit-h2" title="H2 — 7 младших бит: они лежат в контрольном байте слота">
          H2 <b>0x{hex2(h2(hash))}</b>
        </span>
      </span>
    )
  } else if (hash !== undefined && w.legacy) {
    const B = w.legacy.B
    segments = [
      { from: 0, to: 8, cls: 'bit-h2', label: 'tophash' },
      { from: 32 - B, to: 32, cls: 'bit-h1', label: 'корзина' },
    ]
    legend = (
      <span className="map-legend">
        <span className="bit-h2" title="Старшие 8 бит хэша — tophash: быстрая отбраковка слотов, по одному байту">
          tophash <b>0x{hex2(tophash(hash))}</b>
        </span>
        <span className="bit-h1" title="Младшие B бит выбирают корзину">
          корзина <b>{hash & ((1 << B) - 1)}</b>
        </span>
      </span>
    )
  }

  return (
    <div className="map-op">
      <code className="map-op-code">{trace.code}</code>
      <span className="map-op-outcome">{trace.outcome}</span>
      {hash !== undefined && trace.key !== undefined && (
        <span className="map-hash">
          <span className="map-hash-key">
            hash({trace.key === 'NaN' ? 'NaN' : `"${trace.key}"`}) = {hex8(hash)}
          </span>
          <HashBits hash={hash} segments={segments} />
          {legend}
        </span>
      )}
    </div>
  )
}

/* ───────────────────────── Swiss table ───────────────────────── */

function ctrlLabel(c: number): string {
  if (c === CTRL_EMPTY) return '80'
  if (c === CTRL_DELETED) return 'fe'
  return hex2(c)
}

function SwissCell({ s, i, probes, hit, iterAt }: {
  s: Slot
  i: number
  probes: { n: number; p: Probe }[]
  hit: boolean
  iterAt: boolean
}) {
  const matched = probes.some(({ p }) => p.match.includes(i))
  const cmp = probes.flatMap(({ p }) => p.compared).find((c) => c.slot === i)
  const state = s.ctrl === CTRL_EMPTY ? 'empty' : s.ctrl === CTRL_DELETED ? 'deleted' : 'full'
  const cls = [
    'map-cell',
    `is-${state}`,
    matched && 'is-match',
    cmp && (cmp.eq ? 'is-eq' : 'is-neq'),
    hit && 'is-hit',
    iterAt && 'is-iter',
  ]
    .filter(Boolean)
    .join(' ')
  const title =
    state === 'empty'
      ? `слот ${i}: empty (0x80)`
      : state === 'deleted'
        ? `слот ${i}: надгробие deleted (0xfe) — поиск идёт через него дальше`
        : `слот ${i}: ${s.key} = ${s.val}\nконтрольный байт = H2 = 0x${hex2(s.ctrl)}\nхэш ${hex8(s.hash)}` +
          (matched ? '\nH2 совпал с искомым' : '') +
          (cmp ? (cmp.eq ? '\nключ совпал' : '\nложное совпадение: ключ другой') : '')
  return (
    <span className={cls} title={title}>
      <i>{state === 'deleted' ? '✝' : ctrlLabel(s.ctrl)}</i>
      <span>{state === 'full' ? short(s.key) : ''}</span>
    </span>
  )
}

function GroupRow({ g, gi, probes, hitSlot, iterSlot, label }: {
  g: Group
  gi: number
  probes: { n: number; p: Probe }[]
  hitSlot: number | null
  iterSlot: number | null
  label?: string
}) {
  const stop = probes.some(({ p }) => p.stop)
  return (
    <div className={`map-group ${probes.length > 0 ? 'is-probed' : ''}`}>
      <span className="map-group-label" title={`группа ${gi}`}>
        {label ?? `g${gi}`}
        {probes.map(({ n }) => (
          <b key={n} className="map-probe-n" title={`проба №${n}${stop ? ' — здесь есть empty, поиск закончился' : ''}`}>
            {n}
          </b>
        ))}
      </span>
      <span className="map-slots">
        {g.slots.map((s, i) => (
          <SwissCell key={i} s={s} i={i} probes={probes} hit={hitSlot === i} iterAt={iterSlot === i} />
        ))}
      </span>
    </div>
  )
}

/** Где сейчас стоит итератор — чтобы подсветить слот. */
function iterSpot(it: (Iter & { id: number }) | null, trace: OpTrace | null) {
  if (!it || !trace?.iter || !trace.hit) return null
  return trace.hit
}

function TableCard({ t, m, trace, iterHit, held }: {
  t: Table
  m: SwissMap
  trace: OpTrace | null
  iterHit: OpTrace['hit'] | null
  held: boolean
}) {
  const tombs = tombstones(t)
  const rows = m.dir.flatMap((id, i) => (id === t.id ? [i] : []))
  const hit = trace?.hit && !iterHit && trace.hit.table === t.id ? trace.hit : null
  return (
    <section className={`map-table ${held ? 'is-old' : ''}`} style={{ '--c': `var(--wl-${t.id % 8})` } as React.CSSProperties}>
      <header>
        <strong>T{t.id}</strong>
        <span title="Ёмкость таблицы: степень двойки, кратная 8">{t.capacity} слотов</span>
        {held ? (
          <span className="map-badge">заменена — её держит итератор</span>
        ) : (
          <span title="localDepth: сколько старших бит хэша определяют ключи таблицы. Меньше globalDepth — на таблицу указывают несколько строк каталога">
            глубина {t.localDepth}
            {rows.length > 1 ? ` · строки ${rows[0]}–${rows.at(-1)}` : rows.length === 1 ? ` · строка ${rows[0]}` : ''}
          </span>
        )}
      </header>
      <div className="map-groups">
        {t.groups.map((g, gi) => (
          <GroupRow
            key={gi}
            g={g}
            gi={gi}
            probes={probeIndex(trace, (p) => p.table === t.id && p.group === gi)}
            hitSlot={hit && hit.group === gi ? hit.slot : null}
            iterSlot={iterHit && iterHit.table === t.id && iterHit.group === gi ? iterHit.slot : null}
          />
        ))}
      </div>
      <footer className="map-fill">
        <span className="map-fill-bar" title={`занято ${t.used}, надгробий ${tombs}, запас ${t.growthLeft} из ${maxGrowth(t.capacity)} (7/8 ёмкости)`}>
          <span className="fill-used" style={{ width: `${(t.used / t.capacity) * 100}%` }} />
          <span className="fill-tomb" style={{ width: `${(tombs / t.capacity) * 100}%` }} />
          <span className="fill-left" style={{ width: `${(t.growthLeft / t.capacity) * 100}%` }} />
        </span>
        <span>
          занято <b>{t.used}</b>
          {tombs > 0 && (
            <>
              {' '}
              · надгробий <b className="map-tomb-n">{tombs}</b>
            </>
          )}{' '}
          · запас <b>{t.growthLeft}</b>
        </span>
      </footer>
    </section>
  )
}

function Directory({ m, trace }: { m: SwissMap; trace: OpTrace | null }) {
  const cur = trace?.hash !== undefined && !trace.iter ? dirIndex(trace.hash, m.globalDepth) : null
  return (
    <div className="map-dir" title="Каталог: массив указателей на таблицы. Строку выбирают старшие globalDepth бит хэша">
      <span className="board-label">каталог · 2^{m.globalDepth}</span>
      {m.dir.map((id, i) => (
        <span
          key={i}
          className={`map-dir-row ${cur === i ? 'is-hl' : ''}`}
          style={{ '--c': `var(--wl-${id % 8})` } as React.CSSProperties}
        >
          <code>{m.globalDepth === 0 ? '—' : bin(i, m.globalDepth)}</code>→ T{id}
        </span>
      ))}
    </div>
  )
}

function SwissBoard({ w, m, trace }: { w: MapWorld; m: SwissMap; trace: OpTrace | null }) {
  const iterHit = iterSpot(w.iter, trace)
  const it = w.iter?.impl === 'swiss' ? w.iter : null

  if (m.dir.length === 0) {
    const held = it && it.smallGroup
    const hit = trace?.hit && !iterHit && trace.hit.table === null ? trace.hit : null
    return (
      <div className="map-area">
        <section className="map-table map-small">
          <header>
            <strong>маленькая карта</strong>
            <span>одна группа, без таблицы и каталога · dirLen = 0</span>
          </header>
          {m.small ? (
            <div className="map-groups">
              <GroupRow
                g={m.small}
                gi={0}
                label="группа"
                probes={probeIndex(trace, (p) => p.table === null)}
                hitSlot={hit ? hit.slot : null}
                iterSlot={!held && iterHit && iterHit.table === null ? iterHit.slot : null}
              />
            </div>
          ) : (
            <p className="board-empty">группа не выделена — появится при первой записи</p>
          )}
        </section>
      </div>
    )
  }

  const live = m.dir.filter((id, i) => m.dir.indexOf(id) === i).map((id) => m.tables[id]!)
  const old = Object.values(m.tables).filter((t) => t.replaced)
  return (
    <div className="map-area map-area-dir">
      <Directory m={m} trace={trace} />
      <div className="map-tables">
        {live.map((t) => (
          <TableCard key={t.id} t={t} m={m} trace={trace} iterHit={iterHit} held={false} />
        ))}
        {old.map((t) => (
          <TableCard key={t.id} t={t} m={m} trace={trace} iterHit={iterHit} held />
        ))}
        {it?.smallGroup && (
          <section className="map-table map-small is-old">
            <header>
              <strong>старая группа</strong>
              <span className="map-badge">маленькая карта до роста — её держит итератор</span>
            </header>
            <div className="map-groups">
              <GroupRow
                g={it.smallGroup}
                gi={0}
                label="группа"
                probes={[]}
                hitSlot={null}
                iterSlot={iterHit && iterHit.table === null ? iterHit.slot : null}
              />
            </div>
          </section>
        )}
      </div>
    </div>
  )
}

/* ───────────────────────── Старая карта ───────────────────────── */

function topLabel(t: number): string {
  switch (t) {
    case TOP_EMPTY_REST:
      return '·'
    case TOP_EMPTY_ONE:
      return '∘'
    case TOP_EVAC_X:
      return 'X'
    case TOP_EVAC_Y:
      return 'Y'
    case TOP_EVAC_EMPTY:
      return '—'
    default:
      return hex2(t)
  }
}

function LegacyCell({ s, i, probes, hit, iterAt }: {
  s: LSlot
  i: number
  probes: { n: number; p: Probe }[]
  hit: boolean
  iterAt: boolean
}) {
  const full = s.top > TOP_EVAC_EMPTY
  const evac = s.top === TOP_EVAC_X || s.top === TOP_EVAC_Y
  const matched = probes.some(({ p }) => p.match.includes(i))
  const cmp = probes.flatMap(({ p }) => p.compared).find((c) => c.slot === i)
  const state = full ? 'full' : evac ? 'evac' : s.top === TOP_EMPTY_ONE ? 'emptyone' : 'empty'
  const cls = ['map-cell', `is-${state}`, matched && 'is-match', cmp && (cmp.eq ? 'is-eq' : 'is-neq'), hit && 'is-hit', iterAt && 'is-iter']
    .filter(Boolean)
    .join(' ')
  const title = full
    ? `слот ${i}: ${s.key} = ${s.val}\ntophash 0x${hex2(s.top)}`
    : evac
      ? `слот ${i}: ${s.key} переехал в ${s.top === TOP_EVAC_X ? 'X — корзину с тем же номером' : 'Y — корзину с номером + размер старого массива'}`
      : s.top === TOP_EMPTY_ONE
        ? `слот ${i}: emptyOne — пусто, но дальше в цепочке могут быть записи`
        : s.top === TOP_EVAC_EMPTY
          ? `слот ${i}: был пуст при эвакуации`
          : `слот ${i}: emptyRest — отсюда и до конца цепочки пусто, поиск останавливается`
  return (
    <span className={cls} title={title}>
      <i>{topLabel(s.top)}</i>
      <span>{full || evac ? short(s.key) : ''}</span>
    </span>
  )
}

function BucketRow({ b, bi, arr, trace, iterHit, B }: {
  b: Bucket
  bi: number
  arr: number
  trace: OpTrace | null
  iterHit: OpTrace['hit'] | null
  B: number
}) {
  const hit = trace?.hit && !iterHit && trace.hit.arr === arr && trace.hit.bucket === bi ? trace.hit : null
  return (
    <div className={`map-bucket ${b.evacuated ? 'is-evacuated' : ''}`}>
      <span className="map-group-label" title={`корзина ${bi}`}>
        <code>{B === 0 ? '0' : bin(bi, B)}</code>
      </span>
      <span className="map-chain">
        {b.pages.map((page, pi) => {
          const probes = probeIndex(trace, (p) => p.arr === arr && p.bucket === bi && p.page === pi)
          return (
            <span key={pi} className={`map-page ${probes.length > 0 ? 'is-probed' : ''} ${pi > 0 ? 'is-overflow' : ''}`}>
              {pi > 0 && <em aria-hidden="true">→</em>}
              {probes.map(({ n }) => (
                <b key={n} className="map-probe-n">
                  {n}
                </b>
              ))}
              <span className="map-slots">
                {page.map((s, i) => (
                  <LegacyCell
                    key={i}
                    s={s}
                    i={i}
                    probes={probes}
                    hit={Boolean(hit && hit.page === pi && hit.slot === i)}
                    iterAt={Boolean(iterHit && iterHit.arr === arr && iterHit.bucket === bi && iterHit.page === pi && iterHit.slot === i)}
                  />
                ))}
              </span>
            </span>
          )
        })}
      </span>
    </div>
  )
}

function BucketArray({ h, arr, title, trace, iterHit, B, nevacuate }: {
  h: LegacyMap
  arr: number
  title: string
  trace: OpTrace | null
  iterHit: OpTrace['hit'] | null
  B: number
  nevacuate?: number
}) {
  return (
    <section className={`map-table map-buckets ${arr !== h.buckets ? 'is-old' : ''}`}>
      <header>
        <strong>{title}</strong>
        <span>
          {1 << B} корзин · B = {B}
          {nevacuate !== undefined && ` · перенесено до ${nevacuate}`}
        </span>
      </header>
      <div className="map-groups">
        {h.arrays[arr]!.map((b, bi) => (
          <BucketRow key={bi} b={b} bi={bi} arr={arr} trace={trace} iterHit={iterHit} B={B} />
        ))}
      </div>
    </section>
  )
}

function LegacyBoard({ w, h, trace }: { w: MapWorld; h: LegacyMap; trace: OpTrace | null }) {
  const iterHit = iterSpot(w.iter, trace)
  const it = w.iter?.impl === 'legacy' ? w.iter : null
  if (h.buckets === null) {
    return (
      <div className="map-area">
        <p className="board-empty">B = 0: корзина не выделена — появится при первой записи</p>
      </div>
    )
  }
  const oldB = h.sameSizeGrow ? h.B : h.B - 1
  const extra = it && it.arr !== h.buckets && it.arr !== h.oldbuckets && h.arrays[it.arr] ? it.arr : null
  return (
    <div className="map-area map-area-legacy">
      {h.oldbuckets !== null && (
        <BucketArray
          h={h}
          arr={h.oldbuckets}
          title={h.sameSizeGrow ? 'старый массив (рост того же размера)' : 'старый массив'}
          trace={trace}
          iterHit={iterHit}
          B={oldB}
          nevacuate={h.nevacuate}
        />
      )}
      <BucketArray h={h} arr={h.buckets} title={h.oldbuckets !== null ? 'новый массив' : 'корзины'} trace={trace} iterHit={iterHit} B={h.B} />
      {extra !== null && it && (
        <BucketArray h={h} arr={extra} title="массив, который держит итератор" trace={trace} iterHit={iterHit} B={it.B} />
      )}
    </div>
  )
}

/* ───────────────────────── Итератор и шапка ───────────────────────── */

function IterPanel({ w }: { w: MapWorld }) {
  const it = w.iter
  if (!it) return null
  const shown = it.yielded.slice(-24)
  return (
    <div className={`map-iter ${it.done ? 'is-done' : ''}`}>
      <span className="board-label">
        for range m {it.done ? '· закончен' : ''} · отдано {it.yielded.length} · проверено слотов {it.visited}
      </span>
      <span className="map-iter-seq">
        {it.yielded.length > shown.length && <span className="board-more">…</span>}
        {shown.map((y, i) => (
          <span key={i} className={`map-yield ${y.reread ? 'is-reread' : ''}`} title={y.reread ? 'ключ из старой таблицы, значение перечитано в новой' : undefined}>
            {y.key}:{y.val}
          </span>
        ))}
        {it.yielded.length === 0 && <i className="board-empty">ещё ничего не отдано</i>}
      </span>
    </div>
  )
}

function Header({ w }: { w: MapWorld }) {
  if (w.swiss) {
    const m = w.swiss
    const tables = new Set(m.dir).size
    return (
      <div className="map-struct" title="Поля структуры maps.Map">
        <span>
          used <b>{m.used}</b>
        </span>
        <span title="Соль хэша — своя у каждой карты, меняется, когда карта опустела">
          seed <b>{hex8(m.seed)}</b>
        </span>
        <span title="Длина каталога; 0 — маленькая карта">
          dirLen <b>{m.dir.length}</b>
        </span>
        <span>
          globalDepth <b>{m.globalDepth}</b>
        </span>
        <span>
          таблиц <b>{tables}</b>
        </span>
        <span title="Флаг «идёт запись» — ловит одновременные записи" className={m.writing ? 'metric-bad' : undefined}>
          writing <b>{m.writing}</b>
        </span>
        <span>
          clearSeq <b>{m.clearSeq}</b>
        </span>
      </div>
    )
  }
  if (w.legacy) {
    const h = w.legacy
    return (
      <div className="map-struct" title="Поля структуры hmap (до Go 1.24)">
        <span>
          count <b>{h.count}</b>
        </span>
        <span>
          B <b>{h.B}</b>
        </span>
        <span title="Корзин переполнения — грубый счётчик рантайма">
          noverflow <b>{h.noverflow}</b>
        </span>
        <span>
          hash0 <b>{hex8(h.hash0)}</b>
        </span>
        <span title="Средняя заполненность корзины; рост — при 6,5">
          нагрузка <b>{(h.count / (1 << h.B)).toFixed(1)}</b> / 6,5
        </span>
        {h.oldbuckets !== null && (
          <span className="metric-bad">
            {h.sameSizeGrow ? 'рост того же размера' : 'идёт рост'} · nevacuate <b>{h.nevacuate}</b>
          </span>
        )}
        <span className={h.writing ? 'metric-bad' : undefined}>
          hashWriting <b>{h.writing}</b>
        </span>
      </div>
    )
  }
  return (
    <div className="map-struct">
      <span>
        m == <b>nil</b>
      </span>
    </div>
  )
}

export function Board({ world: w, trace }: { world: MapWorld; trace: OpTrace | null }) {
  return (
    <div className="board map-board">
      <OpCard w={w} trace={trace} />
      <Header w={w} />
      {w.swiss && <SwissBoard w={w} m={w.swiss} trace={trace} />}
      {w.legacy && <LegacyBoard w={w} h={w.legacy} trace={trace} />}
      {!w.swiss && !w.legacy && (
        <p className="board-empty map-area">
          Карта не создана: в переменной nil. Читать можно, писать — паника.
        </p>
      )}
      <IterPanel w={w} />
      <div className="map-key" aria-label="Обозначения">
        {w.legacy ? (
          <>
            <span><i className="map-cell is-full"><i>a7</i></i> tophash занятого слота</span>
            <span><i className="map-cell is-empty"><i>·</i></i> emptyRest</span>
            <span><i className="map-cell is-emptyone"><i>∘</i></i> emptyOne</span>
            <span><i className="map-cell is-evac"><i>X</i></i> переехал</span>
          </>
        ) : (
          <>
            <span><i className="map-cell is-full"><i>1e</i></i> контрольный байт = H2</span>
            <span><i className="map-cell is-empty"><i>80</i></i> empty</span>
            <span><i className="map-cell is-deleted"><i>✝</i></i> надгробие</span>
          </>
        )}
        <span><i className="map-cell is-full is-match"><i>··</i></i> H2 совпал</span>
        <span><i className="map-cell is-full is-neq"><i>··</i></i> ложное совпадение</span>
        <span><i className="map-cell is-full is-hit"><i>··</i></i> результат</span>
        <span><b className="map-probe-n">1</b> порядок проб</span>
      </div>
    </div>
  )
}
