import type { MapConfig } from '../engine/types.ts'
import type { ParseError } from '../engine/script.ts'

/**
 * Ручки стенда. Всё, что здесь меняется, превращается в новый сценарий —
 * движок о существовании UI не знает.
 */

const MAX_TABLE = [16, 32, 64, 128, 1024]

export function ConfigKnobs({ config: c, onChange }: { config: MapConfig; onChange: (c: MapConfig) => void }) {
  const set = <K extends keyof MapConfig>(k: K, v: MapConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <div className="knobs">
      <label className="knob" title="Swiss table — карта Go с версии 1.24. Корзины — классическая реализация, которую до сих пор спрашивают на собеседованиях">
        <span className="knob-head">
          <span>Реализация</span>
        </span>
        <select value={c.impl} onChange={(e) => set('impl', e.target.value as MapConfig['impl'])}>
          <option value="swiss">Swiss table — Go 1.24+</option>
          <option value="legacy">корзины — до Go 1.24</option>
        </select>
      </label>
      <label
        className="knob"
        title="Потолок ёмкости одной таблицы. В рантайме — 1024 слота; меньше — чтобы деление таблиц и рост каталога было видно на десятках ключей"
      >
        <span className="knob-head">
          <span>Потолок таблицы</span>
        </span>
        <select value={c.maxTableSlots} disabled={c.impl !== 'swiss'} onChange={(e) => set('maxTableSlots', Number(e.target.value))}>
          {MAX_TABLE.map((n) => (
            <option key={n} value={n}>
              {n} слотов{n === 1024 ? ' (как в рантайме)' : ''}
            </option>
          ))}
        </select>
      </label>
      <label className="knob" title="make(map[string]int, hint): 0 — без подсказки. С подсказкой всё выделяется сразу, и вставки до hint идут без роста">
        <span className="knob-head">
          <span>make(…, hint)</span>
          <b>{c.hint || 'нет'}</b>
        </span>
        <input type="range" min={0} max={200} step={1} value={c.hint} onChange={(e) => set('hint', Number(e.target.value))} />
      </label>
      <label className="knob knob-toggle" title="var m map[string]int — карту объявили, но не создали через make">
        <input type="checkbox" checked={c.nilMap} onChange={(e) => set('nilMap', e.target.checked)} />
        <span>nil-карта (без make)</span>
      </label>
    </div>
  )
}

export function ScriptEditor({ script, errors, onChange }: {
  script: string
  errors: ParseError[]
  onChange: (s: string) => void
}) {
  return (
    <div className="map-script">
      <textarea
        value={script}
        onChange={(e) => onChange(e.target.value)}
        rows={Math.min(16, Math.max(6, script.split('\n').length + 1))}
        spellCheck={false}
        aria-label="Сценарий: по команде на строку"
      />
      {errors.length > 0 && (
        <ul className="map-errors">
          {errors.slice(0, 5).map((e) => (
            <li key={e.line}>
              строка {e.line}: {e.message}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
