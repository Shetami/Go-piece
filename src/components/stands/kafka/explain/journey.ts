import type { Acks, Rec, StepKind, TraceStep } from '../engine/types.ts'
import { num, str, ticks } from './events.ts'

/**
 * Путь одного сообщения — от send() до коммита оффсета.
 *
 * Лента событий отвечает на вопрос «что происходит в кластере», а путь — на
 * вопрос «где сейчас моё сообщение и почему оно там». Каждый шаг пути привязан
 * к ВИДУ шага: описание строится из данных, записанных движком в trace.
 */

export interface StepInfo {
  /** Короткое название остановки — для схемы пути. */
  label: string
  /** Где это происходит. */
  where: 'продюсер' | 'сеть' | 'лидер' | 'реплики' | 'потребитель' | 'группа'
  /** Что происходит, если шаг ещё впереди. */
  ahead: string
}

export const STEP_INFO: Record<StepKind, StepInfo> = {
  send: { label: 'send()', where: 'продюсер', ahead: 'Приложение вызывает send(): сообщение сериализуется, вызов сразу возвращается.' },
  partition: { label: 'партиция', where: 'продюсер', ahead: 'Партиционер выбирает партицию: по хешу ключа или липко, если ключа нет.' },
  batch: { label: 'в пакет', where: 'продюсер', ahead: 'Сообщение ложится в открытый пакет своей партиции и ждёт попутчиков.' },
  seal: { label: 'пакет закрыт', where: 'продюсер', ahead: 'Пакет закроется, когда наполнится или истечёт linger.ms.' },
  request: { label: 'запрос', where: 'сеть', ahead: 'Sender отправит пакет лидеру партиции запросом ProduceRequest.' },
  append: { label: 'лог лидера', where: 'лидер', ahead: 'Лидер допишет пакет в конец лога и назначит оффсеты.' },
  replicate: { label: 'реплики', where: 'реплики', ahead: 'Фолловеры скопируют пакет своими запросами Fetch.' },
  commit: { label: 'HW', where: 'лидер', ahead: 'Когда пакет будет на всех репликах ISR, HW перейдёт его — сообщение закоммичено.' },
  ack: { label: 'ack', where: 'продюсер', ahead: 'Продюсер получит ответ, и сработает колбэк send().' },
  fetch: { label: 'poll', where: 'потребитель', ahead: 'Потребитель группы прочитает сообщение — но только после того, как его перейдёт HW.' },
  process: { label: 'обработка', where: 'потребитель', ahead: 'Код приложения обработает сообщение.' },
  offset: { label: 'коммит', where: 'группа', ahead: 'Группа закоммитит оффсет — и больше это сообщение не прочитает.' },
  retry: { label: 'повтор', where: 'продюсер', ahead: '' },
  'resp-lost': { label: 'ответ потерян', where: 'сеть', ahead: '' },
  dedup: { label: 'повтор отброшен', where: 'лидер', ahead: '' },
  duplicate: { label: 'дубль', where: 'лидер', ahead: '' },
  truncated: { label: 'обрезано', where: 'реплики', ahead: '' },
  lost: { label: 'потеряно', where: 'лидер', ahead: '' },
  failed: { label: 'не доставлено', where: 'продюсер', ahead: '' },
  reprocess: { label: 'повторная обработка', where: 'потребитель', ahead: '' },
  skipped: { label: 'пропущено', where: 'группа', ahead: '' },
}

/** Обычный путь. Порядок ack и репликации зависит от acks — в этом весь смысл настройки. */
export function canonicalPath(acks: Acks): StepKind[] {
  const head: StepKind[] = ['send', 'partition', 'batch', 'seal', 'request']
  const tail: StepKind[] = ['fetch', 'process', 'offset']
  if (acks === 0) return [...head, 'ack', 'append', 'replicate', 'commit', ...tail]
  if (acks === 1) return [...head, 'append', 'ack', 'replicate', 'commit', ...tail]
  return [...head, 'append', 'replicate', 'commit', 'ack', ...tail]
}

/** Шаги, которые означают, что с сообщением что-то пошло не так. */
export const BAD_STEPS = new Set<StepKind>(['retry', 'resp-lost', 'duplicate', 'truncated', 'lost', 'failed', 'reprocess', 'skipped'])

const reasons: Record<string, string> = {
  NOT_LEADER_OR_FOLLOWER: 'брокер больше не лидер',
  NOT_ENOUGH_REPLICAS: 'в ISR меньше min.insync.replicas',
  NOT_ENOUGH_REPLICAS_AFTER_APPEND: 'записано, но ISR сжался ниже min.insync.replicas',
  REQUEST_TIMEOUT: 'ответа нет дольше request.timeout.ms',
}

/** Заголовок и пояснение для уже случившегося шага. */
export function describeStep(t: TraceStep, rec: Rec): { title: string; body: string } {
  const d = t.data
  switch (t.step) {
    case 'send':
      return {
        title: `${str(d.producer)} вызвал send(m${rec.id})${d.key === null ? ' без ключа' : ` с ключом «${str(d.key)}»`}`,
        body: 'Ключ и значение превращаются в байты сериализатором. По сети ещё ничего не ушло: send() асинхронный и возвращается сразу.',
      }
    case 'partition':
      return {
        title: `Партиция ${num(d.partition)} из ${num(d.partitions)}`,
        body:
          d.how === 'hash'
            ? `murmur2(«${str(d.key)}») % ${num(d.partitions)} = ${num(d.partition)}. Этот ключ всегда будет в p${num(d.partition)} — пока число партиций не изменится.`
            : 'Ключа нет — липкий партиционер держит все такие сообщения в одной партиции, пока её пакет не уйдёт.',
      }
    case 'batch':
      return {
        title: `В пакет #${num(d.batch)} — ${num(d.size)}-е по счёту из ${num(d.batchSize)}`,
        body: 'Пакет живёт в памяти продюсера (RecordAccumulator). Сообщения одной партиции копятся вместе, чтобы уйти одним запросом.',
      }
    case 'seal':
      return {
        title: d.reason === 'size' ? `Пакет полон (${num(d.size)})` : `Пакет ушёл по linger: ${num(d.size)} шт. за ${ticks(num(d.waited))}`,
        body:
          d.reason === 'size'
            ? 'Набран batch.size — пакет готов к отправке немедленно.'
            : 'Пакет прождал linger.ms, и Sender забрал его. Больше в него ничего не добавится.',
      }
    case 'request':
      return {
        title: `ProduceRequest → B${num(d.broker)}${num(d.attempt) > 1 ? `, попытка ${num(d.attempt)}` : ''}`,
        body: `Пакет из ${num(d.size)} шт. летит лидеру партиции. ${d.acks === 'all' ? 'acks=all: ответ придёт после репликации на ISR.' : d.acks === 1 ? 'acks=1: ответ придёт, как только лидер запишет.' : 'acks=0: ответа не будет.'}`,
      }
    case 'append':
      return {
        title: `B${num(d.broker)} записал на оффсет ${num(d.offset)} (эпоха ${num(d.epoch)})`,
        body: 'Теперь у сообщения есть адрес: топик, партиция, оффсет. Запись лежит в страничном кэше лидера — пока в единственном экземпляре.',
      }
    case 'replicate':
      return {
        title: `Копия на B${num(d.broker)}`,
        body: `Фолловер сам забрал запись у лидера B${num(d.leader)} запросом Fetch. Лидер узнает об этом из его следующего запроса.`,
      }
    case 'commit':
      return {
        title: `Закоммичено: HW p${num(d.partition)} = ${num(d.hw)}`,
        body: `Запись есть на всех репликах ISR (${(d.isr as number[]).map((b) => `B${b}`).join(', ')}). С этого момента её видят потребители, и она переживёт падение любого из этих брокеров.`,
      }
    case 'ack':
      return d.acks === 0
        ? { title: 'Считается доставленным', body: 'acks=0: продюсер отметил сообщение доставленным в момент отправки. Никто ничего не подтверждал.' }
        : {
            title: `Колбэк: успешно за ${ticks(num(d.latency))}`,
            body:
              d.acks === 1
                ? 'acks=1: лидер ответил сразу после записи. Реплик в этот момент могло ещё не быть — это и есть риск acks=1.'
                : 'acks=all: ответ пришёл после коммита. Сообщение гарантированно на нескольких брокерах.',
          }
    case 'fetch':
      return {
        title: `${str(d.consumer)} («${str(d.group)}») прочитал с B${num(d.broker)}`,
        body: 'poll() вернул сообщение приложению. Лог при этом не изменился: чтение в Kafka ничего не удаляет.',
      }
    case 'process':
      return {
        title: `${str(d.consumer)} обработал — ${ticks(num(d.latency))} от send()`,
        body: 'Обработано, но оффсет ещё не закоммичен: упади потребитель сейчас — сообщение обработают снова.',
      }
    case 'offset':
      return {
        title: `Группа «${str(d.group)}» закоммитила ${num(d.committed)}`,
        body: d.processed
          ? 'Оффсет записан в __consumer_offsets. Путь окончен: группа больше не прочитает это сообщение.'
          : 'Оффсет закоммичен раньше, чем сообщение обработано. Если потребитель упадёт сейчас, оно будет пропущено.',
      }
    case 'retry':
      return {
        title: `Ошибка: ${reasons[str(d.error)] ?? str(d.error)}`,
        body: `Продюсер вернул пакет в очередь и повторит его${d.leader === null ? ', когда появится лидер' : ` — лидеру B${num(d.leader)}`}. Приложение этого не видит.`,
      }
    case 'resp-lost':
      return {
        title: 'Ответ потерян в сети',
        body: `Брокер B${num(d.broker)} ответил, но ответ не дошёл. Продюсер будет ждать до request.timeout.ms и повторит — не зная, что запись уже в логе.`,
      }
    case 'dedup':
      return {
        title: `Повтор узнан: остаётся оффсет ${num(d.offset)}`,
        body: 'Идемпотентный продюсер пронумеровал пакет, и брокер увидел, что этот номер уже записан. Второй копии нет.',
      }
    case 'duplicate':
      return {
        title: `Дубль на оффсете ${num(d.offset)} (оригинал — ${num(d.first)})`,
        body: 'Повтор без идемпотентности: брокер не может узнать пакет и пишет его второй раз. Потребитель увидит сообщение дважды.',
      }
    case 'truncated':
      return {
        title: `Удалено с B${num(d.broker)} (оффсет ${num(d.offset)})`,
        body: 'Реплика сверилась с новым лидером и обрезала лог: у лидера этой записи нет.',
      }
    case 'lost':
      return {
        title: 'Потеряно',
        body:
          d.acks === 0
            ? `Запрос не дошёл (${str(d.why)}), а продюсер давно считает сообщение доставленным.`
            : `Продюсер получил подтверждение, но нового лидера B${num(d.leader)} этой записи нет. Она была только на старом лидере.`,
      }
    case 'failed':
      return {
        title: `Не доставлено: ${num(d.attempts)} попыток за ${ticks(num(d.age))}`,
        body: 'Истёк delivery.timeout.ms, колбэк получил ошибку. Сообщение при этом могло уже лежать в логе — продюсер этого не знает.',
      }
    case 'reprocess':
      return {
        title: `${str(d.consumer)} обработал снова (${num(d.times)}-й раз)`,
        body:
          num(d.offset) !== rec.offset
            ? `Это копия с оффсета ${num(d.offset)} — дубль, записанный повтором продюсера.`
            : 'Прежний владелец партиции обработал сообщение, но не закоммитил. Новый начал с закоммиченного оффсета.',
      }
    case 'skipped':
      return {
        title: `Группа «${str(d.group)}» пропустила`,
        body: `Закоммиченный оффсет ${num(d.committed)} уже выше, а сообщение так и не обработано. Его больше никто не прочитает.`,
      }
  }
}
