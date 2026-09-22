# Go-piece

Понятные лекции по технологиям с интерактивными стендами. Курсы: рантайм Go — планировщик (G, M, P), сборщик мусора, память и каналы; брокеры сообщений — Kafka от `send()` до коммита оффсета.

Сайт статический: Astro + MDX + React-острова + Tailwind, поиск — Pagefind. Бэкенда нет.

## Запуск

```sh
pnpm install
pnpm dev        # http://localhost:4321
pnpm test       # unit-тесты движка и разборов
pnpm build      # сборка в dist/ + индекс поиска
pnpm preview    # посмотреть собранный сайт (поиск работает только здесь)
pnpm demo skew-and-stealing   # текстовый прогон сценария планировщика
pnpm demo:gc write-barrier    # то же для сборщика мусора
pnpm demo:mem contention      # и для аллокатора
pnpm demo:chan buffer         # и для каналов
pnpm demo:kafka journey 80 1  # и для Kafka; третий аргумент — путь сообщения m1
```

Нужен Node 20+ (в `mise.toml` закреплены Node 24 и pnpm 10).

## Устройство

```
src/
  content.config.ts            схемы коллекций lectures и glossary
  content/
    lectures/<курс>/<тема>.mdx лекции; id = путь, курс = первый сегмент
    glossary/<id>.mdx          справочник: один термин — один файл
  components/
    content/                   Term, Predict, Callout — доступны в лекциях без import
    stands/gmp/                стенд «Планировщик Go»
      engine/                  модель планировщика: чистый TS, без DOM, с тестами
      explain/                 разбор событий: текст привязан к ТИПУ события
      ui/                      React-остров: проекция снимка + лента событий
      GmpStand.astro           обёртка для MDX: подтягивает термины, проверяет сценарий
    stands/gc/                 стенд «Сборщик мусора» — то же устройство
      engine/                  трёхцветная разметка, барьер записи, пейсер, подметание
    stands/mem/                стенд «Память»
      engine/                  классы размеров, mcache/mcentral/mheap, стеки горутин
    stands/chan/               стенд «Каналы»
      engine/                  hchan: кольцевой буфер, sendq и recvq, select, закрытие
    stands/kafka/              стенд «Kafka»
      engine/                  продюсер с пакетами, лидер и фолловеры, ISR, HW, группы, сбои
      explain/                 разбор событий и описание каждого шага пути сообщения
    diagrams/kafka/            статичные схемы для лекции (SVG на токенах темы)
  data/courses.ts              курсы и анонсы будущих тем
  pages/                       лекции, справочник, лаборатория, поиск
```

Стенды независимы друг от друга: у каждого свой движок, свои разборы и свой остров. Повторяющиеся части UI (лента, перемотка, ручки) дублируются сознательно — так стенд можно менять, не оглядываясь на соседний.

Главное правило — **движок отделён от отрисовки**. `Simulation` прогоняет сценарий по тикам и хранит полный снимок мира на каждом тике. UI только двигает курсор по этой истории. Отсюда пауза, шаг назад, перемотка и ссылка на прогон в адресе. Объяснения привязаны к типам событий, а не к кадрам.

## Как добавить…

**Термин.** Файл `src/content/glossary/<id>.mdx` с полями `title`, `short` (одна строка для подсказки), `category`, `related`. В лекции: `<Term id="<id>" />` или `<Term id="<id>">своя подпись</Term>`. Опечатка в id роняет сборку.

**Лекцию.** Файл `src/content/lectures/<курс>/<тема>.mdx` с `title`, `description`, `order`. В тексте доступны `<Term>`, `<Predict>`, `<Callout>`, `<GmpStand scenario="…" />`, `<GcStand scenario="…" />`, `<MemStand scenario="…" />`, `<ChanStand scenario="…" />` и `<KafkaStand scenario="…" />`. Страница, навигация, оглавление и список терминов появятся сами. Из `planned` в `src/data/courses.ts` уберите анонс этой темы.

**Сценарий стенда.** Объект сценария в `engine/scenarios.ts` нужного стенда и в массиве `SCENARIOS` / `GC_SCENARIOS` / `MEM_SCENARIOS` / `CHAN_SCENARIOS` / `KAFKA_SCENARIOS`. Поле `claim` обязательно: одна мысль, которую сценарий доказывает. Утверждение из `claim` стоит закрепить тестом в `engine.test.ts`.

**Новый тип события.** Добавить в `EventType` (или `GcEventType`, `MemEventType`, `ChanEventType`, `KafkaEventType`) и `EVENT_IMPORTANCE`, затем разбор в `explain/events.ts`. Без разбора TypeScript сборку не пропустит, а тест проверит, что термины из разбора есть в справочнике.

## Деплой

Vercel: импортировать репозиторий, настройки подхватятся из `vercel.json`. Каждый push в `main` публикует сайт, каждая ветка получает превью-ссылку.

Если адрес будет не `go-piece.vercel.app`, поправьте `site` в `astro.config.mjs`.
