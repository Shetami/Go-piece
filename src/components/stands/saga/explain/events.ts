import type { SagaEvent, SagaEventType } from '../engine/types.ts'

/**
 * Разбор событий стенда. Объяснения привязаны к ТИПУ события: какой бы
 * сценарий ни крутился, клик по «event.lost» откроет один и тот же разбор,
 * подставив в него конкретные заказы и числа из payload.
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
const rows = (n: number) => `${n} ${plural(n, 'строка', 'строки', 'строк')}`
const order = (v: unknown) => `заказ №${num(v)}`

const KIND: Record<string, string> = {
  'order.created': 'заказ создан',
  'order.paid': 'заказ оплачен',
  'payment.refund': 'вернуть деньги',
}

type Explainer = (e: SagaEvent) => Explanation

const EXPLAIN: Record<SagaEventType, Explainer> = {
  'order.created': (e) => ({
    title: `${order(e.payload.order)} записан в базу`,
    body: [
      e.payload.publish === 'outbox'
        ? 'Заказ и событие о нём пишутся в одной транзакции: либо есть и то и другое, либо ничего.'
        : 'Заказ записан. Теперь его нужно отправить в брокер — вторым, отдельным действием. Между этими двумя действиями и живут все беды этой лекции.',
    ],
    terms: ['transactional-outbox'],
  }),

  'outbox.write': (e) => ({
    title: `${order(e.payload.order)}: событие записано в outbox`,
    body: [
      `Строка легла в ту же транзакцию, что и сам заказ. В outbox ждут отправки ${rows(num(e.payload.rows))}.`,
      'Брокер об этом ещё ничего не знает — и это нормально: доставка стала задачей отдельного процесса, а согласованность обеспечена транзакцией базы.',
    ],
    go: 'INSERT INTO orders …; INSERT INTO outbox (aggregate_id, type, payload) … — в одной транзакции',
    terms: ['transactional-outbox'],
  }),

  'event.publish': (e) => ({
    title: `${order(e.payload.order)}: событие отправлено в брокер`,
    body: ['Запись в базу и отправка в брокер прошли обе. В обычной жизни так и бывает — проблемы начинаются, когда между ними что-то случается.'],
    terms: ['transactional-outbox'],
  }),

  'event.lost': (e) => ({
    title: `${order(e.payload.order)}: событие потеряно навсегда`,
    body: [
      e.payload.reason === 'crash'
        ? 'Сервис перезапустился между записью в базу и отправкой в брокер. Заказ в базе есть, а события о нём нет и уже не будет: отправлять его было некому и неоткуда.'
        : 'Брокер недоступен, а записать событие больше некуда: при двух записях подряд оно существует только в памяти процесса. Повторять тоже нечего — перезапуск сервиса о нём не вспомнит.',
      'Для остальных сервисов этого заказа просто не существует. Деньги не списаны, доставка не создана, а в базе заказов он висит в состоянии «создан».',
    ],
    go: 'лечится transactional outbox: событие пишется в ту же транзакцию и отправляется отдельным процессом',
    terms: ['transactional-outbox', 'eventual-consistency'],
  }),

  'relay.send': (e) => ({
    title: `Отправщик забрал из outbox ${rows(num(e.payload.rows))}`,
    body: [
      `Он опрашивает таблицу раз в ${ticks(num(e.payload.every))}; самая старая строка ждала ${ticks(num(e.payload.waited))}.`,
      'Это и есть цена outbox: между записью и публикацией проходит время. Уменьшить его можно частым опросом или чтением журнала базы (CDC) вместо опроса.',
    ],
    go: 'опрос SELECT … FOR UPDATE SKIP LOCKED или Debezium, читающий WAL',
    terms: ['transactional-outbox', 'eventual-consistency'],
  }),

  'event.deliver': (e) => ({
    title: `${order(e.payload.order)}: ${KIND[str(e.payload.kind)] ?? str(e.payload.kind)}${e.payload.skipped ? ' — пропущено как уже обработанное' : ''}`,
    body: [
      e.payload.skipped
        ? 'Потребитель узнал ключ идемпотентности и не стал делать работу второй раз. Именно ради этого ключ и нужен.'
            : `Потребитель взял событие в обработку${e.payload.duplicate ? '. Это дубль: то же событие уже приходило' : ''}.`,
    ],
    terms: ['idempotency-key', 'delivery-semantics'],
  }),

  'event.duplicate': (e) => ({
    title: `${order(e.payload.order)}: событие доставлено дважды`,
    body: [
      'Брокеры дают доставку «хотя бы один раз»: если подтверждение потерялось, сообщение придёт снова. Это не сбой, а нормальный режим работы.',
      e.payload.idempotent
        ? 'Потребитель идемпотентен: второй раз он работу не сделает.'
        : 'Потребитель об этом не знает и обработает событие второй раз — со всеми последствиями.',
    ],
    terms: ['delivery-semantics', 'idempotency-key'],
  }),

  'pay.ok': (e) => ({
    title: `${order(e.payload.order)}: деньги списаны`,
    body: [`С создания заказа прошло ${ticks(num(e.payload.waited))}. Дальше платежи расскажут об этом доставке — тем же способом, что и заказы.`],
    terms: ['eventual-consistency'],
  }),

  'pay.double': (e) => ({
    title: `${order(e.payload.order)}: деньги списаны второй раз`,
    body: [
      `Списаний уже ${num(e.payload.charges)}. ${e.payload.duplicate ? 'Причина — дубль доставки' : 'Причина — повторная обработка'}: потребитель не проверяет, делал ли он эту работу раньше.`,
      'Это самая дорогая ошибка в этой лекции: её видно не в метриках, а в обращениях пользователей. Лечится ключом идемпотентности — уникальной строкой в базе, которую вторая попытка не сможет вставить.',
    ],
    go: 'INSERT INTO processed(key) VALUES ($1) ON CONFLICT DO NOTHING — и работать дальше, только если вставка прошла',
    terms: ['idempotency-key', 'delivery-semantics'],
  }),

  'pay.fail': (e) => ({
    title: `${order(e.payload.order)}: списание не удалось (попытка ${num(e.payload.attempt)})`,
    body: [
      num(e.payload.retriesLeft) > 0 ? 'Потребитель повторит обработку через несколько тиков.' : 'Повторов больше нет.',
      'Обработчик события обязан уметь повторяться: любой сбой в середине означает, что то же событие придёт снова.',
    ],
    terms: ['idempotency-key'],
  }),

  'pay.retry': (e) => ({
    title: `${order(e.payload.order)}: повтор обработки через ${ticks(num(e.payload.after))}`,
    body: [
      `Событие «${KIND[str(e.payload.kind)] ?? ''}» будет обработано ещё раз, попытка ${num(e.payload.attempt)}.`,
      'Повтор безопасен ровно настолько, насколько идемпотентен обработчик. Если нет — каждый повтор рискует удвоить эффект.',
    ],
    terms: ['idempotency-key', 'retry-backoff'],
  }),

  'pay.giveup': (e) => ({
    title: `${order(e.payload.order)}: потребитель сдался после ${num(e.payload.attempts)} ${plural(num(e.payload.attempts), 'попытки', 'попыток', 'попыток')}`,
    body: [
      `Шаг «${KIND[str(e.payload.kind)] ?? ''}» так и не удался. Дальше есть два пути: компенсировать уже сделанное или оставить заказ в промежуточном состоянии.`,
      'В жизни такое событие обычно уезжает в очередь неудачных сообщений (dead letter queue), где его разбирают люди или отдельный процесс.',
    ],
    terms: ['saga', 'delivery-semantics'],
  }),

  'ship.ok': (e) => ({
    title: `${order(e.payload.order)}: доставка создана`,
    body: [`Заказ дошёл до конечного состояния за ${ticks(num(e.payload.waited))}. Все сервисы согласны между собой — с опозданием, но согласны.`],
    terms: ['eventual-consistency'],
  }),

  'ship.fail': (e) => ({
    title: `${order(e.payload.order)}: доставку создать не удалось (попытка ${num(e.payload.attempt)})`,
    body: [
      'Деньги уже списаны — отменить это одной транзакцией нельзя, её нет: шаги прошли в разных сервисах и разных базах.',
      e.payload.compensate
        ? num(e.payload.retriesLeft) > 0
          ? 'Сначала будут повторы, а если не поможет — компенсация: деньги вернутся отдельным шагом.'
          : 'Повторы кончились — запускается компенсация.'
        : 'Компенсация выключена: если повторы не помогут, заказ так и останется оплаченным без доставки.',
    ],
    terms: ['saga'],
  }),

  'saga.compensate': (e) => ({
    title: `${order(e.payload.order)}: компенсация — ${str(e.payload.action)}`,
    body: [
      `Шаг «${str(e.payload.step)}» не удался, и сага откатывает то, что успела сделать: отдельным действием, а не откатом транзакции.`,
      'Компенсация — это не «отмена», а новое бизнес-действие: возврат денег, снятие резерва, письмо с извинением. Оно тоже может не удаться, поэтому его тоже повторяют и делают идемпотентным.',
    ],
    go: 'шаги саги и их компенсации описывают явно; состояние саги хранят в базе, чтобы пережить перезапуск',
    terms: ['saga', 'idempotency-key'],
  }),

  'saga.stuck': (e) => ({
    title: `${order(e.payload.order)} застрял в состоянии «${str(e.payload.state) === 'paid' ? 'оплачен' : 'создан'}»`,
    body: [
      e.payload.reason === 'lost'
        ? 'Событие о заказе потеряно: никто и никогда его не обработает. Заказ будет лежать в базе как немой укор.'
        : 'Шаг не удался, повторы кончились, компенсации нет. Деньги списаны, доставки не будет.',
      'Такие заказы — главный признак проблем с согласованностью. Их ищут сверками: регулярным сравнением состояний в разных сервисах.',
    ],
    terms: ['eventual-consistency', 'saga'],
  }),

  'svc.crash': (e) => ({
    title: `Сервис заказов перезапустился`,
    body: [
      e.payload.publish === 'dual-write'
        ? `Всё, что он собирался отправить в брокер, но не успел, пропало: ${num(e.payload.lost)} ${plural(num(e.payload.lost), 'событие', 'события', 'событий')}.`
        : `Outbox лежит в базе и падение пережил: ${rows(num(e.payload.outbox))} ждут отправки и уедут, когда отправщик вернётся.`,
      'Разница между двумя способами публикации видна именно в такие моменты, а не в обычной жизни.',
    ],
    terms: ['transactional-outbox'],
  }),

  'broker.down': (e) => ({
    title: 'Брокер недоступен',
    body: [
      e.payload.publish === 'dual-write'
        ? 'Публиковать некуда, а хранить событие негде: вторая запись в базу не делалась. Всё, что создаётся сейчас, потеряется.'
        : 'Публиковать некуда, но это не страшно: события лежат в базе и дождутся возвращения брокера.',
    ],
    terms: ['transactional-outbox'],
  }),

  'broker.up': (e) => ({
    title: 'Брокер вернулся',
    body: [
      num(e.payload.outbox) > 0
        ? `В outbox накопилось ${rows(num(e.payload.outbox))} — отправщик выгрузит их разом. Потребители получат всплеск событий: об этом стоит помнить, планируя их ёмкость.`
        : 'Накопленного нет.',
    ],
    terms: ['transactional-outbox', 'eventual-consistency'],
  }),
}

export function explain(e: SagaEvent): Explanation {
  return EXPLAIN[e.type](e)
}
