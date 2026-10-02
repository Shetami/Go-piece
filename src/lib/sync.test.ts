import { test } from 'node:test'
import assert from 'node:assert/strict'
import { isSyncedKey, mergeState, mergeValue } from './sync.ts'

test('в аккаунт уезжает прогресс, но не черновики и не тема', () => {
  assert.ok(isSyncedKey('go-piece:progress'))
  assert.ok(isSyncedKey('go-piece:achievements'))
  assert.ok(isSyncedKey('go-piece:quiz:go-runtime/scheduler'))
  assert.ok(isSyncedKey('go-piece:practice:go/slices-append:solved'))
  assert.ok(!isSyncedKey('go-piece:practice:go/slices-append:code'))
  assert.ok(!isSyncedKey('theme'))
  assert.ok(!isSyncedKey('go-piece:owner'))
})

test('решённая задача остаётся решённой', () => {
  const k = 'go-piece:practice:t:solved'
  assert.equal(mergeValue(k, '1', ''), '1')
  assert.equal(mergeValue(k, '', '1'), '1')
  assert.equal(mergeValue(k, undefined, '1'), '1')
})

test('у теста остаётся лучший результат', () => {
  const k = 'go-piece:quiz:l'
  const low = JSON.stringify({ best: 3, total: 8 })
  const high = JSON.stringify({ best: 7, total: 8 })
  assert.equal(mergeValue(k, low, high), high)
  assert.equal(mergeValue(k, high, low), high)
})

test('прогресс объединяется, а время берётся более раннее', () => {
  const a = JSON.stringify({ read: { x: 5 }, stands: {}, terms: {}, solves: { t1: 10 }, days: ['2026-01-02'], flags: {} })
  const b = JSON.stringify({ read: { x: 3, y: 4 }, stands: { gmp: 1 }, terms: {}, solves: {}, days: ['2026-01-01'], flags: { night: 2 } })
  const merged = JSON.parse(mergeValue('go-piece:progress', a, b)!)
  assert.deepEqual(merged.read, { x: 3, y: 4 })
  assert.deepEqual(merged.stands, { gmp: 1 })
  assert.deepEqual(merged.solves, { t1: 10 })
  assert.deepEqual(merged.days, ['2026-01-01', '2026-01-02'])
  assert.deepEqual(merged.flags, { night: 2 })
})

test('слияние коммутативно и идемпотентно', () => {
  const a = {
    'go-piece:progress': JSON.stringify({ read: { b: 2, a: 1 }, days: ['2026-01-02'] }),
    'go-piece:achievements': JSON.stringify({ z: 9, y: 1 }),
    'go-piece:quiz:l': JSON.stringify({ best: 2, total: 5 }),
  }
  const b = {
    'go-piece:progress': JSON.stringify({ solves: { t: 3 }, read: { a: 0 }, days: ['2026-01-01'] }),
    'go-piece:achievements': JSON.stringify({ y: 5 }),
    'go-piece:practice:t:solved': '1',
  }
  const ab = mergeState(a, b)
  assert.deepEqual(ab, mergeState(b, a))
  assert.deepEqual(mergeState(ab, a), ab)
  assert.deepEqual(mergeState(ab, ab), ab)
})

test('битое значение не затирает целое', () => {
  const k = 'go-piece:achievements'
  const ok = JSON.stringify({ a: 1 })
  assert.equal(mergeValue(k, ok, '{oops'), ok)
  assert.equal(mergeValue(k, '{oops', ok), ok)
})
