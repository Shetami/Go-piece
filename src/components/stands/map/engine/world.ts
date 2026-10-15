import type { Ctx } from './ctx.ts'
import { emit } from './ctx.ts'
import { h1, hashKey } from './hash.ts'
import {
  createLegacy,
  legacyCapacity,
  legacyClear,
  legacyCollect,
  legacyDelete,
  legacyFind,
  legacyHas,
  legacyHash,
  legacyIterNext,
  legacyIterStart,
  legacyPut,
} from './legacy.ts'
import { opCode } from './script.ts'
import {
  createSwiss,
  swissCapacity,
  swissClear,
  swissCollect,
  swissDelete,
  swissFind,
  swissHas,
  swissHash,
  swissIterNext,
  swissIterStart,
  swissPut,
  tableFor,
} from './swiss.ts'
import type { LocatedOp, MapConfig, MapScenario, MapStats, MapWorld } from './types.ts'
import { DEFAULT_MAP_CONFIG } from './types.ts'

export function resolveConfig(partial: Partial<MapConfig> | undefined): MapConfig {
  return { ...DEFAULT_MAP_CONFIG, ...partial }
}

const emptyStats = (): MapStats => ({
  ops: 0,
  probes: 0,
  compares: 0,
  falseMatches: 0,
  maxProbe: 0,
  moved: 0,
  grows: 0,
  splits: 0,
  dirGrows: 0,
  prunes: 0,
  overflows: 0,
  evacuated: 0,
  iterVisited: 0,
  iterYields: 0,
})

export function createWorld(scenario: MapScenario): MapWorld {
  return {
    tick: 0,
    config: resolveConfig(scenario.config),
    swiss: null,
    legacy: null,
    iter: null,
    nextId: 1,
    fillNext: {},
    collideNext: 1,
    lastKeys: [],
    stats: emptyStats(),
    finished: false,
  }
}

/** Есть ли ключ в карте — без следа и счётчиков. */
export function hasKey(w: MapWorld, key: string): boolean {
  if (w.swiss) return swissHas(w.swiss, key)
  if (w.legacy) return legacyHas(w.legacy, key)
  return false
}

export function mapLen(w: MapWorld): number {
  return w.swiss?.used ?? w.legacy?.count ?? 0
}

export function mapCapacity(w: MapWorld): number {
  if (w.swiss) return swissCapacity(w.swiss)
  if (w.legacy) return legacyCapacity(w.legacy)
  return 0
}

/** Разворачивает команду-пачку в простые. collide нужен мир: он подбирает ключи под текущую таблицу. */
function expand(ctx: Ctx, op: LocatedOp): LocatedOp[] | null {
  const w = ctx.world
  const line = op.line
  switch (op.kind) {
    case 'fill': {
      const keys: string[] = []
      let n = w.fillNext[op.prefix] ?? 1
      while (keys.length < op.count) {
        const k = `${op.prefix}${n++}`
        if (!hasKey(w, k)) keys.push(k)
      }
      w.fillNext[op.prefix] = n
      w.lastKeys = keys
      return keys.map((key) => ({ kind: 'put', key, line }))
    }
    case 'collide': {
      const keys = collideKeys(w, op.count, op.target)
      w.lastKeys = keys
      return keys.map((key) => ({ kind: 'put', key, line }))
    }
    case 'last': {
      const keys = op.n === null ? w.lastKeys : op.n > 0 ? w.lastKeys.slice(0, op.n) : w.lastKeys.slice(op.n)
      return keys.map((key) => ({ kind: op.op, key, line }))
    }
    case 'next':
      if (op.count === 'all' || op.count === 1) return null
      return Array.from({ length: op.count }, () => ({ kind: 'next', count: 1, line }))
    default:
      return null
  }
}

/**
 * Подбор ключей c1, c2, … с одной стартовой группой (корзиной).
 * Так злоумышленник строил бы hash flooding, если бы знал соль.
 */
function collideKeys(w: MapWorld, count: number, target?: number): string[] {
  let where: ((hash: number) => string) | null = null
  if (w.swiss) {
    const m = w.swiss
    where =
      m.dir.length === 0
        ? () => 'small'
        : (hash) => {
            const t = tableFor(m, hash)
            return `${t.id}:${h1(hash) & (t.groups.length - 1)}`
          }
  } else if (w.legacy) {
    const h = w.legacy
    where = (hash) => String(hash & ((1 << h.B) - 1))
  }
  if (!where) return Array.from({ length: count }, () => `c${w.collideNext++}`)
  const seed = w.swiss?.seed ?? w.legacy!.hash0

  let want: string | null = null
  if (target !== undefined) {
    if (w.swiss && w.swiss.dir.length > 0) {
      const t = w.swiss.tables[w.swiss.dir[0]!]!
      want = `${t.id}:${target % t.groups.length}`
    } else if (w.legacy) want = String(target % (1 << w.legacy.B))
  }

  const keys: string[] = []
  for (let guard = 0; keys.length < count && guard < 500_000; guard++) {
    const k = `c${w.collideNext++}`
    if (hasKey(w, k)) continue
    const at = where(hashKey(k, seed))
    want ??= at
    if (at === want) keys.push(k)
  }
  return keys
}

const nilMap = (w: MapWorld) => !w.swiss && !w.legacy

/** Выполнить одну простую команду. */
function execute(ctx: Ctx, op: LocatedOp): void {
  const w = ctx.world
  ctx.trace = { code: opCode(op), probes: [], outcome: '' }
  if (op.kind === 'put') ctx.trace.code = opCode({ ...op, val: op.val ?? w.tick })

  if (op.kind === 'make') {
    const { impl, hint, nilMap: isNil, maxTableSlots } = w.config
    if (isNil) {
      ctx.trace.code = 'var m map[string]int'
      ctx.trace.outcome = 'карта не создана — это nil'
      emit(ctx, 'map.make', { nil: true, impl })
      return
    }
    ctx.trace.code = hint > 0 ? `m := make(map[string]int, ${hint})` : 'm := make(map[string]int)'
    if (impl === 'swiss') {
      const m = (w.swiss = createSwiss(ctx, hint))
      const tables = m.dir.length
      ctx.trace.outcome = tables > 0 ? `сразу ${tables} табл. по ${m.tables[m.dir[0]!]!.capacity} слотов` : 'маленькая карта: группа появится при первой записи'
      emit(ctx, 'map.make', { impl, hint, tables, capacity: swissCapacity(m), globalDepth: m.globalDepth, max: maxTableSlots })
    } else {
      const h = (w.legacy = createLegacy(ctx, hint))
      ctx.trace.outcome = h.buckets === null ? 'B = 0: корзина появится при первой записи' : `B = ${h.B}: сразу ${1 << h.B} корзин(ы)`
      emit(ctx, 'map.make', { impl, hint, B: h.B, buckets: h.buckets === null ? 0 : 1 << h.B })
    }
    return
  }

  if (nilMap(w)) {
    switch (op.kind) {
      case 'put':
        ctx.trace.outcome = 'паника: запись в nil-карту'
        emit(ctx, 'nil.write', { key: op.key })
        w.finished = true
        w.finishReason = 'panic'
        w.panic = 'assignment to entry in nil map'
        return
      case 'race':
        ctx.trace.outcome = 'паника: запись в nil-карту'
        emit(ctx, 'nil.write', { key: 'k' })
        w.finished = true
        w.finishReason = 'panic'
        w.panic = 'assignment to entry in nil map'
        return
      default: {
        const what = op.kind === 'get' ? 'get' : op.kind === 'del' ? 'del' : op.kind === 'clear' ? 'clear' : 'range'
        ctx.trace.outcome =
          what === 'get' ? 'нулевое значение и ok == false' : what === 'range' ? 'ноль итераций' : 'ничего не делает'
        emit(ctx, 'nil.read', { op: what, key: 'key' in op ? op.key : null })
        return
      }
    }
  }

  switch (op.kind) {
    case 'put': {
      const val = op.val ?? w.tick
      if (w.swiss) swissPut(ctx, w.swiss, op.key, val)
      else legacyPut(ctx, w.legacy!, op.key, val)
      break
    }
    case 'get': {
      ctx.trace.key = op.key
      let val: number | null = null
      if (w.swiss) {
        const m = w.swiss
        const loc = swissFind(ctx, m, op.key, (ctx.trace.hash = swissHash(ctx, m, op.key)))
        if (loc) {
          val = (loc.table ? loc.table.groups[loc.group]! : m.small!).slots[loc.slot]!.val
          ctx.trace.hit = { table: loc.table?.id ?? null, group: loc.group, slot: loc.slot }
        }
      } else {
        const h = w.legacy!
        const loc = legacyFind(ctx, h, op.key, (ctx.trace.hash = legacyHash(ctx, h, op.key)))
        if (loc) {
          val = h.arrays[loc.arr]![loc.bucket]!.pages[loc.page]![loc.slot]!.val
          ctx.trace.hit = loc
        }
      }
      if (val !== null) {
        ctx.trace.outcome = `нашли: ${val}`
        emit(ctx, 'get.hit', { key: op.key, val, probes: ctx.trace.probes.length, legacy: !w.swiss })
      } else {
        ctx.trace.outcome = 'нет такого ключа: 0, false'
        const last = ctx.trace.probes.at(-1)
        emit(ctx, 'get.miss', {
          key: op.key,
          probes: ctx.trace.probes.length,
          stoppedOnEmpty: Boolean(last?.stop),
          empty: mapLen(w) === 0,
          nan: op.key === 'NaN',
          legacy: !w.swiss,
        })
      }
      break
    }
    case 'del':
      if (w.swiss) swissDelete(ctx, w.swiss, op.key)
      else legacyDelete(ctx, w.legacy!, op.key)
      break
    case 'clear':
      if (w.swiss) swissClear(ctx, w.swiss)
      else legacyClear(ctx, w.legacy!)
      break
    case 'range': {
      const it = w.swiss ? swissIterStart(ctx, w.swiss) : legacyIterStart(ctx, w.legacy!)
      const id = w.nextId++
      w.iter = { ...it, id, visited: 0, yielded: [], done: false, startTick: w.tick }
      ctx.trace.outcome = 'итератор выбрал случайную точку старта'
      emit(ctx, 'iter.start', {
        len: mapLen(w),
        capacity: mapCapacity(w),
        legacy: !w.swiss,
        ...(it.impl === 'swiss'
          ? { entryOffset: it.entryOffset & 0xff, small: it.small, tables: w.swiss!.dir.length }
          : { startBucket: it.startBucket, offset: it.offset, buckets: 1 << it.B }),
      })
      break
    }
    case 'next':
      iterNext(ctx)
      break
    case 'race': {
      const msg =
        op.mode === 'write'
          ? 'concurrent map writes'
          : op.mode === 'read'
            ? 'concurrent map read and map write'
            : 'concurrent map iteration and map write'
      if (w.swiss) w.swiss.writing ^= 1
      else w.legacy!.writing ^= 1
      ctx.trace.outcome = 'вторая горутина видит флаг writing — fatal error'
      emit(ctx, 'race', { mode: op.mode, msg })
      w.finished = true
      w.finishReason = 'fatal'
      w.panic = msg
      break
    }
    default:
      break
  }
}

function iterNext(ctx: Ctx): void {
  const w = ctx.world
  const it = w.iter
  ctx.trace.code = '    // следующая итерация'
  ctx.trace.iter = true
  if (!it || it.done) {
    ctx.trace.outcome = 'обход не начат — сначала range'
    return
  }
  const step = it.impl === 'swiss' ? swissIterNext(w.swiss!, it as never) : legacyIterNext(w.legacy!, it as never)
  if (step.kind === 'done') {
    it.done = true
    w.stats.iterVisited += step.visited
    ctx.trace.outcome = `обход закончен: ${it.yielded.length} записей за ${it.visited} проверенных слотов`
    emit(ctx, 'iter.done', {
      yielded: it.yielded.length,
      visited: it.visited,
      len: mapLen(w),
      capacity: mapCapacity(w),
      order: it.yielded.map((y) => y.key).join(' '),
    })
    return
  }
  const where =
    'table' in step
      ? { table: step.table, group: step.group, slot: step.slot }
      : { arr: step.arr, bucket: step.bucket, page: step.page, slot: step.slot }
  ctx.trace.hit = where
  if (step.kind === 'skip') {
    ctx.trace.key = step.key
    ctx.trace.outcome = `${step.key} удалён, пока шёл обход — пропускаем`
    emit(ctx, 'iter.skip', { key: step.key, legacy: it.impl === 'legacy' })
    return
  }
  it.yielded.push(step.y)
  w.stats.iterYields++
  w.stats.iterVisited += step.visited
  ctx.trace.key = step.y.key
  ctx.trace.code = `    k, v = ${step.y.key === 'NaN' ? 'NaN' : `"${step.y.key}"`}, ${step.y.val}`
  ctx.trace.outcome = step.y.reread ? 'ключ из старой таблицы, значение перечитано в новой' : `проверено слотов: ${step.visited}`
  emit(ctx, step.y.reread ? 'iter.reread' : 'iter.next', {
    key: step.y.key,
    val: step.y.val,
    n: it.yielded.length,
    visited: step.visited,
    legacy: it.impl === 'legacy',
  })
}

export interface Runner {
  queue: LocatedOp[]
}

/** Один тик: одна простая команда. Пачки разворачиваются, не тратя тика. */
export function tick(ctx: Ctx, run: Runner): void {
  const w = ctx.world
  let op: LocatedOp | undefined
  for (;;) {
    op = run.queue.shift()
    if (!op) {
      w.finished = true
      w.finishReason = 'all-done'
      return
    }
    // next без активного обхода — пропускаем, не тратя тик.
    if (op.kind === 'next' && (!w.iter || w.iter.done)) continue
    const more = expand(ctx, op)
    if (!more) break
    run.queue.unshift(...more)
  }

  w.tick++
  w.stats.ops++
  ctx.line = op.line
  ctx.events = []
  execute(ctx, op)
  w.stats.maxProbe = Math.max(w.stats.maxProbe, ctx.trace.probes.length)

  // `next all` возвращается в очередь, пока обход не кончится.
  if (op.kind === 'next' && op.count === 'all' && w.iter && !w.iter.done) run.queue.unshift(op)

  if (w.swiss) swissCollect(ctx, w.swiss)
  if (w.legacy) legacyCollect(ctx, w.legacy)
  if (w.iter?.done && w.tick > w.iter.startTick) {
    // Закончившийся обход ещё тик висит на схеме, потом исчезает.
    const finishedAt = ctx.events.some((e) => e.type === 'iter.done')
    if (!finishedAt) w.iter = null
  }
}

