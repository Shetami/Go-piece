import type { KafkaEvent, KafkaEventType } from '../engine/types.ts'

/**
 * Разбор событий стенда.
 *
 * Объяснения привязаны к ТИПУ события, а не к кадру: какой бы сценарий ни
 * крутился, клик по «isr.shrink» откроет один и тот же разбор, подставив в него
 * конкретные брокеры, партиции и оффсеты из payload.
 */

export interface Explanation {
  title: string
  body: string[]
  /** Что делает настоящая Kafka — классы и настройки. */
  kafka?: string
  /** Где модель расходится с реальностью, если это важно для понимания. */
  model?: string
  terms: string[]
}

export const num = (v: unknown) => (typeof v === 'number' ? v : Number(v ?? 0))
export const str = (v: unknown) => (typeof v === 'string' ? v : String(v ?? ''))
const arr = (v: unknown): number[] => (Array.isArray(v) ? (v as number[]) : [])

export function plural(n: number, one: string, few: string, many: string): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few
  return many
}
export const ticks = (n: number) => `${n} ${plural(n, 'тик', 'тика', 'тиков')}`
export const msgs = (n: number) => `${n} ${plural(n, 'сообщение', 'сообщения', 'сообщений')}`
/** «m3, m4, m5» — не длиннее шести. */
export const recList = (v: unknown) => {
  const ids = arr(v)
  const head = ids.slice(0, 6).map((id) => `m${id}`).join(', ')
  return ids.length > 6 ? `${head} и ещё ${ids.length - 6}` : head
}
const brokers = (v: unknown) => arr(v).map((b) => `B${b}`).join(', ') || '—'
const acksLabel = (v: unknown) => (v === 'all' ? 'acks=all' : `acks=${str(v)}`)

type Explainer = (e: KafkaEvent) => Explanation

const EXPLAIN: Record<KafkaEventType, Explainer> = {
  'rec.send': (e) => ({
    title: `send(m${num(e.payload.rec)}) → партиция ${num(e.payload.partition)}`,
    body: [
      e.payload.key === null
        ? 'У сообщения нет ключа, поэтому партицию выбирает липкий партиционер: все сообщения без ключа идут в одну партицию, пока её пакет не закроется, а потом переключаются на следующую. Так пакеты получаются полнее, чем при раскладке по кругу.'
        : `Ключ «${str(e.payload.key)}» хешируется murmur2 и берётся по модулю числа партиций. Один и тот же ключ всегда попадает в одну и ту же партицию — только поэтому у Kafka есть порядок «по ключу».`,
      `send() не отправляет ничего по сети. Сообщение сериализуется и ложится в аккумулятор — открытый пакет этой партиции, где теперь ${msgs(num(e.payload.inBatch))}. Вызов возвращается сразу, результат придёт позже, в колбэк.`,
    ],
    kafka: 'KafkaProducer.send → partitioner → RecordAccumulator.append',
    terms: ['kafka-partitioner', 'record-accumulator'],
  }),

  'batch.seal': (e) => ({
    title:
      e.payload.reason === 'size'
        ? `Пакет #${num(e.payload.batch)} в p${num(e.payload.partition)} полон: ${msgs(num(e.payload.size))}`
        : `Пакет #${num(e.payload.batch)} в p${num(e.payload.partition)} закрыт по linger: ${msgs(num(e.payload.size))} за ${ticks(num(e.payload.waited))}`,
    body: [
      e.payload.reason === 'size'
        ? `Пакет набрал batch.size (в модели — ${num(e.payload.batchSize)} записи) и больше не принимает сообщения. Он готов к отправке немедленно, а новые сообщения этой партиции пойдут в новый пакет.`
        : `Пакет прождал linger.ms (${ticks(num(e.payload.linger))}) и Sender его забрал. До этой секунды пакет принимал попутчиков — поэтому под нагрузкой пакеты растут даже при linger.ms=0: пока предыдущие запросы в пути, новые сообщения копятся.`,
      'Пакет — единица всего дальнейшего пути: он уходит одним запросом, пишется в лог одним куском, реплицируется и подтверждается целиком. Чем крупнее пакет, тем меньше запросов и тем лучше сжатие.',
    ],
    kafka: 'RecordAccumulator.ready → RecordAccumulator.drain → ProducerBatch.close',
    model: 'batch.size в Kafka считается в байтах (по умолчанию 16 КБ), в модели — в записях.',
    terms: ['record-accumulator'],
  }),

  'produce.request': (e) => ({
    title: `Запрос записи: пакет #${num(e.payload.batch)} → лидер p${num(e.payload.partition)} на B${num(e.payload.broker)}${num(e.payload.attempt) > 1 ? `, попытка ${num(e.payload.attempt)}` : ''}`,
    body: [
      `Продюсер знает из метаданных, что лидер партиции ${num(e.payload.partition)} — брокер B${num(e.payload.broker)}, и отправляет ProduceRequest именно ему. Писать можно только в лидера: фолловеры запросы записи не принимают.`,
      `На этом соединении теперь ${num(e.payload.inFlight)} из ${num(e.payload.maxInFlight)} запросов в пути (max.in.flight.requests.per.connection). Продюсер не ждёт ответа на один запрос, чтобы послать следующий, — иначе пропускная способность упиралась бы в сетевую задержку.`,
      e.payload.acks === 0
        ? 'acks=0: ответа не будет вовсе. Продюсер считает сообщения доставленными в момент отправки.'
        : `${acksLabel(e.payload.acks)}: брокер ответит, ${e.payload.acks === 1 ? 'как только запишет пакет в свой лог' : 'когда пакет окажется на всех репликах из ISR'}.`,
    ],
    kafka: 'Sender.sendProduceRequests → NetworkClient.send (ProduceRequest)',
    terms: ['kafka-broker', 'acks'],
  }),

  'produce.append': (e) => ({
    title: `B${num(e.payload.broker)} записал ${recList(e.payload.recs)} в p${num(e.payload.partition)}: оффсеты ${num(e.payload.from)}–${num(e.payload.to)}`,
    body: [
      'Лидер дописал пакет в конец лога партиции. Оффсет — это просто номер записи в этом логе: он назначается здесь и сейчас, продюсер его заранее не знает.',
      'Запись идёт в страничный кэш ОС, а не на диск: Kafka не вызывает fsync на каждое сообщение. Надёжность обеспечивает не диск одного брокера, а копии на нескольких.',
      e.payload.acks === 1
        ? 'При acks=1 лидер ответит продюсеру прямо сейчас — хотя кроме него записи нет больше нигде.'
        : e.payload.acks === 'all'
          ? `При acks=all ответ откладывается: запрос ждёт в «чистилище», пока фолловеры из ISR (${brokers(e.payload.isr)}) не скопируют пакет.`
          : 'При acks=0 отвечать некому: продюсер давно пошёл дальше.',
    ],
    kafka: 'ReplicaManager.appendRecords → Partition.appendRecordsToLeader → UnifiedLog.appendAsLeader',
    terms: ['kafka-offset', 'kafka-segment', 'page-cache'],
  }),

  'produce.duplicate': (e) => ({
    title: `Дубль: ${recList(e.payload.recs)} записаны в p${num(e.payload.partition)} второй раз`,
    body: [
      `Это повтор пакета, который уже лежит в логе с оффсета ${num(e.payload.first)}. Брокер этого не знает: без идемпотентности у пакета нет ничего, по чему его можно узнать, и он честно пишет его ещё раз — с новыми оффсетами ${num(e.payload.from)}–${num(e.payload.to)}.`,
      'Продюсер повторил запрос, потому что не получил ответа. А не получить ответ можно и тогда, когда запись удалась, — сеть потеряла подтверждение. Со стороны продюсера эти два случая неразличимы.',
      'Потребитель прочитает оба экземпляра и обработает сообщение дважды. Лечится это с двух сторон: идемпотентный продюсер не даёт дублю появиться, идемпотентная обработка делает его безвредным.',
    ],
    kafka: 'enable.idempotence=false — брокер не хранит состояние продюсера',
    terms: ['idempotent-producer', 'delivery-semantics'],
  }),

  'produce.dedup': (e) => ({
    title: `B${num(e.payload.broker)} узнал повтор пакета #${num(e.payload.batch)} и не стал писать его снова`,
    body: [
      'У идемпотентного продюсера есть producer id, а у каждого пакета — номер последовательности в партиции. Брокер помнит последние номера каждого продюсера и видит: этот пакет уже записан.',
      `Вместо второй копии брокер отвечает «успешно» с исходным оффсетом ${num(e.payload.baseOffset)}. Продюсер получает подтверждение, в логе ровно одна копия.`,
      'С Kafka 3.0 идемпотентность включена по умолчанию. Но она защищает только от повторов внутри одной сессии продюсера: перезапущенный процесс получит новый producer id, и защиту от «отправил, упал, отправил снова» даёт уже транзакционный продюсер.',
    ],
    kafka: 'ProducerStateManager: producerId + epoch + sequence; повтор одного из 5 последних пакетов отвечается исходным оффсетом',
    model: 'Брокер в модели узнаёт повтор по содержимому лога; в Kafka — по номеру последовательности, результат тот же.',
    terms: ['idempotent-producer', 'kafka-transactions'],
  }),

  'produce.ack': (e) => ({
    title:
      e.payload.acks === 0
        ? `acks=0: ${recList(e.payload.recs)} «доставлены» без единого подтверждения`
        : `Продюсер получил ответ: ${recList(e.payload.recs)} подтверждены за ${ticks(num(e.payload.latency))}`,
    body: [
      e.payload.acks === 0
        ? 'Никто ничего не подтверждал: продюсер отметил пакет доставленным, как только выпустил его в сеть. Если брокер его не получит, об этом не узнает никто.'
        : `Колбэк send() наконец вызывается. ${acksLabel(e.payload.acks)} ${e.payload.acks === 'all' ? 'означает, что пакет есть на всех репликах из ISR и закоммичен: он переживёт падение любого из них.' : 'означает только то, что пакет есть у лидера. Реплик ещё может не быть.'}`,
      num(e.payload.attempt) > 1
        ? `Подтверждение пришло с ${num(e.payload.attempt)}-й попытки. Это нормально: ретраи продюсера — обычная часть жизни, их не видно приложению, пока не истечёт delivery.timeout.ms.`
        : 'Задержка от send() до подтверждения складывается из ожидания в пакете (linger.ms), пути по сети, записи и — для acks=all — репликации. Сравните её со стендами, где acks другой.',
    ],
    kafka: 'Sender.handleProduceResponse → ProducerBatch.complete → Callback.onCompletion',
    terms: ['acks'],
  }),

  'produce.error': (e) => {
    const err = str(e.payload.error)
    const why: Record<string, string> = {
      NOT_LEADER_OR_FOLLOWER: `B${num(e.payload.broker)} больше не лидер партиции ${num(e.payload.partition)}. Продюсер отправил запрос по устаревшим метаданным — теперь он их обновит и пошлёт пакет новому лидеру${e.payload.leader === null ? ', как только тот появится' : ` B${num(e.payload.leader)}`}.`,
      NOT_ENOUGH_REPLICAS: `В ISR партиции ${num(e.payload.partition)} сейчас ${arr(e.payload.isr).length} ${plural(arr(e.payload.isr).length, 'реплика', 'реплики', 'реплик')} (${brokers(e.payload.isr)}), а min.insync.replicas = ${num(e.payload.minIsr)}. При acks=all лидер отказывается принимать запись заранее, до записи в лог: гарантию «переживёт падение брокера» он дать не может.`,
      NOT_ENOUGH_REPLICAS_AFTER_APPEND: `Пакет уже записан в лог и даже закоммичен, но пока продюсер ждал, ISR сжался ниже min.insync.replicas (${num(e.payload.minIsr)}). Лидер честно сообщает: обещанной надёжности нет. Запись при этом осталась в логе — потребители её прочитают.`,
      REQUEST_TIMEOUT: `Ответа от B${num(e.payload.broker)} нет дольше request.timeout.ms. Брокер мог упасть, запрос мог потеряться, мог потеряться ответ — продюсер не знает, что из этого случилось, и повторяет.`,
    }
    return {
      title: `Ошибка ${err} для пакета #${num(e.payload.batch)} — продюсер повторит`,
      body: [
        why[err] ?? `Брокер ответил ошибкой ${err}.`,
        `Пакет возвращается в голову очереди и уйдёт снова через пару тиков. Пакету уже ${ticks(num(e.payload.age))}; если за delivery.timeout.ms (${ticks(num(e.payload.deliveryTimeout))}) доставить его не удастся, ошибку получит приложение.`,
        err === 'REQUEST_TIMEOUT' || err === 'NOT_ENOUGH_REPLICAS_AFTER_APPEND'
          ? 'Опасный момент: запись, возможно, уже в логе. Повтор без идемпотентности создаст дубль.'
          : 'Такую ошибку Kafka считает временной (retriable): приложение о ней не узнает, если повтор удастся.',
      ],
      kafka: 'Sender.completeBatch → canRetry → RecordAccumulator.reenqueue',
      terms: err.startsWith('NOT_ENOUGH') ? ['min-insync-replicas', 'isr'] : ['leader-election', 'idempotent-producer'],
    }
  },

  'produce.failed': (e) => ({
    title: `delivery.timeout истёк: ${recList(e.payload.recs)} не доставлены`,
    body: [
      `Пакет прожил ${ticks(num(e.payload.age))} и сделал ${num(e.payload.attempts)} ${plural(num(e.payload.attempts), 'попытку', 'попытки', 'попыток')}. Продюсер сдаётся: колбэк send() получает ошибку, и что делать дальше — решает приложение.`,
      'Внимание: «не доставлено» не значит «не записано». Если одна из попыток дошла до лога, а ответ — нет, сообщение уже лежит в партиции и будет прочитано. Kafka не откатывает записи, которые продюсер считает неудачными.',
      'Типичная ошибка — логировать и забывать. Если сообщение важно, неудачу надо обрабатывать: складывать в отдельное хранилище, повторять позже, останавливать приём запросов.',
    ],
    kafka: 'delivery.timeout.ms (по умолчанию 120 с) → TimeoutException в колбэке',
    terms: ['delivery-semantics', 'min-insync-replicas'],
  }),

  'resp.lost': (e) => ({
    title: `Сеть потеряла ответ на пакет #${num(e.payload.batch)}`,
    body: [
      `Брокер B${num(e.payload.broker)} ${e.payload.wasError ? `ответил ошибкой ${str(e.payload.wasError)}` : 'записал пакет и ответил «успешно»'}, но ответ до продюсера не дошёл. Для продюсера это выглядит так же, как если бы запрос не дошёл вовсе.`,
      `Продюсер будет ждать ещё ${ticks(Math.max(0, num(e.payload.timeoutIn)))} до request.timeout.ms, а потом повторит запрос. Что будет в логе после повтора, решает одна настройка — enable.idempotence.`,
    ],
    kafka: 'обрыв соединения или таймаут до ответа — NetworkException / TimeoutException',
    model: 'Потерю ответа модель задаёт сценарием. В реальности так проявляются обрывы соединений, перезапуски балансировщиков и паузы GC на брокере.',
    terms: ['idempotent-producer', 'delivery-semantics'],
  }),

  'repl.fetch': (e) => ({
    title: `B${num(e.payload.follower)} скопировал ${recList(e.payload.recs)} из p${num(e.payload.partition)} (лидер B${num(e.payload.leader)})`,
    body: [
      'Лидер ничего не рассылает сам: фолловеры постоянно шлют ему Fetch — тот же запрос, которым читают потребители, — с оффсетом, с которого им нужны данные. Этот оффсет одновременно сообщает лидеру, докуда у фолловера всё есть.',
      `Лог фолловера теперь заканчивается на ${num(e.payload.leo)}. Лидер узнает об этом только из следующего запроса фолловера — поэтому HW отстаёт от репликации на один круг.`,
      num(e.payload.slow) > 1
        ? `Этот фолловер реплицирует в ${num(e.payload.slow)} раза медленнее обычного — как реплика в далёкой зоне или брокер с забитым диском. При acks=all его медлительность платит каждая запись, пока он в ISR.`
        : 'Реплики хранят побайтно тот же лог, что и лидер: те же оффсеты, те же пакеты. Поэтому любой фолловер из ISR может стать лидером без пересчёта чего-либо.',
    ],
    kafka: 'ReplicaFetcherThread → FetchRequest(replicaId) → Partition.updateFollowerFetchState',
    terms: ['kafka-replication', 'high-watermark'],
  }),

  'hw.advance': (e) => ({
    title: `HW p${num(e.payload.partition)}: ${num(e.payload.from)} → ${num(e.payload.to)} — закоммичены ${recList(e.payload.recs)}`,
    body: [
      `High watermark — минимальный конец лога среди реплик ISR (${brokers(e.payload.isr)}). Всё, что ниже него, есть на каждой синхронной реплике: это и называется «закоммичено».`,
      'Именно здесь сообщение становится видно потребителям: читать разрешено только до HW. Иначе потребитель мог бы прочитать запись, которая потом исчезнет при смене лидера.',
      'И здесь же отвечают продюсеры с acks=all, чьи пакеты ждали в чистилище: HW перешёл их последнюю запись.',
    ],
    kafka: 'Partition.maybeIncrementLeaderHW → DelayedProduce.tryComplete',
    terms: ['high-watermark', 'isr'],
  }),

  'isr.shrink': (e) => ({
    title:
      e.payload.reason === 'dead'
        ? `B${num(e.payload.broker)} вычеркнут из ISR p${num(e.payload.partition)}: брокер мёртв`
        : `B${num(e.payload.broker)} вылетел из ISR p${num(e.payload.partition)}: отстаёт ${ticks(num(e.payload.lagged))}`,
    body: [
      e.payload.reason === 'dead'
        ? 'Контроллер заметил, что брокер пропал, и убрал его из ISR всех партиций, где он был фолловером. Ждать его подтверждений больше никто не будет.'
        : `Фолловер ни разу не догнал лидера дольше replica.lag.time.max.ms и отстаёт на ${num(e.payload.behind)} ${plural(num(e.payload.behind), 'запись', 'записи', 'записей')}. Лидер убирает его из ISR — и перестаёт ждать при acks=all.`,
      `ISR теперь: ${brokers(e.payload.isr)}. ${arr(e.payload.isr).length < num(e.payload.minIsr) ? `Это меньше min.insync.replicas (${num(e.payload.minIsr)}): запись с acks=all будет отклоняться, пока ISR не восстановится.` : `min.insync.replicas = ${num(e.payload.minIsr)} ещё выполняется, записи продолжаются — но страховка стала тоньше.`}`,
      'ISR — компромисс, придуманный Kafka: ждать всех — значит зависеть от самого медленного, ждать никого — значит терять данные. ISR ждёт только тех, кто успевает, и честно сужается, когда кто-то не успевает.',
    ],
    kafka: 'Partition.maybeShrinkIsr (раз в replica.lag.time.max.ms / 2) → AlterPartition к контроллеру',
    terms: ['isr', 'min-insync-replicas'],
  }),

  'isr.expand': (e) => ({
    title: `B${num(e.payload.broker)} догнал лидера и вернулся в ISR p${num(e.payload.partition)}`,
    body: [
      `Лог фолловера дошёл до конца лога лидера (оффсет ${num(e.payload.leo)}). Лидер возвращает его в ISR: теперь его подтверждение снова учитывается в HW и при acks=all.`,
      `ISR: ${brokers(e.payload.isr)}. Если запись стояла из-за min.insync.replicas, она возобновится.`,
    ],
    kafka: 'Partition.maybeExpandIsr → AlterPartition',
    terms: ['isr'],
  }),

  'broker.down': (e) => ({
    title: `Брокер B${num(e.payload.broker)} упал${arr(e.payload.leads).length > 0 ? ` — он лидер ${arr(e.payload.leads).map((p) => `p${p}`).join(', ')}` : ''}`,
    body: [
      `Процесс брокера умер. Контроллер узнает об этом не сразу — через ${ticks(num(e.payload.detectIn))}, когда истечёт сессия брокера. Всё это время продюсеры и потребители шлют ему запросы и ждут ответа до таймаута.`,
      arr(e.payload.leads).length > 0
        ? 'Ожидавшие подтверждения запросы acks=all пропали вместе с ним. Что было записано только на нём и не успело доехать до фолловеров, — вопрос, на который ответит выбор нового лидера.'
        : 'Лидером он не был: его партиции продолжают работать, только ISR у них станет меньше.',
    ],
    kafka: 'KRaft: брокер перестал слать heartbeat контроллеру → broker.session.timeout.ms',
    model: 'Лог упавшего брокера в модели сохраняется, как на диске: вернувшись, брокер сверит его с новым лидером.',
    terms: ['kafka-broker', 'kraft'],
  }),

  'broker.up': (e) => {
    const t = (e.payload.truncated as { partition: number; count: number }[] | undefined) ?? []
    return {
      title: `Брокер B${num(e.payload.broker)} вернулся`,
      body: [
        t.length > 0
          ? `Первым делом он сверил свои логи с лидерами и обрезал то, чего у них нет: ${t.map((x) => `${x.count} ${plural(x.count, 'запись', 'записи', 'записей')} в p${x.partition}`).join(', ')}. Эти записи существовали только на нём — и перестали существовать.`
          : 'Его логи совпадают с логами лидеров: обрезать нечего.',
        'Дальше он работает как фолловер: догоняет лидера и, догнав, возвращается в ISR. Лидерство само к нему не вернётся — для этого Kafka время от времени проводит выборы предпочтительного лидера.',
      ],
      kafka: 'ReplicaFetcher: OffsetsForLeaderEpoch → truncate → fetch',
      terms: ['leader-election', 'isr'],
    }
  },

  'broker.slow': (e) => ({
    title:
      num(e.payload.factor) > 1
        ? `B${num(e.payload.broker)} реплицирует в ${num(e.payload.factor)} раз медленнее`
        : `B${num(e.payload.broker)} снова реплицирует с обычной скоростью`,
    body: [
      num(e.payload.factor) > 1
        ? 'Так выглядит реплика в соседнем дата-центре, брокер с медленным диском или забитой сетью. Брокер жив, отвечает — просто медленно.'
        : 'Причина медлительности ушла: теперь фолловер сможет догнать лидера и вернуться в ISR.',
      'Смотрите на задержку подтверждений при acks=all: пока медленный фолловер в ISR, каждая запись ждёт именно его.',
    ],
    model: 'Медлительность в модели — множитель сетевой задержки запросов репликации.',
    terms: ['kafka-replication', 'isr'],
  }),

  'leader.elect': (e) => {
    const offline = e.payload.to === null
    const missing = arr(e.payload.missing)
    return {
      title: offline
        ? `p${num(e.payload.partition)} без лидера: в ISR не осталось живых реплик`
        : `Новый лидер p${num(e.payload.partition)}: B${num(e.payload.to)}${e.payload.unclean ? ' — нечистые выборы' : ''}, эпоха ${num(e.payload.epoch)}`,
      body: offline
        ? [
            `Живые реплики есть (${brokers(e.payload.candidates)}), но ни одна не входит в ISR — у каждой может не хватать закоммиченных записей. Выбрать такую — значит потерять данные, и Kafka по умолчанию этого не делает.`,
            'Партиция недоступна ни для записи, ни для чтения, пока не вернётся кто-то из ISR. Это осознанный выбор надёжности против доступности; включается обратный — unclean.leader.election.enable.',
          ]
        : [
            e.payload.unclean
              ? `Живых реплик в ISR не было, и включён unclean.leader.election.enable: лидером стал B${num(e.payload.to)}, который не был синхронным. Всё, чего у него нет, потеряно — даже закоммиченное.`
              : `Контроллер выбрал первую живую реплику из ISR. У неё гарантированно есть всё закоммиченное: именно ради этого ISR и существует.`,
            missing.length > 0
              ? `Но у старого лидера были записи, которых у нового нет: ${recList(missing)}. Их не успели скопировать — значит, они не закоммичены. Если продюсер уже получил на них подтверждение (так бывает при acks=1), это потеря.`
              : 'Всё, что было у старого лидера, у нового тоже есть — ничего не пропало.',
            `Эпоха лидера выросла до ${num(e.payload.epoch)}. По эпохам реплики потом находят место, где их истории разошлись, и обрезают лишнее.`,
          ],
      kafka: 'Контроллер KRaft: PartitionChangeBuilder → PartitionChangeRecord в журнал метаданных, leader epoch + 1',
      terms: ['leader-election', 'isr', 'kraft'],
    }
  },

  'log.truncate': (e) => ({
    title: `B${num(e.payload.broker)} обрезал лог p${num(e.payload.partition)} с оффсета ${num(e.payload.from)}: ${recList(e.payload.recs)}`,
    body: [
      `Реплика сверилась с лидером${e.payload.leader === null ? '' : ` B${num(e.payload.leader)}`} и нашла место, где их логи расходятся. Всё после него удалено: у лидера этих записей нет, а лог лидера — единственная правда.`,
      'Удалённые записи не были закоммичены — их не было на всех синхронных репликах. Если продюсер получил на них подтверждение, то только при acks=1 или acks=0. Именно так acks=1 теряет данные.',
      'На освободившихся оффсетах уже могут лежать другие сообщения. Оффсет однозначен только вместе с эпохой лидера.',
    ],
    kafka: 'OffsetsForLeaderEpoch (KIP-101) → UnifiedLog.truncateTo',
    terms: ['leader-election', 'high-watermark'],
  }),

  'rec.lost': (e) => ({
    title: `Потеряно подтверждённое: ${recList(e.payload.recs)} (${acksLabel(e.payload.acks)})`,
    body: [
      e.payload.acks === 0
        ? `Продюсер отправил и забыл, а запрос не дошёл: ${str(e.payload.why)}. Никто никогда не узнает, что эти сообщения были.`
        : `Продюсер получил на эти сообщения «успешно», а в логе нового лидера${e.payload.leader === null ? '' : ` B${num(e.payload.leader)}`} их нет. Они были только на старом лидере и умерли вместе с ним.`,
      'Потребители их не прочитают, повторов не будет: с точки зрения продюсера всё давно доставлено. Это самый неприятный вид потерь — тихий.',
      'Лечение известно: acks=all, фактор репликации 3 и min.insync.replicas=2. Тогда подтверждение означает «есть минимум на двух брокерах», и падение одного ничего не теряет.',
    ],
    terms: ['acks', 'min-insync-replicas', 'delivery-semantics'],
  }),

  'consumer.join': (e) => ({
    title: `${str(e.payload.consumer)} вступает в группу «${str(e.payload.group)}»`,
    body: [
      'Потребитель находит координатора своей группы — брокер, который хранит её оффсеты, — и просит принять его. Координатор запускает ребалансировку: партиции надо разделить заново с учётом новичка.',
      'Группа — это то, что превращает топик из «общей ленты» в очередь задач: каждое сообщение группа обрабатывает один раз, раздавая партиции между участниками.',
    ],
    kafka: 'FindCoordinator → JoinGroup / ConsumerGroupHeartbeat (KIP-848)',
    terms: ['consumer-group', 'rebalance'],
  }),

  'consumer.crash': (e) => ({
    title: `${str(e.payload.consumer)} упал${num(e.payload.lostInMemory) > 0 ? `, в памяти пропало ${msgs(num(e.payload.lostInMemory))}` : ''}`,
    body: [
      `Процесс умер посреди работы. Его партиции (${arr(e.payload.partitions).map((p) => `p${p}`).join(', ') || '—'}) никто не читает: координатор узнает о смерти только через session.timeout.ms (${ticks(num(e.payload.detectIn))}), когда перестанут приходить heartbeat'ы. Всё это время лаг этих партиций растёт.`,
      e.payload.mode === 'auto'
        ? 'Коммит здесь автоматический: он фиксировал всё, что poll успел отдать приложению. Прочитанное, но не обработанное, группа посчитает обработанным.'
        : 'Коммит здесь после обработки: всё, что потребитель успел обработать после последнего коммита, группа обработает ещё раз.',
    ],
    kafka: 'session.timeout.ms (по умолчанию 45 с), heartbeat.interval.ms',
    terms: ['rebalance', 'offset-commit'],
  }),

  'group.rebalance': (e) => {
    const asg = (e.payload.assignment as Record<string, number[]>) ?? {}
    const idle = (e.payload.idle as string[]) ?? []
    const lines = Object.entries(asg)
      .map(([c, ps]) => `${c}: ${ps.length > 0 ? ps.map((p) => `p${p}`).join(', ') : 'ничего'}`)
      .join('; ')
    return {
      title: `Группа «${str(e.payload.group)}» поделила партиции (поколение ${num(e.payload.generation)})`,
      body: [
        `${e.payload.first ? 'Первое распределение' : `Причина: ${str(e.payload.reason)}`}. Теперь ${lines || 'участников нет'}.`,
        idle.length > 0
          ? `${idle.join(', ')} без партиций: участников больше, чем партиций, и лишние просто ждут. Партиция — единица параллелизма группы, больше потребителей, чем партиций, не бывает полезно.`
          : 'Каждая партиция у ровно одного участника группы — поэтому порядок внутри партиции сохраняется и при обработке.',
        e.payload.first
          ? 'Читать каждый начнёт с закоммиченного оффсета группы, а если его нет — с начала лога (auto.offset.reset=earliest).'
          : `Новые владельцы начнут с закоммиченных оффсетов группы — не с того места, где остановился прежний владелец. Всё между ними будет прочитано заново. Пока шла ребалансировка, группа стояла ${ticks(num(e.payload.paused))}.`,
      ],
      kafka: 'RangeAssignor / CooperativeStickyAssignor; с Kafka 4.0 — серверное распределение KIP-848',
      model: 'Модель использует eager-протокол: на время ребалансировки все участники отдают все партиции. Кооперативный протокол и KIP-848 переносят только то, что действительно меняет владельца.',
      terms: ['rebalance', 'consumer-group'],
    }
  },

  'consumer.fetch': (e) => ({
    title: `${str(e.payload.consumer)} прочитал ${recList(e.payload.recs)} из p${num(e.payload.partition)} с оффсета ${num(e.payload.from)}`,
    body: [
      `Потребитель сам просит данные у лидера партиции — Kafka ничего не «проталкивает». Лидер отдаёт записи только до high watermark (${num(e.payload.hw)}), хотя лог у него может быть длиннее (${num(e.payload.leo)}): незакоммиченное читать нельзя.`,
      'Чтение ничего не меняет в логе: сообщение не удаляется и не помечается. Другая группа прочитает те же записи независимо, а эта группа может перечитать их, просто сдвинув оффсет назад.',
    ],
    kafka: 'KafkaConsumer.poll → Fetcher → FetchRequest (до HW; с read_committed — до LSO)',
    terms: ['high-watermark', 'kafka-offset', 'zero-copy'],
  }),

  'consumer.process': (e) => ({
    title: `${str(e.payload.consumer)} обработал m${num(e.payload.rec)}${num(e.payload.times) > 1 ? ` — в ${num(e.payload.times)}-й раз` : ''}`,
    body: [
      `Путь сообщения от send() до обработки занял ${ticks(num(e.payload.latency))}. Kafka здесь уже ни при чём: это код приложения.`,
      'Оффсет ещё не закоммичен. Если потребитель упадёт прямо сейчас, это сообщение прочитает и обработает кто-то другой.',
    ],
    terms: ['offset-commit'],
  }),

  'offset.commit': (e) => {
    const offs = (e.payload.offsets as Record<string, [number, number]>) ?? {}
    const list = Object.entries(offs)
      .map(([p, [a, b]]) => `p${p}: ${a} → ${b}`)
      .join(', ')
    const unprocessed = arr(e.payload.unprocessed)
    return {
      title: `Группа «${str(e.payload.group)}» закоммитила оффсеты: ${list}`,
      body: [
        `Коммит — это запись в служебный топик __consumer_offsets: «группа ${str(e.payload.group)} обработала партицию до такого-то оффсета». Число — следующий оффсет, который надо прочитать, а не последний прочитанный.`,
        e.payload.mode === 'auto'
          ? 'Автокоммит сработал по таймеру и зафиксировал позицию — всё, что poll уже отдал приложению.'
          : e.payload.mode === 'revoke'
            ? 'Перед тем как отдать партиции при ребалансировке, потребитель коммитит то, что успел. Это onPartitionsRevoked.'
            : 'Потребитель разобрал пачку целиком и только потом закоммитил. Упадёт между обработкой и коммитом — сообщения обработают ещё раз: это at-least-once.',
        unprocessed.length > 0
          ? `Среди закоммиченных есть ещё не обработанные: ${recList(unprocessed)}. Если потребитель упадёт до их обработки, группа их пропустит.`
          : 'Все закоммиченные сообщения уже обработаны.',
      ],
      kafka: 'OffsetCommitRequest → координатор группы → __consumer_offsets',
      terms: ['offset-commit', 'delivery-semantics'],
    }
  },

  'rec.reprocess': (e) => ({
    title: `m${num(e.payload.rec)} обработано повторно (${num(e.payload.times)}-й раз)`,
    body: [
      e.payload.cause === 'duplicate'
        ? `В логе две копии этого сообщения: на оффсете ${num(e.payload.firstOffset)} и на ${num(e.payload.offset)}. Потребитель не может отличить дубль от нового сообщения — для него это две записи.`
        : `Сообщение уже обработал прежний владелец партиции, но не успел закоммитить оффсет. Новый владелец начал с закоммиченного — и прошёл этот участок заново.`,
      'Это нормальный режим работы Kafka, а не авария: at-least-once — гарантия по умолчанию. Обработчик должен быть идемпотентным: проверять ключ операции, делать upsert вместо insert, хранить обработанные id.',
    ],
    terms: ['delivery-semantics', 'offset-commit', 'idempotent-producer'],
  }),

  'rec.skipped': (e) => ({
    title: `Группа «${str(e.payload.group)}» пропустила ${recList(e.payload.recs)} навсегда`,
    body: [
      `Закоммиченный оффсет p${num(e.payload.partition)} — ${num(e.payload.committed)}, а эти сообщения ниже него и ни разу не обработаны. Новый владелец партиции начнёт с ${num(e.payload.committed)} и их не увидит.`,
      'Так бывает, когда оффсет коммитят раньше, чем закончена обработка: автокоммит по таймеру при асинхронной обработке, коммит сразу после poll, передача сообщений в пул горутин без учёта того, какие из них закончились. Это at-most-once — и обычно не то, что нужно.',
    ],
    terms: ['offset-commit', 'delivery-semantics'],
  }),
}

export function explain(e: KafkaEvent): Explanation {
  return EXPLAIN[e.type](e)
}

/** Человеческие названия типов — для фильтров и легенды. */
export const EVENT_LABEL: Record<KafkaEventType, string> = {
  'rec.send': 'send()',
  'batch.seal': 'пакет закрыт',
  'produce.request': 'запрос записи',
  'produce.append': 'запись в лог',
  'produce.ack': 'подтверждение',
  'produce.error': 'ошибка записи',
  'produce.failed': 'не доставлено',
  'produce.dedup': 'повтор отброшен',
  'produce.duplicate': 'дубль в логе',
  'resp.lost': 'ответ потерян',
  'repl.fetch': 'репликация',
  'hw.advance': 'HW сдвинулся',
  'isr.shrink': 'ISR сжался',
  'isr.expand': 'ISR расширился',
  'broker.down': 'брокер упал',
  'broker.up': 'брокер вернулся',
  'broker.slow': 'медленная реплика',
  'leader.elect': 'выбор лидера',
  'log.truncate': 'лог обрезан',
  'rec.lost': 'потеря',
  'consumer.join': 'вступление в группу',
  'consumer.crash': 'потребитель упал',
  'group.rebalance': 'ребалансировка',
  'consumer.fetch': 'чтение',
  'consumer.process': 'обработка',
  'offset.commit': 'коммит оффсета',
  'rec.reprocess': 'повторная обработка',
  'rec.skipped': 'пропуск',
}
