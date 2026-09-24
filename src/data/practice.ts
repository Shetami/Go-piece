/**
 * Справочник раздела «Практика»: в каком виде спрашивают и о чём спрашивают.
 *
 * Типы задач названы по формату вопроса на собеседовании, а не по теме: одна и
 * та же тема спрашивается четырьмя разными способами, и готовиться к ним надо
 * по-разному. У каждого типа своя механика проверки — она описана в `check`
 * и реализована в src/pages/api/check.ts.
 */

/** Как спрашивают. Порядок здесь — порядок фильтров на странице списка. */
export const TASK_KINDS = [
  {
    id: 'output',
    title: 'Что выведет?',
    /** Заголовок над полем ответа на странице задачи. */
    prompt: 'Что напечатает эта программа?',
    promptSql: 'Что вернёт этот запрос?',
    blurb:
      'Код рабочий, вопрос в семантике языка. Вы пишете ожидаемый вывод (для SQL — строки результата), потом код исполняется по-настоящему и ответы сверяются.',
    check: 'Ваш ответ сверяется с настоящим выводом программы.',
  },
  {
    id: 'bug',
    title: 'Найди баг',
    prompt: 'В какой строке ошибка?',
    promptSql: 'В какой строке ошибка?',
    blurb:
      'Код компилируется и что-то делает, но делает не то. Нужно ткнуть в строку, где сломано, — как на собеседовании пальцем в экран.',
    check: 'Вы отмечаете строку в редакторе, проверка говорит, та ли это строка.',
  },
  {
    id: 'fix',
    title: 'Почини',
    prompt: 'Почините код так, чтобы он вёл себя правильно',
    promptSql: 'Почините запрос так, чтобы он возвращал правильный результат',
    blurb:
      'Программа падает, зависает или печатает не то, запрос возвращает не те строки. Правьте код в редакторе, пока результат не совпадёт с ожидаемым.',
    check: 'Ваш код запускается, вывод сверяется с эталонным.',
  },
  {
    id: 'implement',
    title: 'Реализуй',
    prompt: 'Допишите реализацию',
    promptSql: 'Напишите запрос',
    blurb:
      'Есть заготовка функции и её контракт — или схема базы и вопрос к ней. Надо написать тело функции или запрос; проверяют скрытые тесты или сверка с эталонным результатом.',
    check: 'К вашему коду дописываются скрытые тесты и запускаются через go test.',
  },
] as const

export type TaskKindId = (typeof TASK_KINDS)[number]['id']

export const TASK_KIND_IDS = TASK_KINDS.map((k) => k.id) as unknown as [TaskKindId, ...TaskKindId[]]

export function taskKind(id: TaskKindId) {
  const kind = TASK_KINDS.find((k) => k.id === id)
  if (!kind) throw new Error(`неизвестный тип задачи: ${id}`)
  return kind
}

/**
 * На чём задача. От языка зависит почти всё: какие файлы лежат в папке задачи
 * (`.go` или `.sql`), чем исполняется код (песочница go.dev или PostgreSQL в
 * браузере) и какой остров рисует страницу. Язык задаёт тема, а не задача:
 * так его невозможно перепутать в отдельном фронтматтере.
 */
export const TASK_LANGS = [
  { id: 'go', title: 'Go' },
  { id: 'sql', title: 'SQL' },
] as const

export type TaskLangId = (typeof TASK_LANGS)[number]['id']

/**
 * О чём спрашивают. Это классика собеседований по Go и SQL, а не оглавление
 * курсов сайта: пересечения с лекциями есть, но тем тут заведомо больше. Задача
 * может сослаться на лекцию или термин сама — полем `related` во фронтматтере.
 */
export const TASK_TOPICS = [
  { lang: 'go', id: 'slices', title: 'Слайсы и массивы', blurb: 'Длина и вместимость, общий массив под двумя слайсами, append и перевыделение.' },
  { lang: 'go', id: 'maps', title: 'Мапы', blurb: 'Нулевое значение, порядок обхода, отсутствующий ключ, мапа под конкурентным доступом.' },
  { lang: 'go', id: 'strings', title: 'Строки и руны', blurb: 'Байты против рун, обход range, склейка и Builder.' },
  { lang: 'go', id: 'interfaces', title: 'Интерфейсы', blurb: 'Пара (тип, значение), nil-интерфейс с ненулевым типом, приведение типа, методы на значении и на указателе.' },
  { lang: 'go', id: 'defer-panic', title: 'defer, panic, recover', blurb: 'Порядок вызовов, момент вычисления аргументов, именованный результат, перехват паники.' },
  { lang: 'go', id: 'closures', title: 'Замыкания', blurb: 'Что именно захватывает функция, переменная цикла, отложенный вызов внутри итерации.' },
  { lang: 'go', id: 'errors', title: 'Ошибки', blurb: 'Обёртки и %w, errors.Is и errors.As, своя ошибка как значение и как указатель.' },
  { lang: 'go', id: 'concurrency', title: 'Конкурентность', blurb: 'Горутины, пулы воркеров, разветвление и сборка, отмена и утечки.' },
  { lang: 'go', id: 'channels', title: 'Каналы и select', blurb: 'Буфер и его отсутствие, закрытие, чтение из закрытого, nil-канал, select с default.' },
  { lang: 'go', id: 'sync', title: 'sync и гонки', blurb: 'WaitGroup, Mutex, Once, атомарные операции и цена копирования примитивов синхронизации.' },
  { lang: 'go', id: 'context', title: 'context', blurb: 'Отмена по дереву, дедлайн, передача значений и что происходит с горутинами после Done.' },
  { lang: 'go', id: 'generics', title: 'Дженерики', blurb: 'Параметры типа, ограничения, вывод типа и во что это обходится.' },
  { lang: 'go', id: 'runtime', title: 'Рантайм', blurb: 'Планировщик, сборщик мусора, escape-анализ — то, что разобрано в лекциях курса о рантайме.' },
  { lang: 'sql', id: 'sql-basics', title: 'SQL: выборка и NULL', blurb: 'WHERE и трёхзначная логика, сравнение с NULL, приоритет AND и OR, сортировка, деление целых.' },
  { lang: 'sql', id: 'sql-joins', title: 'SQL: соединения', blurb: 'INNER и LEFT JOIN, условие в ON против WHERE, размножение строк, анти-соединение, самосоединение.' },
  { lang: 'sql', id: 'sql-grouping', title: 'SQL: группировка', blurb: 'GROUP BY и HAVING, агрегаты и NULL, условная агрегация, дыры в рядах по датам.' },
  { lang: 'sql', id: 'sql-subqueries', title: 'SQL: подзапросы и CTE', blurb: 'IN и NOT IN, EXISTS, коррелированные подзапросы, рекурсивные CTE.' },
  { lang: 'sql', id: 'sql-window', title: 'SQL: оконные функции', blurb: 'ROW_NUMBER, RANK и DENSE_RANK, нарастающий итог, рамка окна, LAG, острова и промежутки.' },
] as const

export type TaskTopicId = (typeof TASK_TOPICS)[number]['id']

export const TASK_TOPIC_IDS = TASK_TOPICS.map((t) => t.id) as unknown as [TaskTopicId, ...TaskTopicId[]]

export function taskLang(topic: TaskTopicId): TaskLangId {
  return taskTopic(topic).lang
}

export function taskTopic(id: TaskTopicId) {
  const topic = TASK_TOPICS.find((t) => t.id === id)
  if (!topic) throw new Error(`неизвестная тема задачи: ${id}`)
  return topic
}

/** Насколько трудно. Отдельно от типа: «что выведет» бывает и лёгким, и злым. */
export const TASK_LEVELS = [
  { id: 'easy', title: 'Разминка' },
  { id: 'medium', title: 'Обычная' },
  { id: 'hard', title: 'Злая' },
] as const

export type TaskLevelId = (typeof TASK_LEVELS)[number]['id']

export const TASK_LEVEL_IDS = TASK_LEVELS.map((l) => l.id) as unknown as [TaskLevelId, ...TaskLevelId[]]

export function taskLevel(id: TaskLevelId) {
  const level = TASK_LEVELS.find((l) => l.id === id)
  if (!level) throw new Error(`неизвестная сложность: ${id}`)
  return level
}
