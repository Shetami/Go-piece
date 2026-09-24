/**
 * Сборка редактора: CodeMirror 6 плюс настоящая грамматика Go.
 *
 * Подсветка идёт не по регулярным выражениям, а по разбору: `@codemirror/lang-go`
 * — это парсер Lezer, он же даёт отступы, парные скобки и складывание блоков.
 * Своего мнения о синтаксисе у нас нет и не надо: за ошибки отвечает настоящий
 * компилятор, его сообщения приезжают сюда подчёркиваниями (`goDiagnostics`).
 *
 * Цвета берутся из CSS-переменных сайта, а не задаются в теме числами: тему
 * страницы переключает класс на <html>, и редактор обязан переключиться вместе
 * с ней, не пересоздаваясь.
 */

import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
import { bracketMatching, foldGutter, foldKeymap, indentOnInput, syntaxHighlighting, HighlightStyle } from '@codemirror/language'
import { go } from '@codemirror/lang-go'
import { PostgreSQL, sql } from '@codemirror/lang-sql'
import { setDiagnostics, type Diagnostic } from '@codemirror/lint'
import { Compartment, EditorState, StateEffect, StateField, type Extension, type Text } from '@codemirror/state'
import {
  Decoration,
  EditorView,
  gutter,
  GutterMarker,
  highlightActiveLine,
  highlightActiveLineGutter,
  keymap,
  lineNumbers,
  type DecorationSet,
} from '@codemirror/view'
import { tags as t } from '@lezer/highlight'

/** Переключается на лету: задачи «что выведет» и «найди баг» открываются запертыми. */
export const readOnlyCompartment = new Compartment()

/** Подсветка синтаксиса. Значения — переменные из global.css, они же меняются с темой. */
const goHighlight = HighlightStyle.define([
  { tag: [t.keyword, t.moduleKeyword, t.controlKeyword, t.operatorKeyword], color: 'var(--cm-keyword)' },
  { tag: [t.typeName, t.standard(t.typeName), t.namespace], color: 'var(--cm-type)' },
  { tag: [t.function(t.variableName), t.function(t.propertyName)], color: 'var(--cm-func)' },
  { tag: [t.string, t.special(t.string), t.character], color: 'var(--cm-string)' },
  { tag: [t.number, t.bool, t.null], color: 'var(--cm-number)' },
  { tag: [t.comment, t.lineComment, t.blockComment], color: 'var(--cm-comment)', fontStyle: 'italic' },
  { tag: [t.operator, t.punctuation, t.separator, t.bracket], color: 'var(--cm-punct)' },
  { tag: t.propertyName, color: 'var(--cm-ink)' },
  { tag: t.invalid, color: 'var(--cm-error)' },
])

const theme = EditorView.theme({
  '&': {
    color: 'var(--cm-ink)',
    backgroundColor: 'var(--cm-bg)',
    fontSize: '13px',
  },
  '&.cm-focused': { outline: 'none' },
  '.cm-scroller': {
    fontFamily: 'var(--font-mono, ui-monospace, monospace)',
    lineHeight: '1.6',
  },
  '.cm-content': { caretColor: 'var(--cm-caret)', padding: '12px 0' },
  '.cm-cursor, .cm-dropCursor': { borderLeftColor: 'var(--cm-caret)' },
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
    backgroundColor: 'var(--cm-selection)',
  },
  '.cm-activeLine': { backgroundColor: 'var(--cm-active-line)' },
  '.cm-gutters': {
    backgroundColor: 'var(--cm-gutter-bg)',
    color: 'var(--cm-gutter-ink)',
    border: 'none',
    borderRight: '1px solid var(--cm-gutter-line)',
  },
  '.cm-activeLineGutter': { backgroundColor: 'var(--cm-active-line)', color: 'var(--cm-ink)' },
  '.cm-lineNumbers .cm-gutterElement': { padding: '0 8px 0 12px', minWidth: '2.5ch' },
  '.cm-foldGutter .cm-gutterElement': { padding: '0 4px' },
  '.cm-tooltip': {
    backgroundColor: 'var(--cm-tooltip-bg)',
    border: '1px solid var(--cm-gutter-line)',
    color: 'var(--cm-ink)',
    borderRadius: '6px',
  },
  '.cm-diagnostic': { padding: '4px 8px', fontFamily: 'var(--font-mono, monospace)', fontSize: '12px' },
  '.cm-diagnostic-error': { borderLeft: '3px solid var(--cm-error)' },
  '.cm-lintRange-error': {
    backgroundImage: 'none',
    borderBottom: '2px wavy var(--cm-error)',
    textDecoration: 'underline wavy var(--cm-error)',
  },
})

// ─── Выбор строки: механика задач «найди баг» ────────────────────────────────

/** Отметить строку (с единицы) или снять отметку нулём. */
export const setPickedLine = StateEffect.define<number>()

const pickedMark = Decoration.line({ class: 'cm-pickedLine' })

/** Какая строка отмечена сейчас. Ноль — никакая. */
export const pickedLineField = StateField.define<{ line: number; deco: DecorationSet }>({
  create: () => ({ line: 0, deco: Decoration.none }),
  update(value, tr) {
    for (const e of tr.effects) {
      if (!e.is(setPickedLine)) continue
      const line = e.value
      if (line < 1 || line > tr.state.doc.lines) return { line: 0, deco: Decoration.none }
      const pos = tr.state.doc.line(line).from
      return { line, deco: Decoration.set([pickedMark.range(pos)]) }
    }
    // Правка документа сбивает нумерацию — отметка перестаёт что-либо значить.
    if (tr.docChanged) return { line: 0, deco: Decoration.none }
    return value
  },
  provide: (f) => EditorView.decorations.from(f, (v) => v.deco),
})

/**
 * Маркер рисуется для КАЖДОЙ строки, а не только для выбранной: пустая колонка,
 * о которой сказано лишь в подсказке, — это загадка, а не механика. Полый
 * кружок говорит «сюда можно ткнуть», закрашенный — «выбрано здесь».
 */
class PickMarker extends GutterMarker {
  constructor(private readonly picked: boolean) {
    super()
  }

  override eq(other: PickMarker) {
    return other.picked === this.picked
  }

  override toDOM() {
    const el = document.createElement('span')
    el.className = this.picked ? 'cm-pickDot is-picked' : 'cm-pickDot'
    el.textContent = this.picked ? '●' : '○'
    return el
  }
}

const pickedMarker = new PickMarker(true)
const emptyMarker = new PickMarker(false)

/**
 * Колонка слева от номеров: показывает отметку и принимает клик.
 * Клик по самим номерам строк CodeMirror тратит на выделение строки,
 * поэтому под выбор отведена отдельная колонка.
 */
function pickGutter(onPick: (line: number) => void): Extension {
  return gutter({
    class: 'cm-pickGutter',
    lineMarker: (view, line) => {
      const picked = view.state.field(pickedLineField, false)?.line
      return picked && view.state.doc.lineAt(line.from).number === picked ? pickedMarker : emptyMarker
    },
    lineMarkerChange: (update) => update.transactions.some((tr) => tr.effects.some((e) => e.is(setPickedLine))),
    initialSpacer: () => emptyMarker,
    domEventHandlers: {
      mousedown(view, line) {
        const n = view.state.doc.lineAt(line.from).number
        const current = view.state.field(pickedLineField, false)?.line ?? 0
        view.dispatch({ effects: setPickedLine.of(n === current ? 0 : n) })
        onPick(n === current ? 0 : n)
        return true
      },
    },
  })
}

// ─── Ошибки от настоящего Go ─────────────────────────────────────────────────

/**
 * Разбор сообщений компилятора, vet и gofmt. Формат у всех один:
 * `prog.go:9:36: текст` (иногда с `./` впереди). Строки, которые в него не
 * укладываются, — продолжение предыдущего сообщения, и они к нему и цепляются.
 */
export function goDiagnostics(text: string, doc: Text, severity: 'error' | 'warning' = 'error'): Diagnostic[] {
  const out: Diagnostic[] = []
  for (const raw of text.split('\n')) {
    const m = raw.match(/^\.?\/?\S*?\.go:(\d+)(?::(\d+))?:\s*(.+)$/)
    if (!m) {
      const last = out[out.length - 1]
      if (last && raw.trim()) last.message += `\n${raw.trim()}`
      continue
    }
    const lineNo = Math.min(Number(m[1]), doc.lines)
    const line = doc.line(lineNo)
    const col = m[2] ? Math.max(0, Number(m[2]) - 1) : 0
    out.push({
      from: Math.min(line.from + col, line.to),
      to: line.to,
      severity,
      message: m[3] ?? raw,
    })
  }
  return out
}

/**
 * Ошибка Postgres в редакторе. Postgres сообщает позицию в символах с единицы;
 * подчёркиваем слово, на которое она указывает, — обычно это и есть виновник:
 * неизвестная колонка, лишняя запятая, опечатка в ключевом слове.
 */
export function sqlDiagnostic(message: string, position: number | undefined, doc: Text): Diagnostic[] {
  if (doc.length === 0) return []
  if (position === undefined) {
    const first = doc.line(1)
    return [{ from: first.from, to: first.to, severity: 'error', message }]
  }
  const from = Math.min(Math.max(0, position - 1), doc.length)
  const line = doc.lineAt(from)
  const word = doc.sliceString(from, line.to).match(/^[\p{L}\p{N}_."]+/u)?.[0].length ?? 1
  return [{ from, to: Math.min(from + word, line.to), severity: 'error', message }]
}

/** Показать (или стереть, если список пуст) сообщения Go в редакторе. */
export function showDiagnostics(view: EditorView, diagnostics: Diagnostic[]): void {
  view.dispatch(setDiagnostics(view.state, diagnostics))
}

// ─── Сборка ──────────────────────────────────────────────────────────────────

export interface EditorOptions {
  /** Ctrl/Cmd+Enter. */
  onRun: () => void
  /** Ctrl/Cmd+S — вместо «сохранить страницу», как в любом редакторе кода. */
  onFormat: () => void
  onChange: (code: string) => void
  /** Задачам «найди баг» нужна колонка выбора строки. */
  onPick?: (line: number) => void
  readOnly: boolean
  /** Грамматика подсветки. По умолчанию Go. */
  language?: 'go' | 'sql'
}

export function editorExtensions(opts: EditorOptions): Extension[] {
  return [
    // Колонка выбора идёт первой, чтобы оказаться левее номеров строк:
    // подсказка на странице обещает именно «слева от номера».
    ...(opts.onPick ? [pickGutter(opts.onPick)] : []),
    lineNumbers(),
    highlightActiveLineGutter(),
    highlightActiveLine(),
    foldGutter(),
    history(),
    indentOnInput(),
    bracketMatching(),
    opts.language === 'sql' ? sql({ dialect: PostgreSQL }) : go(),
    syntaxHighlighting(goHighlight),
    pickedLineField,
    theme,
    EditorView.lineWrapping,
    readOnlyCompartment.of([EditorState.readOnly.of(opts.readOnly), EditorView.editable.of(!opts.readOnly)]),
    EditorView.updateListener.of((u) => {
      if (u.docChanged) opts.onChange(u.state.doc.toString())
    }),
    keymap.of([
      {
        key: 'Mod-Enter',
        preventDefault: true,
        run: () => {
          opts.onRun()
          return true
        },
      },
      {
        key: 'Mod-s',
        preventDefault: true,
        run: () => {
          opts.onFormat()
          return true
        },
      },
      // Tab отдан отступу: в Go отступы — табы, и это не предмет для спора.
      // В SQL тоже удобнее, чем прыжок фокуса прочь из редактора.
      indentWithTab,
      ...defaultKeymap,
      ...historyKeymap,
      ...foldKeymap,
    ]),
  ]
}
