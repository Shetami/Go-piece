/**
 * Вывод прогона.
 *
 * Песочница go.dev возвращает вывод не одним куском, а списком событий с
 * паузами между ними: часы там поддельные, `time.Sleep` не стоит ничего, но
 * сколько бы прошло на настоящих часах — известно. Мы эти паузы проигрываем.
 * На конкурентных задачах это единственное, что видно снаружи: кто за кем
 * успел и где программа на самом деле ждала.
 */

import { useEffect, useRef, useState } from 'react'
import type { OutputChunk } from '../protocol.ts'

/** Длиннее держать читателя незачем: смысл в порядке событий, а не в ожидании. */
const MAX_PAUSE_MS = 1500

interface Props {
  /** Ошибки компиляции: программа не запускалась. */
  errors: string
  /** Замечания go vet: запустилась, но, возможно, делает не то. */
  vet: string
  output: OutputChunk[]
  /** К коду дописали пустой main — об этом надо сказать, иначе сюрприз. */
  stubbedMain: boolean
  running: boolean
  /** Ошибка самого запроса, а не программы. */
  failure?: string
}

/**
 * Сколько кусков вывода уже можно показать. Таймеры живут на своём поколении
 * (`runId`), иначе хвост прошлого прогона дорисовался бы поверх нового.
 */
function usePlayback(output: OutputChunk[]): number {
  const [shown, setShown] = useState(output.length)
  const timers = useRef<ReturnType<typeof setTimeout>[]>([])

  useEffect(() => {
    for (const id of timers.current) clearTimeout(id)
    timers.current = []

    const paused = output.some((c) => c.delay > 0)
    if (!paused) {
      setShown(output.length)
      return
    }

    setShown(0)
    let at = 0
    output.forEach((chunk, i) => {
      at += Math.min(chunk.delay / 1e6, MAX_PAUSE_MS)
      timers.current.push(setTimeout(() => setShown(i + 1), at))
    })

    return () => {
      for (const id of timers.current) clearTimeout(id)
      timers.current = []
    }
  }, [output])

  return shown
}

export default function Console({ errors, vet, output, stubbedMain, running, failure }: Props) {
  const shown = usePlayback(output)
  const visible = output.slice(0, shown)
  const playing = shown < output.length

  const empty = !errors && !vet && !failure && output.length === 0

  return (
    <div className="pr-console">
      <div className="pr-console-head">
        <span>Вывод</span>
        {running && <span className="pr-spinner">компилируется на go.dev…</span>}
        {playing && !running && <span className="pr-spinner">идёт…</span>}
      </div>

      <div className="pr-console-body">
        {failure && <div className="pr-msg pr-msg-fail">{failure}</div>}

        {errors && (
          <div className="pr-msg pr-msg-error">
            <div className="pr-msg-title">Не компилируется</div>
            <pre>{errors.trimEnd()}</pre>
          </div>
        )}

        {vet && (
          <div className="pr-msg pr-msg-vet">
            <div className="pr-msg-title">go vet</div>
            <pre>{vet.trimEnd()}</pre>
          </div>
        )}

        {stubbedMain && !errors && (
          <div className="pr-note">
            В коде нет <code>func main</code> — дописан пустой, чтобы было что компилировать.
          </div>
        )}

        {visible.length > 0 && (
          <pre className="pr-output">
            {visible.map((chunk, i) => (
              <span key={i} className={chunk.kind === 'stderr' ? 'pr-stderr' : undefined}>
                {chunk.text}
              </span>
            ))}
            {playing && <span className="pr-caret">▋</span>}
          </pre>
        )}

        {empty && !running && <div className="pr-empty">Пусто. Запустите код — сюда приедет то, что он напечатает.</div>}
      </div>
    </div>
  )
}
