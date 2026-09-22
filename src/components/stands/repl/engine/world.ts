import type {
  Client,
  NodeState,
  ReplConfig,
  ReplEvent,
  ReplEventType,
  ReplScenario,
  ReplWorld,
  WalRecord,
} from './types.ts'
import { DEFAULT_REPL_CONFIG } from './types.ts'

export interface Ctx {
  world: ReplWorld
  events: ReplEvent[]
  scenario: ReplScenario
}

export function resolveConfig(partial: Partial<ReplConfig> | undefined): ReplConfig {
  return { ...DEFAULT_REPL_CONFIG, ...(partial ?? {}) }
}

export function createWorld(scenario: ReplScenario): ReplWorld {
  const config = resolveConfig(scenario.config)
  const kv = () => Object.fromEntries(scenario.keys.map((k) => [k, { value: 0, lsn: 0 }]))
  const node = (id: number, name: string, latency: number, replayRate: number): NodeState => ({
    id,
    name,
    role: id === 0 ? 'primary' : 'replica',
    up: true,
    broken: false,
    latency,
    replayRate,
    writeLsn: 0,
    flushLsn: 0,
    replayLsn: 0,
    kv: kv(),
    queries: [],
    conflictSince: null,
    slowUntil: 0,
    slowRate: replayRate,
    promotedAt: id === 0 ? 0 : null,
  })
  const nodes: NodeState[] = [node(0, 'primary', 0, config.replayRate)]
  for (let i = 1; i <= config.replicas; i++) {
    const spec = scenario.replicas[i - 1]
    nodes.push(node(i, spec?.name ?? `r${i}`, spec?.latency ?? config.netLatency, spec?.replayRate ?? config.replayRate))
  }
  const clients: Client[] = scenario.clients.map((c, id) => ({
    id,
    name: c.name,
    state: 'idle',
    opIdx: 0,
    nextAt: Math.max(1, c.at),
    iter: 0,
    repeatsLeft: c.repeat === 'forever' ? Number.POSITIVE_INFINITY : Math.max(0, (c.repeat ?? 1) - 1),
    waitLsn: null,
    waitSince: null,
    waitNode: null,
    ownAcked: {},
    seen: {},
    lastRead: null,
    rr: id,
    query: null,
    acked: 0,
    errors: 0,
  }))
  return {
    tick: 0,
    config,
    nodes,
    primary: 0,
    wal: [],
    nextLsn: 1,
    walOldest: 1,
    dead: [],
    view: Object.fromEntries(
      nodes.slice(1).map((n) => [n.id, { sent: 0, write: 0, flush: 0, replay: 0, feedback: null, lastReply: 0 }]),
    ),
    net: [],
    clients,
    primaryDownSince: null,
    ackedLsns: [],
    marks: clients.map(() => null),
    stats: {
      commits: 0,
      acked: 0,
      lost: 0,
      unknown: 0,
      writeErrors: 0,
      reads: 0,
      staleReads: 0,
      ownStale: 0,
      backwards: 0,
      readErrors: 0,
      commitWaitTicks: 0,
      maxCommitWait: 0,
      cancels: 0,
      maxLag: 0,
      maxRetained: 0,
      maxDead: 0,
      downTicks: 0,
    },
    finished: false,
  }
}

export function emit(ctx: Ctx, type: ReplEventType, actors: ReplEvent['actors'] = {}, payload: Record<string, unknown> = {}): void {
  ctx.events.push({ tick: ctx.world.tick, type, actors, payload })
}

export const primaryOf = (w: ReplWorld): NodeState => w.nodes[w.primary]!

/** Ведущий на месте и принимает запись. */
export const primaryUp = (w: ReplWorld) => primaryOf(w).up

export const lastLsn = (w: ReplWorld) => w.nextLsn - 1

export function recordAt(w: ReplWorld, lsn: number): WalRecord | undefined {
  // Записи идут подряд с LSN 1, поэтому индекс известен заранее.
  const r = w.wal[lsn - 1]
  return r && r.lsn === lsn ? r : w.wal.find((x) => x.lsn === lsn)
}

export const replicasOf = (w: ReplWorld) => w.nodes.filter((n) => n.id !== w.primary)

/** Отставание реплики: сколько записей WAL ведущего она ещё не проиграла. */
export const lagOf = (w: ReplWorld, n: NodeState) => Math.max(0, lastLsn(w) - n.replayLsn)

/** Отставание во времени: сколько тиков назад ведущий записал самую старую непроигранную запись. */
export function timeLagOf(w: ReplWorld, n: NodeState): number {
  const r = recordAt(w, n.replayLsn + 1)
  return r ? w.tick - r.tick : 0
}

/** Сколько записей WAL ведущий хранит сейчас. */
export const retainedOf = (w: ReplWorld) => Math.max(0, w.nextLsn - w.walOldest)
