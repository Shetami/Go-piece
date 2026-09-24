import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { TASK_TOPICS } from '../../../data/practice.ts'
import { openDb, runQuery } from './engine.ts'
import { canonCell, compareTables, parseAnswer, type SqlOutcome, type SqlTable } from './result.ts'

/**
 * SQL-задачи прогоняются по-настоящему: PGlite — тот же Postgres, что работает
 * у читателя в браузере, только запущенный в Node. Поэтому здесь проверяется не
 * только целостность файлов, но и смысл задачи: эталон выполняется, неверная
 * заготовка действительно неверна, а ответ в разборе совпадает с тем, что
 * вернёт Postgres.
 */

// ─── Сверка результатов ──────────────────────────────────────────────────────

const t = (rows: (string | null)[][], columns = rows[0]?.map((_, i) => `c${i}`) ?? ['c0']): SqlTable => ({ columns, rows })

test('numeric с разным масштабом и булевы в двух записях — одно значение', () => {
  assert.equal(canonCell('2.50'), canonCell('2.5'))
  assert.equal(canonCell('3.7500000000000000'), canonCell('3.75'))
  assert.equal(canonCell('t'), canonCell('true'))
  assert.notEqual(canonCell(null), canonCell('NULL'))
  assert.notEqual(canonCell(''), canonCell(null))
})

test('порядок строк важен только при ordered', () => {
  const a = t([['1'], ['2']])
  const b = t([['2'], ['1']])
  assert.equal(compareTables(a, b, false).pass, true)
  assert.equal(compareTables(a, b, true).pass, false)
})

test('имена колонок не важны, число — важно', () => {
  assert.equal(compareTables(t([['1']], ['count']), t([['1']], ['total']), true).pass, true)
  assert.equal(compareTables(t([['1']]), t([['1', '2']]), true).pass, false)
})

test('дубликаты считаются: две одинаковые строки — не одна', () => {
  assert.equal(compareTables(t([['1'], ['1']]), t([['1']]), false).pass, false)
})

test('ответ «что вернёт» прощает заголовок, черту psql и «пусто»', () => {
  const psql = ' name | staff \n------+-------\n Аня  |     3\n(1 row)\n'
  assert.deepEqual(parseAnswer(psql, ['name', 'staff']), [['Аня', '3']])
  assert.deepEqual(parseAnswer('пусто', ['name']), [])
  assert.deepEqual(parseAnswer('Боря | NULL', ['a', 'b']), [['Боря', null]])
})

// ─── Сами задачи ─────────────────────────────────────────────────────────────

const tasksDir = fileURLToPath(new URL('../../../content/tasks/', import.meta.url))
const sqlTopics = TASK_TOPICS.filter((topic) => topic.lang === 'sql').map((topic) => topic.id as string)

interface SqlTask {
  id: string
  kind: string
  ordered: boolean
  bugLine?: number
  body: string
  schema: string
  starter: string
  solution: string
}

function collect(): SqlTask[] {
  const out: SqlTask[] = []
  for (const topic of sqlTopics) {
    const topicDir = join(tasksDir, topic)
    if (!statSync(topicDir, { throwIfNoEntry: false })?.isDirectory()) continue
    for (const slug of readdirSync(topicDir)) {
      const dir = join(topicDir, slug)
      const read = (name: string) => readFileSync(join(dir, name), 'utf8')
      const mdx = read('index.mdx')
      const front = mdx.match(/^---\n([\s\S]*?)\n---\n/)?.[1] ?? ''
      const field = (name: string) => front.match(new RegExp(`^${name}:\\s*(.+)$`, 'm'))?.[1]?.trim()
      out.push({
        id: `${topic}/${slug}`,
        kind: field('kind') ?? '',
        ordered: field('ordered') === 'true',
        bugLine: field('bugLine') ? Number(field('bugLine')) : undefined,
        body: mdx.slice(mdx.indexOf('\n---\n', 4) + 5),
        schema: read('schema.sql'),
        starter: read('starter.sql'),
        solution: read('solution.sql'),
      })
    }
  }
  return out
}

const TASKS = collect()

function table(outcome: SqlOutcome, what: string): SqlTable {
  assert.ok(outcome.ok, `${what}: ${outcome.ok ? '' : outcome.error}`)
  assert.ok(outcome.table, `${what}: запрос не вернул строк`)
  return outcome.table
}

/** Первый блок кода после `<Reveal>` — там разбор начинается с правильного ответа. */
function revealAnswer(body: string): string | undefined {
  const reveal = body.slice(body.indexOf('<Reveal>'))
  return reveal.match(/```[a-z]*\n([\s\S]*?)```/)?.[1]
}

test('SQL-задачи вообще есть', () => {
  assert.ok(TASKS.length > 0)
})

for (const task of TASKS) {
  test(`SQL: ${task.id}`, async () => {
    const db = await openDb(task.schema)
    try {
      const expected = table(await runQuery(db, task.solution), `${task.id}/solution.sql`)
      assert.ok(expected.rows.length > 0 || task.kind === 'output', `${task.id}: эталон вернул ноль строк — проверять нечего`)

      if (task.ordered) {
        assert.match(task.solution, /order\s+by/i, `${task.id}: ordered: true, а в эталоне нет ORDER BY`)
      }

      const starter = await runQuery(db, task.starter)

      switch (task.kind) {
        case 'output': {
          assert.equal(task.starter, task.solution, `${task.id}: у «что вернёт» заготовка и есть эталон`)
          const got = table(starter, `${task.id}/starter.sql`)
          const answer = revealAnswer(task.body)
          assert.ok(answer !== undefined, `${task.id}: разбор должен начинаться с ответа в блоке кода`)
          const parsed = parseAnswer(answer, got.columns)
          const cmp = compareTables(got, { columns: got.columns, rows: parsed }, task.ordered)
          assert.ok(cmp.pass, `${task.id}: ответ в разборе не совпадает с Postgres: ${cmp.detail}`)
          break
        }

        case 'bug': {
          const got = table(starter, `${task.id}/starter.sql`)
          assert.equal(compareTables(expected, got, task.ordered).pass, false, `${task.id}: баг ни на что не влияет`)
          // Эталон — та же заготовка с исправленной отмеченной строкой, и больше ничем.
          const a = task.starter.split('\n')
          const b = task.solution.split('\n')
          assert.equal(a.length, b.length, `${task.id}: починка бага меняет число строк — bugLine не проверить`)
          const changed = a.map((line, i) => (line === b[i] ? 0 : i + 1)).filter(Boolean)
          assert.deepEqual(changed, [task.bugLine], `${task.id}: эталон отличается от заготовки не только в строке bugLine`)
          break
        }

        case 'fix':
        case 'implement': {
          if (starter.ok && starter.table) {
            const cmp = compareTables(expected, starter.table, task.ordered)
            assert.equal(cmp.pass, false, `${task.id}: заготовка уже проходит проверку`)
          }
          break
        }

        default:
          assert.fail(`${task.id}: неизвестный kind ${task.kind}`)
      }
    } finally {
      await db.close()
    }
  })
}

test('прогон не меняет данные задачи', async () => {
  const db = await openDb('CREATE TABLE t (x int); INSERT INTO t VALUES (1), (2);')
  try {
    const del = await runQuery(db, 'DELETE FROM t')
    assert.ok(del.ok)
    const after = table(await runQuery(db, 'SELECT count(*) FROM t'), 'count')
    assert.equal(after.rows[0]?.[0], '2')

    const commit = await runQuery(db, 'DELETE FROM t; COMMIT;')
    assert.equal(commit.ok, false, 'COMMIT внутри запроса должен отклоняться')
  } finally {
    await db.close()
  }
})

test('ошибка Postgres приходит с позицией', async () => {
  const db = await openDb('CREATE TABLE t (x int);')
  try {
    const bad = await runQuery(db, 'SELECT y FROM t')
    assert.equal(bad.ok, false)
    if (!bad.ok) assert.equal(bad.position, 8)
  } finally {
    await db.close()
  }
})
