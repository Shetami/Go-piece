import type { LoadScenario } from './types.ts'

/**
 * Пресеты стенда. У каждого одна мысль, которую он доказывает, — поле claim.
 * Во всех сервис один и тот же: 4 воркера, запрос в среднем 4 тика, то есть
 * ёмкость — один запрос за тик. Меняется нагрузка и поведение клиентов.
 */

export const little: LoadScenario = {
  id: 'little',
  title: 'Закон Литтла',
  claim:
    'Сколько запросов одновременно внутри сервиса, столько же, сколько приходит за тик, умноженное на время, которое каждый там проводит: L = λ·W. Это верно для любой очереди, если она не растёт бесконечно.',
  config: { rate: 0.6 },
  phases: [],
  watchFor: [],
  stopAfter: 200,
}

export const knee: LoadScenario = {
  id: 'knee',
  title: 'Колено',
  claim:
    'Сервис загружен на 90%, запас ещё есть — а очередь то и дело вырастает до десятка запросов. Нагрузка выросла в 1,8 раза по сравнению с 50%, а среднее ожидание в очереди — почти в двадцать раз.',
  config: { rate: 0.9, seed: 3 },
  phases: [],
  watchFor: ['queue.grow'],
  stopAfter: 240,
}

export const tail: LoadScenario = {
  id: 'tail',
  title: 'Медленный хвост',
  claim:
    'Три запроса из ста в 10 раз медленнее остальных. Медиана не меняется, а p99 вырастает в несколько раз. И страдают не только медленные: обычные запросы стоят в очереди, пока медленные занимают воркеры.',
  config: { rate: 0.6, slowShare: 3, slowFactor: 10, seed: 9 },
  phases: [],
  watchFor: ['req.slow', 'queue.grow'],
  stopAfter: 240,
}

export const burst: LoadScenario = {
  id: 'burst',
  title: 'Всплеск',
  claim:
    'На 20 тиков нагрузка удваивается. Всплеск кончился, а очередь рассасывается ещё дольше, чем он длился: разбирать её можно только запасом ёмкости, а он небольшой. Хуже всего задержка уже после всплеска.',
  config: { rate: 0.7, seed: 3 },
  phases: [
    { at: 40, x: 2 },
    { at: 60, x: 1 },
  ],
  watchFor: ['load.change', 'queue.grow', 'queue.drained'],
  stopAfter: 160,
}

export const retryStorm: LoadScenario = {
  id: 'retry-storm',
  title: 'Шторм повторов',
  claim:
    'Короткий всплеск, клиенты с таймаутом и тремя повторами. Всплеск давно кончился, а сервис загружен на 100% и почти не отвечает: он обрабатывает запросы, которых уже никто не ждёт, а их повторы стоят в очереди.',
  config: { rate: 0.8, timeout: 12, retries: 3, seed: 7 },
  phases: [
    { at: 30, x: 2 },
    { at: 45, x: 1 },
  ],
  watchFor: ['load.change', 'req.timeout', 'req.wasted', 'overload.collapse', 'overload.recover'],
  stopAfter: 200,
}

export const shedding: LoadScenario = {
  id: 'shedding',
  title: 'Сброс нагрузки',
  claim:
    'Нагрузка на 30% выше ёмкости. Без предела очередь растёт, и каждый следующий запрос ждёт дольше предыдущего. С ограниченной очередью лишние сразу получают отказ, а принятые обслуживаются быстро.',
  config: { rate: 1.3, queueLimit: 0 },
  phases: [],
  watchFor: ['queue.grow', 'req.reject'],
  stopAfter: 160,
}

export const LOAD_SCENARIOS: LoadScenario[] = [little, knee, tail, burst, retryStorm, shedding]

export function loadScenarioById(id: string): LoadScenario | undefined {
  return LOAD_SCENARIOS.find((s) => s.id === id)
}
