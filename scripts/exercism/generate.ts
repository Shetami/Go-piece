/**
 * Задачи темы «Алгоритмы» из exercism/problem-specifications (MIT).
 *
 *   pnpm tasks:exercism            все упражнения из exercises.ts
 *   pnpm tasks:exercism knapsack   только перечисленные (по slug задачи)
 *
 * Берём из репозитория только тест-кейсы: canonical-data.json превращается в
 * check.go — табличные тесты, которые /api/check дописывает к коду
 * пользователя. Условие, заготовку и эталон генератор не трогает, если они уже
 * есть: это авторский текст на русском, его пишут руками. Для нового
 * упражнения он кладёт черновик (draft: true), чтобы было с чего начать.
 *
 * Репозиторий клонируется в node_modules/.cache на коммит SPECS_REF —
 * check.go от этого воспроизводим. EXERCISM_SPECS=<путь> — взять локальную копию.
 */

import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { EXERCISES, SPECS_REF, SPECS_REPO, type Exercise, type GoType, type Operation } from './exercises.ts'

const root = fileURLToPath(new URL('../../', import.meta.url))
const tasksDir = join(root, 'src/content/tasks/algorithms')
const cacheDir = join(root, 'node_modules/.cache/problem-specifications')

interface Case {
  uuid: string
  description: string
  property: string
  input: Record<string, unknown>
  expected: unknown
  reimplements?: string
}

function git(...args: string[]): string {
  return execFileSync('git', ['-C', cacheDir, ...args], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim()
}

/** Папка exercises нужного коммита: своя копия или кэш, который подтягиваем до SPECS_REF. */
function specsDir(): string {
  const local = process.env.EXERCISM_SPECS
  if (local) return join(local, 'exercises')

  if (!existsSync(join(cacheDir, '.git'))) {
    mkdirSync(cacheDir, { recursive: true })
    git('init', '-q')
    git('remote', 'add', 'origin', SPECS_REPO)
  }
  let head = ''
  try {
    head = git('rev-parse', 'HEAD')
  } catch {
    // Свежий репозиторий без коммитов.
  }
  if (head !== SPECS_REF) {
    console.log(`problem-specifications → ${SPECS_REF.slice(0, 7)}`)
    git('fetch', '-q', '--depth', '1', 'origin', SPECS_REF)
    git('checkout', '-q', '--detach', 'FETCH_HEAD')
  }
  return join(cacheDir, 'exercises')
}

/**
 * Кейсы упражнения плоским списком. Группы раскрываются, а кейс, который
 * переписан новым (`reimplements`), выбрасывается: в силе только последняя версия.
 */
function loadCases(dir: string, ex: Exercise): Case[] {
  const data = JSON.parse(readFileSync(join(dir, ex.exercise, 'canonical-data.json'), 'utf8'))
  const flat = (list: any[]): Case[] => list.flatMap((c) => (Array.isArray(c.cases) ? flat(c.cases) : [c]))
  const all = flat(data.cases)
  const replaced = new Set(all.map((c) => c.reimplements).filter(Boolean))
  return all.filter((c) => !replaced.has(c.uuid) && !ex.skip?.[c.uuid])
}

// --- Go: типы и литералы -----------------------------------------------------

function typeName(t: GoType): string {
  if (typeof t === 'string') return t
  if ('slice' in t) return `[]${typeName(t.slice)}`
  if ('array' in t) return `[${t.len}]${typeName(t.array)}`
  if ('map' in t) return `map[${typeName(t.map[0])}]${typeName(t.map[1])}`
  return t.struct
}

function isScalar(t: GoType): boolean {
  return t === 'int' || t === 'string' || t === 'bool'
}

/**
 * Значение из JSON как литерал Go. `inner` — литерал стоит внутри другого
 * составного, и тип элемента можно опустить: `[][]int{{1, 2}}`.
 */
function literal(t: GoType, v: unknown, inner = false): string {
  const fail = () => {
    throw new Error(`${JSON.stringify(v)} не ложится в ${typeName(t)}`)
  }
  if (t === 'int') return Number.isInteger(v) ? String(v) : fail()
  // Экранирование JSON — подмножество экранирования строк Go.
  if (t === 'string') return typeof v === 'string' ? JSON.stringify(v) : fail()
  if (t === 'bool') return typeof v === 'boolean' ? String(v) : fail()

  const prefix = inner ? '' : typeName(t)
  if ('slice' in t || 'array' in t) {
    if (!Array.isArray(v)) return fail()
    if ('array' in t && v.length !== t.len) return fail()
    const elem = 'slice' in t ? t.slice : t.array
    return `${prefix}{${v.map((x) => literal(elem, x, true)).join(', ')}}`
  }
  if ('map' in t) {
    if (typeof v !== 'object' || v === null || Array.isArray(v)) return fail()
    // Ключи по порядку: иначе check.go менялся бы от перестановки в JSON.
    const entries = Object.entries(v).sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0))
    return `${prefix}{${entries.map(([k, x]) => `${literal(t.map[0], k, true)}: ${literal(t.map[1], x, true)}`).join(', ')}}`
  }
  if (typeof v !== 'object' || v === null) return fail()
  const fields = Object.entries(t.fields).map(([key, [name, ft]]) => {
    if (!(key in v)) fail()
    return `${name}: ${literal(ft, (v as Record<string, unknown>)[key])}`
  })
  return `${prefix}{${fields.join(', ')}}`
}

function zero(t: GoType): string {
  if (t === 'int') return '0'
  if (t === 'string') return '""'
  if (t === 'bool') return 'false'
  if ('struct' in t) return `${t.struct}{}`
  if ('array' in t) return `${typeName(t)}{}`
  return 'nil'
}

/** Глагол fmt для значения в сообщении о провале: строки — в кавычках, чтобы были видны пробелы. */
function verb(t: GoType): string {
  if (t === 'string' || (typeof t === 'object' && 'slice' in t && t.slice === 'string')) return '%q'
  return '%v'
}

const KEYWORDS = new Set(
  'break case chan const continue default defer else fallthrough for func go goto if import interface map package range return select struct switch type var'.split(' '),
)
/** Служебные поля строки таблицы — ключи input не должны с ними совпасть. */
const RESERVED = new Set(['name', 'want', 'wantErr'])

function field(key: string): string {
  return KEYWORDS.has(key) || RESERVED.has(key) ? `${key}_` : key
}

// --- check.go ----------------------------------------------------------------

function isError(expected: unknown): boolean {
  return typeof expected === 'object' && expected !== null && !Array.isArray(expected) && 'error' in expected
}

function testFunc(ex: Exercise, op: Operation, cases: Case[]): string {
  if (cases.length === 0) throw new Error(`${ex.slug}: у ${op.property} нет кейсов`)

  const rows = cases.map((c) => {
    const args = op.args.map(([key, t]) => {
      if (!(key in c.input)) throw new Error(`${ex.slug}: в кейсе «${c.description}» нет ${key}`)
      return literal(t, c.input[key])
    })
    let want: string
    let wantErr = false
    if (isError(c.expected)) {
      if (op.error) {
        want = zero(op.result)
        wantErr = true
      } else if (op.errorValue !== undefined) {
        want = literal(op.result, op.errorValue)
      } else {
        throw new Error(`${ex.slug}: кейс «${c.description}» ждёт ошибку, а ${op.func} её не возвращает`)
      }
    } else {
      want = literal(op.result, c.expected)
    }
    const cells = [JSON.stringify(c.description), ...args, want]
    if (op.error) cells.push(String(wantErr))
    return `\t\t{${cells.join(', ')}},`
  })

  const fields = [
    '\t\tname string',
    ...op.args.map(([key, t]) => `\t\t${field(key)} ${typeName(t)}`),
    `\t\twant ${typeName(op.result)}`,
    ...(op.error ? ['\t\twantErr bool'] : []),
  ]

  const call = op.call
    ? op.call.replace(/\$(\w+)/g, (_, key: string) => `c.${field(key)}`)
    : `${op.func}(${op.args.map(([key]) => `c.${field(key)}`).join(', ')})`

  // Сообщение повторяет вызов с настоящими аргументами — по нему кейс можно воспроизвести.
  const shown = op.call
    ? op.call.replace(/\$(\w+)/g, (_, key: string) => verb(op.args.find(([k]) => k === key)![1]))
    : `${op.func}(${op.args.map(([, t]) => verb(t)).join(', ')})`
  const shownArgs = op.args.map(([key]) => `c.${field(key)}`).join(', ')
  const same = isScalar(op.result) ? 'got != c.want' : '!sameResult(got, c.want)'

  const body = op.error
    ? [
        `\t\t\tgot, err := ${call}`,
        `\t\t\tif c.wantErr {`,
        `\t\t\t\tif err == nil {`,
        `\t\t\t\t\tt.Fatalf("${shown} = ${verb(op.result)}, ожидали ошибку", ${shownArgs}, got)`,
        `\t\t\t\t}`,
        `\t\t\t\treturn`,
        `\t\t\t}`,
        `\t\t\tif err != nil {`,
        `\t\t\t\tt.Fatalf("${shown}: неожиданная ошибка %v", ${shownArgs}, err)`,
        `\t\t\t}`,
      ]
    : [`\t\t\tgot := ${call}`]

  return [
    `func Test${op.test}(t *testing.T) {`,
    `\tcases := []struct {`,
    ...fields,
    `\t}{`,
    ...rows,
    `\t}`,
    `\tfor _, c := range cases {`,
    `\t\tt.Run(c.name, func(t *testing.T) {`,
    ...body,
    `\t\t\tif ${same} {`,
    `\t\t\t\tt.Fatalf("${shown} = ${verb(op.result)}, ожидали ${verb(op.result)}", ${shownArgs}, got, c.want)`,
    `\t\t\t}`,
    `\t\t})`,
    `\t}`,
    `}`,
  ].join('\n')
}

const SAME_RESULT = `// sameResult — reflect.DeepEqual, который не отличает nil от пустого слайса
// или мапы: для ответа это одно и то же.
func sameResult(got, want any) bool {
	g, w := reflect.ValueOf(got), reflect.ValueOf(want)
	if g.Kind() == w.Kind() && (g.Kind() == reflect.Slice || g.Kind() == reflect.Map) && g.Len() == 0 && w.Len() == 0 {
		return true
	}
	return reflect.DeepEqual(got, want)
}`

function checkFile(ex: Exercise, cases: Case[]): string {
  const known = new Set(ex.operations.map((op) => op.property))
  const unknown = [...new Set(cases.map((c) => c.property))].filter((p) => !known.has(p))
  if (unknown.length > 0) throw new Error(`${ex.slug}: операции ${unknown.join(', ')} не описаны в exercises.ts`)

  const tests = ex.operations.map((op) =>
    testFunc(
      ex,
      op,
      cases.filter((c) => c.property === op.property),
    ),
  )
  const needsSame = ex.operations.some((op) => !isScalar(op.result))

  // Без import: /api/check дописывает файл к коду пользователя, импорты проставит goimports.
  return [
    `// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications`,
    `// (exercises/${ex.exercise}, MIT, © Exercism). Руками не править.`,
    'package main',
    '',
    ...tests.flatMap((t) => [t, '']),
    ...(needsSame ? [SAME_RESULT, ''] : []),
  ].join('\n')
}

// --- черновик новой задачи ---------------------------------------------------

function starterFile(ex: Exercise): string {
  const funcs = ex.operations
    .filter((op) => !op.call)
    .map((op) => {
      const params = op.args.map(([key, t]) => `${field(key)} ${typeName(t)}`).join(', ')
      const result = op.error ? `(${typeName(op.result)}, error)` : typeName(op.result)
      const ret = op.error ? `${zero(op.result)}, nil` : zero(op.result)
      return `// ${op.func} — TODO: контракт.\nfunc ${op.func}(${params}) ${result} {\n\t// ваш код\n\treturn ${ret}\n}`
    })
  return ['package main', '', ...(ex.types ? [ex.types, ''] : []), funcs.join('\n\n'), ''].join('\n')
}

function draftIndex(ex: Exercise, order: number): string {
  return `---
title: ${ex.slug}
description: TODO
kind: implement
topic: algorithms
level: medium
order: ${order}
draft: true
---

TODO: условие. Исходное — exercises/${ex.exercise}/instructions.md в problem-specifications.

<Reveal>

TODO: разбор.

</Reveal>
`
}

// --- запуск ------------------------------------------------------------------

function writeIfChanged(path: string, content: string): boolean {
  if (existsSync(path) && readFileSync(path, 'utf8') === content) return false
  writeFileSync(path, content)
  return true
}

function gofmt(path: string) {
  try {
    execFileSync('gofmt', ['-w', path], { stdio: ['ignore', 'ignore', 'pipe'] })
  } catch (err) {
    throw new Error(`gofmt ${path}: ${(err as { stderr?: Buffer }).stderr?.toString() ?? err}`)
  }
}

function main() {
  const only = new Set(process.argv.slice(2))
  const unknown = [...only].filter((s) => !EXERCISES.some((e) => e.slug === s))
  if (unknown.length > 0) throw new Error(`нет в exercises.ts: ${unknown.join(', ')}`)

  const dir = specsDir()
  EXERCISES.forEach((ex, i) => {
    if (only.size > 0 && !only.has(ex.slug)) return
    const taskDir = join(tasksDir, ex.slug)
    mkdirSync(taskDir, { recursive: true })

    const drafted: string[] = []
    const draft = (name: string, content: string) => {
      const path = join(taskDir, name)
      if (existsSync(path)) return
      writeFileSync(path, content)
      if (name.endsWith('.go')) gofmt(path)
      drafted.push(name)
    }
    draft('index.mdx', draftIndex(ex, (i + 1) * 10))
    draft('starter.go', starterFile(ex))
    draft('solution.go', starterFile(ex))

    // check.go пишется во временное имя и форматируется там: так «не изменился» — честное сравнение.
    const cases = loadCases(dir, ex)
    const tmp = join(taskDir, '.check.go.tmp')
    writeFileSync(tmp, checkFile(ex, cases))
    gofmt(tmp)
    const changed = writeIfChanged(join(taskDir, 'check.go'), readFileSync(tmp, 'utf8'))
    rmSync(tmp)

    const notes = [`${cases.length} кейсов`, changed ? 'check.go обновлён' : 'без изменений']
    if (drafted.length > 0) notes.push(`черновик: ${drafted.join(', ')}`)
    console.log(`${ex.slug.padEnd(22)} ${notes.join(' · ')}`)
  })
}

main()
