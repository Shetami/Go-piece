import type { Acks, CommitMode, ConsumerSpec, Fault, KafkaConfig, ProducerSpec } from '../engine/types.ts'

/**
 * Ручки стенда. Всё, что здесь меняется, превращается в новый сценарий —
 * движок о существовании UI не знает.
 */

function Toggle({ label, hint, checked, onChange }: {
  label: string
  hint: string
  checked: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <label className="knob knob-toggle" title={hint}>
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      <span>{label}</span>
    </label>
  )
}

function Range({ label, hint, value, min, max, step = 1, display, onChange }: {
  label: string
  hint: string
  value: number
  min: number
  max: number
  step?: number
  display?: string
  onChange: (v: number) => void
}) {
  return (
    <label className="knob" title={hint}>
      <span className="knob-head">
        <span>{label}</span>
        <b>{display ?? value}</b>
      </span>
      <input type="range" min={min} max={max} step={step} value={value} onChange={(e) => onChange(Number(e.target.value))} />
    </label>
  )
}

function Choice<T extends string | number>({ label, hint, value, options, onChange }: {
  label: string
  hint: string
  value: T
  options: { value: T; label: string }[]
  onChange: (v: T) => void
}) {
  return (
    <label className="knob" title={hint}>
      <span className="knob-head">
        <span>{label}</span>
      </span>
      <span className="kf-choice" role="radiogroup" aria-label={label}>
        {options.map((o) => (
          <button
            key={String(o.value)}
            type="button"
            role="radio"
            aria-checked={o.value === value}
            className={o.value === value ? 'is-now' : ''}
            onClick={() => onChange(o.value)}
          >
            {o.label}
          </button>
        ))}
      </span>
    </label>
  )
}

const t = (n: number) => `${n} т.`

export function ConfigKnobs({ config: c, lab, onChange }: { config: KafkaConfig; lab: boolean; onChange: (c: KafkaConfig) => void }) {
  const set = <K extends keyof KafkaConfig>(k: K, v: KafkaConfig[K]) => onChange({ ...c, [k]: v })
  return (
    <>
      <h4 className="settings-sub">Продюсер</h4>
      <div className="knobs">
        <Choice<Acks>
          label="acks"
          hint="Сколько подтверждений ждать: 0 — никаких, 1 — только лидера, all — всех реплик из ISR"
          value={c.acks}
          options={[
            { value: 0, label: '0' },
            { value: 1, label: '1' },
            { value: 'all', label: 'all' },
          ]}
          onChange={(v) => set('acks', v)}
        />
        <Range
          label="linger.ms"
          hint="Сколько пакет ждёт попутчиков. Больше — меньше запросов и выше задержка"
          value={c.lingerTicks}
          min={0}
          max={20}
          display={t(c.lingerTicks)}
          onChange={(v) => set('lingerTicks', v)}
        />
        <Range
          label="batch.size"
          hint="Предельный размер пакета. В Kafka — в байтах, в модели — в записях"
          value={c.batchSize}
          min={1}
          max={10}
          onChange={(v) => set('batchSize', v)}
        />
        <Range
          label="max.in.flight"
          hint="Сколько запросов на одно соединение может быть в пути одновременно"
          value={c.maxInFlight}
          min={1}
          max={5}
          onChange={(v) => set('maxInFlight', v)}
        />
        <Toggle
          label="Идемпотентность"
          hint="enable.idempotence: брокер узнаёт повтор пакета по номеру последовательности и не пишет его второй раз"
          checked={c.idempotence}
          onChange={(v) => set('idempotence', v)}
        />
        {lab && (
          <>
            <Range
              label="request.timeout"
              hint="Сколько ждать ответа, прежде чем повторить запрос"
              value={c.requestTimeout}
              min={4}
              max={40}
              display={t(c.requestTimeout)}
              onChange={(v) => set('requestTimeout', v)}
            />
            <Range
              label="delivery.timeout"
              hint="Сколько всего пытаться доставить пакет, прежде чем отдать ошибку приложению"
              value={c.deliveryTimeout}
              min={10}
              max={150}
              step={5}
              display={t(c.deliveryTimeout)}
              onChange={(v) => set('deliveryTimeout', v)}
            />
          </>
        )}
      </div>

      <h4 className="settings-sub">Кластер</h4>
      <div className="knobs">
        <Range
          label="Брокеров"
          hint="Сколько серверов в кластере"
          value={c.brokers}
          min={1}
          max={5}
          onChange={(v) => onChange({ ...c, brokers: v, replicationFactor: Math.min(c.replicationFactor, v) })}
        />
        <Range
          label="Партиций"
          hint="Партиция — единица параллелизма: в группе её читает ровно один потребитель"
          value={c.partitions}
          min={1}
          max={6}
          onChange={(v) => set('partitions', v)}
        />
        <Range
          label="Фактор репликации"
          hint="Сколько копий у каждой партиции, включая лидера"
          value={c.replicationFactor}
          min={1}
          max={c.brokers}
          onChange={(v) => set('replicationFactor', v)}
        />
        <Range
          label="min.insync.replicas"
          hint="При acks=all запись отклоняется, если в ISR меньше реплик"
          value={c.minInsyncReplicas}
          min={1}
          max={3}
          onChange={(v) => set('minInsyncReplicas', v)}
        />
        <Range
          label="replica.lag.time.max"
          hint="Сколько тиков фолловер может не догонять лидера, прежде чем вылетит из ISR"
          value={c.replicaLagMax}
          min={4}
          max={40}
          display={t(c.replicaLagMax)}
          onChange={(v) => set('replicaLagMax', v)}
        />
        <Toggle
          label="Нечистые выборы"
          hint="unclean.leader.election.enable: разрешить лидера не из ISR. Партиция оживёт, но данные могут пропасть"
          checked={c.uncleanElection}
          onChange={(v) => set('uncleanElection', v)}
        />
        {lab && (
          <Range
            label="Задержка сети"
            hint="Сколько тиков любой запрос летит в одну сторону"
            value={c.netLatency}
            min={1}
            max={4}
            display={t(c.netLatency)}
            onChange={(v) => set('netLatency', v)}
          />
        )}
      </div>

      <h4 className="settings-sub">Потребители</h4>
      <div className="knobs">
        <Choice<CommitMode>
          label="Коммит оффсета"
          hint="auto — по таймеру коммитится всё, что poll уже отдал; after-process — после обработки пачки"
          value={c.commitMode}
          options={[
            { value: 'after-process', label: 'после обработки' },
            { value: 'auto', label: 'авто' },
          ]}
          onChange={(v) => set('commitMode', v)}
        />
        <Range
          label="Обработка"
          hint="Сколько тиков приложение обрабатывает одно сообщение"
          value={c.processTicks}
          min={1}
          max={8}
          display={t(c.processTicks)}
          onChange={(v) => set('processTicks', v)}
        />
        <Range
          label="max.poll.records"
          hint="Сколько сообщений poll отдаёт за раз"
          value={c.maxPollRecords}
          min={1}
          max={10}
          onChange={(v) => set('maxPollRecords', v)}
        />
        <Range
          label="session.timeout"
          hint="Через сколько тиков без heartbeat группа отдаёт партиции упавшего"
          value={c.sessionTimeout}
          min={2}
          max={40}
          display={t(c.sessionTimeout)}
          onChange={(v) => set('sessionTimeout', v)}
        />
        {c.commitMode === 'auto' && (
          <Range
            label="auto.commit.interval"
            hint="Как часто автокоммит фиксирует позицию"
            value={c.autoCommitInterval}
            min={2}
            max={30}
            display={t(c.autoCommitInterval)}
            onChange={(v) => set('autoCommitInterval', v)}
          />
        )}
      </div>
    </>
  )
}

/* ─────────────────────────── нагрузка ─────────────────────────── */

const parseKeys = (s: string) => s.split(',').map((x) => x.trim()).filter(Boolean)

export function ProducersEditor({ producers, onChange }: { producers: ProducerSpec[]; onChange: (p: ProducerSpec[]) => void }) {
  const set = (i: number, next: ProducerSpec) => onChange(producers.map((x, j) => (j === i ? next : x)))
  return (
    <div className="chan-editor">
      {producers.map((p, i) => {
        const [lo, hi] = typeof p.every === 'number' ? [p.every, p.every] : p.every
        return (
          <div className="chan-edit" key={i}>
            <input className="wl-name" value={p.name} onChange={(e) => set(i, { ...p, name: e.target.value })} aria-label="Имя продюсера" />
            <label title="Сколько сообщений отправит за прогон">
              сообщений
              <input type="number" min={1} max={80} value={p.messages} onChange={(e) => set(i, { ...p, messages: Math.max(1, Math.min(80, Number(e.target.value))) })} />
            </label>
            <label title="Пауза между вызовами send(), в тиках">
              раз в
              <input type="number" min={1} max={20} value={lo} onChange={(e) => set(i, { ...p, every: [Math.max(1, Number(e.target.value)), Math.max(hi, Number(e.target.value))] })} />
              –
              <input type="number" min={1} max={20} value={hi} onChange={(e) => set(i, { ...p, every: [Math.min(lo, Number(e.target.value)), Math.max(1, Number(e.target.value))] })} />
            </label>
            <label title="Ключи через запятую. Пусто — сообщения без ключа">
              ключи
              <input
                className="kf-keys"
                value={(p.keys ?? []).join(', ')}
                placeholder="без ключа"
                onChange={(e) => {
                  const keys = parseKeys(e.target.value)
                  set(i, { ...p, keys, weights: undefined })
                }}
              />
            </label>
            {producers.length > 1 && (
              <button type="button" className="btn-ghost" onClick={() => onChange(producers.filter((_, j) => j !== i))} title="Удалить продюсера">
                ✕
              </button>
            )}
          </div>
        )
      })}
      <button
        type="button"
        className="btn"
        onClick={() => onChange([...producers, { name: `p${producers.length + 1}`, messages: 10, every: [2, 3] }])}
      >
        + продюсер
      </button>
    </div>
  )
}

export function ConsumersEditor({ consumers, onChange }: { consumers: ConsumerSpec[]; onChange: (c: ConsumerSpec[]) => void }) {
  const set = (i: number, next: ConsumerSpec) => onChange(consumers.map((x, j) => (j === i ? next : x)))
  return (
    <div className="chan-editor">
      {consumers.map((c, i) => (
        <div className="chan-edit" key={i}>
          <input className="wl-name" value={c.name} onChange={(e) => set(i, { ...c, name: e.target.value })} aria-label="Имя потребителя" />
          <label title="Потребители с одинаковой группой делят партиции">
            группа
            <input className="kf-keys" value={c.group} onChange={(e) => set(i, { ...c, group: e.target.value || 'group' })} />
          </label>
          <label title="На каком тике потребитель вступает в группу">
            с тика
            <input type="number" min={1} max={200} value={c.joinAt ?? 1} onChange={(e) => set(i, { ...c, joinAt: Math.max(1, Number(e.target.value)) })} />
          </label>
          {consumers.length > 1 && (
            <button type="button" className="btn-ghost" onClick={() => onChange(consumers.filter((_, j) => j !== i))} title="Удалить потребителя">
              ✕
            </button>
          )}
        </div>
      ))}
      <button
        type="button"
        className="btn"
        onClick={() => onChange([...consumers, { name: `c${consumers.length + 1}`, group: consumers[0]?.group ?? 'group' }])}
      >
        + потребитель
      </button>
    </div>
  )
}

const FAULT_LABEL: Record<Fault['kind'], string> = {
  'broker.down': 'брокер падает',
  'broker.up': 'брокер возвращается',
  'broker.slow': 'брокер тормозит',
  'consumer.crash': 'потребитель падает',
  'drop.response': 'сеть теряет ответ',
}

function defaultFault(kind: Fault['kind'], at: number, consumers: ConsumerSpec[], producers: ProducerSpec[]): Fault {
  switch (kind) {
    case 'broker.down':
    case 'broker.up':
      return { at, kind, broker: 0 }
    case 'broker.slow':
      return { at, kind, broker: 1, factor: 4 }
    case 'consumer.crash':
      return { at, kind, consumer: consumers[0]?.name ?? 'c1' }
    case 'drop.response':
      return { at, kind, producer: producers[0]?.name ?? 'p1' }
  }
}

export function FaultsEditor({ faults, brokers, consumers, producers, onChange }: {
  faults: Fault[]
  brokers: number
  consumers: ConsumerSpec[]
  producers: ProducerSpec[]
  onChange: (f: Fault[]) => void
}) {
  const set = (i: number, next: Fault) => onChange(faults.map((x, j) => (j === i ? next : x)))
  return (
    <div className="chan-editor">
      {faults.length === 0 && <p className="kf-muted">Сбоев нет — всё работает идеально. Добавьте, чтобы посмотреть, что сломается.</p>}
      {faults.map((f, i) => (
        <div className="chan-edit" key={i}>
          <label title="Тик, на котором случится сбой">
            тик
            <input type="number" min={1} max={300} value={f.at} onChange={(e) => set(i, { ...f, at: Math.max(1, Number(e.target.value)) })} />
          </label>
          <select value={f.kind} onChange={(e) => set(i, defaultFault(e.target.value as Fault['kind'], f.at, consumers, producers))} aria-label="Что случается">
            {Object.entries(FAULT_LABEL).map(([k, l]) => (
              <option key={k} value={k}>
                {l}
              </option>
            ))}
          </select>
          {(f.kind === 'broker.down' || f.kind === 'broker.up' || f.kind === 'broker.slow') && (
            <select value={f.broker} onChange={(e) => set(i, { ...f, broker: Number(e.target.value) })} aria-label="Брокер">
              {Array.from({ length: brokers }, (_, b) => (
                <option key={b} value={b}>
                  B{b}
                </option>
              ))}
            </select>
          )}
          {f.kind === 'broker.slow' && (
            <label title="Во сколько раз медленнее реплицирует. 1 — вернуть нормальную скорость">
              ×
              <input type="number" min={1} max={30} value={f.factor} onChange={(e) => set(i, { ...f, factor: Math.max(1, Number(e.target.value)) })} />
            </label>
          )}
          {f.kind === 'consumer.crash' && (
            <select value={f.consumer} onChange={(e) => set(i, { ...f, consumer: e.target.value })} aria-label="Потребитель">
              {consumers.map((c) => (
                <option key={c.name} value={c.name}>
                  {c.name}
                </option>
              ))}
            </select>
          )}
          {f.kind === 'drop.response' && (
            <select value={f.producer} onChange={(e) => set(i, { ...f, producer: e.target.value })} aria-label="Продюсер">
              {producers.map((p) => (
                <option key={p.name} value={p.name}>
                  {p.name}
                </option>
              ))}
            </select>
          )}
          <button type="button" className="btn-ghost" onClick={() => onChange(faults.filter((_, j) => j !== i))} title="Убрать сбой">
            ✕
          </button>
        </div>
      ))}
      <button
        type="button"
        className="btn"
        onClick={() => onChange([...faults, defaultFault('broker.down', 20, consumers, producers)])}
      >
        + сбой
      </button>
    </div>
  )
}
