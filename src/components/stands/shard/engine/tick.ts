import type { Request, Sample, Shard } from './types.ts'
import type { Ctx } from './world.ts'
import { FIRST_MILESTONE, busyOf, emit, emptySample, geometric, hash, loadOf, multiplierAt, newShard, ownerOf, poisson, random, sampleKey } from './world.ts'

/**
 * Один тик. Порядок фаз:
 *
 *   1. решардирование по расписанию: часть ключей меняет владельца;
 *   2. переезд ключей — по migrateRate за тик;
 *   3. шарды отдают готовые ответы, веерный запрос ждёт последнего из них;
 *   4. приходят запросы и раскладываются по шардам;
 *   5. свободные воркеры берут запросы из очередей;
 *   6. тик пишется в журнал.
 */
export function tick(ctx: Ctx): void {
  const w = ctx.world
  w.tick++
  const x = multiplierAt(ctx.scenario, w.tick)
  if (x !== w.x) {
    emit(ctx, 'load.change', {}, { from: w.x, to: x, rate: w.config.rate * x, shards: w.shards.length })
    w.x = x
  }
  const sample = emptySample(x, w.shards.length)
  reshard(ctx)
  migrate(ctx)
  complete(ctx, sample)
  arrive(ctx, sample)
  for (const s of w.shards) dispatch(ctx, s)
  account(ctx, sample)
}

/* ─────────────────────────── решардирование ─────────────────────────── */

function reshard(ctx: Ctx): void {
  const w = ctx.world
  for (const f of ctx.scenario.faults) {
    if (f.at !== w.tick) continue
    const from = w.shards.length
    const to = f.kind === 'add' ? from + 1 : Math.max(1, from - 1)
    if (to === from) continue
    if (to > from) {
      w.shards.push(newShard(from, w.config.workers))
      w.qMilestone.push(FIRST_MILESTONE)
    }
    const dropped = to < from ? w.shards.pop()! : null
    let moved = 0
    for (let k = 0; k < w.config.keys; k++) {
      const next = ownerOf(w.config, k, to)
      if (next === w.owner[k]) continue
      // Ключ переезжает: пока он в пути, запрос к нему стоит дороже.
      if (w.moving[k] === undefined) w.moving[k] = w.owner[k]!
      w.owner[k] = next
      moved++
    }
    for (const s of w.shards) s.owns = 0
    for (const o of w.owner) w.shards[o]!.owns++
    // Запросы убранного шарда возвращаются тем, кто теперь владеет их ключами.
    if (dropped) {
      for (const q of [...dropped.queue, ...dropped.workers.filter((x): x is Request => x !== null)]) {
        const target = q.key < 0 ? w.shards[0]! : w.shards[w.owner[q.key]!]!
        q.startAt = null
        q.finishAt = null
        target.queue.push(q)
      }
      w.qMilestone.pop()
    }
    emit(ctx, 'reshard.start', {}, {
      from,
      to,
      moved,
      share: moved / w.config.keys,
      scheme: w.config.scheme,
      vnodes: w.config.vnodes,
      rate: w.config.migrateRate,
      kind: f.kind,
    })
  }
}

function migrate(ctx: Ctx): void {
  const w = ctx.world
  // Переезжают вперемешку, а не по порядку ключей: иначе самые популярные
  // уезжали бы первыми и переезд был бы незаметен.
  const keys = Object.keys(w.moving).sort((a, b) => hash(Number(a), 99) - hash(Number(b), 99))
  if (keys.length === 0) return
  const n = Math.min(w.config.migrateRate, keys.length)
  for (let i = 0; i < n; i++) {
    const k = Number(keys[i])
    delete w.moving[k]
    w.stats.moved++
    emit(ctx, 'reshard.move', { key: [k], shard: [w.owner[k]!] }, { key: k, to: w.shards[w.owner[k]!]?.name ?? '', left: keys.length - i - 1 })
  }
  if (Object.keys(w.moving).length === 0) {
    emit(ctx, 'reshard.done', {}, { moved: w.stats.moved, shards: w.shards.length })
  }
}

/* ─────────────────────────────── ответы ─────────────────────────────── */

function complete(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  for (const s of w.shards) {
    s.workers.forEach((q, i) => {
      if (!q || q.finishAt !== w.tick) return
      s.workers[i] = null
      s.served++
      if (q.group === null) {
        w.stats.done++
        sample.done++
        ctx.log.done.push({ t: w.tick, lat: w.tick - q.bornAt, shard: s.id, kind: 'point', migrating: q.migrating })
        emit(ctx, 'req.done', { shard: [s.id], key: [q.key] }, { shard: s.name, key: q.key, lat: w.tick - q.bornAt, wait: (q.startAt ?? w.tick) - q.arriveAt, migrating: q.migrating })
        return
      }
      const g = w.groups.find((x) => x.id === q.group)
      if (!g) return
      g.left--
      if (g.left > 0) return
      w.groups = w.groups.filter((x) => x !== g)
      w.stats.done++
      w.stats.scatter++
      sample.done++
      ctx.log.done.push({ t: w.tick, lat: w.tick - g.bornAt, shard: s.id, kind: 'scatter', migrating: false })
      emit(ctx, 'scatter.done', { shard: [s.id] }, { shards: g.shards, lat: w.tick - g.bornAt, slowest: s.name })
    })
  }
}

/* ─────────────────────────────── запросы ─────────────────────────────── */

function push(ctx: Ctx, shard: Shard, key: number, group: number | null, bornAt: number, migrating: boolean): void {
  const w = ctx.world
  shard.queue.push({ id: w.nextId++, key, group, bornAt, arriveAt: w.tick, startAt: null, finishAt: null, migrating })
  w.stats.subRequests++
}

function arrive(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  const c = w.config
  const n = poisson(w, c.rate * w.x)
  for (let i = 0; i < n; i++) {
    const key = sampleKey(w)
    const scatter = random(w, 'key') < c.scatterShare / 100
    sample.arrived++
    w.stats.requests++
    if (scatter) {
      const g = { id: w.nextGroup++, bornAt: w.tick, left: w.shards.length, shards: w.shards.length }
      w.groups.push(g)
      for (const s of w.shards) push(ctx, s, -1, g.id, w.tick, false)
      emit(ctx, 'scatter.start', { shard: w.shards.map((s) => s.id) }, { shards: w.shards.length, share: c.scatterShare })
      continue
    }
    const shard = w.shards[w.owner[key]!]!
    const migrating = w.moving[key] !== undefined
    if (migrating) {
      w.stats.migratingHits++
      emit(ctx, 'req.migrating', { shard: [shard.id], key: [key] }, {
        key,
        to: shard.name,
        from: w.shards[w.moving[key]!]?.name ?? `s${w.moving[key]! + 1}`,
        left: Object.keys(w.moving).length,
      })
    } else {
      emit(ctx, 'req.route', { shard: [shard.id], key: [key] }, { key, shard: shard.name, scheme: c.scheme, queue: shard.queue.length })
    }
    push(ctx, shard, key, null, w.tick, migrating)
  }
}

function dispatch(ctx: Ctx, s: Shard): void {
  const w = ctx.world
  s.workers.forEach((busy, i) => {
    if (busy) return
    const q = s.queue.shift()
    if (!q) return
    q.startAt = w.tick
    // Пока ключ не переехал, шард сначала забирает его у старого владельца — это вдвое дороже.
    q.finishAt = w.tick + geometric(w, w.config.serviceTime) * (q.migrating ? 2 : 1)
    s.workers[i] = q
  })
}

/* ─────────────────────────────── учёт ─────────────────────────────── */

function account(ctx: Ctx, sample: Sample): void {
  const w = ctx.world
  sample.load = w.shards.map((s) => loadOf(s))
  sample.migrating = Object.keys(w.moving).length
  ctx.log.series.push(sample)

  const loads = sample.load
  const max = Math.max(...loads)
  const hot = loads.indexOf(max)
  w.stats.maxQueue = Math.max(w.stats.maxQueue, ...w.shards.map((s) => s.queue.length))
  if (!w.skewed && max >= 8 && Math.min(...loads) <= 1 && loads.length > 1) {
    w.skewed = true
    emit(ctx, 'shard.skew', { shard: [hot] }, {
      shard: w.shards[hot]?.name ?? '',
      load: max,
      others: loads.filter((_, i) => i !== hot).join(', '),
      scheme: w.config.scheme,
      hotShare: w.config.hotShare,
    })
  } else if (w.skewed && max < 4) {
    w.skewed = false
  }
  w.shards.forEach((s, i) => {
    const q = s.queue.length
    if (q === 0) w.qMilestone[i] = FIRST_MILESTONE
    else if (q >= (w.qMilestone[i] ?? FIRST_MILESTONE)) {
      emit(ctx, 'queue.grow', { shard: [s.id] }, { shard: s.name, len: q, owns: s.owns, busy: busyOf(s), migrating: sample.migrating })
      while ((w.qMilestone[i] ?? FIRST_MILESTONE) <= q) w.qMilestone[i] = (w.qMilestone[i] ?? FIRST_MILESTONE) * 2
    }
  })
}
