package main

// ShardedMap — потокобезопасная мапа string → V, разбитая на n шардов,
// у каждого свой sync.RWMutex: операции над разными шардами не мешают друг
// другу. Шард выбирается по хэшу ключа. n < 1 считается как 1.
//
//   - Get, Set, Delete, Len — обычные.
//   - Update атомарно меняет значение ключа: fn получает текущее значение
//     (ok=false, если ключа нет) и возвращает новое; keep=false удаляет ключ.
//     Между чтением и записью никто не должен вклиниться. fn не обращается
//     к этой же мапе.
//   - Range вызывает fn для пар, пока fn возвращает true. Внутри fn можно
//     вызывать любые методы этой мапы (Set, Delete, Update…) — без дедлока.
//     Порядок обхода не важен.
type ShardedMap[V any] struct {
	// ваши поля
}

func NewShardedMap[V any](n int) *ShardedMap[V] {
	return &ShardedMap[V]{}
}

func (s *ShardedMap[V]) Get(key string) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

func (s *ShardedMap[V]) Set(key string, v V) {
	// ваш код
}

func (s *ShardedMap[V]) Delete(key string) {
	// ваш код
}

func (s *ShardedMap[V]) Update(key string, fn func(v V, ok bool) (nv V, keep bool)) {
	// ваш код
}

func (s *ShardedMap[V]) Len() int {
	// ваш код
	return 0
}

func (s *ShardedMap[V]) Range(fn func(key string, v V) bool) {
	// ваш код
}
