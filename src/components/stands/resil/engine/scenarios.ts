import type { ResilScenario } from './types.ts'

/**
 * Пресеты стенда. У каждого одна мысль, которую он доказывает, — поле claim.
 * Сервис A: 6 воркеров, своя работа 2 тика. Сервис B: 4 воркера, ответ 4 тика,
 * то есть один запрос за тик. Половине запросов нужен B.
 */

export const cascade: ResilScenario = {
  id: 'cascade',
  title: 'Каскад',
  claim:
    'Зависимость стала вшестеро медленнее — и сервис встал целиком. Его воркеры заняты ожиданием ответа, поэтому запросы, которым зависимость вообще не нужна, тоже стоят в очереди.',
  config: { timeout: 0, breaker: false },
  phases: [],
  faults: [{ kind: 'slow', at: 40, factor: 6 }],
  watchFor: ['dep.fault', 'pool.saturated', 'queue.grow'],
  stopAfter: 160,
}

export const breaker: ResilScenario = {
  id: 'breaker',
  title: 'Предохранитель',
  claim:
    'Зависимость перестала отвечать вовсе. С таймаутом сервис жив, но каждый запрос к ней честно ждёт таймаут и только потом падает. Предохранитель после нескольких неудач перестаёт в неё ходить и раз в 20 тиков проверяет, не ожила ли она.',
  config: { timeout: 8, retries: 1, breaker: false, breakerFails: 5, breakerOpen: 20 },
  phases: [],
  faults: [{ kind: 'hang', at: 40, until: 120 }],
  watchFor: ['dep.fault', 'dep.timeout', 'breaker.open', 'breaker.close', 'dep.recover'],
  stopAfter: 180,
}

export const bulkhead: ResilScenario = {
  id: 'bulkhead',
  title: 'Переборки',
  claim:
    'Та же медленная зависимость, но ждать её могут не больше двух воркеров из шести. Запросы к зависимости страдают, остальные — нет: перегородка не даёт воде залить весь корпус.',
  config: { timeout: 0, bulkhead: 0, fallback: true },
  phases: [],
  faults: [{ kind: 'slow', at: 40, factor: 6 }],
  watchFor: ['dep.fault', 'bulkhead.block', 'pool.saturated'],
  stopAfter: 160,
}

export const rateLimit: ResilScenario = {
  id: 'rate-limit',
  title: 'Ограничитель частоты',
  claim:
    'Нагрузка внезапно выросла втрое и держится. Без ограничителя очередь растёт, и плохо становится всем. Ограничитель пропускает столько, сколько сервис может, а лишним сразу отказывает.',
  config: { rateLimit: 0, burst: 5, timeout: 12 },
  phases: [
    { at: 40, x: 3 },
    { at: 100, x: 1 },
  ],
  faults: [],
  watchFor: ['load.change', 'queue.grow', 'req.reject'],
  stopAfter: 180,
}

export const degrade: ResilScenario = {
  id: 'degrade',
  title: 'Деградация',
  claim:
    'Зависимость лежит надолго. С предохранителем и заглушкой сервис отвечает всем — половине полноценно, половине урезанно. Без заглушки половина запросов превращается в ошибки.',
  config: { timeout: 8, breaker: true, breakerFails: 5, breakerOpen: 20, fallback: false },
  phases: [],
  faults: [{ kind: 'hang', at: 30, until: 140 }],
  watchFor: ['breaker.open', 'req.fallback', 'req.fail', 'dep.recover'],
  stopAfter: 180,
}

export const RESIL_SCENARIOS: ResilScenario[] = [cascade, breaker, bulkhead, rateLimit, degrade]

export function resilScenarioById(id: string): ResilScenario | undefined {
  return RESIL_SCENARIOS.find((s) => s.id === id)
}
