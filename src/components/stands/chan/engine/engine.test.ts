import { test, describe } from 'node:test'
import assert from 'node:assert/strict'

import { Simulation } from './simulation.ts'
import {
  CHAN_SCENARIOS,
  buffered,
  closing,
  deadlock,
  leak,
  pipeline,
  rendezvous,
  selectScenario,
} from './scenarios.ts'
import type { ChanScenario, ChanSpec } from './types.ts'

/**
 * Каждый тест закрепляет claim одного из сценариев: если стенд однажды начнёт
 * доказывать не то, что написано в лекции, тест это заметит.
 */

const run = (s: ChanScenario, seed = 1, ticks?: number) => new Simulation(s, seed).runToEnd(ticks ?? s.stopAfter)

const withConfig = (s: ChanScenario, config: Partial<ChanScenario['config']>): ChanScenario => ({
  ...s,
  config: { ...s.config, ...config },
})

const withChan = (s: ChanScenario, name: string, patch: Partial<ChanSpec>): ChanScenario => ({
  ...s,
  chans: s.chans.map((c) => (c.name === name ? { ...c, ...patch } : c)),
})

/** Сколько значений приняли горутины конкретной нагрузки. */
const receivedBy = (sim: Simulation, workload: number) =>
  sim.world.gs.filter((g) => g.workload === workload).reduce((n, g) => n + g.received, 0)

describe('общие свойства движка', () => {
  test('один seed даёт посимвольно одинаковую ленту событий', () => {
    for (const sc of CHAN_SCENARIOS) {
      const a = run(sc, 7)
      const b = run(sc, 7)
      assert.equal(
        JSON.stringify(a.events),
        JSON.stringify(b.events),
        `${sc.id}: два прогона с одним seed разошлись`,
      )
    }
  })

  test('ни один сценарий не падает и доходит до внятного конца', () => {
    for (const sc of CHAN_SCENARIOS) {
      const sim = run(sc)
      assert.ok(sim.finished, `${sc.id}: прогон не завершился`)
      assert.ok(sim.world.finishReason, `${sc.id}: не сказано, почему прогон кончился`)
    }
  })

  test('передачи сходятся: всё, что ушло, ушло либо из рук в руки, либо через буфер', () => {
    for (const sc of CHAN_SCENARIOS) {
      const s = run(sc).world.stats
      assert.equal(s.transfers, s.direct + s.buffered, `${sc.id}: счётчики передач разошлись`)
    }
  })

  test('в очереди канала горутина стоит не больше одного раза', () => {
    for (const sc of CHAN_SCENARIOS) {
      for (const snap of run(sc).history) {
        for (const c of snap.world.chans) {
          const ids = [...c.sendq, ...c.recvq].map((x) => x.g)
          assert.equal(new Set(ids).size, ids.length, `${sc.id}: дубль в очередях канала ${c.name}`)
        }
      }
    }
  })
})

describe('небуферизованный канал — точка встречи', () => {
  test('хранить негде: каждая передача идёт мимо буфера', () => {
    const sim = run(rendezvous)
    assert.ok(sim.world.stats.transfers > 10, 'передач слишком мало, чтобы о чём-то судить')
    assert.equal(sim.world.stats.buffered, 0, 'у канала без буфера появились буферизованные передачи')
    assert.equal(sim.directShare, 1)
  })

  test('первым паркуется тот, кто пришёл раньше, — и отправитель, и получатель', () => {
    const sim = run(rendezvous)
    assert.ok(sim.countEvents('send.block') > 0, 'отправитель ни разу не ждал получателя')
    assert.ok(sim.countEvents('recv.block') > 0, 'получатель ни разу не ждал отправителя')
  })

  test('парковка всегда кончается пробуждением: спящих в конце нет', () => {
    const sim = run(rendezvous)
    assert.equal(sim.world.stats.parks, sim.world.stats.wakeups)
  })
})

describe('буфер и обратное давление', () => {
  test('буфер побольше не приносит получателю больше работы', () => {
    const small = run(buffered)
    const big = run(withChan(buffered, 'jobs', { cap: 16 }))
    const a = receivedBy(small, 1)
    const b = receivedBy(big, 1)
    assert.ok(a > 5, `получатель почти ничего не сделал: ${a}`)
    assert.ok(
      Math.abs(a - b) <= 2,
      `буфер 4 → ${a} принятых, буфер 16 → ${b}: разница слишком велика, чтобы claim был честным`,
    )
  })

  test('но отправители начинают ждать позже — и это всё, что покупает буфер', () => {
    const small = run(buffered)
    const big = run(withChan(buffered, 'jobs', { cap: 16 }))
    const firstBlock = (sim: Simulation) => sim.events.find((e) => e.type === 'send.block')?.tick ?? Infinity
    assert.ok(
      firstBlock(big) > firstBlock(small),
      `с большим буфером отправитель уснул не позже: ${firstBlock(big)} против ${firstBlock(small)}`,
    )
    assert.ok(small.countEvents('buf.full') > 0, 'буфер ни разу не заполнился')
  })

  test('освободившуюся ячейку занимает спящий отправитель — в хвост, порядок не ломается', () => {
    const sim = run(buffered)
    assert.ok(sim.countEvents('recv.wake') > 0, 'ни разу не случился разбор полного буфера')
    for (const snap of sim.history) {
      for (const c of snap.world.chans) {
        // Пока в sendq кто-то есть, буфер обязан быть полон: иначе он бы туда положил.
        if (c.sendq.length > 0 && c.cap > 0) {
          assert.equal(c.qcount, c.cap, `тик ${snap.world.tick}: отправитель спит, а в буфере «${c.name}» есть место`)
        }
      }
    }
  })
})

describe('конвейер', () => {
  test('буфер перед медленным звеном переполняется, буфер после него простаивает', () => {
    const sim = run(pipeline)
    const before = sim.world.chans.find((c) => c.name === 'in')!
    const after = sim.world.chans.find((c) => c.name === 'out')!
    assert.equal(before.stats.maxQcount, before.cap, 'буфер перед узким местом так и не заполнился')
    assert.ok(
      after.stats.maxQcount < after.cap,
      `буфер после узкого места тоже заполнился (${after.stats.maxQcount} из ${after.cap}) — узкое место не там, где сказано`,
    )
  })

  test('ускорение звена поднимает пропускную способность всего конвейера', () => {
    const faster: ChanScenario = {
      ...pipeline,
      workloads: pipeline.workloads.map((w) =>
        w.name === 'stage' ? { ...w, phases: [w.phases[0]!, { kind: 'cpu', ticks: 1 }, w.phases[2]!] } : w,
      ),
    }
    assert.ok(
      run(faster).world.stats.transfers > run(pipeline).world.stats.transfers * 1.3,
      'убрали узкое место, а пропускная способность почти не выросла',
    )
  })
})

describe('select', () => {
  test('уснув, горутина стоит в очередях сразу всех своих каналов', () => {
    const sim = run(selectScenario)
    const blocked = sim.events.filter((e) => e.type === 'select.block')
    assert.ok(blocked.length > 0, 'select ни разу не уснул')

    // Разбудить уснувший select могут в том же тике, и к концу тика очередей уже
    // не будет. Берём первый случай, когда горутина действительно осталась спать.
    const e = blocked.find((x) => {
      const g = sim.history[x.tick]?.world.gs[x.actors.g![0]! - 1]
      return g?.state === 'waiting'
    })
    assert.ok(e, 'select ни разу не остался спать до конца тика')
    const world = sim.history[e.tick]!.world
    const gid = e.actors.g![0]!
    const queued = world.chans.filter((c) => c.recvq.some((x) => x.g === gid) || c.sendq.some((x) => x.g === gid))
    assert.equal(queued.length, Number(e.payload.queued), 'горутина встала не во все очереди')
    assert.ok(queued.length >= 3, `case было три, а очередей ${queued.length}`)
    assert.ok(queued.every((c) => c.recvq.some((x) => x.g === gid && x.fromSelect)))
  })

  test('проснувшись по одному каналу, горутина снимается со всех остальных', () => {
    const sim = run(selectScenario)
    for (const snap of sim.history) {
      for (const g of snap.world.gs) {
        if (g.state === 'waiting') continue
        for (const c of snap.world.chans) {
          assert.ok(
            !c.sendq.some((x) => x.g === g.id) && !c.recvq.some((x) => x.g === g.id),
            `тик ${snap.world.tick}: G${g.id} не спит, но осталась в очереди «${c.name}»`,
          )
        }
      }
    }
  })

  test('выбор между готовыми case случаен, а не по порядку в исходнике', () => {
    const picked = new Set<unknown>()
    for (let seed = 1; seed <= 6; seed++) {
      for (const e of run(selectScenario, seed).events) {
        if (e.type === 'select.ready') picked.add(e.payload.chan)
      }
    }
    assert.ok(picked.size > 1, `select всегда выбирал один и тот же канал: ${[...picked].join(', ')}`)
  })

  test('default не даёт уснуть: с ним горутина не паркуется ни разу', () => {
    const sim = run(selectScenario)
    const poller = sim.world.gs.find((g) => g.name === 'poller')!
    assert.equal(poller.blocks, 0, 'select с default всё-таки уснул')
    assert.ok(sim.countEvents('select.default') > 0, 'ветка default ни разу не сработала')
  })
})

describe('закрытие', () => {
  test('закрытие будит всех получателей разом', () => {
    const sim = run(closing)
    const close = sim.events.find((e) => e.type === 'chan.close')!
    const woken = Number(close.payload.receivers)
    assert.ok(woken >= 3, `закрытие разбудило всего ${woken} — сценарий перестал показывать широковещание`)
    const readyAtSameTick = sim.events.filter((e) => e.type === 'g.ready' && e.tick === close.tick)
    assert.equal(readyAtSameTick.length, woken, 'разбужены не все, кто стоял в очереди приёма')
  })

  test('приём из закрытого и пустого канала не блокирует', () => {
    const sim = run(closing)
    const close = sim.events.find((e) => e.type === 'chan.close')!
    assert.ok(sim.countEvents('recv.closed') > 0, 'никто не получил нулевое значение')
    assert.ok(
      !sim.events.some((e) => e.type === 'recv.block' && e.tick > close.tick),
      'после закрытия кто-то всё равно уснул на приёме',
    )
  })

  test('отправка в закрытый канал заканчивает программу паникой', () => {
    const sim = run(closing)
    assert.equal(sim.world.finishReason, 'panic')
    assert.equal(sim.world.panic, 'send on closed channel')
  })
})

describe('взаимная блокировка и утечка', () => {
  test('симметричные «сначала отправить» не встречаются никогда', () => {
    const sim = run(deadlock)
    assert.equal(sim.world.finishReason, 'deadlock')
    assert.equal(sim.world.stats.transfers, 0)
  })

  test('операция на nil-канале не попадает ни в одну очередь — будить оттуда некому', () => {
    const sim = run(deadlock)
    assert.ok(sim.countEvents('chan.nil') > 0, 'nil-канал никого не поймал')
    const never = sim.world.chans.find((c) => c.name === 'never')!
    assert.equal(never.sendq.length + never.recvq.length, 0, 'у nil-канала завелась очередь')
    const ghost = sim.world.gs.find((g) => g.name === 'ghost')!
    assert.equal(ghost.state, 'waiting')
  })

  test('без обнаружения deadlock программа не падает, а просто стоит', () => {
    const sim = run(withConfig(deadlock, { deadlockDetect: false }))
    assert.equal(sim.world.finishReason, 'stop-after')
    assert.equal(sim.world.gs.filter((g) => g.state === 'waiting').length, 3)
  })

  test('забытый получатель оставляет горутины висеть, и рантайм этого не видит', () => {
    const sim = run(leak)
    assert.equal(sim.world.finishReason, 'stop-after', 'утечка не должна ронять программу')
    const ev = sim.events.find((e) => e.type === 'leak')
    assert.ok(ev, 'утечка не обнаружена')
    assert.ok(Number(ev.payload.count) >= 3, `зависло всего ${ev.payload.count} горутин`)
  })

  test('буфер на число ожидаемых ответов убирает утечку целиком', () => {
    const fixed = run(withChan(leak, 'results', { cap: 4 }))
    assert.equal(fixed.countEvents('leak'), 0, 'буфера хватило на всех, а утечка всё равно объявлена')
    assert.equal(
      fixed.world.gs.filter((g) => g.name === 'worker' && g.state !== 'done').length,
      0,
      'не все отправители смогли завершиться',
    )
  })
})

describe('ручки стенда', () => {
  test('без передачи из рук в руки то же самое едет через буфер — с лишним копированием', () => {
    const on = run(pipeline)
    const off = run(withConfig(pipeline, { directHandoff: false }))
    assert.ok(on.world.stats.direct > 20, `передач из рук в руки почти нет: ${on.world.stats.direct}`)
    assert.equal(off.world.stats.direct, 0, 'передача из рук в руки выключена, но всё равно случилась')
    assert.ok(off.world.stats.buffered > on.world.stats.buffered, 'через буфер не поехало больше')
    // На небуферизованном канале выключить её нельзя: хранить значение негде.
    const unbuf = run(withConfig(rendezvous, { directHandoff: false }))
    assert.equal(unbuf.world.stats.buffered, 0)
    assert.ok(unbuf.world.stats.direct > 0)
  })

  test('без runnext разбуженная горутина ждёт дольше', () => {
    const withNext = run(pipeline)
    const without = run(withConfig(pipeline, { runnext: false }))
    assert.ok(
      without.avgWait > withNext.avgWait,
      `без runnext ожидание не выросло: ${without.avgWait.toFixed(2)} против ${withNext.avgWait.toFixed(2)}`,
    )
  })
})
