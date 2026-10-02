/**
 * Достижения. Чистый модуль: ни DOM, ни localStorage, ни import.meta.glob —
 * его импортируют и трекер на каждой странице, и остров /achievements/, и тесты.
 *
 * Достижение не хранит счётчиков: оно вычисляется из прогресса (что прочитано,
 * решено, пройдено) и каталога (что вообще есть на сайте). Поэтому старый
 * прогресс — решённые до появления достижений задачи и пройденные тесты —
 * засчитывается сам, а добавить достижение значит дописать одну функцию.
 *
 * Названия — из One Piece, условия — про сам сайт: что изучено и решено.
 * Награда в белли ничего не открывает, она складывается в общую сумму
 * на плакате «WANTED» и задаёт ранг.
 */
import type { TaskKindId, TaskLevelId, TaskLangId } from '../data/practice.ts'

/** Что есть на сайте. Собирается на сборке — см. src/lib/catalog.ts. */
export interface Catalog {
  courses: { id: string; title: string; lectures: string[] }[]
  /** Лекции, у которых есть тест «Проверь себя». */
  quizzes: string[]
  tasks: { id: string; topic: string; lang: TaskLangId; kind: TaskKindId; level: TaskLevelId }[]
  /** Стенды по имени папки: gmp, gc, chan… */
  stands: string[]
  /** Сколько терминов в справочнике. */
  terms: number
}

export interface QuizResult {
  best: number
  total: number
}

/** Прогресс в этом браузере — его собирает src/lib/progress.ts. */
export interface Progress {
  /** Лекции, дочитанные до конца. */
  read: ReadonlySet<string>
  quizzes: ReadonlyMap<string, QuizResult>
  solved: ReadonlySet<string>
  /** Стенды, которые хоть раз трогали руками. */
  stands: ReadonlySet<string>
  /** Открытые страницы справочника. */
  terms: ReadonlySet<string>
  /** Дни, когда что-то было сделано, YYYY-MM-DD по местному времени. */
  days: readonly string[]
  /** Больше всего задач, решённых за один день. */
  bestDay: number
  /** Разовые события для секретных достижений: lost, night. */
  flags: ReadonlySet<string>
}

export const GROUPS = [
  { id: 'voyage', title: 'Плавание', blurb: 'Лекции и курсы — острова на пути по Гранд Лайн.' },
  { id: 'trials', title: 'Испытания', blurb: 'Тесты «Проверь себя» в конце лекций.' },
  { id: 'battles', title: 'Сражения', blurb: 'Задачи практики — как на собеседовании.' },
  { id: 'egghead', title: 'Эгхед', blurb: 'Лаборатория стендов и справочник.' },
  { id: 'crew', title: 'Дух команды', blurb: 'Регулярность важнее рывков.' },
  { id: 'secret', title: 'Секреты', blurb: 'Условия откроются, когда достижение будет получено.' },
] as const

export type GroupId = (typeof GROUPS)[number]['id']

export interface Achievement {
  id: string
  group: GroupId
  /** Название из мира One Piece. */
  title: string
  /** Строчка-отсылка под названием. */
  quote: string
  /** Что сделать — без отсылок, прямым текстом. */
  task: string
  /** Награда в белли. */
  bounty: number
  /** Условие скрыто, пока достижение не получено. */
  secret?: boolean
  measure: (p: Progress, c: Catalog) => { have: number; need: number }
}

const M = 1_000_000

function count<T>(items: Iterable<T>, pred: (x: T) => boolean): number {
  let n = 0
  for (const x of items) if (pred(x)) n++
  return n
}

const solvedWhere = (p: Progress, c: Catalog, pred: (t: Catalog['tasks'][number]) => boolean) =>
  count(c.tasks, (t) => pred(t) && p.solved.has(t.id))

const tasksSolved = (need: number) => (p: Progress, c: Catalog) => ({
  have: solvedWhere(p, c, () => true),
  need: Math.min(need, c.tasks.length),
})

const kindSolved = (kind: TaskKindId, need: number) => (p: Progress, c: Catalog) => ({
  have: solvedWhere(p, c, (t) => t.kind === kind),
  need: Math.min(need, count(c.tasks, (t) => t.kind === kind)),
})

const topicCleared = (topic: string) => (p: Progress, c: Catalog) => ({
  have: solvedWhere(p, c, (t) => t.topic === topic),
  need: count(c.tasks, (t) => t.topic === topic),
})

const lecturesRead = (p: Progress, c: Catalog) => count(c.courses.flatMap((x) => x.lectures), (l) => p.read.has(l))
const allLectures = (c: Catalog) => c.courses.reduce((s, x) => s + x.lectures.length, 0)

const courseRead = (course: string) => (p: Progress, c: Catalog) => {
  const lectures = c.courses.find((x) => x.id === course)?.lectures ?? []
  return { have: count(lectures, (l) => p.read.has(l)), need: lectures.length }
}

const quizzesPassed = (p: Progress, c: Catalog) => count(c.quizzes, (q) => p.quizzes.has(q))
const quizzesPerfect = (p: Progress, c: Catalog) =>
  count(c.quizzes, (q) => {
    const r = p.quizzes.get(q)
    return r !== undefined && r.best === r.total
  })

const flag = (name: string) => (p: Progress) => ({ have: p.flags.has(name) ? 1 : 0, need: 1 })

/** Самая длинная серия дней подряд. Дни — YYYY-MM-DD, порядок и повторы не важны. */
export function longestStreak(days: readonly string[]): number {
  const nums = [...new Set(days)]
    .map((d) => {
      const [y, m, dd] = d.split('-').map(Number)
      return Date.UTC(y!, m! - 1, dd!) / 86_400_000
    })
    .sort((a, b) => a - b)
  let best = 0
  let run = 0
  for (let i = 0; i < nums.length; i++) {
    run = i > 0 && nums[i] === nums[i - 1]! + 1 ? run + 1 : 1
    best = Math.max(best, run)
  }
  return best
}

/** Курс → остров. Порядок курсов берётся из каталога, здесь только отсылки. */
const COURSE_FLAVOR: Record<string, { title: string; quote: string }> = {
  'go-runtime': {
    title: 'Пробуждение фрукта',
    quote: 'Сила фрукта раскрывается, когда понимаешь, как он устроен изнутри.',
  },
  concurrency: {
    title: 'Гому-Гому но Гатлинг',
    quote: 'Сотня ударов одновременно — и ни одна горутина не потерялась.',
  },
  brokers: {
    title: 'Новостная чайка',
    quote: 'Чайка-почтальон разносит газету всем, кто подписан, — даже тем, кто спал.',
  },
  databases: {
    title: 'Библиотека Охары',
    quote: 'Знание переживает пожар, только если его успели записать на диск.',
  },
  'system-design': {
    title: 'Верфь Галлей-Ла',
    quote: 'Корабль, который выдержит Гранд Лайн, проектируют до того, как спустить на воду.',
  },
  kubernetes: {
    title: 'Рулевой Джимбэй',
    quote: 'Kubernetes по-гречески — рулевой. У Мугивар он тоже есть.',
  },
  production: {
    title: 'Пережить Маринфорд',
    quote: 'Прод — это война, к которой готовятся заранее.',
  },
  observability: {
    title: 'Вечный лог-пос',
    quote: 'Всегда знать, где ты и куда идёт система.',
  },
  cicd: {
    title: 'Кудэ-Бурст',
    quote: 'Рывок Сани: от коммита до прода одним выстрелом.',
  },
}

/** Достижения курсов строятся по каталогу курсов из src/data/courses.ts. */
export function courseAchievements(courses: readonly { id: string; title: string }[]): Achievement[] {
  return courses.map((course) => {
    const flavor = COURSE_FLAVOR[course.id] ?? { title: course.title, quote: 'Ещё один остров на пути.' }
    return {
      id: `course-${course.id}`,
      group: 'voyage',
      title: flavor.title,
      quote: flavor.quote,
      task: `Дочитать все лекции курса «${course.title}».`,
      bounty: 120 * M,
      measure: courseRead(course.id),
    }
  })
}

const BASE: Achievement[] = [
  // Плавание
  {
    id: 'lecture-first',
    group: 'voyage',
    title: 'Я стану Королём пиратов!',
    quote: 'Любое плавание начинается с того, что кто-то отчаливает от пристани.',
    task: 'Дочитать первую лекцию до конца.',
    bounty: 5 * M,
    measure: (p, c) => ({ have: lecturesRead(p, c), need: 1 }),
  },
  {
    id: 'lectures-10',
    group: 'voyage',
    title: 'Лог-пос настроен',
    quote: 'Стрелка нашла следующий остров — дальше по курсу.',
    task: 'Дочитать 10 лекций.',
    bounty: 30 * M,
    measure: (p, c) => ({ have: lecturesRead(p, c), need: Math.min(10, allLectures(c)) }),
  },
  {
    id: 'lectures-25',
    group: 'voyage',
    title: 'Вход в Гранд Лайн',
    quote: 'Половина команд разворачивается у Обратной горы. Вы — нет.',
    task: 'Дочитать 25 лекций.',
    bounty: 80 * M,
    measure: (p, c) => ({ have: lecturesRead(p, c), need: Math.min(25, allLectures(c)) }),
  },
  // Сюда встают достижения курсов — см. achievementsFor.
  {
    id: 'lectures-all',
    group: 'voyage',
    title: 'Рио-Понеглиф',
    quote: 'Вся история мира, собранная в одном камне.',
    task: 'Дочитать все лекции на сайте.',
    bounty: 500 * M,
    measure: (p, c) => ({ have: lecturesRead(p, c), need: allLectures(c) }),
  },

  // Испытания
  {
    id: 'quiz-first',
    group: 'trials',
    title: 'Испытания жрецов Скайпии',
    quote: 'Первое испытание пройдено. Впереди ещё три жреца и бог.',
    task: 'Пройти первый тест «Проверь себя».',
    bounty: 5 * M,
    measure: (p, c) => ({ have: quizzesPassed(p, c), need: 1 }),
  },
  {
    id: 'quiz-perfect',
    group: 'trials',
    title: 'Ни царапины',
    quote: '«Ничего не случилось». — Зоро',
    task: 'Пройти тест без единой ошибки.',
    bounty: 20 * M,
    measure: (p, c) => ({ have: quizzesPerfect(p, c), need: 1 }),
  },
  {
    id: 'quiz-10',
    group: 'trials',
    title: 'Тропа испытаний',
    quote: 'Каждый остров Гранд Лайн проверяет на прочность по-своему.',
    task: 'Пройти 10 тестов.',
    bounty: 50 * M,
    measure: (p, c) => ({ have: quizzesPassed(p, c), need: Math.min(10, c.quizzes.length) }),
  },
  {
    id: 'quiz-perfect-10',
    group: 'trials',
    title: 'Воля наблюдения',
    quote: 'Видишь правильный ответ раньше, чем дочитал вопрос.',
    task: 'Пройти 10 тестов без единой ошибки.',
    bounty: 150 * M,
    measure: (p, c) => ({ have: quizzesPerfect(p, c), need: Math.min(10, c.quizzes.length) }),
  },
  {
    id: 'quiz-all',
    group: 'trials',
    title: 'Завоеватель Скайпии',
    quote: 'Все испытания позади, колокол Шандоры звонит.',
    task: 'Пройти все тесты.',
    bounty: 300 * M,
    measure: (p, c) => ({ have: quizzesPassed(p, c), need: c.quizzes.length }),
  },
  {
    id: 'quiz-all-perfect',
    group: 'trials',
    title: 'Воля короля',
    quote: 'Редкий дар: ни один вопрос не устоял.',
    task: 'Пройти все тесты без единой ошибки.',
    bounty: 600 * M,
    measure: (p, c) => ({ have: quizzesPerfect(p, c), need: c.quizzes.length }),
  },

  // Сражения
  {
    id: 'task-first',
    group: 'battles',
    title: 'Первая награда',
    quote: '30 000 000 белли за голову Луффи — с этого всё началось.',
    task: 'Решить первую задачу практики.',
    bounty: 30 * M,
    measure: tasksSolved(1),
  },
  {
    id: 'tasks-10',
    group: 'battles',
    title: 'Гир Секонд',
    quote: 'Кровь бежит быстрее — и задачи тоже.',
    task: 'Решить 10 задач.',
    bounty: 50 * M,
    measure: tasksSolved(10),
  },
  {
    id: 'tasks-50',
    group: 'battles',
    title: 'Гир Сёрд',
    quote: 'Кулак размером с великана.',
    task: 'Решить 50 задач.',
    bounty: 120 * M,
    measure: tasksSolved(50),
  },
  {
    id: 'tasks-150',
    group: 'battles',
    title: 'Гир Фоурс',
    quote: 'Воля вооружения поверх резины — броня из опыта.',
    task: 'Решить 150 задач.',
    bounty: 300 * M,
    measure: tasksSolved(150),
  },
  {
    id: 'tasks-300',
    group: 'battles',
    title: 'Гир Фифс',
    quote: 'Свобода: код подчиняется воображению.',
    task: 'Решить 300 задач.',
    bounty: 600 * M,
    measure: tasksSolved(300),
  },
  {
    id: 'tasks-all',
    group: 'battles',
    title: 'Ван Пис существует!',
    quote: 'Сокровище найдено: решена каждая задача.',
    task: 'Решить все задачи практики.',
    bounty: 1500 * M,
    measure: tasksSolved(Infinity),
  },
  {
    id: 'all-kinds',
    group: 'battles',
    title: 'Команда в сборе',
    quote: 'Капитан, мечник, снайпер, доктор — каждый на своём месте.',
    task: 'Решить хотя бы по одной задаче каждого типа: «Что выведет?», «Найди баг», «Почини», «Реализуй».',
    bounty: 40 * M,
    measure: (p, c) => {
      const kinds = new Set(c.tasks.map((t) => t.kind))
      return { have: count(kinds, (k) => solvedWhere(p, c, (t) => t.kind === k) > 0), need: kinds.size }
    },
  },
  {
    id: 'kind-output',
    group: 'battles',
    title: 'Видеть будущее, как Катакури',
    quote: 'Знаешь, что напечатает программа, ещё до запуска.',
    task: 'Решить 10 задач «Что выведет?».',
    bounty: 80 * M,
    measure: kindSolved('output', 10),
  },
  {
    id: 'kind-bug',
    group: 'battles',
    title: 'Король снайперов Согекинг',
    quote: 'Одна строка, один выстрел — точно в баг.',
    task: 'Решить 10 задач «Найди баг».',
    bounty: 80 * M,
    measure: kindSolved('bug', 10),
  },
  {
    id: 'kind-fix',
    group: 'battles',
    title: 'Доктор Чоппер',
    quote: 'Нет такой болезни, которую нельзя вылечить. И такого кода.',
    task: 'Решить 10 задач «Почини».',
    bounty: 80 * M,
    measure: kindSolved('fix', 10),
  },
  {
    id: 'kind-implement',
    group: 'battles',
    title: 'Су-у-упер! Франки',
    quote: 'Не можешь найти нужную деталь — построй её сам.',
    task: 'Решить 25 задач «Реализуй».',
    bounty: 100 * M,
    measure: kindSolved('implement', 25),
  },
  {
    id: 'hard-first',
    group: 'battles',
    title: 'Против Шичибукая',
    quote: 'Первый противник, который бьёт в полную силу.',
    task: 'Решить первую задачу сложности «Злая».',
    bounty: 40 * M,
    measure: (p, c) => ({ have: solvedWhere(p, c, (t) => t.level === 'hard'), need: 1 }),
  },
  {
    id: 'hard-25',
    group: 'battles',
    title: 'Против Ёнко',
    quote: 'Император моря — это уже не разминка.',
    task: 'Решить 25 задач сложности «Злая».',
    bounty: 250 * M,
    measure: (p, c) => ({
      have: solvedWhere(p, c, (t) => t.level === 'hard'),
      need: Math.min(25, count(c.tasks, (t) => t.level === 'hard')),
    }),
  },
  {
    id: 'sql-10',
    group: 'battles',
    title: 'Читать понеглифы, как Робин',
    quote: 'Древние таблицы отвечают тому, кто умеет спросить.',
    task: 'Решить 10 задач на SQL.',
    bounty: 80 * M,
    measure: (p, c) => ({
      have: solvedWhere(p, c, (t) => t.lang === 'sql'),
      need: Math.min(10, count(c.tasks, (t) => t.lang === 'sql')),
    }),
  },
  {
    id: 'topic-clear',
    group: 'battles',
    title: 'Остров покорён',
    quote: 'Флаг Мугивар над первым островом.',
    task: 'Решить все задачи какой-нибудь одной темы.',
    bounty: 100 * M,
    measure: (p, c) => {
      // Ближайшая к закрытию тема — по ней и показываем прогресс.
      const topics = [...new Set(c.tasks.map((t) => t.topic))].map((t) => topicCleared(t)(p, c))
      if (topics.length === 0) return { have: 0, need: 1 }
      return topics.reduce((a, b) => (b.need - b.have < a.need - a.have ? b : a))
    },
  },
  {
    id: 'topic-channels',
    group: 'battles',
    title: 'Дэн-дэн муси',
    quote: 'Улитка-телефон передаёт слова, только если на том конце сняли трубку.',
    task: 'Решить все задачи темы «Каналы и select».',
    bounty: 150 * M,
    measure: topicCleared('channels'),
  },
  {
    id: 'topic-defer-panic',
    group: 'battles',
    title: 'Йоми-Йоми но Ми',
    quote: 'Брук умер и вернулся. recover тоже так умеет.',
    task: 'Решить все задачи темы «defer, panic, recover».',
    bounty: 150 * M,
    measure: topicCleared('defer-panic'),
  },
  {
    id: 'topic-maps',
    group: 'battles',
    title: 'Карты Нами',
    quote: 'Хороший навигатор знает: порядок обхода никто не обещал.',
    task: 'Решить все задачи темы «Мапы».',
    bounty: 150 * M,
    measure: topicCleared('maps'),
  },

  // Эгхед
  {
    id: 'stand-first',
    group: 'egghead',
    title: 'Гость Эгхеда',
    quote: 'Остров будущего: здесь всё можно потрогать руками.',
    task: 'Запустить или покрутить любой интерактивный стенд.',
    bounty: 10 * M,
    measure: (p, c) => ({ have: count(c.stands, (s) => p.stands.has(s)), need: 1 }),
  },
  {
    id: 'stands-all',
    group: 'egghead',
    title: 'Лаборатория Вегапанка',
    quote: 'Гений не верит на слово — он ставит эксперимент.',
    task: 'Покрутить каждый стенд лаборатории.',
    bounty: 200 * M,
    measure: (p, c) => ({ have: count(c.stands, (s) => p.stands.has(s)), need: c.stands.length }),
  },
  {
    id: 'terms-25',
    group: 'egghead',
    title: 'Ученик Охары',
    quote: 'Сначала выучи слова — потом читай камни.',
    task: 'Открыть 25 статей справочника.',
    bounty: 20 * M,
    measure: (p, c) => ({ have: p.terms.size, need: Math.min(25, c.terms) }),
  },
  {
    id: 'terms-100',
    group: 'egghead',
    title: 'Профессор Кловер',
    quote: 'Учёные Охары знали: знание — это сила, которой боятся.',
    task: 'Открыть 100 статей справочника.',
    bounty: 100 * M,
    measure: (p, c) => ({ have: p.terms.size, need: Math.min(100, c.terms) }),
  },

  // Дух команды
  {
    id: 'streak-3',
    group: 'crew',
    title: 'Утренняя тренировка Зоро',
    quote: 'Тысяча отжиманий не делается за один день.',
    task: 'Заниматься три дня подряд: читать, решать или проходить тесты.',
    bounty: 20 * M,
    measure: (p) => ({ have: longestStreak(p.days), need: 3 }),
  },
  {
    id: 'streak-7',
    group: 'crew',
    title: 'Неделя у Михока',
    quote: 'Учиться у сильнейшего — каждый день, без выходных.',
    task: 'Заниматься семь дней подряд.',
    bounty: 100 * M,
    measure: (p) => ({ have: longestStreak(p.days), need: 7 }),
  },
  {
    id: 'days-30',
    group: 'crew',
    title: 'Два года на Русукайне',
    quote: 'Луффи тренировался два года. Вам хватит тридцати дней.',
    task: 'Заниматься в 30 разных дней.',
    bounty: 300 * M,
    measure: (p) => ({ have: new Set(p.days).size, need: 30 }),
  },

  // Секреты
  {
    id: 'lost',
    group: 'secret',
    title: 'Зоро снова заблудился',
    quote: '«Я точно знал, что это направо».',
    task: 'Попасть на несуществующую страницу.',
    bounty: 1 * M,
    secret: true,
    measure: flag('lost'),
  },
  {
    id: 'night',
    group: 'secret',
    title: 'Ночная вахта Брука',
    quote: 'Йо-хо-хо! Ни сна, ни ошибок компиляции.',
    task: 'Решить задачу между полуночью и пятью утра.',
    bounty: 10 * M,
    secret: true,
    measure: flag('night'),
  },
  {
    id: 'feast',
    group: 'secret',
    title: 'Ещё мяса!',
    quote: 'Аппетит капитана не знает дна.',
    task: 'Решить 20 задач за один день.',
    bounty: 50 * M,
    secret: true,
    measure: (p) => ({ have: p.bestDay, need: 20 }),
  },
]

/** Мета-достижение: получить все остальные. Считается последним. */
export const FINALE: Omit<Achievement, 'measure'> = {
  id: 'joy-boy',
  group: 'crew',
  title: 'Джой Бой вернулся',
  quote: 'Барабаны освобождения звучат снова.',
  task: 'Получить все остальные достижения.',
  bounty: 2000 * M,
}

/** Полный список: базовые + курсы (после lectures-25) + финал. */
export function achievementsFor(catalog: { courses: readonly { id: string; title: string }[] }): Achievement[] {
  const at = BASE.findIndex((a) => a.id === 'lectures-all')
  const list = [...BASE.slice(0, at), ...courseAchievements(catalog.courses), ...BASE.slice(at)]
  const others = list.length
  list.push({ ...FINALE, measure: () => ({ have: 0, need: others }) })
  return list
}

export interface Evaluated {
  achievement: Achievement
  have: number
  need: number
  unlocked: boolean
  /** Когда получено, мс — если уже записано в хранилище. */
  unlockedAt?: number
}

/**
 * Посчитать все достижения. `earned` — уже полученные раньше (id → когда):
 * полученное однажды не отбирается, даже если каталог вырос и «все задачи»
 * больше не все.
 */
export function evaluate(progress: Progress, catalog: Catalog, earned: Readonly<Record<string, number>> = {}): Evaluated[] {
  const list = achievementsFor(catalog)
  const out: Evaluated[] = list.map((achievement) => {
    const { have, need } = achievement.measure(progress, catalog)
    const unlockedAt = earned[achievement.id]
    return {
      achievement,
      have: Math.min(have, need),
      need,
      unlocked: unlockedAt !== undefined || (need > 0 && have >= need),
      unlockedAt,
    }
  })
  const finale = out.at(-1)!
  const rest = out.length - 1
  finale.have = count(out.slice(0, -1), (e) => e.unlocked)
  finale.need = rest
  finale.unlocked = finale.unlockedAt !== undefined || finale.have >= rest
  return out
}

/** Ранги по сумме наград. Последний — только за все достижения сразу. */
export const RANKS = [
  { min: 0, title: 'Юнга' },
  { min: 30 * M, title: 'Пират-новичок' },
  { min: 150 * M, title: 'Сверхновая' },
  { min: 500 * M, title: 'Худшее поколение' },
  { min: 1500 * M, title: 'Шичибукай' },
  { min: 4000 * M, title: 'Ёнко' },
] as const

export const KING = 'Король пиратов'

export function rankFor(results: readonly Evaluated[]): { title: string; bounty: number; next?: { title: string; min: number } } {
  const bounty = results.reduce<number>((s, e) => s + (e.unlocked ? e.achievement.bounty : 0), 0)
  if (results.length > 0 && results.every((e) => e.unlocked)) return { title: KING, bounty }
  let idx = 0
  for (let i = 0; i < RANKS.length; i++) if (bounty >= RANKS[i]!.min) idx = i
  const next: { title: string; min: number } = RANKS[idx + 1] ?? {
    title: KING,
    min: results.reduce<number>((s, e) => s + e.achievement.bounty, 0),
  }
  return { title: RANKS[idx]!.title, bounty, next }
}

/** 1500000000 → «1 500 000 000». Пробелы неразрывные: сумма не должна переноситься. */
export function formatBerry(n: number): string {
  return String(Math.round(n)).replace(/\B(?=(\d{3})+(?!\d))/g, ' ')
}
