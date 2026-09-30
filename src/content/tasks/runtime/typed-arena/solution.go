package main

// Arena раздаёт *T из заранее выделенных блоков по chunkSize элементов —
// как арена на один запрос: тысячи мелких объектов без тысяч аллокаций,
// а в конце запроса Reset освобождает всё разом.
//
// Контракт:
//   - New возвращает указатель на обнулённый T, не пересекающийся с
//     другими выданными; указатель действителен до Reset;
//   - Reset делает все выданные указатели недействительными, но память
//     оставляет арене: повторить ту же нагрузку после Reset можно без
//     единой аллокации;
//   - после Reset арена не удерживает объекты, на которые ссылались
//     выданные T (сборщик мусора может их собрать).
//
// Не обязана быть потокобезопасной.
type Arena[T any] struct {
	chunks    [][]T // каждый блок — ровно chunkSize элементов, никогда не растёт
	cur       int   // индекс текущего блока
	used      int   // занято элементов в текущем блоке
	chunkSize int
	n         int
}

// NewArena создаёт арену; chunkSize >= 1.
func NewArena[T any](chunkSize int) *Arena[T] {
	return &Arena[T]{chunkSize: chunkSize}
}

func (a *Arena[T]) New() *T {
	if len(a.chunks) == 0 || a.used == a.chunkSize {
		if len(a.chunks) > 0 {
			a.cur++
		}
		// Новый блок выделяем, только если после Reset не осталось старого.
		// Блоки не растут через append — иначе адреса выданных T
		// указывали бы в брошенный старый массив.
		if a.cur == len(a.chunks) {
			a.chunks = append(a.chunks, make([]T, a.chunkSize))
		}
		a.used = 0
	}
	p := &a.chunks[a.cur][a.used]
	a.used++
	a.n++
	return p
}

func (a *Arena[T]) Reset() {
	// Обнуляем сразу, а не при повторной выдаче: иначе до следующего New
	// блоки держат указатели на чужие объекты, и GC их не соберёт.
	for i := 0; i < len(a.chunks) && i <= a.cur; i++ {
		clear(a.chunks[i])
	}
	a.cur, a.used, a.n = 0, 0, 0
}

// Len — сколько объектов выдано с последнего Reset.
func (a *Arena[T]) Len() int { return a.n }
