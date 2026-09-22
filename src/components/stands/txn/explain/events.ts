import type { Isolation, TxnEvent, TxnEventType } from '../engine/types.ts'

/**
 * Разбор событий стенда.
 *
 * Объяснения привязаны к ТИПУ события, а не к кадру: какой бы сценарий ни
 * крутился, клик по «lock.wait» откроет один и тот же разбор, подставив в него
 * конкретные транзакции, строки и числа из payload.
 */

export interface Explanation {
  title: string
  body: string[]
  /** Что делает настоящий PostgreSQL — функции из src/backend. */
  pg?: string
  /** Где модель расходится с реальностью, если это важно для понимания. */
  model?: string
  terms: string[]
}

const num = (v: unknown) => (typeof v === 'number' ? v : Number(v ?? 0))
const str = (v: unknown) => (typeof v === 'string' ? v : String(v ?? ''))
const txn = (e: TxnEvent) => str(e.payload.txn)

function plural(n: number, one: string, few: string, many: string): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few
  return many
}
const ticks = (n: number) => `${n} ${plural(n, 'тик', 'тика', 'тиков')}`
const versions = (n: number) => `${n} ${plural(n, 'версию', 'версии', 'версий')}`
const records = (n: number) => `${n} ${plural(n, 'запись', 'записи', 'записей')}`
const commits = (n: number) => `${n} ${plural(n, 'коммит', 'коммита', 'коммитов')}`
const rows = (n: number) => `${n} ${plural(n, 'строку', 'строки', 'строк')}`

export const ISOLATION_LABEL: Record<Isolation, string> = {
  'read-uncommitted': 'READ UNCOMMITTED',
  'read-committed': 'READ COMMITTED',
  'repeatable-read': 'REPEATABLE READ',
  serializable: 'SERIALIZABLE',
}
const level = (v: unknown) => ISOLATION_LABEL[v as Isolation] ?? str(v)
const val = (v: unknown) => (v === null || v === undefined ? 'NULL' : str(v))

const ABORT_TITLE: Record<string, string> = {
  serialization: 'ошибка сериализации: строку уже изменили',
  ssi: 'ошибка сериализации: цикл rw-зависимостей',
  deadlock: 'жертва взаимной блокировки',
  unique: 'нарушен первичный ключ',
  crash: 'сервер упал',
}

type Explainer = (e: TxnEvent) => Explanation

const EXPLAIN: Record<TxnEventType, Explainer> = {
  'txn.begin': (e) => ({
    title: num(e.payload.retries) > 0 ? `${txn(e)}: повтор транзакции, попытка ${num(e.payload.retries) + 1}` : `${txn(e)}: BEGIN, ${level(e.payload.isolation)}`,
    body: [
      'Транзакция началась с первым оператором. Номера (xid) у неё пока нет и снимка тоже: PostgreSQL выдаёт xid только тем, кто начал писать, а снимок берёт при первом запросе, а не при BEGIN.',
      e.payload.isolation === 'read-committed'
        ? 'Уровень READ COMMITTED — значение по умолчанию. Каждый оператор получит свой свежий снимок.'
        : e.payload.isolation === 'read-uncommitted'
          ? 'Уровень READ UNCOMMITTED в этой модели учебный: транзакция видит чужие незакоммиченные версии. В PostgreSQL такого не бывает — он молча выполняет этот уровень как READ COMMITTED.'
          : `Уровень ${level(e.payload.isolation)}: снимок, взятый первым оператором, будет действовать до конца транзакции.`,
    ],
    pg: 'StartTransaction → в PGPROC пока только виртуальный номер (vxid)',
    terms: ['transaction', 'isolation-level'],
  }),

  'txn.xid': (e) => ({
    title: `${txn(e)} получила xid ${num(e.payload.xid)}`,
    body: [
      'Транзакция собирается изменить строку, и ей нужен номер: он попадёт в заголовок каждой версии, которую она создаст или удалит, — в поля xmin и xmax.',
      'Номера выдаются по возрастанию из общего счётчика. По ним снимки решают, что видно: всё, что закоммитила транзакция с меньшим номером, завершившаяся до снимка, — видно; всё, что начато позже или ещё идёт, — нет.',
      'Номер 32-битный, поэтому счётчик однажды обойдёт круг. Чтобы старые строки не «уехали в будущее», VACUUM замораживает их — это отдельная обязанность очистки.',
    ],
    pg: 'GetNewTransactionId → ExtendCLOG; номер пишется в MyProc->xid',
    terms: ['xmin-xmax', 'xid-wraparound'],
  }),

  'snap.take': (e) => ({
    title: e.payload.perStatement
      ? `${txn(e)}: снимок для оператора — xmin ${num(e.payload.xmin)}, xmax ${num(e.payload.xmax)}`
      : `${txn(e)}: снимок на всю транзакцию — xmin ${num(e.payload.xmin)}, xmax ${num(e.payload.xmax)}`,
    body: [
      `Снимок — это не копия данных, а три числа: xmin ${num(e.payload.xmin)} (всё, что меньше, уже завершено), xmax ${num(e.payload.xmax)} (всё, что начиная с него, ещё не началось) и список идущих транзакций${str(e.payload.xip) ? `: ${str(e.payload.xip)}` : ' — сейчас он пуст'}. Любую версию строки можно проверить по её xmin и xmax: видна она этому снимку или нет.`,
      e.payload.perStatement
        ? 'На READ COMMITTED снимок берётся заново для каждого оператора. Поэтому следующий запрос той же транзакции увидит всё, что успели закоммитить за это время, — отсюда неповторяемое чтение и фантомы.'
        : 'На REPEATABLE READ и SERIALIZABLE этот снимок будет действовать до конца транзакции. Что бы ни закоммитили другие, эта транзакция видит базу такой, какой та была сейчас.',
    ],
    pg: 'GetSnapshotData — обход ProcArray; GetTransactionSnapshot решает, новый снимок или прежний',
    terms: ['db-snapshot', 'mvcc'],
  }),

  'row.read': (e) => ({
    title: e.payload.forUpdate
      ? `${txn(e)}: SELECT … FOR UPDATE — ${str(e.payload.key)} = ${val(e.payload.value)}`
      : `${txn(e)}: SELECT — ${str(e.payload.key)} = ${val(e.payload.value)}`,
    body: [
      e.payload.value === null
        ? 'Ни одна версия этой строки не видна снимку транзакции: для неё строки нет.'
        : e.payload.own
          ? 'Транзакция видит собственную версию: свои изменения видны сразу, до коммита.'
          : `Среди версий строки видна та, что создала транзакция ${str(e.payload.xmin)}: её коммит попал в снимок, а заменившей её версии для этого снимка ещё нет.`,
      'Чтение ничего не блокирует и ничего не ждёт. Писатель может менять эту строку прямо сейчас — читатель просто продолжит видеть свою версию. Это главное обещание MVCC: читатели не мешают писателям, писатели не мешают читателям.',
      e.payload.forUpdate
        ? 'FOR UPDATE — исключение: строка заблокирована, как будто транзакция её уже меняет. Другие UPDATE и FOR UPDATE на ней встанут в очередь до конца этой транзакции.'
        : 'Прочитанное значение приложение запоминает. Если дальше оно посчитает по нему новое значение и запишет, это «прочитать — изменить — записать», и именно здесь теряются обновления.',
    ],
    pg: 'ExecScan → heap_getnext / index_fetch_heap → HeapTupleSatisfiesMVCC для каждой версии',
    terms: e.payload.forUpdate ? ['row-lock', 'mvcc'] : ['mvcc', 'xmin-xmax'],
  }),

  'row.scan': (e) => ({
    title: `${txn(e)}: SELECT count(*) … WHERE v >= ${e.payload.min === null ? '−∞' : num(e.payload.min)} — ${num(e.payload.count)}`,
    body: [
      `Под условие подошли ${rows(num(e.payload.count))}${str(e.payload.keys) ? `: ${str(e.payload.keys)}` : ''}, сумма ${num(e.payload.sum)}. Это чтение по условию, а не по ключу: результат зависит не только от существующих строк, но и от тех, которых ещё нет.`,
      'Заблокировать строку, которой нет, нельзя. Поэтому блокировки строк от фантомов не спасают — нужен снимок, который новых строк просто не увидит, или предикатные блокировки, как в SERIALIZABLE.',
    ],
    pg: 'SeqScan: каждая версия страница за страницей проходит проверку видимости',
    terms: ['phantom-read', 'db-snapshot'],
  }),

  'row.update': (e) => ({
    title: `${txn(e)}: UPDATE ${str(e.payload.key)} ${num(e.payload.from)} → ${num(e.payload.to)} — новая версия`,
    body: [
      `Строка не перезаписана. Старой версии ${str(e.payload.oldCtid)} проставлен xmax = ${num(e.payload.xid)}, а рядом появилась новая, ${str(e.payload.newCtid)}, с xmin = ${num(e.payload.xid)}. Пока ${txn(e)} не закоммитилась, все остальные видят старую — а сама она уже новую.`,
      'xmax на старой версии работает ещё и как блокировка. Любой, кто захочет изменить эту строку, увидит в xmax идущую транзакцию и встанет ждать её конца. Отдельной таблицы блокировок строк нет — блокировка лежит в самой строке.',
      e.payload.expr === 'read'
        ? `Значение посчитало приложение: прочитанное раньше плюс ${num(e.payload.d)}. База не знает, откуда взялось число, и запишет его, даже если строка с тех пор изменилась.`
        : e.payload.expr === 'delta'
          ? `SET v = v + ${num(e.payload.d)} считает база, по той версии, которую сейчас меняет. Если строку за время ожидания изменили, прибавка ляжет поверх свежего значения.`
          : 'Новое значение задано константой.',
      e.payload.samePage
        ? 'Новая версия легла в ту же страницу — в PostgreSQL это шанс на HOT-обновление, когда индексы трогать не нужно.'
        : 'Новая версия легла в другую страницу. В PostgreSQL это значит новые записи во всех индексах таблицы — даже если индексированные колонки не менялись.',
    ],
    pg: 'ExecUpdate → heap_update: t_xmax старой версии, t_ctid → новая, XLOG_HEAP_UPDATE в WAL',
    terms: ['xmin-xmax', 'row-lock', 'hot-update'],
  }),

  'row.insert': (e) => ({
    title: `${txn(e)}: INSERT ${str(e.payload.key)} = ${num(e.payload.value)}`,
    body: [
      `Новая версия ${str(e.payload.ctid)} с xmin = ${num(e.payload.xid)} и пустым xmax. До коммита её видит только сама ${txn(e)}; для остальных строки как будто нет.`,
      'Если такую строку вставят двое сразу, второй будет ждать конца первого на проверке первичного ключа — и получит ошибку, если первый закоммитится.',
    ],
    pg: 'ExecInsert → heap_insert, затем вставка в индексы и проверка уникальности',
    terms: ['xmin-xmax'],
  }),

  'row.delete': (e) => ({
    title: `${txn(e)}: DELETE ${str(e.payload.key)}`,
    body: [
      `DELETE тоже ничего не удаляет: он ставит xmax = ${num(e.payload.xid)} на видимую версию. После коммита она станет мёртвой для новых снимков, но физически останется на месте, пока её не уберёт VACUUM.`,
    ],
    pg: 'ExecDelete → heap_delete',
    terms: ['xmin-xmax', 'vacuum'],
  }),

  'row.lock': (e) => ({
    title: `${txn(e)} заблокировала ${str(e.payload.key)} (FOR UPDATE)`,
    body: [
      'Строка не изменена, но её xmax теперь указывает на эту транзакцию с пометкой «только блокировка». Для читателей ничего не поменялось, а вот UPDATE, DELETE и другие FOR UPDATE встанут в очередь.',
      'Так «прочитать — изменить — записать» становится безопасным и на READ COMMITTED: между чтением и записью строку никто не изменит. Цена — ожидание для всех, кому нужна та же строка.',
    ],
    pg: 'ExecLockRows → heap_lock_tuple: t_xmax + HEAP_XMAX_LOCK_ONLY, запись в WAL',
    terms: ['row-lock', 'lost-update'],
  }),

  'row.recheck': (e) => ({
    title: `${txn(e)}: строку изменила ${str(e.payload.by)} — перепроверка на свежей версии (${num(e.payload.from)} → ${num(e.payload.to)})`,
    body: [
      `UPDATE нашёл строку по своему снимку, но её уже заменила закоммиченная ${str(e.payload.by)}. На READ COMMITTED PostgreSQL не сдаётся: он переходит по цепочке к самой свежей версии, заново проверяет на ней условие WHERE и меняет уже её.`,
      'Это нарушает «один оператор — один снимок»: остальную таблицу оператор видит по старому снимку, а эту строку — по новому. Зато `SET v = v + 10` не теряет чужих изменений.',
      'Если же новое значение посчитало приложение из прочитанного раньше, перепроверка его не спасёт: база прибавлять ничего не будет, она запишет готовое число — и затрёт то, что только что сделала другая транзакция.',
    ],
    pg: 'ExecUpdate → table_tuple_lock → EvalPlanQual: условие перепроверяется на новейшей версии',
    terms: ['eval-plan-qual', 'row-lock'],
  }),

  'stmt.skip': (e) => ({
    title:
      e.payload.why === 'cond'
        ? `${txn(e)}: проверка не прошла (${str(e.payload.of)} = ${num(e.payload.sum)} < ${num(e.payload.gte)}) — UPDATE не выполняется`
        : e.payload.why === 'deleted'
          ? `${txn(e)}: строку ${str(e.payload.key)} удалили, пока ждали, — UPDATE 0`
          : `${txn(e)}: строки ${str(e.payload.key)} нет — UPDATE 0`,
    body: [
      e.payload.why === 'cond'
        ? 'Приложение проверило условие по прочитанным значениям и решило ничего не менять. Эта проверка — логика приложения, база о ней ничего не знает: она не может гарантировать, что условие останется верным к моменту коммита. Это умеет только SERIALIZABLE.'
        : 'Оператор не нашёл подходящей строки и ничего не изменил. Ошибки нет: UPDATE, затронувший ноль строк, — нормальный результат, и приложению стоит его проверять.',
    ],
    terms: e.payload.why === 'cond' ? ['write-skew'] : ['eval-plan-qual'],
  }),

  'lock.wait': (e) => ({
    title: e.payload.behind
      ? `${txn(e)} встала в очередь к ${str(e.payload.key)} за ${str(e.payload.holder)}`
      : `${txn(e)} ждёт ${str(e.payload.holder)}: строка ${str(e.payload.key)} занята`,
    body: [
      e.payload.behind
        ? `Строка сейчас свободна, но её уже ждёт ${str(e.payload.holder)}. Очередь честная: пришедший позже встаёт в хвост, а не проскакивает вперёд, пока первый просыпается.`
        : `В xmax нужной версии записан xid ${num(e.payload.xid)} — транзакция ${str(e.payload.holder)}, и она ещё идёт${e.payload.lockOnly ? ' (строка взята FOR UPDATE)' : ''}. Менять строку, пока та не закончится, нельзя: неизвестно, закоммитит она свою версию или откатит.`,
      'Ждут не строку, а транзакцию целиком: блокировка снимется только на её COMMIT или ROLLBACK. Поэтому длинная транзакция, успевшая что-то изменить, держит очередь за собой до самого конца — даже если остальное время она просто думает.',
      'Читатели этой очереди не видят и в неё не встают: SELECT без FOR UPDATE спокойно читает старую версию.',
    ],
    pg: 'heap_update → XactLockTableWait(xmax): ожидание блокировки на xid владельца; очередь — через блокировку кортежа',
    terms: ['row-lock', 'db-deadlock'],
  }),

  'lock.granted': (e) => ({
    title: `${txn(e)} дождалась ${str(e.payload.key)} (ждала ${ticks(num(e.payload.waited))})`,
    body: [
      'Транзакция, державшая строку, закончилась, и оператор продолжил с того места, где остановился. Дальше решает уровень изоляции: READ COMMITTED перепроверит свежую версию, REPEATABLE READ получит ошибку, если та транзакция закоммитила изменение, и продолжит спокойно, если откатила.',
    ],
    pg: 'XactLockTableWait возвращается → heap_update повторяет проверку t_xmax',
    terms: ['row-lock'],
  }),

  'commit.wait': (e) => ({
    title: `${txn(e)}: COMMIT — ждёт, пока WAL дойдёт до диска`,
    body: [
      `Запись о коммите получила номер ${num(e.payload.lsn)} в WAL, но на диске пока всё только до ${num(e.payload.flushed)}. С synchronous_commit = on клиент не получит ответа, пока fsync не подтвердит, что запись на диске.`,
      e.payload.flushing !== null
        ? `Диск сейчас занят сбросом до ${num(e.payload.flushing)}. Этот коммит поедет следующим сбросом — вместе со всеми, кто успеет подойти. Это групповой коммит: один fsync на много транзакций.`
        : 'Сброс начнётся в этом же тике. Все коммиты, которые успеют встать в очередь, уедут одним fsync.',
      'Пока идёт ожидание, транзакция для всех ещё «идёт»: её изменений никто не видит, строки остаются заблокированы.',
    ],
    pg: 'RecordTransactionCommit → XLogFlush(lsn) — ожидание WALWriteLock, групповой сброс',
    terms: ['wal', 'synchronous-commit', 'group-commit'],
  }),

  'txn.commit': (e) => ({
    title: e.payload.readOnly
      ? `${txn(e)}: COMMIT — только читала, записывать нечего`
      : e.payload.sync
        ? `${txn(e)}: COMMIT подтверждён — запись на диске`
        : `${txn(e)}: COMMIT подтверждён сразу — WAL ещё в памяти`,
    body: [
      e.payload.readOnly
        ? 'Транзакция ничего не меняла, у неё даже нет xid. Её коммит — просто освобождение снимка: ни WAL, ни fsync.'
        : `В pg_xact для xid ${num(e.payload.xid)} записано «закоммичена». С этого момента все её версии видны новым снимкам — одним движением, строки перебирать не нужно.`,
      e.payload.readOnly
        ? `Транзакция длилась ${ticks(num(e.payload.took))}. Пока она шла, её снимок удерживал горизонт очистки.`
        : e.payload.sync
          ? `Ожидание диска заняло ${ticks(num(e.payload.waited))}. Зато клиент знает твёрдо: что бы ни случилось с сервером дальше, эта транзакция переживёт перезапуск.`
          : 'synchronous_commit = off: клиент получил «COMMIT» раньше, чем запись о коммите легла на диск. Если сервер упадёт в ближайшие мгновения, транзакция исчезнет — хотя клиенту сказали, что всё хорошо. Данные при этом не портятся: база восстановится в согласованное, но чуть более раннее состояние.',
    ],
    pg: 'CommitTransaction → RecordTransactionCommit → TransactionIdCommitTree (pg_xact) → ProcArrayEndTransaction',
    terms: e.payload.readOnly ? ['transaction'] : ['pg-xact', 'wal', 'synchronous-commit'],
  }),

  'txn.rollback': (e) => ({
    title: `${txn(e)}: ROLLBACK`,
    body: [
      'Откат в PostgreSQL мгновенный при любом объёме изменений: в pg_xact транзакция помечается как откаченная, и все её версии разом становятся невидимыми. Ничего не нужно возвращать на место — старые версии и так никуда не делись.',
      e.payload.wrote ? `Версии, которые она успела создать (${str(e.payload.wrote)}), теперь мёртвые и ждут VACUUM.` : 'Она ничего не успела изменить.',
    ],
    pg: 'AbortTransaction → RecordTransactionAbort → TransactionIdAbortTree',
    terms: ['pg-xact', 'vacuum'],
  }),

  'txn.abort': (e) => ({
    title: `${txn(e)} прервана: ${ABORT_TITLE[str(e.payload.reason)] ?? str(e.payload.reason)}`,
    body: [
      e.payload.reason === 'serialization'
        ? `На ${level(e.payload.isolation)} транзакция не может изменить строку ${str(e.payload.key)}: её изменила ${str(e.payload.by)} уже после снимка. Перепроверить свежую версию, как READ COMMITTED, нельзя — это значило бы увидеть то, чего снимок не видит. Остаётся одно: ошибка.`
        : e.payload.reason === 'ssi'
          ? `Транзакция не увидела изменений ${str(e.payload.out)}, а ${str(e.payload.in)} не увидела её собственных. Две такие зависимости подряд — признак того, что никакой последовательный порядок не даст такого же результата. ${str(e.payload.out)} уже закоммичена, отменять её поздно, поэтому ошибку получает эта.`
          : e.payload.reason === 'deadlock'
            ? 'База нашла цикл ожиданий и прервала одну из транзакций, чтобы остальные могли продолжить. Кого именно — решает то, чей таймер deadlock_timeout сработал первым.'
            : e.payload.reason === 'unique'
              ? `Строка с ключом ${str(e.payload.key)} уже есть. Вставка невозможна.`
              : 'Сервер упал посреди транзакции. Клиент получил обрыв соединения, а всё, что транзакция успела сделать, откатится при восстановлении.',
      e.payload.message ? `Клиент получил: ERROR: ${str(e.payload.message)}.` : '',
      e.payload.reason === 'serialization' || e.payload.reason === 'ssi' || e.payload.reason === 'deadlock'
        ? e.payload.willRetry
          ? 'Это не авария, а нормальная работа строгих уровней. Клиент повторит транзакцию целиком — с новым снимком она увидит свежие данные.'
          : 'Это не авария, а нормальная работа строгих уровней. Правильный ответ приложения — повторить транзакцию целиком (SQLSTATE 40001 или 40P01). Включите «повторять при ошибке», чтобы увидеть, чем кончится повтор.'
        : '',
    ].filter(Boolean),
    pg:
      e.payload.reason === 'ssi'
        ? 'PreCommit_CheckForSerializationFailure / OnConflict_CheckForSerializationFailure (predicate.c)'
        : e.payload.reason === 'deadlock'
          ? 'CheckDeadLock → DeadLockCheck (deadlock.c)'
          : e.payload.reason === 'serialization'
            ? 'heap_update вернул TM_Updated → ereport(ERROR, ERRCODE_T_R_SERIALIZATION_FAILURE)'
            : undefined,
    terms:
      e.payload.reason === 'deadlock'
        ? ['db-deadlock']
        : e.payload.reason === 'ssi'
          ? ['ssi', 'serialization-failure']
          : e.payload.reason === 'crash'
            ? ['crash-recovery']
            : ['serialization-failure'],
  }),

  'txn.retry': (e) => ({
    title: `${str(e.payload.prev)} упала — клиент повторяет её как ${txn(e)}`,
    body: [
      'Повторяется вся транзакция с самого начала, а не упавший оператор: прочитанное в прошлый раз уже неактуально, и решения, принятые по нему, нужно принять заново.',
      'Новая попытка получит новый снимок и увидит всё, что закоммитили за это время. Поэтому она, скорее всего, пройдёт — или корректно откажется делать то, что стало неверным.',
      'В Go это обычный цикл вокруг функции транзакции с проверкой кода ошибки: 40001 — ошибка сериализации, 40P01 — дедлок.',
    ],
    terms: ['serialization-failure'],
  }),

  'ssi.conflict': (e) => ({
    title: `rw-зависимость: ${str(e.payload.reader)} не видит запись ${str(e.payload.writer)}${str(e.payload.key) === '*' ? '' : ` в ${str(e.payload.key)}`}`,
    body: [
      `${str(e.payload.reader)} прочитала ${str(e.payload.key) === '*' ? 'таблицу по условию' : `строку ${str(e.payload.key)}`}, а ${str(e.payload.writer)} её меняет, и снимки у них разные. Значит, в любом последовательном порядке ${str(e.payload.reader)} должна идти раньше ${str(e.payload.writer)}.`,
      'Сама по себе такая зависимость безвредна — это обычная параллельность. Опасна цепочка из двух подряд: A не видит B, а B не видит C. Если при этом C закоммитилась первой, порядок уже не выстроить, и SERIALIZABLE прервёт одну из транзакций.',
      'Для этого PostgreSQL помнит, кто что читал, — SIREAD-блокировки. Они никого не блокируют, они только следят.',
    ],
    pg: 'CheckForSerializableConflictOut / CheckForSerializableConflictIn → FlagRWConflict (predicate.c)',
    terms: ['ssi'],
  }),

  'deadlock.found': (e) => ({
    title: `Взаимная блокировка: ${str(e.payload.cycle)}`,
    body: [
      `Цикл из ${num(e.payload.size)} транзакций: каждая держит строку, которую ждёт следующая. Сами они из него не выйдут никогда.`,
      `${txn(e)} прождала ${ticks(num(e.payload.waited))} — это deadlock_timeout. Только после него PostgreSQL строит граф ожиданий: искать цикл при каждом ожидании было бы слишком дорого, а большинство ожиданий кончаются сами.`,
      'Лечится порядком: если все транзакции берут строки в одном и том же порядке — например, по возрастанию id счёта, — цикла не может быть в принципе.',
    ],
    pg: 'ProcSleep → CheckDeadLock → DeadLockCheck — поиск цикла в графе ожидания',
    terms: ['db-deadlock', 'row-lock'],
  }),

  'anomaly.dirty': (e) => ({
    title: `Грязное чтение: ${txn(e)} видит ${str(e.payload.key)} = ${val(e.payload.value)} до коммита ${str(e.payload.by)}`,
    body: [
      `${str(e.payload.by)} ещё не закоммитилась — может откатиться, а ${txn(e)} уже прочитала её изменение и может принять по нему решение.`,
      'SQL-стандарт разрешает это на READ UNCOMMITTED. PostgreSQL — никогда: на MVCC чужая незакоммиченная версия просто не видна снимку, и READ UNCOMMITTED там работает как READ COMMITTED.',
    ],
    model: 'READ UNCOMMITTED в модели учебный: так вели бы себя базы на блокировках. В PostgreSQL этого события не бывает.',
    terms: ['dirty-read', 'isolation-level'],
  }),

  'anomaly.nonrepeatable': (e) => ({
    title: `Неповторяемое чтение: ${txn(e)} — ${str(e.payload.key)} было ${val(e.payload.before)}, стало ${val(e.payload.after)}`,
    body: [
      'Тот же запрос в той же транзакции вернул другое значение: между двумя чтениями кто-то закоммитил изменение, а новый снимок его увидел.',
      e.payload.isolation === 'read-committed'
        ? 'На READ COMMITTED это нормальное поведение. Если отчёт или расчёт читает одни данные несколько раз и должен видеть их согласованными, нужен REPEATABLE READ — один снимок на всю транзакцию.'
        : 'Здесь это нормальное поведение уровня изоляции.',
    ],
    terms: ['nonrepeatable-read', 'db-snapshot'],
  }),

  'anomaly.phantom': (e) => ({
    title: `Фантом: ${txn(e)} — было ${rows(num(e.payload.before))}, стало ${num(e.payload.after)}`,
    body: [
      'Второй запрос по тому же условию нашёл другое число строк: кто-то вставил или удалил подходящую строку и закоммитил. Каждая отдельная строка при этом прочитана честно.',
      'Стандарт SQL допускает фантомы и на REPEATABLE READ. В PostgreSQL их там нет: снимок один на транзакцию, и новые строки ему не видны. Переключите уровень и проверьте.',
    ],
    terms: ['phantom-read', 'isolation-level'],
  }),

  'anomaly.lost': (e) => ({
    title: `Потерянное обновление: ${txn(e)} записывает ${num(e.payload.wrote)}, затирая изменение ${str(e.payload.by)}`,
    body: [
      `${txn(e)} прочитала ${str(e.payload.key)} = ${num(e.payload.read)} и посчитала новое значение в приложении. Но сейчас в строке уже ${num(e.payload.actual)} — это закоммитила ${str(e.payload.by)}. Запись ${num(e.payload.wrote)} молча выбрасывает её изменение.`,
      'Ни база, ни приложение ошибки не видят. Лечится одним из трёх способов: считать в самой базе (SET v = v + d), заблокировать строку при чтении (SELECT … FOR UPDATE) или поднять уровень до REPEATABLE READ и повторять транзакцию при ошибке.',
    ],
    terms: ['lost-update', 'row-lock'],
  }),

  'invariant.broken': (e) => ({
    title: `Нарушено правило «${str(e.payload.label)}»`,
    body: [
      `Закоммиченные данные сейчас: ${str(e.payload.detail)}.`,
      e.payload.kind === 'min-sum'
        ? 'Каждая транзакция по отдельности проверила условие и поступила правильно. Неверным стал только их общий результат — такое не ловится ни блокировками строк, ни REPEATABLE READ. Нужен SERIALIZABLE или явная блокировка общего ресурса.'
        : 'Клиенты получили «COMMIT» на изменения, которых в базе нет. Ни одной ошибки при этом не было — поэтому такие потери так трудно найти в проде.',
    ],
    terms: e.payload.kind === 'min-sum' ? ['write-skew', 'ssi'] : ['lost-update', 'acid'],
  }),

  'wal.flush': (e) => ({
    title: e.payload.background
      ? `WAL сброшен на диск до ${num(e.payload.upTo)} (фоновый walwriter)`
      : `WAL сброшен до ${num(e.payload.upTo)} — подтверждено ${commits(num(e.payload.commits))}`,
    body: [
      e.payload.background
        ? 'Коммитов, которые ждали бы диска, нет — WAL сбросил фоновый процесс по таймеру wal_writer_delay. С synchronous_commit = off только он и делает записи о коммитах долговечными.'
        : `Один fsync сделал долговечными сразу ${commits(num(e.payload.commits))}: ${str(e.payload.names)}. Все они встали в очередь, пока диск был занят, и уехали вместе. Чем больше параллельных коммитов, тем дешевле каждый.`,
      'WAL пишется последовательно, в конец файла, — поэтому один сброс дешевле, чем разбросанная запись изменённых страниц. Сами страницы данных уйдут на диск позже, на контрольной точке.',
    ],
    pg: 'XLogFlush / XLogBackgroundFlush → issue_xlog_fsync',
    terms: ['wal', 'group-commit'],
  }),

  checkpoint: (e) => ({
    title: `Контрольная точка на LSN ${num(e.payload.lsn)}`,
    body: [
      'Все изменённые страницы из буферного кэша записаны на диск. Восстановлению после падения больше не нужны записи WAL до этой точки — их можно удалить или отправить в архив.',
      `До этой точки восстановлению пришлось бы проиграть ${records(num(e.payload.replayWas))}. Чем реже контрольные точки, тем дольше восстановление; чем чаще — тем больше записи на диск и тем больше полных страниц в WAL.`,
    ],
    pg: 'CheckPointGuts → CheckPointBuffers (запись грязных страниц), запись XLOG_CHECKPOINT_ONLINE',
    model: 'Контрольная точка в модели мгновенная. В PostgreSQL она растянута во времени (checkpoint_completion_target), чтобы не создавать всплеск записи.',
    terms: ['db-checkpoint', 'wal'],
  }),

  'vacuum.run': (e) => ({
    title:
      num(e.payload.removed) > 0
        ? `VACUUM убрал ${versions(num(e.payload.removed))}${num(e.payload.kept) > 0 ? `, ${num(e.payload.kept)} пришлось оставить` : ''}`
        : num(e.payload.kept) > 0
          ? `VACUUM не смог убрать ничего: ${num(e.payload.kept)} мёртвых ещё кому-то видны`
          : 'VACUUM: убирать нечего',
    body: [
      `Горизонт очистки — xid ${num(e.payload.horizon)}${e.payload.holder ? `, его держит ${str(e.payload.holder)}` : ''}. Мёртвая версия, которую заменили раньше горизонта, не видна уже ни одному снимку, и её можно убрать. Всё, что моложе, может понадобиться кому-то из идущих транзакций.`,
      'Место убранных версий не отдаётся операционной системе: оно помечается свободным в карте свободного места, и следующие версии лягут туда. Таблица перестаёт расти, но и не сжимается.',
      `В таблице ${e.payload.pages} ${plural(num(e.payload.pages), 'страница', 'страницы', 'страниц')}.`,
    ],
    pg: 'lazy_scan_heap → heap_page_prune (HeapTupleSatisfiesVacuum по GlobalVisState) → RecordPageWithFreeSpace',
    model: 'Автовакуум в модели приходит по таймеру. В PostgreSQL его запускает число изменённых строк: autovacuum_vacuum_scale_factor (20% таблицы) плюс порог.',
    terms: ['vacuum', 'xmin-horizon'],
  }),

  'vacuum.blocked': (e) => ({
    title: `${str(e.payload.holder)} мешает очистке: ${num(e.payload.kept)} мёртвых ${plural(num(e.payload.kept), 'версия остаётся', 'версии остаются', 'версий остаются')}`,
    body: [
      `${str(e.payload.holder)} (${level(e.payload.isolation)}) идёт уже ${ticks(num(e.payload.age))}. Её ${e.payload.isolation === 'read-committed' ? 'номер' : 'снимок'} определяет горизонт: всё, что умерло после ${num(e.payload.horizon)}, она в теории может захотеть прочитать, и VACUUM обязан это сохранить.`,
      e.payload.idle
        ? 'При этом она ничего не делает — все её запросы уже выполнены, а COMMIT так и не пришёл. В pg_stat_activity такая сессия видна как «idle in transaction». Одна забытая транзакция держит очистку всей базы, а не только своей таблицы.'
        : 'Горизонт один на всю базу: длинный отчёт по одной таблице мешает чистить все остальные.',
      'Защита — idle_in_transaction_session_timeout, короткие транзакции и мониторинг возраста самой старой транзакции (age(backend_xmin) в pg_stat_activity).',
    ],
    pg: 'ComputeXidHorizons — минимум xmin по всем backend в ProcArray',
    terms: ['xmin-horizon', 'table-bloat', 'idle-in-transaction'],
  }),

  'db.crash': (e) => ({
    title: `Сервер упал: на диске WAL до ${num(e.payload.flushed)}, ${records(num(e.payload.lostRecords))} пропали`,
    body: [
      'Всё, что было в памяти, — буферный кэш со страницами данных и хвост WAL, не успевший на диск, — исчезло. Но страница данных не может попасть на диск раньше записи WAL, которая её изменила, поэтому на диске нет ничего, чего нельзя объяснить по WAL.',
      `Идущие транзакции (${num(e.payload.active)}) оборваны. ${e.payload.sync ? 'С synchronous_commit = on ни одна из подтверждённых транзакций не пропадёт: клиент получал «COMMIT» только после fsync.' : 'С synchronous_commit = off часть подтверждённых коммитов могла не успеть на диск — сейчас станет видно, какие.'}`,
      `Восстановление проиграет ${records(num(e.payload.replay))} с последней контрольной точки и займёт ${ticks(num(e.payload.down) - 1)}.`,
    ],
    pg: 'StartupXLOG → восстановление с redo-точки последнего checkpoint',
    terms: ['crash-recovery', 'wal'],
  }),

  'db.recovered': (e) => ({
    title: 'Сервер восстановился',
    body: [
      `Все записи WAL на диске проиграны. Транзакции без записи о коммите помечены откаченными — их версии стали мёртвыми. Данные сейчас: ${str(e.payload.rows)}.`,
      num(e.payload.lost) > 0
        ? `Потеряно ${commits(num(e.payload.lost))}, о которых клиенты получили «COMMIT». База при этом согласованна — просто она вернулась в состояние на несколько мгновений раньше.`
        : 'Ни один подтверждённый клиенту коммит не потерян.',
    ],
    pg: 'StartupXLOG: redo до конца WAL, затем контрольная точка конца восстановления',
    terms: ['crash-recovery'],
  }),

  'commit.lost': (e) => ({
    title: `${txn(e)} потеряна: клиент получил «COMMIT» на тике ${num(e.payload.ackTick)}, а на диск запись не попала`,
    body: [
      `Запись о коммите получила LSN ${num(e.payload.lsn)}, а на диске WAL был только до ${num(e.payload.flushed)}. С synchronous_commit = off клиенту ответили, не дожидаясь fsync, — и падение унесло эту транзакцию${str(e.payload.wrote) ? ` вместе с изменениями ${str(e.payload.wrote)}` : ''}.`,
      'Это осознанный обмен, а не баг: synchronous_commit = off ускоряет коммиты в разы, но окно потери — до трёх wal_writer_delay. Для счётчиков просмотров и логов это приемлемо, для платежей — нет. Параметр можно менять для отдельной транзакции: SET LOCAL synchronous_commit = off.',
    ],
    pg: 'synchronous_commit = off: RecordTransactionCommit не вызывает XLogFlush, только XLogSetAsyncXactLSN',
    terms: ['synchronous-commit', 'wal'],
  }),
}

export function explain(e: TxnEvent): Explanation {
  return EXPLAIN[e.type](e)
}
