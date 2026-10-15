import type { MapScenario } from './types.ts'

/**
 * Пресеты. У каждого одна мысль, которую он доказывает, — поле claim.
 * Сценарий — обычный текст команд: его же можно править в лаборатории.
 */

export const small: MapScenario = {
  id: 'small',
  title: 'Маленькая карта и контрольное слово',
  claim:
    'До восьми ключей карта — одна группа без таблицы. В контрольном байте занятого слота лежат 7 бит хэша (H2): поиск сравнивает их сразу у восьми слотов и трогает ключ, только если они совпали.',
  config: {},
  script: `# make без подсказки: группа появится при первой записи
put ann 1
put bob 2
put cid 3
put dan 4
get bob
get eve      # нет такого — ни одного сравнения ключей
put bob 20   # перезапись существующего
del cid
get cid`,
  watchFor: ['small.alloc', 'get.hit', 'get.miss', 'put.update'],
}

export const promote: MapScenario = {
  id: 'promote',
  title: 'Девятый ключ: из группы — в таблицу',
  claim:
    'Девятому ключу в одной группе места нет: карта превращается в таблицу на 16 слотов, и все восемь записей перекладываются заново — в группы, которые выбирает H1.',
  config: {},
  script: `fill 8
get k3
put k9 9     # девятый — рост в таблицу
get k3       # тот же ключ теперь ищется по пробам`,
  watchFor: ['small.promote'],
}

export const probing: MapScenario = {
  id: 'probing',
  title: 'Коллизии и пробы по группам',
  claim:
    'Ключи с одной стартовой группой не помещаются в восемь слотов и уходят дальше по квадратичной последовательности проб. Поиск идёт тем же путём, а отсутствующий ключ ищется до первой группы, где есть empty.',
  config: { hint: 40 },
  script: `# таблица на 64 слота сразу — роста не будет
fill 10
collide 20   # двадцать ключей с одной стартовой группой
get @last:1  # первый из них — в стартовой группе
get @last:-1 # последний — третья группа по пробам
get zzz      # нет такого: до группы с empty`,
  watchFor: ['h2.false', 'get.hit', 'get.miss'],
}

export const tombs: MapScenario = {
  id: 'tombs',
  title: 'Удаление и надгробия',
  claim:
    'Удаление из полной группы оставляет надгробие: пустой слот оборвал бы поиск ключей, которые при вставке проскочили эту группу. Из группы, где уже есть empty, слот освобождается сразу. Надгробие переиспользуется первой же вставкой.',
  config: { hint: 12 },
  script: `# таблица на 16 слотов: две группы
collide 10 0  # группа 0 забита, два ключа ушли в группу 1
del @last:2   # из полной группы — надгробия
del @last:-1  # из группы с empty — просто empty
get @last:-2  # поиск проходит сквозь надгробия
collide 1 0   # новый ключ с той же стартовой группой — на место надгробия`,
  watchFor: ['tomb.set', 'tomb.reuse', 'del.hit'],
}

export const prune: MapScenario = {
  id: 'prune',
  title: 'Прополка надгробий вместо роста',
  claim:
    'Когда запас таблицы кончается, рантайм сначала пробует вычистить надгробия, через которые никто не проходит, — и только если это не освобождает 10% места, растёт. Карта с потоком вставок и удалений не пухнет бесконечно.',
  config: { hint: 12, maxTableSlots: 16 },
  script: `# одна таблица на 16 слотов, расти ей некуда — только делиться
collide 8 0   # группа 0 забита целиком
del @last:4   # четыре надгробия: группа 0 была полной
collide 6 1   # шесть ключей в группу 1 — запас growthLeft кончился
collide 1 1   # места нет: сначала прополка, и она помогает
fill 8        # надгробий больше нет — таблица делится`,
  watchFor: ['tomb.set', 'tomb.prune', 'table.grow', 'table.split'],
}

export const growth: MapScenario = {
  id: 'growth',
  title: 'Рост: удвоение, деление, каталог',
  claim:
    'Таблица заполняется до 7/8 и удваивается, пока не упрётся в потолок; дальше она делится на две по следующему биту хэша, а каталог удваивается, когда бит не хватает. Одна вставка перестраивает не больше одной таблицы.',
  config: { maxTableSlots: 32 },
  script: `fill 9      # маленькая карта → таблица на 16
fill 20     # 16 → 32, потом деление
fill 40     # ещё деления, каталог растёт
get k7`,
  watchFor: ['small.promote', 'table.grow', 'table.split', 'dir.grow'],
}

export const hint: MapScenario = {
  id: 'hint',
  title: 'make с подсказкой размера',
  claim:
    'make(map[K]V, n) сразу строит каталог и таблицы под n записей с запасом 7/8: те же вставки проходят без единого роста и перекладывания. Сравните счётчик «переложено» с пресетом «Рост».',
  config: { hint: 69, maxTableSlots: 32 },
  script: `fill 69
get k7`,
  watchFor: ['map.make', 'table.grow', 'table.split'],
}

export const order: MapScenario = {
  id: 'order',
  title: 'Порядок обхода случаен',
  claim:
    'Итератор начинает со случайной таблицы и случайного слота и идёт по кругу. Два обхода одной и той же карты дают разный порядок — у маленькой карты это циклические сдвиги одного порядка.',
  config: {},
  script: `put a 1
put b 2
put c 3
put d 4
range
next all
range
next all
range
next all`,
  watchFor: ['iter.start', 'iter.done'],
}

export const iterGrow: MapScenario = {
  id: 'iter-grow',
  title: 'Рост карты во время обхода',
  claim:
    'Если таблица выросла под итератором, он продолжает идти по старой: позиция в новой ничего не значит. Ключи берёт из старой таблицы, значения перечитывает в новой, удалённые пропускает, а добавленные только в новую — не видит.',
  config: { maxTableSlots: 64 },
  script: `fill 13
range
next 3
fill 10 x   # добавляем — таблица растёт под итератором
put k1 100  # меняем значение — итератор отдаст новое
del k2      # удаляем — итератор не отдаст
next all`,
  watchFor: ['table.grow', 'iter.reread', 'iter.skip', 'iter.done'],
}

export const noShrink: MapScenario = {
  id: 'no-shrink',
  title: 'Карта не уменьшается',
  claim:
    'delete и clear освобождают слоты, но не таблицы: после удаления всего ёмкость остаётся прежней, а обход пустой карты всё равно проверяет каждый слот. Когда карта опустела, меняется соль хэша.',
  config: { maxTableSlots: 32 },
  script: `fill 40
del k1..k38
range
next all
clear
get k40`,
  watchFor: ['seed.reset', 'iter.done', 'clear'],
}

export const nan: MapScenario = {
  id: 'nan',
  title: 'NaN — ключ, который нельзя найти',
  claim:
    'NaN != NaN, поэтому каждая запись по NaN — новая запись, найти её нельзя, а удалить — только clear. Чтобы такие ключи не сваливались в одну группу, рантайм хэширует NaN случайным числом.',
  config: {},
  script: `put NaN 1
put NaN 2
put NaN 3
get NaN
del NaN
range
next all
clear`,
  watchFor: ['put.new', 'get.miss', 'del.miss', 'iter.done'],
}

export const nilMap: MapScenario = {
  id: 'nil',
  title: 'nil-карта',
  claim:
    'Объявленная, но не созданная карта читается как пустая: чтение, len, range и даже delete работают. А запись паникует — assignment to entry in nil map.',
  config: { nilMap: true },
  script: `get ann
del ann
range
put ann 1`,
  watchFor: ['nil.read', 'nil.write'],
}

export const race: MapScenario = {
  id: 'race',
  title: 'Две записи одновременно',
  claim:
    'Карта не потокобезопасна. Запись переключает флаг writing; вторая горутина, увидевшая его, вызывает fatal error, а не panic: recover её не ловит, defer не выполняются.',
  config: {},
  script: `put ann 1
put bob 2
race write`,
  watchFor: ['race'],
}

export const legacyBuckets: MapScenario = {
  id: 'legacy-buckets',
  title: 'До Go 1.24: корзины и переполнение',
  claim:
    'Старая карта — 2^B корзин по 8 слотов. Корзину выбирают младшие B бит хэша, а 8 старших бит (tophash) отсеивают чужие слоты по одному байту. Полная корзина обрастает цепочкой корзин переполнения — поиск идёт по ней звено за звеном.',
  config: { impl: 'legacy', hint: 20 },
  script: `fill 6
collide 12   # двенадцать ключей в одну корзину → переполнение
get c1
get zzz
del k2`,
  watchFor: ['bucket.overflow', 'get.hit', 'get.miss', 'h2.false'],
}

export const legacyEvac: MapScenario = {
  id: 'legacy-evac',
  title: 'До Go 1.24: постепенная эвакуация',
  claim:
    'При заполнении 6,5 на корзину массив корзин удваивается сразу, а записи переезжают постепенно: каждая следующая запись эвакуирует свою старую корзину и ещё одну по порядку. Пока переезд идёт, чтение смотрит в старую корзину, если та не перенесена.',
  config: { impl: 'legacy', hint: 52 },
  script: `fill 52     # B=3: 8 корзин, в среднем по 6,5 записи
put x1      # 53-й ключ: массив удваивается, переезжают две корзины
get k1      # чтение смотрит в старую корзину, если та не переехала
get k2
range       # обход, начатый посреди переезда
next 4
put x2      # каждая запись переносит ещё корзины
put x3
put x4
put x5      # последняя корзина переехала — старый массив больше не нужен
next all`,
  watchFor: ['grow.start', 'evac', 'grow.done', 'iter.reread'],
}

export const legacySameSize: MapScenario = {
  id: 'legacy-samesize',
  title: 'До Go 1.24: рост того же размера',
  claim:
    'Если вставлять и удалять ключи одной корзины, цепочки переполнения остаются, хотя записей мало. Когда корзин переполнения столько же, сколько основных, карта «растёт» в тот же размер — только чтобы переложить записи плотно.',
  config: { impl: 'legacy', hint: 14 },
  script: `fill 4
collide 9 0
del @last
collide 9 1
del @last
collide 9 2
del @last
collide 9 3
del @last
put trigger 1  # новая запись — рост того же размера
fill 3`,
  watchFor: ['bucket.overflow', 'grow.start', 'grow.done'],
}

export const MAP_SCENARIOS: MapScenario[] = [
  small,
  promote,
  probing,
  tombs,
  prune,
  growth,
  hint,
  order,
  iterGrow,
  noShrink,
  nan,
  nilMap,
  race,
  legacyBuckets,
  legacyEvac,
  legacySameSize,
]

export function mapScenarioById(id: string): MapScenario | undefined {
  return MAP_SCENARIOS.find((s) => s.id === id)
}
