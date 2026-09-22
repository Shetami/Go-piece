import type { ReplEvent, ReplEventType } from '../engine/types.ts'

/**
 * Разбор событий стенда. Объяснения привязаны к ТИПУ события: какой бы
 * сценарий ни крутился, клик по «commit.lost» откроет один и тот же разбор,
 * подставив в него конкретные узлы, клиентов и числа из payload.
 */

export interface Explanation {
  title: string
  body: string[]
  /** Что делает настоящий PostgreSQL. */
  pg?: string
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
const records = (n: number) => `${n} ${plural(n, 'запись', 'записи', 'записей')}`
const commits = (n: number) => `${n} ${plural(n, 'коммит', 'коммита', 'коммитов')}`
const val = (v: unknown) => (v === null || v === undefined ? 'NULL' : str(v))

const LEVEL_TEXT: Record<string, string> = {
  off: 'не ждёт даже своего диска',
  local: 'ждёт только своего диска',
  remote_write: 'ждёт, пока реплика получит запись и передаст её ОС',
  on: 'ждёт, пока реплика сбросит запись на диск',
  remote_apply: 'ждёт, пока реплика проиграет запись и её станет видно запросам',
}
const STANDBYS_TEXT: Record<string, string> = {
  none: "synchronous_standby_names пуст — синхронных реплик нет",
  'first-r1': "synchronous_standby_names = 'FIRST 1 (r1)'",
  'any-1': "synchronous_standby_names = 'ANY 1 (*)'",
}

type Explainer = (e: ReplEvent) => Explanation

const EXPLAIN: Record<ReplEventType, Explainer> = {
  'wal.write': (e) => ({
    title: `${str(e.payload.client)}: UPDATE ${str(e.payload.key)} → ${num(e.payload.value)}, запись WAL ${num(e.payload.lsn)}`,
    body: [
      `Ведущий записал изменение в WAL под номером ${num(e.payload.lsn)}. Всё, что знают реплики о данных, они узнают из этого журнала: репликация в PostgreSQL — это пересылка WAL, а не SQL-запросов.`,
      'Отправлять запись walsender будет только после того, как ведущий сбросит её на свой диск: реплика никогда не должна опередить ведущего.',
    ],
    pg: 'XLogInsert → XLogFlush; WalSndWakeup будит walsender',
    model: 'Одна запись WAL в модели — целая транзакция: изменение и её коммит. В PostgreSQL это несколько записей.',
    terms: ['wal', 'streaming-replication'],
  }),

  'replica.receive': (e) => ({
    title: `${str(e.payload.node)} получила WAL ${num(e.payload.from)}–${num(e.payload.to)}`,
    body: [
      `walreceiver на реплике принял записи и записал их в свой pg_wal. Задержка сети до этой реплики — ${ticks(num(e.payload.latency))}. На следующем тике он сбросит их на диск, а startup-процесс проиграет.`,
      'Три позиции реплики — write, flush и replay — ведущий видит в pg_stat_replication. Разница между ними и есть отставание, разложенное по причинам: сеть, диск реплики, скорость проигрывания.',
    ],
    pg: 'walreceiver: XLogWalRcvWrite → XLogWalRcvFlush',
    terms: ['streaming-replication', 'replication-lag'],
  }),

  'replica.replay': (e) => ({
    title: `${str(e.payload.node)} проиграла WAL до ${num(e.payload.to)}${num(e.payload.behind) > 0 ? ` — отстаёт на ${records(num(e.payload.behind))}` : ''}`,
    body: [
      `Проиграть запись — значит применить её к страницам данных, ровно как при восстановлении после сбоя. Только теперь восстановление не кончается: реплика вечно догоняет ведущего. За тик эта реплика проигрывает до ${records(num(e.payload.rate))}${e.payload.slow ? ' — сейчас она работает медленнее обычного' : ''}.`,
      'Запросы на реплике видят данные ровно до этой позиции. Если ведущий пишет быстрее, чем реплика проигрывает, отставание растёт без предела — сколько ни жди.',
    ],
    pg: 'startup-процесс: PerformWalRecovery → ApplyWalRecord; позиция — pg_last_wal_replay_lsn()',
    terms: ['replication-lag', 'crash-recovery'],
  }),

  'commit.wait': (e) => ({
    title: `${str(e.payload.client)}: COMMIT ждёт — ${str(e.payload.waitingFor)}`,
    body: [
      `synchronous_commit = ${str(e.payload.level)}: коммит ${LEVEL_TEXT[str(e.payload.level)] ?? ''}. ${STANDBYS_TEXT[str(e.payload.standbys)] ?? ''}.`,
      e.payload.standbys === 'none'
        ? 'Значит, коммит подтвердится, как только запись окажется на диске ведущего. Реплики получат её позже, асинхронно, — и если ведущий умрёт раньше, она может до них не дойти.'
        : 'Клиент не получит ответа, пока синхронная реплика не сообщит, что дошла до этой записи. Круг до реплики и обратно добавляется к каждому коммиту.',
      'Пока коммит ждёт, другие транзакции его изменения не видят — даже на самом ведущем.',
    ],
    pg: 'RecordTransactionCommit → XLogFlush, затем SyncRepWaitForLSN',
    terms: ['synchronous-replication', 'synchronous-commit'],
  }),

  'commit.ack': (e) => ({
    title: num(e.payload.waited) > 0 ? `${str(e.payload.client)}: COMMIT подтверждён через ${ticks(num(e.payload.waited))}` : `${str(e.payload.client)}: COMMIT подтверждён сразу`,
    body: [
      e.payload.standbys === 'none' || e.payload.level === 'local' || e.payload.level === 'off'
        ? 'Подтверждение означает: запись на диске ведущего. О репликах оно не говорит ничего — они могут отставать на сколько угодно.'
        : `Подтверждение означает: синхронная реплика ${e.payload.level === 'remote_apply' ? 'уже проиграла запись — на ней её видно запросам' : e.payload.level === 'remote_write' ? 'получила запись, хотя на диск ещё не сбросила' : 'сбросила запись на диск'}. Упади ведущий сейчас — эта транзакция не пропадёт.`,
    ],
    terms: ['synchronous-replication'],
  }),

  'commit.hang': (e) => ({
    title: `${str(e.payload.client)}: COMMIT ждёт уже ${ticks(num(e.payload.waited))} — ${str(e.payload.waitingFor)}`,
    body: [
      'Ведущий жив и здоров, запись у него на диске, но подтвердить коммит он не может: синхронной реплики, которую он обязан дождаться, нет. Приложение видит просто зависший запрос — без ошибки.',
      e.payload.standbys === 'first-r1'
        ? "С FIRST 1 (r1) синхронной может быть только r1. Пока её нет, запись в базу стоит целиком. Поэтому синхронных кандидатов всегда перечисляют несколько: 'ANY 1 (r1, r2)' или 'FIRST 1 (r1, r2)'."
        : 'Ни одна из синхронных реплик не отвечает. Больше кандидатов — меньше шансов, что недоступны все сразу.',
      'Ждущий коммит можно отменить (pg_cancel_backend), но запись уже в WAL ведущего: транзакция станет видна, хотя до реплики не дошла. Отмена ожидания — это не откат.',
    ],
    pg: 'SyncRepWaitForLSN — бесконечное ожидание; синхронные реплики выбирает SyncRepGetSyncStandbys',
    terms: ['synchronous-replication'],
  }),

  'read.ok': (e) => ({
    title: `${str(e.payload.client)} прочитал ${str(e.payload.key)} = ${val(e.payload.value)} на ${str(e.payload.node)}`,
    body: [
      e.payload.isPrimary
        ? 'Чтение с ведущего видит всё подтверждённое. Это самый свежий и самый дорогой путь: вся нагрузка на один сервер.'
        : 'Реплика уже проиграла последнее изменение этой строки — чтение совпало с ведущим. Но это удача, а не гарантия.',
    ],
    terms: ['streaming-replication'],
  }),

  'read.stale': (e) => ({
    title: `${str(e.payload.client)}: на ${str(e.payload.node)} ${str(e.payload.key)} = ${val(e.payload.value)}, а на ведущем уже ${val(e.payload.freshValue)}`,
    body: [
      `Реплика отстаёт на ${records(num(e.payload.lag))}: последнее изменение этой строки она ещё не проиграла. Это нормальная работа асинхронной реплики — она показывает базу такой, какой та была недавно.`,
      'Для списка товаров или ленты это приемлемо. Для остатка на счёте перед списанием — нет: такие чтения идут на ведущий.',
    ],
    terms: ['replication-lag'],
  }),

  'read.own-stale': (e) => ({
    title: `${str(e.payload.client)} не видит собственную запись: на ${str(e.payload.node)} ${val(e.payload.value)}, а сохранил ${val(e.payload.ownValue)}`,
    body: [
      `Клиент только что получил подтверждение коммита (LSN ${num(e.payload.own)}) и сразу прочитал с реплики. А реплика эту запись ещё не проиграла: у неё строка на LSN ${num(e.payload.lsn)}. Пользователь нажал «сохранить», страница перезагрузилась — а там старые данные.`,
      'Лечится одним из способов: читать свои изменения с ведущего (например, несколько секунд после записи), запоминать LSN коммита (pg_current_wal_lsn()) и читать с реплики, только когда pg_last_wal_replay_lsn() его догнал, или synchronous_commit = remote_apply — но он гарантирует это только для синхронной реплики, а не для всех.',
    ],
    terms: ['read-your-writes', 'replication-lag'],
  }),

  'read.backwards': (e) => ({
    title: `${str(e.payload.client)}: время пошло назад — было ${val(e.payload.seenValue)}${e.payload.prevNode ? ` (${str(e.payload.prevNode)})` : ''}, теперь ${val(e.payload.value)} на ${str(e.payload.node)}`,
    body: [
      'Прошлое чтение попало на реплику, которая отстаёт меньше, это — на более отстающую. Каждая реплика по отдельности показывает согласованное прошлое, но разное. Для пользователя это выглядит так, будто лайк поставили, а потом он исчез.',
      'Лечится привязкой клиента к одной реплике (по пользователю или сессии) или тем же приёмом с LSN: не читать с реплики, которая отстала больше, чем уже видел клиент.',
    ],
    terms: ['monotonic-reads', 'replication-lag'],
  }),

  'read.error': (e) => ({
    title: `${str(e.payload.client)}: чтение с ${str(e.payload.node)} не удалось`,
    body: [
      e.payload.why === 'broken'
        ? 'Реплика отстала так, что нужного ей WAL на ведущем больше нет. Она не принимает запросы, пока её не пересоздадут из новой копии.'
        : 'Узел недоступен. Балансировщик чтений должен уметь убирать такие узлы из списка — иначе часть запросов будет падать.',
    ],
    terms: ['streaming-replication'],
  }),

  'write.error': (e) => ({
    title: `${str(e.payload.client)}: запись не удалась — ведущего нет`,
    body: [
      e.payload.failover
        ? 'Ведущий умер, а новый ещё не выбран: менеджер кластера сначала должен убедиться, что старый действительно мёртв, а не просто медленно отвечает. Все эти тики запись в базу невозможна.'
        : 'Ведущий умер, а автоматического переключения нет. Реплики живы и отвечают на чтения, но писать некуда, пока человек не сделает одну из них ведущей.',
      'Приложение должно повторять запись с паузой и переподключаться — после переключения ведущий будет по другому адресу.',
    ],
    terms: ['failover'],
  }),

  'primary.down': (e) => ({
    title: `Ведущий ${str(e.payload.node)} упал`,
    body: [
      `Всё, что ведущий не успел отправить (${records(num(e.payload.inflight))} в пути), пропало вместе с ним. Реплики сохранили только то, что уже получили.`,
      num(e.payload.waiting) > 0
        ? `${num(e.payload.waiting)} ${plural(num(e.payload.waiting), 'клиент ждал', 'клиента ждали', 'клиентов ждали')} подтверждения коммита и получили обрыв соединения. Закоммичена ли их транзакция, они не знают.`
        : 'Подтверждения коммита в этот момент никто не ждал.',
      e.payload.failover
        ? `Менеджер кластера заметит смерть через ${ticks(num(e.payload.detect))} и сделает ведущей самую свежую реплику.`
        : 'Автоматического переключения нет: запись встала до вмешательства человека.',
    ],
    terms: ['failover'],
  }),

  'replica.down': (e) => ({
    title: `Реплика ${str(e.payload.node)} недоступна`,
    body: [
      e.payload.sync
        ? 'Это синхронная реплика. Коммиты, которым нужно её подтверждение, будут ждать, пока она не вернётся.'
        : 'Реплика асинхронная: запись на ведущем продолжается как ни в чём не бывало.',
      'Что будет с WAL, который реплика пропустит, решает слот репликации: со слотом ведущий сохранит всё до её возвращения, без слота — только wal_keep_size последних записей.',
    ],
    terms: ['replication-slot', 'synchronous-replication'],
  }),

  'replica.up': (e) => ({
    title: `Реплика ${str(e.payload.node)} вернулась, отстаёт на ${records(num(e.payload.behind))}`,
    body: [
      `walreceiver подключается к ведущему и просит WAL начиная с ${num(e.payload.from) + 1}. Если ведущий его ещё хранит (сейчас самая старая запись — ${num(e.payload.oldest)}), реплика догонит, получая записи пачками. Если нет — нужна новая копия всей базы.`,
    ],
    pg: 'START_REPLICATION с позиции реплики; walsender читает WAL из pg_wal',
    terms: ['replication-slot'],
  }),

  'replica.slow': (e) => ({
    title: `${str(e.payload.node)} проигрывает медленнее: ${num(e.payload.rate)} вместо ${num(e.payload.normal)} за тик`,
    body: [
      'Проигрывание WAL в PostgreSQL идёт в один поток. Медленный диск реплики, тяжёлые запросы на ней или просто большая запись (CREATE INDEX, массовый UPDATE) — и реплика начинает отставать, хотя сеть в порядке.',
    ],
    terms: ['replication-lag'],
  }),

  'failover.promote': (e) => ({
    title: `${str(e.payload.node)} стала ведущей${num(e.payload.lost) > 0 ? ` — ${commits(num(e.payload.lost))} потеряно` : ''}`,
    body: [
      `Ведущего не было ${ticks(num(e.payload.downFor))}. Менеджер кластера выбрал реплику, получившую больше всех, — она проиграла весь свой WAL до ${num(e.payload.end)} и начала принимать запись. Остальные реплики теперь получают WAL от неё.`,
      num(e.payload.discarded) > 0
        ? `${records(num(e.payload.discarded))} старого ведущего до неё так и не дошли. ${num(e.payload.lost) > 0 ? `Из них ${commits(num(e.payload.lost))} клиенты успели получить подтверждение — эти данные потеряны, хотя клиентам сказали «готово».` : 'Подтверждения по ним никто не получал.'}`
        : 'Все записи старого ведущего есть на новом. Ничего не потеряно.',
      `Старый ведущий, ${str(e.payload.old)}, если вернётся, не сможет просто стать репликой: у него могут быть записи, которых нет у нового. Его возвращают через pg_rewind или пересоздают.`,
    ],
    pg: 'pg_promote() / Patroni: выбор самой свежей реплики, новый таймлайн WAL',
    model: 'Переключение делает идеальный менеджер кластера: он точно знает, какая реплика свежее всех, и никогда не ошибается в том, что ведущий мёртв.',
    terms: ['failover', 'synchronous-replication'],
  }),

  'commit.lost': (e) => ({
    title: `Потерян подтверждённый коммит ${str(e.payload.client)} (LSN ${num(e.payload.lsn)})`,
    body: [
      `Клиент получил «COMMIT», но новая ведущая, ${str(e.payload.node)}, получила WAL только до ${num(e.payload.end)}. Запись ${num(e.payload.lsn)} существовала только на старом ведущем.`,
      'Это не баг, а обещание асинхронной репликации: подтверждение означает «на диске ведущего», не больше. Синхронная реплика закрывает эту дыру — ценой круга по сети в каждом коммите.',
    ],
    terms: ['failover', 'synchronous-replication'],
  }),

  'commit.unknown': (e) => ({
    title: `${str(e.payload.client)}: соединение оборвалось во время COMMIT — исход неизвестен`,
    body: [
      'Запись о коммите уже была в WAL ведущего, но подтверждение клиент не получил. Транзакция могла выжить (если запись успела на новую ведущую) или пропасть. Клиент не знает — и узнать может, только проверив данные.',
      'Поэтому повтор записи после обрыва должен быть идемпотентным: уникальный id операции и INSERT … ON CONFLICT DO NOTHING, а не слепой повтор.',
    ],
    terms: ['failover'],
  }),

  'wal.retained': (e) => ({
    title: `Ведущий хранит уже ${records(num(e.payload.kept))} WAL — их держит слот ${str(e.payload.node)}`,
    body: [
      `Слот репликации помнит, докуда дошла реплика, и не даёт удалить WAL, который ей ещё нужен. ${e.payload.up ? 'Реплика отстаёт.' : 'Реплики нет — а слот ждёт её.'}`,
      'Пока реплика не вернётся, WAL будет расти. Забытый слот умершей реплики — классическая причина, по которой на ведущем кончается диск. Защита — max_slot_wal_keep_size: сверх него слот станет недействительным, а реплику придётся пересоздать.',
    ],
    pg: 'ReplicationSlotsComputeRequiredLSN; мониторинг — pg_replication_slots.restart_lsn',
    terms: ['replication-slot'],
  }),

  'wal.removed': (e) => ({
    title: `Ведущий удалил WAL, который нужен ${str(e.payload.node)}`,
    body: [
      `Без слота ведущий хранит только wal_keep_size последних записей (здесь ${num(e.payload.keep)}). Реплике нужна запись ${num(e.payload.need)}, а самая старая оставшаяся — ${num(e.payload.oldest)}.`,
      e.payload.up ? 'Реплика не успевает забирать WAL — и уже не догонит.' : 'Когда реплика вернётся, догнать ведущего она не сможет.',
    ],
    terms: ['replication-slot'],
  }),

  'replica.broken': (e) => ({
    title: `${str(e.payload.node)} не может догнать ведущего: нужного WAL больше нет`,
    body: [
      `Реплика просит WAL с ${num(e.payload.need)}, а ведущий хранит только с ${num(e.payload.oldest)}. «requested WAL segment has already been removed» — реплику придётся пересоздать из новой копии (pg_basebackup), на большой базе это часы.`,
      e.payload.slots
        ? 'Слоты включены, но реплика отстала настолько, что слот её не спас.'
        : 'Со слотом этого бы не случилось — ценой роста WAL на ведущем, пока реплики не было. Выбор между «может кончиться диск» и «может сломаться реплика».',
    ],
    terms: ['replication-slot'],
  }),

  'vacuum.run': (e) => ({
    title: num(e.payload.removed) > 0 ? `Вакуум на ведущем убрал ${num(e.payload.removed)} мёртвых версий` : 'Вакуум на ведущем: убирать нечего',
    body: [
      'Очистка на ведущем записывается в WAL и приезжает на реплики как обычное изменение: реплика физически удалит те же версии в своих страницах. Реплика — побайтовая копия, у неё не может быть своих версий строк.',
      e.payload.feedback
        ? `Горизонт очистки учитывает снимки запросов на репликах (hot_standby_feedback): ${num(e.payload.kept)} версий пришлось оставить.`
        : 'Горизонт считается только по транзакциям ведущего — о запросах на репликах он ничего не знает.',
    ],
    terms: ['vacuum', 'recovery-conflict'],
  }),

  'feedback.hold': (e) => ({
    title: `hot_standby_feedback: ${str(e.payload.node)} держит ${num(e.payload.kept)} мёртвых версий на ведущем`,
    body: [
      'Реплика сообщила ведущему, что у неё идёт запрос со старым снимком. Ведущий не убирает версии, которые этот запрос может читать, — и конфликта на реплике не будет.',
      'Цена — та же, что у долгой транзакции на самом ведущем: мёртвые версии копятся, таблицы раздуваются. Долгий отчёт на реплике с feedback — это долгий отчёт на ведущем, только его не видно в pg_stat_activity ведущего.',
    ],
    pg: 'ProcessStandbyHSFeedbackMessage — xmin реплики попадает в горизонт ведущего',
    terms: ['hot-standby-feedback', 'xmin-horizon', 'table-bloat'],
  }),

  'conflict.wait': (e) => ({
    title: `${str(e.payload.node)} остановила проигрывание: запись очистки мешает запросу ${str(e.payload.clients)}`,
    body: [
      `Запись WAL ${num(e.payload.lsn)} удаляет ${num(e.payload.removed)} версий, которые ещё видит снимок запроса. Реплика не может одновременно проиграть её и сохранить данные для запроса — это конфликт восстановления.`,
      num(e.payload.delay) < 0
        ? 'max_standby_streaming_delay = -1: реплика будет ждать запрос сколько угодно. Отставание растёт, пока запрос не кончится.'
        : `Реплика подождёт ${ticks(num(e.payload.delay))} (max_standby_streaming_delay), а потом отменит запрос. Всё это время WAL копится и отставание растёт.`,
    ],
    pg: 'ResolveRecoveryConflictWithSnapshot → ожидание до max_standby_streaming_delay',
    terms: ['recovery-conflict', 'replication-lag'],
  }),

  'conflict.cancel': (e) => ({
    title: `Запрос ${str(e.payload.client)} на ${str(e.payload.node)} отменён: конфликт с восстановлением`,
    body: [
      `Запрос проработал ${ticks(num(e.payload.ran))} из ${num(e.payload.planned)} и получил «canceling statement due to conflict with recovery». Реплика ждала ${ticks(num(e.payload.waited))} и выбрала WAL, а не запрос.`,
      'Три выхода, и у каждого цена: hot_standby_feedback (мусор на ведущем), больше max_standby_streaming_delay или -1 (растёт отставание — плохо, если реплика нужна для переключения) или отдельная реплика для отчётов со своими настройками.',
    ],
    terms: ['recovery-conflict', 'hot-standby-feedback'],
  }),

  'query.start': (e) => ({
    title: `${str(e.payload.client)}: долгий отчёт на ${str(e.payload.node)} (снимок на LSN ${num(e.payload.snapshot)})`,
    body: [
      `Запрос берёт снимок реплики — данные, проигранные до LSN ${num(e.payload.snapshot)}, — и будет читать их ${ticks(num(e.payload.ticks))}. Всё это время ему нужны версии строк, которые на ведущем уже могут считаться мёртвыми.`,
      e.payload.feedback ? 'hot_standby_feedback включён: реплика сообщит ведущему о своём снимке.' : 'hot_standby_feedback выключен: ведущий о запросе ничего не знает.',
    ],
    terms: ['recovery-conflict'],
  }),

  'query.done': (e) => ({
    title: `${str(e.payload.client)}: отчёт на ${str(e.payload.node)} закончился`,
    body: ['Снимок освобождён. Если проигрывание стояло из-за конфликта, оно продолжится, а ведущий с hot_standby_feedback снова сможет чистить.'],
    terms: ['recovery-conflict'],
  }),
}

export function explain(e: ReplEvent): Explanation {
  return EXPLAIN[e.type](e)
}
