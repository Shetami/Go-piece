import type { LbEvent, LbEventType } from '../engine/types.ts'

/**
 * Разбор событий стенда. Объяснения привязаны к ТИПУ события: какой бы
 * сценарий ни крутился, клик по «health.eject» откроет один и тот же разбор,
 * подставив в него конкретные реплики и числа из payload.
 */

export interface Explanation {
  title: string
  body: string[]
  /** Как это выглядит в сервисе на Go или в его окружении. */
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
const pct = (v: unknown) => `${Math.round(num(v) * 100)}%`

const ALGO_TEXT: Record<string, string> = {
  'round-robin': 'по кругу',
  random: 'случайно',
  'least-conn': 'наименьшее число соединений',
  p2c: 'два случайных',
}
const algo = (v: unknown) => ALGO_TEXT[str(v)] ?? str(v)

const ALGO_WHY: Record<string, string> = {
  'round-robin': 'По кругу балансировщик ничего не смотрит: реплики получают запросы строго по очереди, хоть свободны они, хоть завалены.',
  random: 'Случайный выбор ничего не смотрит: иногда несколько запросов подряд попадают на одну реплику, пока соседняя простаивает.',
  'least-conn': 'Балансировщик выбрал реплику, от которой ждёт меньше всего ответов. Это единственное, что он знает о её загрузке, — но этого обычно хватает.',
  p2c: 'Балансировщик взял две случайные реплики и отправил на ту, от которой ждёт меньше ответов. Почти так же хорошо, как сравнить все, и не нужно знать о каждой.',
}

const STATE_TEXT: Record<string, string> = {
  up: 'работает',
  slow: 'тормозит',
  errors: 'отвечает ошибками',
  crash: 'упала',
  hang: 'зависла',
  booting: 'запускается',
}

type Explainer = (e: LbEvent) => Explanation

const EXPLAIN: Record<LbEventType, Explainer> = {
  'req.route': (e) => ({
    title: `Запрос → ${str(e.payload.replica)} (${algo(e.payload.algo)})${num(e.payload.attempt) > 0 ? ', повтор' : ''}`,
    body: [
      ALGO_WHY[str(e.payload.algo)] ?? '',
      `Незавершённых запросов балансировщика к ${str(e.payload.replica)} теперь: ${num(e.payload.outstanding)}.`,
    ],
    terms: ['load-balancer'],
  }),

  'req.done': (e) => ({
    title: `${str(e.payload.replica)} ответила за ${ticks(num(e.payload.lat))}${num(e.payload.attempt) > 0 ? ' — со второй попытки' : ''}`,
    body: [
      num(e.payload.attempt) > 0
        ? `Пользователь ждал ${ticks(num(e.payload.lat))}: первая попытка ушла на сломанную реплику, и только повтор дошёл сюда.`
        : `${num(e.payload.wait)} ${plural(num(e.payload.wait), 'тик', 'тика', 'тиков')} в очереди реплики и ${num(e.payload.service)} работы.`,
    ],
    terms: ['tail-latency'],
  }),

  'req.error': (e) => ({
    title: `${str(e.payload.replica)}: ${e.payload.reason === 'reset' ? 'соединение оборвалось' : e.payload.reason === 'crash' ? 'соединение отклонено' : 'ответ 500'}${num(e.payload.inRow) > 1 ? ` — ${num(e.payload.inRow)}-я ошибка подряд` : ''}`,
    body: [
      e.payload.reason === 'errors'
        ? 'Реплика ответила ошибкой сразу, даже не попытавшись поработать: скорее всего, у неё сломана зависимость — пул соединений с базой, конфиг, сертификат. Такой ответ приходит быстрее любого настоящего.'
        : e.payload.reason === 'crash'
          ? 'Процесс не принимает соединения. Балансировщик узнаёт об этом сразу, но исключить реплику сможет только проверка здоровья или пассивное исключение.'
          : 'Реплика перезапустилась или сломалась, пока держала запрос. Клиент получил обрыв соединения.',
      e.payload.algo === 'least-conn' && e.payload.reason === 'errors'
        ? 'Для «наименьшего числа соединений» это ловушка: быстро отвечающая реплика всегда выглядит самой свободной, и балансировщик шлёт на неё всё больше. Это чёрная дыра.'
        : num(e.payload.retriesLeft) > 0
          ? 'Клиент повторит запрос на другой реплике.'
          : 'Повторов нет — пользователь получит ошибку.',
    ],
    go: 'в Envoy и gRPC это лечит outlier detection: consecutive_5xx, base_ejection_time',
    terms: e.payload.algo === 'least-conn' ? ['least-connections', 'outlier-detection'] : ['outlier-detection'],
  }),

  'req.timeout': (e) => ({
    title: `Таймаут на ${str(e.payload.replica)}: клиент ждал ${ticks(num(e.payload.timeout))}`,
    body: [
      e.payload.state === 'hang'
        ? 'Реплика зависла: соединение приняла, а отвечать не будет никогда. Узнать об этом можно только одним способом — дождаться таймаута. Поэтому зависшая реплика хуже упавшей: упавшая отвечает ошибкой сразу.'
        : `Запрос простоял ${e.payload.where === 'queue' ? 'в очереди' : 'в работе'} дольше, чем клиент готов ждать. Реплика ${STATE_TEXT[str(e.payload.state)] ?? ''}, но перегружена.`,
      num(e.payload.retriesLeft) > 0 ? 'Клиент повторит запрос на другой реплике — ответ придёт, но пользователь уже потерял время таймаута.' : 'Повторов нет — пользователь получит ошибку.',
    ],
    terms: ['request-deadline', 'health-check'],
  }),

  'req.retry': (e) => ({
    title: `Повтор в обход ${str(e.payload.avoid)}: ${num(e.payload.attempt)}-я попытка`,
    body: [
      `Пользователь ждёт уже ${ticks(num(e.payload.waited))}. Клиент (или прокси) повторяет запрос и просит балансировщик не выбирать реплику, где провалилась прошлая попытка.`,
      'Повтор на другой реплике — дешёвый способ скрыть отказ одной реплики. Но он скрывает и саму поломку: без метрик ошибок по репликам можно долго не замечать, что одна из них больна.',
    ],
    go: 'Envoy: retry_policy с retry_host_predicate previous_hosts; в Go — свой цикл с исключением адреса',
    terms: ['retry-backoff'],
  }),

  'req.fail': (e) => ({
    title: `Пользователь получил ошибку после ${num(e.payload.attempts)} ${plural(num(e.payload.attempts), 'попытки', 'попыток', 'попыток')}`,
    body: [
      `Прождал ${ticks(num(e.payload.waited))}; ${e.payload.reason === 'timeout' ? 'последняя попытка кончилась таймаутом' : e.payload.reason === 'nobackend' ? 'балансировщику некуда было отправить запрос' : 'последняя попытка кончилась ошибкой'}.`,
    ],
    terms: ['load-balancer'],
  }),

  'req.nobackend': (e) => ({
    title: 'Некуда отправить: ни одной реплики в ротации',
    body: [
      `Реплик всего ${num(e.payload.replicas)}, но ни одна не принимает запросы: все исключены или запускаются. Балансировщик отвечает 503 сам.`,
      'Поэтому пассивное исключение ограничивают: Envoy по умолчанию не исключает больше 10% реплик сразу. Лучше слать трафик на подозрительные реплики, чем никуда.',
    ],
    terms: ['outlier-detection'],
  }),

  'load.change': (e) => ({
    title: num(e.payload.to) > num(e.payload.from) ? `Нагрузка выросла в ${num(e.payload.to)} раза` : 'Нагрузка вернулась к обычной',
    body: [
      `Теперь приходит ${num(e.payload.rate).toFixed(2)} запроса за тик, а работающие реплики вместе успевают ${num(e.payload.capacity).toFixed(2)}. ${num(e.payload.rate) > num(e.payload.capacity) ? 'Этого мало: очереди будут расти, пока не добавятся реплики.' : 'Этого хватает.'}`,
    ],
    terms: ['utilization', 'autoscaling'],
  }),

  'replica.fault': (e) => {
    const kind = str(e.payload.kind)
    const text: Record<string, string> = {
      slow: `${str(e.payload.replica)} стала в ${num(e.payload.factor)} раза медленнее: шумный сосед на хосте, деградировавший диск, прогрев кэша после деплоя. Процесс жив, на /healthz отвечает — снаружи всё выглядит нормально.`,
      errors: `${str(e.payload.replica)} сломалась «тихо»: процесс жив, /healthz отвечает 200, но каждый настоящий запрос сразу получает 500 — например, сломан пул соединений с базой, а проверка в базу не ходит.`,
      crash: `${str(e.payload.replica)} упала: процесс не принимает соединения. Это самая простая поломка — её видно и запросам, и проверке здоровья.`,
      hang: `${str(e.payload.replica)} зависла: соединения принимает, запросы берёт, но не отвечает — дедлок, бесконечная пауза, исчерпанный пул без таймаута. Это самая неприятная поломка: узнать о ней можно только по таймауту.`,
    }
    return {
      title: `${str(e.payload.replica)} ${STATE_TEXT[kind] ?? kind}${kind === 'slow' ? ` ×${num(e.payload.factor)}` : ''}`,
      body: [
        text[kind] ?? '',
        kind === 'slow' && e.payload.algo === 'round-robin'
          ? 'Балансировка по кругу продолжит слать ей четверть запросов. Если это больше, чем она теперь успевает, её очередь будет расти без предела.'
          : kind === 'errors' && e.payload.algo === 'least-conn'
            ? '«Наименьшее число соединений» сейчас начнёт отправлять на неё всё больше: ошибка приходит мгновенно, и незавершённых запросов у неё всегда ноль.'
            : 'Посмотрим, как отреагирует балансировщик.',
      ],
      terms: kind === 'slow' ? ['health-check', 'least-connections'] : kind === 'errors' ? ['health-check', 'outlier-detection'] : ['health-check'],
    }
  },

  'replica.recover': (e) => ({
    title: `${str(e.payload.replica)} снова ${e.payload.was === 'hang' ? 'работает после перезапуска' : 'в порядке'}`,
    body: [
      e.payload.inRotation
        ? 'Балансировщик и не исключал её, так что запросы пойдут сразу.'
        : 'Но балансировщик пока считает её мёртвой: вернуть в ротацию её должны несколько успешных проверок здоровья подряд.',
    ],
    terms: ['health-check'],
  }),

  'health.fail': (e) => ({
    title: `${str(e.payload.replica)} не прошла проверку здоровья (${num(e.payload.failed)} из ${num(e.payload.need)})`,
    body: [
      `Реплика ${STATE_TEXT[str(e.payload.state)] ?? ''}: на /healthz нет ответа. Но одна проваленная проверка ещё ничего не решает — сеть моргнула, сборщик мусора остановил мир. Реплику исключат после ${num(e.payload.need)} провалов подряд.`,
      'Всё это время она остаётся в ротации и получает свою долю запросов.',
    ],
    terms: ['health-check'],
  }),

  'health.eject': (e) => ({
    title: `${str(e.payload.replica)} исключена из ротации проверкой здоровья`,
    body: [
      `${num(e.payload.checks)} ${plural(num(e.payload.checks), 'проверка', 'проверки', 'проверок')} подряд с интервалом ${ticks(num(e.payload.every))} провалились. От поломки до исключения прошло до ${ticks(num(e.payload.checks) * num(e.payload.every))} — всё это время часть запросов уходила в никуда.`,
      'Остальные реплики получили её долю нагрузки. Если запаса не было, они тоже начнут тормозить.',
    ],
    go: 'Kubernetes: readinessProbe с periodSeconds и failureThreshold убирает под из Endpoints',
    terms: ['health-check', 'load-balancer'],
  }),

  'health.restore': (e) => ({
    title: `${str(e.payload.replica)} вернулась в ротацию`,
    body: [`${num(e.payload.checks)} проверки подряд прошли. Балансировщик снова шлёт на неё запросы.`],
    terms: ['health-check'],
  }),

  'outlier.eject': (e) => ({
    title: `${str(e.payload.replica)} исключена по ошибкам на ${ticks(num(e.payload.forTicks))}`,
    body: [
      `${num(e.payload.errors)} ошибки подряд на настоящих запросах. Балансировщику не нужно ждать проверки здоровья: он смотрит на ответы, которые и так через него идут.`,
      'Это пассивная проверка. Она ловит то, что /healthz пропускает, — реплику, которая жива, но отвечает ошибками. Через время реплику вернут и посмотрят снова.',
    ],
    go: 'Envoy outlier_detection: consecutive_5xx: 3, base_ejection_time, max_ejection_percent',
    terms: ['outlier-detection'],
  }),

  'outlier.return': (e) => ({
    title: `${str(e.payload.replica)} возвращена после исключения по ошибкам`,
    body: [
      e.payload.state === 'errors'
        ? 'Реплика всё ещё сломана: первые же запросы получат ошибки, и её исключат снова. Envoy в таком случае каждый раз увеличивает время исключения.'
        : 'Время исключения прошло — реплика снова получает запросы.',
    ],
    terms: ['outlier-detection'],
  }),

  'queue.grow': (e) => ({
    title: `Очередь на ${str(e.payload.replica)}: ${reqs(num(e.payload.len))}`,
    body: [
      e.payload.state === 'slow' && e.payload.algo === 'round-robin'
        ? `Реплика стала в ${num(e.payload.factor)} раза медленнее, а по кругу получает столько же, сколько остальные. Её загрузка выше 100%, очередь будет расти, пока балансировщик не начнёт учитывать её состояние.`
        : e.payload.algo === 'random'
          ? 'Случайный выбор неравномерен: на эту реплику попало несколько запросов подряд, а соседние в это время простаивали.'
          : 'Реплике не хватает ёмкости: запросов приходит больше, чем она успевает.',
    ],
    terms: ['least-connections', 'power-of-two-choices'],
  }),

  'scale.decide': (e) => ({
    title: `Автомасштабирование: +${num(e.payload.adding)} ${plural(num(e.payload.adding), 'реплика', 'реплики', 'реплик')}`,
    body: [
      `Средняя загрузка за последние 10 тиков — ${pct(e.payload.util)} при цели ${pct(e.payload.target)}. Чтобы загрузка стала целевой, нужно ${num(e.payload.desired)} ${plural(num(e.payload.desired), 'реплика', 'реплики', 'реплик')} вместо ${num(e.payload.active)}.`,
      `Новые реплики будут готовы через ${ticks(num(e.payload.boot))}: планирование пода, скачивание образа, запуск, прогрев. Всё это время нагрузку держат старые.`,
    ],
    go: 'Kubernetes HPA: desired = ceil(current × currentUtilization / targetUtilization)',
    terms: ['autoscaling', 'utilization'],
  }),

  'scale.ready': (e) => ({
    title: `${str(e.payload.replica)} запустилась и вошла в ротацию`,
    body: [
      `Запуск занял ${ticks(num(e.payload.boot))}. Теперь работающих реплик ${num(e.payload.total)}.`,
      'Чем дольше запускается реплика, тем больше запаса нужно держать заранее: автомасштабирование помогает от роста, который длится дольше времени запуска, но не от внезапного всплеска.',
    ],
    terms: ['autoscaling'],
  }),
}

export function explain(e: LbEvent): Explanation {
  return EXPLAIN[e.type](e)
}
