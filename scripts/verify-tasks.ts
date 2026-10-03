/**
 * Прогон задач «реализуй» на Go локально, без песочницы go.dev:
 *
 *   pnpm tasks:verify              все задачи
 *   pnpm tasks:verify algorithms   только те, чей id начинается с префикса
 *
 * Для каждой задачи два прогона, склеенные так же, как это делает /api/check:
 *   solution.go + check.go — тесты обязаны пройти;
 *   starter.go  + check.go — обязано собраться: иначе пользователь получит
 *                            ошибку компиляции в наших тестах, а не в своём коде.
 *
 * Песочница расставляет импорты через goimports. Здесь его может не быть,
 * поэтому импорты check.go добавляются по тому, какие пакеты в нём упомянуты.
 */

import { execFile } from 'node:child_process'
import { mkdtempSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { promisify } from 'node:util'

const run = promisify(execFile)
const tasksDir = fileURLToPath(new URL('../src/content/tasks/', import.meta.url))

/** Пакеты, которые встречаются в check.go: имя в коде → путь импорта. */
const STD: Record<string, string> = {
  errors: 'errors',
  fmt: 'fmt',
  maps: 'maps',
  math: 'math',
  reflect: 'reflect',
  slices: 'slices',
  sort: 'sort',
  strconv: 'strconv',
  strings: 'strings',
  sync: 'sync',
  testing: 'testing',
  time: 'time',
  unicode: 'unicode',
  utf8: 'unicode/utf8',
  atomic: 'sync/atomic',
  context: 'context',
}

interface Task {
  id: string
  dir: string
}

function implementTasks(prefix: string): Task[] {
  const out: Task[] = []
  for (const topic of readdirSync(tasksDir)) {
    const topicDir = join(tasksDir, topic)
    if (!statSync(topicDir).isDirectory()) continue
    for (const slug of readdirSync(topicDir)) {
      const dir = join(topicDir, slug)
      const id = `${topic}/${slug}`
      if (!id.startsWith(prefix) || !statSync(dir).isDirectory()) continue
      const index = readFileSync(join(dir, 'index.mdx'), 'utf8')
      if (/^kind:\s*implement\s*$/m.test(index) && readdirSync(dir).includes('check.go')) out.push({ id, dir })
    }
  }
  return out
}

/** Импорты файла и код без них. */
function splitImports(source: string): { imports: Set<string>; body: string } {
  const imports = new Set<string>()
  const body = source
    .replace(/^import\s*\(([\s\S]*?)\)\s*$/gm, (_, block: string) => {
      for (const m of block.matchAll(/"([^"]+)"/g)) imports.add(m[1]!)
      return ''
    })
    .replace(/^import\s+"([^"]+)"\s*$/gm, (_, path: string) => {
      imports.add(path)
      return ''
    })
  return { imports, body }
}

function merge(code: string, check: string): string {
  const user = splitImports(code)
  const tests = check.replace(/^package\s+\w+\s*/m, '')
  const imports = new Set(user.imports)
  for (const [name, path] of Object.entries(STD)) {
    if (new RegExp(`\\b${name}\\.`).test(tests)) imports.add(path)
  }
  const body = user.body.replace(/^package\s+\w+\s*/m, '')
  const block = [...imports].map((p) => `\t"${p}"`).join('\n')
  return `package main\n\nimport (\n${block}\n)\n\n${body.trim()}\n\n${tests}`
}

/** go test в отдельной папке. Код ошибки 1 — тесты упали; сборка падает с «[build failed]». */
async function goTest(source: string): Promise<{ ok: boolean; built: boolean; output: string }> {
  const dir = mkdtempSync(join(tmpdir(), 'task-'))
  try {
    writeFileSync(join(dir, 'go.mod'), 'module task\n\ngo 1.24\n')
    writeFileSync(join(dir, 'task_test.go'), source)
    try {
      const { stdout } = await run('go', ['test', '-count=1', '.'], { cwd: dir, timeout: 120_000 })
      return { ok: true, built: true, output: stdout }
    } catch (err) {
      const e = err as { stdout?: string; stderr?: string; killed?: boolean; signal?: string }
      if (e.killed) {
        // Не вердикт по задаче: машина занята, и go test не уложился в таймаут.
        return { ok: false, built: false, output: `go test прерван (${e.signal ?? 'таймаут'}) — повторите на свободной машине` }
      }
      const output = `${e.stdout ?? ''}${e.stderr ?? ''}`
      return { ok: false, built: !/build failed|setup failed/.test(output), output }
    }
  } finally {
    rmSync(dir, { recursive: true, force: true })
  }
}

async function verify(task: Task): Promise<string[]> {
  const read = (name: string) => readFileSync(join(task.dir, name), 'utf8')
  const check = read('check.go')
  const problems: string[] = []

  const solved = await goTest(merge(read('solution.go'), check))
  if (!solved.ok) problems.push(`эталон не проходит тесты:\n${solved.output}`)

  const started = await goTest(merge(read('starter.go'), check))
  if (!started.built) problems.push(`заготовка не собирается с тестами:\n${started.output}`)

  return problems
}

async function main() {
  const tasks = implementTasks(process.argv[2] ?? '')
  if (tasks.length === 0) {
    console.error('ни одной задачи «реализуй» под этот префикс')
    process.exit(1)
  }

  let failed = 0
  // Компилятор и так грузит все ядра — больше четырёх сборок разом не ускоряет.
  const queue = [...tasks]
  await Promise.all(
    Array.from({ length: 4 }, async () => {
      for (let task = queue.shift(); task; task = queue.shift()) {
        const problems = await verify(task)
        if (problems.length === 0) {
          console.log(`ok    ${task.id}`)
          continue
        }
        failed++
        console.log(`FAIL  ${task.id}\n${problems.map((p) => p.replace(/^/gm, '      ')).join('\n')}`)
      }
    }),
  )

  console.log(`\n${tasks.length - failed}/${tasks.length} задач в порядке`)
  if (failed > 0) process.exit(1)
}

await main()
