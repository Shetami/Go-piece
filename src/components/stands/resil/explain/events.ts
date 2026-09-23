import type { ResilEvent, ResilEventType } from '../engine/types.ts'

/**
 * Разбор событий стенда. Объяснения привязаны к ТИПУ события: какой бы
 * сценарий ни крутился, клик по «breaker.open» откроет один и тот же разбор,
 * подставив в него конкретные числа из payload.
 */

export interface Explanation {
  title: string
  body: string[]
  /** Как это выглядит в сервисе на Go. */
  go?: string
  model?: string
  terms: string[]
}

const num = (v: unknown) => (typeof v === 'number' ? v : Number(v ?? 0))
const str = (v: unknown) => (typeof v === 'string' ? v : String(v ?? ''))

function plural(n: number, one: string, few: string, many: string): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few
  return many
}
const ticks = (n: number) => `${n} ${plural(n, 'тик', 'тика', 'тиков')}`
const reqs = (n: number) => `${n} ${plural(n, 'запрос', 'запроса', 'запросов')}`

const DEP: Record<string, string> = {
  ok: 'отвечает нормально',
  slow: 'отвечает медленно',
  errors: 'отвечает ошибками',
  hang: 'не отвечает вовсе',
}

type Explainer = (e: ResilEvent) => Explanation

const EXPLAIN: Record<ResilEventType, Explainer> = {
  'req.accept': (e) => ({
    title: `Запрос #${num(e.payload.id)} принят${e.payload.needsDep ? ' — ему нужен B' : ' — он обойдётся без B'}`,
    body: [
      e.payload.needsDep
        ? 'Такому запросу сервис A сначала сделает свою часть работы, а потом пойдёт в B и будет ждать ответа, занимая воркер.'
        : 'Этот запрос A обслужит сам. Но в общей очереди он стоит вместе со всеми — и если воркеры заняты ожиданием B, ждать будет и он.',
      `В очереди сейчас ${reqs(num(e.payload.queue))}.`,
    ],
    terms: ['cascading-failure'],
  }),

  'req.reject': (e) => ({
    title: 'Отказ на входе: ограничитель частоты',
    body: [
      `Ограничитель пропускает ${num(e.payload.limit)} запроса за тик и позволяет взять взаймы до ${num(e.payload.burst)} — этот запас уже исчерпан.`,
      'Отказ стоит почти ничего и приходит сразу. Это лучше, чем принять запрос, подержать его в очереди и отдать таймаут: ёмкость сервиса тратится на тех, кому он успеет ответить.',
    ],
    go: 'golang.org/x/time/rate: if !limiter.Allow() { http.Error(w, "too many requests", 429) }',
    terms: ['rate-limiting', 'load-shedding'],
  }),

  'req.done': (e) => ({
    title: `Ответ на #${num(e.payload.id)} за ${ticks(num(e.payload.lat))}`,
    body: [
      e.payload.needsDep
        ? `Полный путь: очередь, работа A, вызов B${num(e.payload.attempt) > 0 ? ` и ${num(e.payload.attempt)} повтор` : ''}. Задержка такого запроса — сумма своей работы и чужой.`
        : 'Запрос обслужен полностью внутри A. Его задержка зависит только от того, свободны ли воркеры.',
    ],
    terms: ['tail-latency'],
  }),

  'dep.call': (e) => ({
    title: `#${num(e.payload.id)}: вызов B${num(e.payload.attempt) > 1 ? `, попытка ${num(e.payload.attempt)}` : ''}`,
    body: [
      'Воркер A занят и будет занят всё время вызова. Это главный механизм каскада: ёмкость сервиса измеряется не его работой, а временем удержания воркера.',
      num(e.payload.timeout) > 0
        ? `Таймаут вызова — ${ticks(num(e.payload.timeout))}: дольше воркер ждать не будет.`
        : 'Таймаута нет: воркер будет ждать столько, сколько понадобится, — хоть вечно.',
    ],
    go: 'ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond); resp, err := client.Do(req.WithContext(ctx))',
    terms: ['cascading-failure', 'request-deadline'],
  }),

  'dep.ok': (e) => ({
    title: `#${num(e.payload.id)}: B ответил за ${ticks(num(e.payload.waited))}`,
    body: ['Воркер A свободен, ответ уходит клиенту. Пока зависимость отвечает быстро, всё это устройство незаметно.'],
    terms: ['cascading-failure'],
  }),

  'dep.timeout': (e) => ({
    title: `#${num(e.payload.id)}: B не ответил за ${ticks(num(e.payload.timeout))}`,
    body: [
      `B сейчас ${DEP[str(e.payload.dep)] ?? ''}. Таймаут вернул воркеру свободу — без него он ждал бы дальше.`,
      num(e.payload.retriesLeft) > 0
        ? 'Сейчас будет повтор. Но если зависимость лежит, повтор — это ещё один таймаут: лишняя занятая ёмкость и лишняя нагрузка на неё.'
        : 'Повторов не осталось: клиент получит ошибку или заглушку.',
    ],
    go: 'errors.Is(err, context.DeadlineExceeded) — и обязательно отмена вызова, а не просто ответ клиенту',
    terms: ['request-deadline', 'circuit-breaker'],
  }),

  'dep.error': (e) => ({
    title: `#${num(e.payload.id)}: B ответил ошибкой`,
    body: [
      `Ответ пришёл быстро — за ${ticks(num(e.payload.waited))}, — но это ошибка. Быстрые ошибки опаснее медленных: они дёшевы, и на них легко построить бесконечный цикл повторов.`,
      e.payload.breaker === 'closed' ? 'Предохранитель считает неудачи подряд.' : `Предохранитель сейчас: ${str(e.payload.breaker)}.`,
    ],
    terms: ['circuit-breaker'],
  }),

  'dep.retry': (e) => ({
    title: `#${num(e.payload.id)}: повтор вызова B (попытка ${num(e.payload.attempt)})`,
    body: [
      `Прошлая попытка кончилась ${e.payload.reason === 'timeout' ? 'таймаутом' : 'ошибкой'}. Повтор оправдан, когда сбой случайный.`,
      'Когда зависимость лежит целиком, повторы только умножают нагрузку на неё и занимают воркеры вызывающего. Поэтому повторы ограничивают бюджетом и отключают при разомкнутом предохранителе.',
    ],
    terms: ['retry-backoff', 'circuit-breaker'],
  }),

  'req.fallback': (e) => ({
    title: `#${num(e.payload.id)}: ответ-заглушка вместо ошибки`,
    body: [
      `B ${DEP[str(e.payload.dep)] ?? ''}, предохранитель — ${str(e.payload.breaker)}. Вместо ошибки пользователь получает урезанный ответ: страницу без блока рекомендаций, список без персонализации, данные из кэша.`,
      'Так сервис деградирует, а не падает. Главное — чтобы заглушка была честной: пустой блок лучше, чем неверные данные, выданные за настоящие.',
    ],
    go: 'if errors.Is(err, breaker.ErrOpen) { return cachedOrEmpty(ctx) }',
    terms: ['graceful-degradation', 'circuit-breaker'],
  }),

  'req.fail': (e) => ({
    title: `#${num(e.payload.id)}: ошибка после ${num(e.payload.attempts)} ${plural(num(e.payload.attempts), 'попытки', 'попыток', 'попыток')}`,
    body: [
      `Прождал ${ticks(num(e.payload.lat))}. B ${DEP[str(e.payload.dep)] ?? ''}, предохранитель — ${str(e.payload.breaker)}.`,
      'Ошибка — честный исход, если ответить нечем. Плохо, когда она приходит не сразу, а после долгого ожидания: тогда и пользователь ждал зря, и ёмкость потрачена.',
    ],
    terms: ['graceful-degradation'],
  }),

  'breaker.open': (e) => ({
    title: e.payload.probe ? `Предохранитель снова разомкнут: проба не прошла` : `Предохранитель разомкнут: ${num(e.payload.fails)} неудач подряд`,
    body: [
      `Следующие ${ticks(num(e.payload.openFor))} сервис в B не ходит вовсе: запросы, которым нужен B, сразу получают ошибку или заглушку.`,
      'Смысл двойной. Вызывающий не тратит воркеры на заведомо безнадёжные вызовы, а падающая зависимость получает передышку вместо шквала повторов — иногда только это и позволяет ей встать.',
    ],
    go: 'sony/gobreaker, failsafe-go; ReadyToTrip по доле ошибок, а не только по числу подряд',
    terms: ['circuit-breaker', 'metastable-failure'],
  }),

  'breaker.half': (e) => ({
    title: `Предохранитель приоткрылся: пробный вызов`,
    body: [
      `Прошло ${ticks(num(e.payload.after))}. Один запрос пропускается в B, чтобы проверить, ожила ли она. Остальные пока продолжают получать быстрый отказ.`,
      'Проба должна быть одна: если пустить всех разом, едва вставшая зависимость тут же ляжет снова.',
    ],
    terms: ['circuit-breaker'],
  }),

  'breaker.close': () => ({
    title: 'Предохранитель замкнут: зависимость ожила',
    body: [
      'Пробный вызов прошёл успешно, и сервис снова ходит в B как обычно.',
      'Возвращаться стоит осторожно: некоторые реализации после замыкания какое-то время пускают лишь часть трафика, чтобы не обрушить зависимость повторно.',
    ],
    terms: ['circuit-breaker'],
  }),

  'pool.saturated': (e) => ({
    title: `Все воркеры A заняты, ${num(e.payload.waiting)} из них просто ждут B`,
    body: [
      `Сервис не работает — он ждёт. В очереди ${reqs(num(e.payload.queue))}, и среди них есть те, кому B вообще не нужен: они встали за чужим ожиданием.`,
      num(e.payload.timeout) > 0
        ? `Таймаут ${ticks(num(e.payload.timeout))} рано или поздно вернёт воркеры, но пока он не истёк, ёмкость занята.`
        : 'Таймаута нет: воркеры не освободятся, пока B не ответит. Так авария одного сервиса становится аварией другого.',
    ],
    go: 'признак в pprof: сотни горутин в ожидании net/http; лечится таймаутом, переборкой и предохранителем',
    terms: ['cascading-failure', 'bulkhead'],
  }),

  'bulkhead.block': (e) => ({
    title: `#${num(e.payload.id)}: переборка не пустила к B`,
    body: [
      `Ждать B одновременно могут не больше ${num(e.payload.limit)} воркеров из ${num(e.payload.workers)}. Лимит занят, поэтому запрос не встаёт в очередь к зависимости, а сразу получает заглушку или ошибку.`,
      'Так зависимость получает только выделенную ей долю ёмкости. Остальные воркеры остаются для запросов, которым B не нужен.',
    ],
    go: 'семафор на канале вокруг вызова зависимости — по одному на каждую зависимость',
    terms: ['bulkhead', 'cascading-failure'],
  }),

  'queue.grow': (e) => ({
    title: `Очередь на входе: ${reqs(num(e.payload.len))}`,
    body: [
      `${num(e.payload.waiting)} воркеров ждут B, который ${DEP[str(e.payload.dep)] ?? ''}. Пока они заняты, очередь разбирается только оставшимися.`,
      'Эта очередь — уже знакомое колено загрузки: чем меньше свободных воркеров, тем быстрее она растёт.',
    ],
    terms: ['utilization', 'cascading-failure'],
  }),

  'dep.fault': (e) => ({
    title: `Зависимость сломалась: ${DEP[str(e.payload.kind)] ?? str(e.payload.kind)}${e.payload.kind === 'slow' ? ` (×${num(e.payload.factor)})` : ''}`,
    body: [
      e.payload.kind === 'hang'
        ? 'Самый неприятный вид отказа: соединение принимается, ответа нет. Узнать об этом можно только по таймауту.'
        : e.payload.kind === 'errors'
          ? 'Зависимость отвечает быстро, но ошибками. Это дешёвые ответы — и потому опасные: они легко порождают шквал повторов.'
          : 'Зависимость просто замедлилась. Ни ошибок, ни отказов — но каждый вызов теперь держит воркер в разы дольше.',
      num(e.payload.timeout) > 0 ? `Таймаут вызова — ${ticks(num(e.payload.timeout))}.` : 'Таймаута у вызова нет — посмотрим, чем это кончится.',
    ],
    terms: ['cascading-failure'],
  }),

  'dep.recover': (e) => ({
    title: 'Зависимость ожила',
    body: [
      `Она снова отвечает нормально. Предохранитель сейчас ${str(e.payload.breaker)}.`,
      e.payload.breaker === 'open'
        ? 'Сервис пока об этом не знает: он узнает на следующей пробе. Это плата за защиту — небольшая задержка возвращения.'
        : 'Сервис продолжит ходить в неё как обычно.',
    ],
    terms: ['circuit-breaker'],
  }),

  'load.change': (e) => ({
    title: num(e.payload.to) > num(e.payload.from) ? `Нагрузка выросла в ${num(e.payload.to)} раза` : 'Нагрузка вернулась к обычной',
    body: [`Теперь приходит ${num(e.payload.rate).toFixed(1)} запроса за тик на ${num(e.payload.workers)} воркеров.`],
    terms: ['utilization', 'rate-limiting'],
  }),
}

export function explain(e: ResilEvent): Explanation {
  return EXPLAIN[e.type](e)
}
