import type { LocatedOp, MapOp } from './types.ts'

/**
 * Сценарий стенда — текст, по команде на строку. Понимает и короткую форму,
 * и запись, как в Go:
 *
 *   put ann 1           m["ann"] = 1
 *   get ann             m["ann"]          v, ok := m["ann"]
 *   del ann             delete(m, "ann")
 *   put k1..k20         — двадцать ключей подряд
 *   fill 20 [user]      — двадцать новых ключей user1, user2, …
 *   collide 6 [группа]  — шесть ключей с одной стартовой группой (корзиной)
 *   del @last           — удалить всё, что положила последняя fill/collide
 *   del @last:3         — первые три её ключа; get @last:-1 — последний
 *   range               for k, v := range m   — начать обход
 *   next [n|all]        — следующий шаг обхода
 *   clear               clear(m)
 *   race [write|read|iter] — вторая горутина пишет одновременно
 *
 * Комментарии — после # или //.
 */

export interface ParseError {
  line: number
  message: string
}

/** Больше ключей за одну команду стенд не берёт: история хранит снимок на каждый тик. */
export const MAX_BATCH = 200

const unquote = (s: string) => s.trim().replace(/^["'`](.*)["'`]$/, '$1')

/** k1..k20 → k1, k2, …, k20. Обычный ключ — массив из одного. */
function expandKeys(raw: string): string[] | string {
  const key = unquote(raw)
  const m = /^(.*?)(\d+)\.\.\1?(\d+)$/.exec(key)
  if (!m) return key.length === 0 ? 'пустой ключ' : [key]
  const [, prefix, a, b] = m
  const from = Number(a)
  const to = Number(b)
  if (to < from) return `диапазон ${key} пуст`
  if (to - from + 1 > MAX_BATCH) return `не больше ${MAX_BATCH} ключей за раз`
  return Array.from({ length: to - from + 1 }, (_, i) => `${prefix}${from + i}`)
}

const count = (raw: string | undefined, def: number): number | string => {
  if (raw === undefined) return def
  const n = Number(raw)
  if (!Number.isInteger(n) || n < 1) return `«${raw}» — нужно целое число больше нуля`
  if (n > MAX_BATCH) return `не больше ${MAX_BATCH} за раз`
  return n
}

/** @last, @last:3, @last:-1 — ключи последней пачки. */
function lastRef(arg: string, op: 'get' | 'del'): MapOp[] | string | null {
  const m = /^@last(?::(-?\d+))?$/.exec(arg)
  if (!m) return arg.startsWith('@') ? `«${arg}» — есть только @last, @last:N и @last:-N` : null
  const n = m[1] === undefined ? null : Number(m[1])
  if (n === 0) return '@last:0 — пустой выбор'
  return [{ kind: 'last', op, n }]
}

function parseLine(src: string): MapOp[] | string {
  let s = src.trim()

  // Запись как в Go.
  let m = /^m\[(.+)\]\s*=\s*(-?\d+)$/.exec(s)
  if (m) s = `put ${m[1]} ${m[2]}`
  m = /^(?:\w+\s*(?:,\s*\w+)?\s*:?=\s*)?m\[(.+)\]$/.exec(s)
  if (m) s = `get ${m[1]}`
  m = /^delete\(\s*m\s*,\s*(.+)\)$/.exec(s)
  if (m) s = `del ${m[1]}`
  if (/^clear\(\s*m\s*\)$/.test(s)) s = 'clear'
  if (/^for\b.*\brange\s+m\b/.test(s)) s = 'range'

  const [cmd, ...args] = s.split(/\s+/)
  switch (cmd) {
    case 'put':
    case 'set': {
      if (!args[0]) return 'put: не указан ключ'
      const keys = expandKeys(args[0])
      if (typeof keys === 'string') return keys
      let val: number | undefined
      if (args[1] !== undefined) {
        val = Number(args[1])
        if (!Number.isInteger(val)) return `put: значение «${args[1]}» — не целое число`
      }
      return keys.map((key) => ({ kind: 'put', key, val }))
    }
    case 'get': {
      if (!args[0]) return 'get: не указан ключ'
      const last = lastRef(args[0], 'get')
      if (last) return last
      const keys = expandKeys(args[0])
      if (typeof keys === 'string') return keys
      return keys.map((key) => ({ kind: 'get', key }))
    }
    case 'del':
    case 'delete': {
      if (!args[0]) return 'del: не указан ключ'
      const last = lastRef(args[0], 'del')
      if (last) return last
      const keys = expandKeys(args[0])
      if (typeof keys === 'string') return keys
      return keys.map((key) => ({ kind: 'del', key }))
    }
    case 'fill': {
      const n = count(args[0], 10)
      if (typeof n === 'string') return `fill: ${n}`
      const prefix = args[1] ? unquote(args[1]) : 'k'
      if (!/^[\p{L}_][\p{L}\p{N}_-]*$/u.test(prefix)) return `fill: префикс «${prefix}» — нужны буквы`
      return [{ kind: 'fill', count: n, prefix }]
    }
    case 'collide': {
      const n = count(args[0], 6)
      if (typeof n === 'string') return `collide: ${n}`
      let target: number | undefined
      if (args[1] !== undefined) {
        target = Number(args[1])
        if (!Number.isInteger(target) || target < 0) return `collide: номер группы «${args[1]}» — не число`
      }
      return [{ kind: 'collide', count: n, target }]
    }
    case 'range':
    case 'iter':
      return [{ kind: 'range' }]
    case 'next': {
      if (args[0] === 'all') return [{ kind: 'next', count: 'all' }]
      const n = count(args[0], 1)
      if (typeof n === 'string') return `next: ${n}`
      return [{ kind: 'next', count: n }]
    }
    case 'clear':
      return [{ kind: 'clear' }]
    case 'race': {
      const mode = args[0] ?? 'write'
      if (mode !== 'write' && mode !== 'read' && mode !== 'iter') return 'race: write, read или iter'
      return [{ kind: 'race', mode }]
    }
    default:
      return `не понимаю «${src.trim()}»`
  }
}

export function parseScript(text: string): { ops: LocatedOp[]; errors: ParseError[] } {
  const ops: LocatedOp[] = []
  const errors: ParseError[] = []
  text.split('\n').forEach((raw, i) => {
    const line = raw.replace(/(#|\/\/).*$/, '').trim()
    if (!line) return
    const r = parseLine(line)
    if (typeof r === 'string') errors.push({ line: i + 1, message: r })
    else for (const op of r) ops.push({ ...op, line: i + 1 })
  })
  return { ops, errors }
}

const q = (k: string) => (k === 'NaN' ? 'math.NaN()' : `"${k}"`)

/** Как команда выглядит в Go — для схемы и ленты. */
export function opCode(op: MapOp): string {
  switch (op.kind) {
    case 'make':
      return 'm := make(map[string]int)'
    case 'put':
      return `m[${q(op.key)}] = ${op.val ?? '…'}`
    case 'get':
      return `v, ok := m[${q(op.key)}]`
    case 'del':
      return `delete(m, ${q(op.key)})`
    case 'last':
      return `${op.op === 'get' ? 'm[k]' : 'delete(m, k)'} // ключи последней пачки`
    case 'clear':
      return 'clear(m)'
    case 'range':
      return 'for k, v := range m {'
    case 'next':
      return '    // следующая итерация'
    case 'fill':
      return `for i := range ${op.count} { m[fmt.Sprint("${op.prefix}", i)] = i }`
    case 'collide':
      return `// ${op.count} ключей с одной стартовой группой`
    case 'race':
      return op.mode === 'write' ? 'go func() { m[k] = v }() // две записи сразу' : op.mode === 'read' ? 'go func() { _ = m[k] }() // чтение во время записи' : 'go func() { for range m {} }() // обход во время записи'
  }
}
