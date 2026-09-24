import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { TASK_KINDS, TASK_LEVELS, TASK_TOPICS, type TaskTopicId, taskLang } from '../../data/practice.ts'

/**
 * Проверки целостности задач.
 *
 * Задача — это папка из нескольких файлов, и связи между ними держатся на
 * соглашении, а не на типах: у `implement` обязан быть check.go, у `fix` —
 * expect.txt, а bugLine обязан указывать на существующую строку starter.go.
 * Зод в content.config.ts проверяет фронтматтер, но про соседние файлы он
 * ничего не знает — поэтому про них знает этот тест.
 *
 * Здесь нет и не может быть прогона кода на Go: тесты не ходят в сеть.
 * Что задачи действительно решаются, проверено запросами к /api/check руками.
 * SQL-задачи — другое дело: их Postgres работает локально, и sql.test.ts
 * прогоняет каждую по-настоящему.
 */

const tasksDir = fileURLToPath(new URL('../../content/tasks/', import.meta.url))
const glossaryDir = fileURLToPath(new URL('../../content/glossary/', import.meta.url))
const lecturesDir = fileURLToPath(new URL('../../content/lectures/', import.meta.url))

const glossary = new Set(readdirSync(glossaryDir).map((f) => f.replace(/\.mdx$/, '')))

const lectures = new Set(
  readdirSync(lecturesDir)
    .filter((course) => statSync(join(lecturesDir, course)).isDirectory())
    .flatMap((course) => readdirSync(join(lecturesDir, course)).map((f) => `${course}/${f.replace(/\.mdx$/, '')}`)),
)

interface Task {
  id: string
  dir: string
  front: Record<string, string>
  /** Те же поля, но как они записаны в файле — с кавычками, если они есть. */
  raw: Record<string, string>
  body: string
  files: Set<string>
}

/** Разбор фронтматтера — ровно настолько, насколько он здесь нужен: плоские строки и списки. */
function frontmatter(source: string): { front: Record<string, string>; raw: Record<string, string>; body: string } {
  const m = source.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/)
  assert.ok(m?.[1] !== undefined && m[2] !== undefined, 'нет фронтматтера')
  const front: Record<string, string> = {}
  const raw: Record<string, string> = {}
  for (const line of m[1].split('\n')) {
    const kv = line.match(/^([a-zA-Z]+):\s*(.*)$/)
    if (!kv?.[1]) continue
    const value = (kv[2] ?? '').trim()
    raw[kv[1]] = value
    front[kv[1]] = value.replace(/^["'](.*)["']$/, '$1')
  }
  return { front, raw, body: m[2] }
}

function collect(): Task[] {
  const out: Task[] = []
  for (const topic of readdirSync(tasksDir)) {
    const topicDir = join(tasksDir, topic)
    if (!statSync(topicDir).isDirectory()) continue
    for (const slug of readdirSync(topicDir)) {
      const dir = join(topicDir, slug)
      if (!statSync(dir).isDirectory()) continue
      const files = new Set(readdirSync(dir))
      assert.ok(files.has('index.mdx'), `${topic}/${slug}: нет index.mdx`)
      const { front, raw, body } = frontmatter(readFileSync(join(dir, 'index.mdx'), 'utf8'))
      out.push({ id: `${topic}/${slug}`, dir, front, raw, body, files })
    }
  }
  return out
}

const TASKS = collect()

/** Язык задачи задаёт тема. Неизвестную тему ловит отдельный тест — здесь считаем её Go. */
function langOf(task: Task): 'go' | 'sql' {
  const known = TASK_TOPICS.some((t) => t.id === task.front.topic)
  return known ? taskLang(task.front.topic as TaskTopicId) : 'go'
}

/** Имя файла заготовки с поправкой на язык: starter.go или starter.sql. */
function codeFile(task: Task, base: 'starter' | 'solution'): string {
  return `${base}.${langOf(task)}`
}

test('задачи вообще есть', () => {
  assert.ok(TASKS.length > 0, 'в src/content/tasks ни одной задачи')
})

test('фронтматтер не разваливает YAML', () => {
  /*
   * `title: Merge: собрать...` — незакавыченный скаляр с «: » внутри, и весь
   * разбор фронтматтера падает. Ошибка вылезает только на сборке и сообщением
   * про отступы, по которому не догадаешься, в чём дело. Ловим здесь.
   */
  for (const task of TASKS) {
    for (const [key, value] of Object.entries(task.raw)) {
      if (/^["'[]/.test(value)) continue
      assert.ok(!value.includes(': '), `${task.id}: значение ${key} содержит «: » — возьмите его в кавычки`)
      assert.ok(!/^[[{>|*&!%@`]/.test(value), `${task.id}: значение ${key} начинается со служебного символа YAML`)
    }
  }
})

test('у каждой задачи есть заготовка и эталон', () => {
  for (const task of TASKS) {
    for (const name of [codeFile(task, 'starter'), codeFile(task, 'solution')]) {
      assert.ok(task.files.has(name), `${task.id}: нет ${name}`)
    }
    if (langOf(task) === 'sql') {
      assert.ok(task.files.has('schema.sql'), `${task.id}: нет schema.sql`)
      // Файлы чужого языка рядом — почти наверняка задача лежит не в той теме.
      assert.ok(![...task.files].some((f) => f.endsWith('.go')), `${task.id}: .go-файлы в SQL-задаче`)
      continue
    }
    for (const name of ['starter.go', 'solution.go']) {
      const source = readFileSync(join(task.dir, name), 'utf8')
      assert.match(source, /^package main\b/m, `${task.id}/${name}: не package main`)
    }
  }
})

test('ordered имеет смысл только у SQL-задач', () => {
  for (const task of TASKS) {
    if (task.front.ordered === undefined) continue
    assert.equal(langOf(task), 'sql', `${task.id}: ordered только для SQL`)
    assert.match(task.front.ordered, /^(true|false)$/, `${task.id}: ordered — true или false`)
  }
})

test('kind, topic и level взяты из справочника', () => {
  const kinds = new Set(TASK_KINDS.map((k) => k.id))
  const topics = new Set(TASK_TOPICS.map((t) => t.id))
  const levels = new Set(TASK_LEVELS.map((l) => l.id))
  for (const task of TASKS) {
    assert.ok(kinds.has(task.front.kind as never), `${task.id}: неизвестный kind ${task.front.kind}`)
    assert.ok(topics.has(task.front.topic as never), `${task.id}: неизвестный topic ${task.front.topic}`)
    if (task.front.level) assert.ok(levels.has(task.front.level as never), `${task.id}: неизвестный level`)
    // Папка задачи лежит в папке темы — иначе ссылки и фильтры разъедутся.
    assert.equal(task.id.split('/')[0], task.front.topic, `${task.id}: папка не совпадает с topic`)
  }
})

test('у задач «реализуй» есть скрытые тесты, и в заготовке нет main', () => {
  // У SQL-задач скрытых тестов нет: ответ сверяется с результатом solution.sql.
  for (const task of TASKS.filter((t) => t.front.kind === 'implement' && langOf(t) === 'go')) {
    assert.ok(task.files.has('check.go'), `${task.id}: нет check.go`)

    const check = readFileSync(join(task.dir, 'check.go'), 'utf8')
    assert.match(check, /func Test\w+\(t \*testing\.T\)/, `${task.id}/check.go: ни одного теста`)

    // Песочница включает режим go test, только если main в файле нет.
    for (const name of ['starter.go', 'solution.go', 'check.go']) {
      const source = readFileSync(join(task.dir, name), 'utf8')
      assert.doesNotMatch(source, /^func main\(\)/m, `${task.id}/${name}: main помешает запуску через go test`)
    }
  }
})

test('у задач «почини» есть непустой эталонный вывод', () => {
  for (const task of TASKS.filter((t) => t.front.kind === 'fix' && langOf(t) === 'go')) {
    assert.ok(task.files.has('expect.txt'), `${task.id}: нет expect.txt`)
    const expected = readFileSync(join(task.dir, 'expect.txt'), 'utf8')
    assert.notEqual(expected.trim(), '', `${task.id}/expect.txt: пустой`)
  }
})

test('bugLine есть только у «найди баг» и указывает внутрь заготовки', () => {
  for (const task of TASKS) {
    const raw = task.front.bugLine
    if (task.front.kind !== 'bug') {
      assert.equal(raw, undefined, `${task.id}: bugLine имеет смысл только для kind: bug`)
      continue
    }
    assert.ok(raw, `${task.id}: для kind: bug нужен bugLine`)
    const line = Number(raw)
    const starter = codeFile(task, 'starter')
    const lines = readFileSync(join(task.dir, starter), 'utf8').split('\n')
    assert.ok(Number.isInteger(line) && line >= 1 && line <= lines.length, `${task.id}: bugLine=${raw} вне ${starter}`)
    assert.notEqual(lines[line - 1]?.trim(), '', `${task.id}: bugLine указывает на пустую строку`)
  }
})

test('у каждой задачи есть разбор', () => {
  for (const task of TASKS) {
    assert.match(task.body, /<Reveal>/, `${task.id}: нет <Reveal> с разбором`)
    assert.match(task.body, /<\/Reveal>/, `${task.id}: <Reveal> не закрыт`)
  }
})

test('ссылки на термины и лекции ведут в существующие файлы', () => {
  for (const task of TASKS) {
    for (const id of (task.front.related ?? '').replace(/[[\]]/g, '').split(',')) {
      const term = id.trim()
      if (!term) continue
      assert.ok(glossary.has(term), `${task.id}: нет термина ${term}`)
    }
    if (task.front.lecture) {
      assert.ok(lectures.has(task.front.lecture), `${task.id}: нет лекции ${task.front.lecture}`)
    }
  }
})

test('ответы не уезжают на клиент', () => {
  /*
   * check.go и expect.txt читает только src/lib/task-secrets.ts, а его имеет
   * право импортировать только серверный маршрут. Стоит островy или странице
   * дотянуться до него — и правильные ответы окажутся в бандле, который
   * открывается через «посмотреть исходный код».
   */
  const srcDir = fileURLToPath(new URL('../../', import.meta.url))

  function* walk(dir: string): Generator<string> {
    for (const name of readdirSync(dir)) {
      const path = join(dir, name)
      if (statSync(path).isDirectory()) yield* walk(path)
      else if (/\.(ts|tsx|astro)$/.test(name) && !name.endsWith('.test.ts')) yield path
    }
  }

  for (const path of walk(srcDir)) {
    const source = readFileSync(path, 'utf8')
    // Ищем именно импорт, а не упоминание: про файл можно писать в комментариях.
    if (!/from\s+['"][^'"]*task-secrets(\.ts)?['"]/.test(source)) continue
    const relative = path.slice(srcDir.length)
    assert.match(relative, /^pages\/api\//, `${relative} импортирует task-secrets.ts — ответы утекут на клиент`)
  }
})
