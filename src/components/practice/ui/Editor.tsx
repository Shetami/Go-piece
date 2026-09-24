/**
 * Обёртка CodeMirror для React.
 *
 * Редактор создаётся один раз и живёт своей жизнью: перерисовки React его не
 * трогают. Снаружи в него можно только положить новый текст (`code` сменился
 * не из-за набора — сброс, вставка решения, форматирование) и переключить
 * запертость. Обработчики держим в ref, чтобы не пересобирать расширения на
 * каждый рендер и не терять историю отмен.
 */

import { useEffect, useImperativeHandle, useRef, type Ref } from 'react'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import type { Diagnostic } from '@codemirror/lint'
import { editorExtensions, readOnlyCompartment, setPickedLine, showDiagnostics } from './editor.ts'

export interface EditorHandle {
  /** Подчеркнуть места, на которые пожаловался Go. Пустой список стирает подчёркивания. */
  setDiagnostics: (diagnostics: Diagnostic[]) => void
  focus: () => void
}

interface Props {
  code: string
  onChange: (code: string) => void
  onRun: () => void
  onFormat: () => void
  /** Есть — значит у задачи механика «ткни в строку». */
  onPick?: (line: number) => void
  pickedLine?: number
  readOnly: boolean
  /** Грамматика подсветки; задаётся один раз при создании редактора. */
  language?: 'go' | 'sql'
  ref?: Ref<EditorHandle>
}

export default function Editor({ code, onChange, onRun, onFormat, onPick, pickedLine, readOnly, language = 'go', ref }: Props) {
  const host = useRef<HTMLDivElement>(null)
  const view = useRef<EditorView | null>(null)

  // Свежие обработчики без пересборки редактора.
  const handlers = useRef({ onChange, onRun, onFormat, onPick })
  handlers.current = { onChange, onRun, onFormat, onPick }

  useEffect(() => {
    if (!host.current) return

    const state = EditorState.create({
      doc: code,
      extensions: editorExtensions({
        onChange: (value) => handlers.current.onChange(value),
        onRun: () => handlers.current.onRun(),
        onFormat: () => handlers.current.onFormat(),
        onPick: onPick ? (line) => handlers.current.onPick?.(line) : undefined,
        readOnly,
        language,
      }),
    })

    const editor = new EditorView({ state, parent: host.current })
    view.current = editor
    return () => {
      editor.destroy()
      view.current = null
    }
    // Намеренно неполный список зависимостей: пересоздавать редактор можно
    // только при смене самой механики задачи — когда появляется или исчезает
    // колонка выбора строки. Всё остальное меняется через dispatch.
  }, [Boolean(onPick)])

  // Текст сменился снаружи — подменяем документ целиком.
  useEffect(() => {
    const editor = view.current
    if (!editor || editor.state.doc.toString() === code) return
    editor.dispatch({ changes: { from: 0, to: editor.state.doc.length, insert: code } })
  }, [code])

  useEffect(() => {
    view.current?.dispatch({
      effects: readOnlyCompartment.reconfigure([EditorState.readOnly.of(readOnly), EditorView.editable.of(!readOnly)]),
    })
  }, [readOnly])

  useEffect(() => {
    if (pickedLine === undefined) return
    view.current?.dispatch({ effects: setPickedLine.of(pickedLine) })
  }, [pickedLine])

  useImperativeHandle(ref, () => ({
    setDiagnostics: (diagnostics) => {
      if (view.current) showDiagnostics(view.current, diagnostics)
    },
    focus: () => view.current?.focus(),
  }))

  return <div ref={host} className="cm-host" data-readonly={readOnly || undefined} />
}
