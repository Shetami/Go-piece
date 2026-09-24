import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

import type { Question, Quiz } from './types.ts'

/**
 * Проверки тестов «Проверь себя».
 *
 * Формат вопроса проверяет компилятор, а здесь — то, чего типы не видят:
 * ответ указывает на существующий вариант, у каждой лекции есть тест, у теста
 * есть к чему привязаться, и вопросы не вырождаются («отметьте все» с
 * единственным правильным, все «верно/неверно» с одним ответом и т. п.).
 */

const lecturesDir = fileURLToPath(new URL('../../content/lectures/', import.meta.url))
const quizzesDir = fileURLToPath(new URL('../../data/quizzes/', import.meta.url))

function listIds(dir: string, ext: string): string[] {
  const out: string[] = []
  for (const course of readdirSync(dir)) {
    const courseDir = join(dir, course)
    if (!statSync(courseDir).isDirectory()) continue
    for (const f of readdirSync(courseDir)) {
      if (f.endsWith(ext)) out.push(`${course}/${f.slice(0, -ext.length)}`)
    }
  }
  return out.sort()
}

const lectures = listIds(lecturesDir, '.mdx').filter(
  (id) => !/^draft:\s*true/m.test(readFileSync(join(lecturesDir, `${id}.mdx`), 'utf8')),
)

const quizzes = new Map<string, Quiz>()
for (const id of listIds(quizzesDir, '.ts')) {
  const mod = (await import(pathToFileURL(join(quizzesDir, `${id}.ts`)).href)) as { default: Quiz }
  quizzes.set(id, mod.default)
}

function texts(q: Question): string[] {
  switch (q.kind) {
    case 'single':
    case 'multi':
      return [q.q, q.explain, ...q.options]
    case 'truefalse':
      return [q.q, q.explain]
    case 'order':
      return [q.q, q.explain, ...q.items]
  }
}

test('у каждой лекции есть тест, у каждого теста — лекция', () => {
  for (const id of lectures) assert.ok(quizzes.has(id), `нет теста для лекции ${id}: src/data/quizzes/${id}.ts`)
  for (const id of quizzes.keys()) assert.ok(lectures.includes(id), `тест ${id} не привязан ни к одной лекции`)
})

test('в тесте от 6 до 10 вопросов и хотя бы три их вида', () => {
  for (const [id, quiz] of quizzes) {
    const n = quiz.questions.length
    assert.ok(n >= 6 && n <= 10, `${id}: ${n} вопросов, нужно 6–10`)
    const kinds = new Set(quiz.questions.map((q) => q.kind))
    assert.ok(kinds.size >= 3, `${id}: виды вопросов ${[...kinds].join(', ')} — нужно хотя бы три разных`)
  }
})

test('ответы указывают на существующие варианты', () => {
  for (const [id, quiz] of quizzes) {
    quiz.questions.forEach((q, i) => {
      const at = `${id}, вопрос ${i + 1}`
      if (q.kind === 'single') {
        assert.ok(q.options.length >= 3 && q.options.length <= 5, `${at}: вариантов ${q.options.length}, нужно 3–5`)
        assert.ok(Number.isInteger(q.answer) && q.answer >= 0 && q.answer < q.options.length, `${at}: answer вне вариантов`)
      }
      if (q.kind === 'multi') {
        assert.ok(q.options.length >= 3 && q.options.length <= 6, `${at}: вариантов ${q.options.length}, нужно 3–6`)
        assert.equal(new Set(q.answer).size, q.answer.length, `${at}: повторы в answer`)
        for (const a of q.answer) assert.ok(a >= 0 && a < q.options.length, `${at}: answer ${a} вне вариантов`)
        // Один правильный — это single; все правильные — нечего выбирать.
        assert.ok(q.answer.length >= 2, `${at}: в «отметьте все» должно быть хотя бы два правильных`)
        assert.ok(q.answer.length < q.options.length, `${at}: все варианты правильные`)
      }
      if (q.kind === 'order') {
        assert.ok(q.items.length >= 3 && q.items.length <= 6, `${at}: шагов ${q.items.length}, нужно 3–6`)
      }
    })
  }
})

test('варианты внутри вопроса не повторяются', () => {
  for (const [id, quiz] of quizzes) {
    quiz.questions.forEach((q, i) => {
      const list = q.kind === 'order' ? q.items : q.kind === 'truefalse' ? [] : q.options
      assert.equal(new Set(list).size, list.length, `${id}, вопрос ${i + 1}: одинаковые варианты`)
    })
  }
})

test('у каждого вопроса есть разбор и разметка не сломана', () => {
  for (const [id, quiz] of quizzes) {
    quiz.questions.forEach((q, i) => {
      const at = `${id}, вопрос ${i + 1}`
      assert.ok(q.explain.trim().length >= 30, `${at}: разбор слишком короткий`)
      for (const t of texts(q)) {
        assert.ok(t.trim().length > 0, `${at}: пустой текст`)
        assert.equal((t.match(/`/g) ?? []).length % 2, 0, `${at}: непарные обратные кавычки в «${t}»`)
        assert.equal((t.match(/\*\*/g) ?? []).length % 2, 0, `${at}: непарные ** в «${t}»`)
      }
    })
  }
})

test('«верно/неверно» не всегда с одним ответом', () => {
  // Варианты single и multi перемешивает сам остров, а два ответа «верно/неверно»
  // не перемешать — за балансом следит этот тест.
  const tf = [...quizzes.values()].flatMap((q) => q.questions).filter((q) => q.kind === 'truefalse')
  if (tf.length < 4) return
  const share = tf.filter((q) => q.answer).length / tf.length
  assert.ok(share > 0.25 && share < 0.75, `«верно» — ${Math.round(share * 100)}% утверждений, нужен баланс`)
})
