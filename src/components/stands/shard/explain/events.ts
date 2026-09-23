import type { ShardEvent, ShardEventType } from '../engine/types.ts'

/**
 * Разбор событий стенда. Объяснения привязаны к ТИПУ события: какой бы
 * сценарий ни крутился, клик по «reshard.start» откроет один и тот же разбор,
 * подставив в него конкретные шарды, ключи и числа из payload.
 */

export interface Explanation {
  title: string
  body: string[]
  /** Как это выглядит в реальных хранилищах. */
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
const keysN = (n: number) => `${n} ${plural(n, 'ключ', 'ключа', 'ключей')}`
const shardsN = (n: number) => `${n} ${plural(n, 'шард', 'шарда', 'шардов')}`
const pct = (v: unknown) => `${Math.round(num(v) * 100)}%`

const SCHEME: Record<string, string> = {
  mod: 'остаток от деления хеша на число шардов',
  ring: 'кольцо согласованного хеширования',
  range: 'диапазоны ключей',
}

type Explainer = (e: ShardEvent) => Explanation

const EXPLAIN: Record<ShardEventType, Explainer> = {
  'req.route': (e) => ({
    title: `k${num(e.payload.key)} → ${str(e.payload.shard)}`,
    body: [
      `Раскладка — ${SCHEME[str(e.payload.scheme)] ?? ''}. Клиент вычисляет шард сам, по ключу: никакого похода «куда-нибудь» нет, у каждого ключа ровно один владелец.`,
      'Поэтому ключ шардирования выбирают так, чтобы он был в подавляющем большинстве запросов. Если его нет, запрос придётся задать всем шардам сразу.',
    ],
    terms: ['sharding'],
  }),

  'req.done': (e) => ({
    title: `${str(e.payload.shard)} ответил про k${num(e.payload.key)} за ${ticks(num(e.payload.lat))}`,
    body: [
      e.payload.migrating
        ? 'Ключ ещё переезжал, поэтому запрос стоил вдвое дороже: новый владелец сначала забрал значение у старого.'
        : `${num(e.payload.wait)} ${plural(num(e.payload.wait), 'тик', 'тика', 'тиков')} в очереди шарда, остальное — работа. Точечный запрос к своему шарду ничем не отличается от запроса к обычной базе.`,
    ],
    terms: ['sharding'],
  }),

  'scatter.start': (e) => ({
    title: `Запрос без ключа шардирования: спрашиваем все ${shardsN(num(e.payload.shards))}`,
    body: [
      'Такой запрос не знает, где лежит ответ, — например, ищет по полю, которое не входит в ключ шардирования. Приходится задать вопрос каждому шарду и собрать ответы.',
      `Один запрос пользователя превращается в ${num(e.payload.shards)} запроса к хранилищу. При доле таких запросов ${num(e.payload.share)}% они дают заметную часть всей нагрузки.`,
    ],
    terms: ['scatter-gather'],
  }),

  'scatter.done': (e) => ({
    title: `Веерный запрос собран за ${ticks(num(e.payload.lat))}, последним ответил ${str(e.payload.slowest)}`,
    body: [
      `Ответ готов, только когда ответили все ${shardsN(num(e.payload.shards))}. Значит, задержка такого запроса — это максимум из ${num(e.payload.shards)}, а не среднее.`,
      'Чем больше шардов, тем выше шанс, что хотя бы один окажется занят, и тем хуже веерным запросам. Это тот же эффект усиления хвоста, что и у веерных вызовов между сервисами.',
    ],
    terms: ['scatter-gather', 'tail-latency'],
  }),

  'req.migrating': (e) => ({
    title: `k${num(e.payload.key)} ещё переезжает: ${str(e.payload.from)} → ${str(e.payload.to)}`,
    body: [
      'Ключ уже принадлежит новому шарду, но данные ещё не переехали. Новый владелец сначала забирает значение у старого, и запрос стоит вдвое дороже обычного.',
      `Осталось перевезти ${keysN(num(e.payload.left))}. Так работает большинство хранилищ: решардирование идёт фоном, а запросы к непереехавшим ключам обслуживаются через перенаправление.`,
    ],
    go: 'Redis Cluster отвечает -ASK/-MOVED, Vitess держит маршрут через VTGate, Kafka переносит партиции фоном',
    terms: ['resharding', 'sharding'],
  }),

  'reshard.start': (e) => ({
    title: `${e.payload.kind === 'add' ? 'Добавили шард' : 'Убрали шард'}: ${num(e.payload.from)} → ${num(e.payload.to)}, переезжает ${pct(e.payload.share)} ключей`,
    body: [
      `Раскладка — ${SCHEME[str(e.payload.scheme)] ?? ''}. ${
        e.payload.scheme === 'mod'
          ? 'При остатке от деления смена числа шардов меняет владельца почти у всех ключей: было «% 3», стало «% 4».'
          : e.payload.scheme === 'ring'
            ? `Кольцо отдаёт новому шарду только его долю: ${num(e.payload.vnodes)} виртуальных узлов встают между чужими, и переезжают лишь ключи между ними.`
            : 'Диапазоны переезжают границами: часть ключей каждого шарда уходит соседу.'
      }`,
      `${keysN(num(e.payload.moved))} переедут по ${num(e.payload.rate)} за тик. Всё это время данные есть в двух местах, запросы к ним дороже, а сеть и диски заняты переносом — поэтому решардирование стараются делать редко и заранее.`,
    ],
    terms: e.payload.scheme === 'ring' ? ['consistent-hashing', 'resharding'] : ['resharding', 'consistent-hashing'],
  }),

  'reshard.move': (e) => ({
    title: `k${num(e.payload.key)} переехал на ${str(e.payload.to)}`,
    body: [`Осталось ${keysN(num(e.payload.left))}. После переезда запросы к этому ключу снова стоят как обычные.`],
    terms: ['resharding'],
  }),

  'reshard.done': (e) => ({
    title: `Переезд закончен: ${shardsN(num(e.payload.shards))}, перевезено ${keysN(num(e.payload.moved))}`,
    body: [
      'Все ключи на своих местах, лишних расходов больше нет. Скорость переезда — компромисс: быстрее перевозить значит сильнее мешать боевым запросам.',
      'В жизни на этом дело не заканчивается: старые копии нужно удалить, а клиентов — убедить обновить карту шардов.',
    ],
    terms: ['resharding'],
  }),

  'shard.skew': (e) => ({
    title: `Перекос: на ${str(e.payload.shard)} ${num(e.payload.load)} запросов, у соседей ${str(e.payload.others)}`,
    body: [
      num(e.payload.hotShare) > 0
        ? `Один ключ собирает ${num(e.payload.hotShare)}% запросов, и он целиком лежит на этом шарде. Шардирование делит данные, а нагрузку на один ключ разделить не может: как ни считай хеш, ключ один и шард у него один.`
        : `Раскладка — ${SCHEME[str(e.payload.scheme)] ?? ''}: на этот шард попала горячая часть ключей. Данные разделены поровну, а запросы — нет.`,
      'Лечат это не сменой хеша, а работой с самим ключом: разрезать его на части (`post:42:shard3`), положить перед хранилищем кэш или вынести горячую сущность в отдельное хранилище.',
    ],
    terms: ['hot-partition', 'sharding'],
  }),

  'queue.grow': (e) => ({
    title: `Очередь на ${str(e.payload.shard)}: ${num(e.payload.len)}`,
    body: [
      `Шард владеет ${keysN(num(e.payload.owns))} и сейчас занят целиком. Шард — это обычный сервис со своей очередью: всё, что известно про загрузку и колено задержки, работает и здесь.`,
      num(e.payload.migrating) > 0 ? `Идёт переезд: осталось ${keysN(num(e.payload.migrating))}, и часть запросов стоит вдвое дороже.` : 'Соседние шарды при этом могут простаивать: перебросить им нагрузку нельзя, у них нет этих данных.',
    ],
    terms: ['hot-partition', 'utilization'],
  }),

  'load.change': (e) => ({
    title: num(e.payload.to) > num(e.payload.from) ? `Нагрузка выросла в ${num(e.payload.to)} раза` : 'Нагрузка вернулась к обычной',
    body: [`Теперь приходит ${num(e.payload.rate).toFixed(1)} запроса за тик на ${shardsN(num(e.payload.shards))}.`],
    terms: ['utilization'],
  }),
}

export function explain(e: ShardEvent): Explanation {
  return EXPLAIN[e.type](e)
}
