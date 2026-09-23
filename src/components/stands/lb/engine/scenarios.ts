import type { LbScenario } from './types.ts'

/**
 * Пресеты стенда. У каждого одна мысль, которую он доказывает, — поле claim.
 * Реплика — 2 воркера, запрос в среднем 4 тика, то есть ползапроса за тик;
 * четыре реплики вместе — два запроса за тик.
 */

export const slowReplica: LbScenario = {
  id: 'slow-replica',
  title: 'Медленная реплика',
  claim:
    'Одна из четырёх реплик стала вчетверо медленнее, но проверку здоровья проходит. Балансировка по кругу продолжает слать ей четверть запросов — её очередь растёт без предела. Наименьшее число соединений замечает это само.',
  config: { algo: 'round-robin', rate: 1.2 },
  phases: [],
  faults: [{ kind: 'slow', replica: 3, at: 20, factor: 4 }],
  watchFor: ['replica.fault', 'queue.grow'],
  stopAfter: 160,
}

export const twoChoices: LbScenario = {
  id: 'two-choices',
  title: 'Два случайных',
  claim:
    'Реплики одинаковые, нагрузка 80%. Случайный выбор то и дело сваливает несколько запросов на одну реплику, пока соседняя простаивает. Достаточно взять две случайные и отправить на менее занятую — и очереди почти как при полном знании.',
  config: { algo: 'random', rate: 1.6 },
  phases: [],
  faults: [],
  watchFor: ['queue.grow'],
  stopAfter: 200,
}

export const blackHole: LbScenario = {
  id: 'black-hole',
  title: 'Чёрная дыра',
  claim:
    'Реплика сломалась так, что на любой запрос мгновенно отвечает 500, а /healthz проходит. У неё всегда ноль незавершённых запросов — и «наименьшее число соединений» отправляет на неё большую часть трафика.',
  config: { algo: 'least-conn', rate: 1.2 },
  phases: [],
  faults: [{ kind: 'errors', replica: 1, at: 20 }],
  watchFor: ['replica.fault', 'outlier.eject'],
  stopAfter: 120,
}

export const hang: LbScenario = {
  id: 'hang',
  title: 'Зависшая реплика',
  claim:
    'Реплика зависла: соединения принимает, но не отвечает. Пока проверки здоровья её не исключат, каждый четвёртый запрос ждёт до таймаута. Повтор на другой реплике спасает запрос, но не время пользователя.',
  config: { algo: 'round-robin', rate: 1.2, timeout: 12, retries: 1, healthEvery: 10, healthFails: 3 },
  phases: [],
  faults: [{ kind: 'hang', replica: 2, at: 30 }],
  watchFor: ['replica.fault', 'req.timeout', 'health.fail', 'health.eject'],
  stopAfter: 120,
}

export const autoscale: LbScenario = {
  id: 'autoscale',
  title: 'Автомасштабирование',
  claim:
    'Нагрузка вдвое выросла и осталась. Автомасштабирование замечает это через десяток тиков, а новые реплики запускаются ещё 25: всё это время очередь растёт. Масштабирование спасает от долгого роста, но не от внезапного.',
  config: { algo: 'least-conn', replicas: 3, rate: 0.9, autoscale: true, scaleTarget: 70, bootTime: 25 },
  phases: [{ at: 40, x: 2.5 }],
  faults: [],
  watchFor: ['load.change', 'scale.decide', 'scale.ready', 'queue.grow'],
  stopAfter: 160,
}

export const LB_SCENARIOS: LbScenario[] = [slowReplica, twoChoices, blackHole, hang, autoscale]

export function lbScenarioById(id: string): LbScenario | undefined {
  return LB_SCENARIOS.find((s) => s.id === id)
}
