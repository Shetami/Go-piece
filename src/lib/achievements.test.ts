import { test } from 'node:test'
import assert from 'node:assert/strict'

import { COURSES } from '../data/courses.ts'
import {
  GROUPS,
  KING,
  RANKS,
  achievementsFor,
  evaluate,
  formatBerry,
  longestStreak,
  rankFor,
  type Catalog,
  type Progress,
} from './achievements.ts'

const catalog: Catalog = {
  courses: [
    { id: 'go-runtime', title: 'Рантайм Go', lectures: ['go-runtime/a', 'go-runtime/b'] },
    { id: 'brokers', title: 'Брокеры сообщений', lectures: ['brokers/kafka'] },
  ],
  quizzes: ['go-runtime/a', 'brokers/kafka'],
  tasks: [
    { id: 'channels/a', topic: 'channels', lang: 'go', kind: 'output', level: 'easy' },
    { id: 'channels/b', topic: 'channels', lang: 'go', kind: 'bug', level: 'hard' },
    { id: 'maps/a', topic: 'maps', lang: 'go', kind: 'fix', level: 'medium' },
    { id: 'sql-basics/a', topic: 'sql-basics', lang: 'sql', kind: 'implement', level: 'medium' },
  ],
  stands: ['chan', 'gc'],
  terms: 3,
}

function progress(over: Partial<Progress> = {}): Progress {
  return {
    read: new Set(),
    quizzes: new Map(),
    solved: new Set(),
    stands: new Set(),
    terms: new Set(),
    days: [],
    bestDay: 0,
    flags: new Set(),
    ...over,
  }
}

const byId = (p: Progress, earned: Record<string, number> = {}) =>
  new Map(evaluate(p, catalog, earned).map((e) => [e.achievement.id, e]))

test('ids уникальны, группы известны, награда положительна', () => {
  const list = achievementsFor({ courses: COURSES })
  assert.equal(new Set(list.map((a) => a.id)).size, list.length)
  const groups = new Set<string>(GROUPS.map((g) => g.id))
  for (const a of list) {
    assert.ok(groups.has(a.group), `${a.id}: неизвестная группа ${a.group}`)
    assert.ok(a.bounty > 0, `${a.id}: нулевая награда`)
    assert.ok(a.title && a.quote && a.task, `${a.id}: пустой текст`)
  }
})

test('у каждого курса есть своё достижение, а не заглушка', () => {
  const list = achievementsFor({ courses: COURSES })
  for (const c of COURSES) {
    const a = list.find((x) => x.id === `course-${c.id}`)
    assert.ok(a, `нет достижения для курса ${c.id}`)
    assert.notEqual(a.title, c.title, `курс ${c.id}: нет отсылки в COURSE_FLAVOR`)
  }
})

test('без прогресса ничего не получено', () => {
  for (const e of evaluate(progress(), catalog)) assert.equal(e.unlocked, false, e.achievement.id)
})

test('первая решённая задача открывает «Первую награду», но не «Гир Секонд»', () => {
  const r = byId(progress({ solved: new Set(['channels/a']) }))
  assert.equal(r.get('task-first')!.unlocked, true)
  assert.equal(r.get('tasks-10')!.unlocked, false)
})

test('пороги не больше, чем есть на сайте: все 4 задачи — это и 10, и все', () => {
  const r = byId(progress({ solved: new Set(catalog.tasks.map((t) => t.id)) }))
  for (const id of ['tasks-10', 'tasks-50', 'tasks-all', 'all-kinds', 'topic-clear', 'topic-channels', 'topic-maps', 'sql-10']) {
    assert.equal(r.get(id)!.unlocked, true, id)
  }
  // Задач на defer-panic в каталоге нет — пустая тема не засчитывается.
  assert.equal(r.get('topic-defer-panic')!.unlocked, false)
})

test('курс засчитан, только когда прочитаны все его лекции', () => {
  assert.equal(byId(progress({ read: new Set(['go-runtime/a']) })).get('course-go-runtime')!.unlocked, false)
  const r = byId(progress({ read: new Set(['go-runtime/a', 'go-runtime/b']) }))
  assert.equal(r.get('course-go-runtime')!.unlocked, true)
  assert.equal(r.get('course-brokers')!.unlocked, false)
  assert.equal(r.get('lectures-all')!.unlocked, false)
})

test('идеальный тест отличается от просто пройденного', () => {
  const r = byId(progress({ quizzes: new Map([['go-runtime/a', { best: 5, total: 6 }]]) }))
  assert.equal(r.get('quiz-first')!.unlocked, true)
  assert.equal(r.get('quiz-perfect')!.unlocked, false)
  const r2 = byId(progress({ quizzes: new Map([['go-runtime/a', { best: 6, total: 6 }]]) }))
  assert.equal(r2.get('quiz-perfect')!.unlocked, true)
})

test('полученное однажды не отбирается', () => {
  const r = byId(progress(), { 'tasks-all': 123 })
  assert.equal(r.get('tasks-all')!.unlocked, true)
  assert.equal(r.get('tasks-all')!.unlockedAt, 123)
})

test('финал открывается вместе с последним из остальных', () => {
  const list = achievementsFor(catalog)
  const earned = Object.fromEntries(list.slice(0, -2).map((a) => [a.id, 1]))
  const finale = list.at(-1)!.id
  assert.equal(byId(progress(), earned).get(finale)!.unlocked, false)
  earned[list.at(-2)!.id] = 1
  assert.equal(byId(progress(), earned).get(finale)!.unlocked, true)
})

test('серия считает только подряд идущие дни, через границу месяца тоже', () => {
  assert.equal(longestStreak([]), 0)
  assert.equal(longestStreak(['2026-01-31', '2026-02-01', '2026-02-02', '2026-02-04']), 3)
  assert.equal(longestStreak(['2026-03-02', '2026-03-01', '2026-03-01']), 2)
})

test('ранг растёт с наградой, Король пиратов — только за всё', () => {
  assert.equal(rankFor(evaluate(progress(), catalog)).title, RANKS[0].title)
  const list = achievementsFor(catalog)
  const all = Object.fromEntries(list.map((a) => [a.id, 1]))
  assert.equal(rankFor(evaluate(progress(), catalog, all)).title, KING)
  // Все ранги достижимы раньше, чем всё получено.
  const total = achievementsFor({ courses: COURSES }).reduce((s, a) => s + a.bounty, 0)
  assert.ok(RANKS.at(-1)!.min < total)
})

test('сумма в белли разбита на разряды', () => {
  assert.equal(formatBerry(1_500_000_000), '1 500 000 000')
  assert.equal(formatBerry(999), '999')
})
