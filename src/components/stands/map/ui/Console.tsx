import { useState } from 'react'
import { parseScript } from '../engine/script.ts'

/**
 * Консоль: команда выполняется в конце текущего прогона, и стенд проигрывает
 * её у вас на глазах. Тот же язык, что в сценарии.
 */

const QUICK: { label: string; cmd: string; hint: string }[] = [
  { label: '+1 ключ', cmd: 'fill 1', hint: 'Новый ключ k1, k2, …' },
  { label: '+10 ключей', cmd: 'fill 10', hint: 'Десять новых ключей подряд — удобно, чтобы дойти до роста' },
  { label: 'коллизии ×6', cmd: 'collide 6', hint: 'Шесть ключей с одной стартовой группой (корзиной) — видно пробы и переполнение' },
  { label: 'удалить их', cmd: 'del @last', hint: 'Удалить все ключи последней пачки fill или collide' },
  { label: 'range', cmd: 'range', hint: 'Начать обход: for k, v := range m' },
  { label: 'next', cmd: 'next', hint: 'Следующая итерация обхода' },
  { label: 'next all', cmd: 'next all', hint: 'Дообойти до конца' },
  { label: 'clear', cmd: 'clear', hint: 'clear(m) — память при этом не возвращается' },
]

export function Console({ onRun, disabled }: { onRun: (cmd: string) => void; disabled: string | null }) {
  const [text, setText] = useState('')
  const [error, setError] = useState<string | null>(null)

  const run = (cmd: string) => {
    const { ops, errors } = parseScript(cmd)
    if (errors.length > 0) {
      setError(errors[0]!.message)
      return
    }
    if (ops.length === 0) return
    setError(null)
    setText('')
    onRun(cmd.trim())
  }

  return (
    <div className="map-console">
      <form
        className="map-console-line"
        onSubmit={(e) => {
          e.preventDefault()
          run(text)
        }}
      >
        <span aria-hidden="true">›</span>
        <input
          value={text}
          onChange={(e) => {
            setText(e.target.value)
            setError(null)
          }}
          placeholder='m["ann"] = 1 · get ann · delete(m, "ann") · fill 20 · collide 6 · range · next'
          aria-label="Команда для карты"
          disabled={disabled !== null}
          spellCheck={false}
          autoComplete="off"
        />
        <button type="submit" className="btn btn-primary" disabled={disabled !== null || text.trim() === ''}>
          выполнить
        </button>
      </form>
      <div className="map-quick">
        {QUICK.map((q) => (
          <button key={q.cmd} type="button" className="btn" title={q.hint} disabled={disabled !== null} onClick={() => run(q.cmd)}>
            {q.label}
          </button>
        ))}
      </div>
      {(error || disabled) && <p className="map-console-err">{error ?? disabled}</p>}
    </div>
  )
}
