package main

// MultiMap хранит для каждого ключа набор значений без повторов,
// в порядке первого добавления.
//
//   - Add добавляет v к ключу k; false, если такая пара уже есть.
//   - Remove убирает пару; false, если её не было. Ключ, у которого не
//     осталось значений, исчезает совсем.
//   - RemoveAll убирает ключ целиком и возвращает его значения.
//   - Get возвращает значения ключа — копию: изменения результата
//     не должны влиять на мапу. Для отсутствующего ключа — пустой результат.
//   - Len — число ключей, Size — число пар.
type MultiMap[K, V comparable] struct {
	// ваши поля
}

func NewMultiMap[K, V comparable]() *MultiMap[K, V] {
	return &MultiMap[K, V]{}
}

func (m *MultiMap[K, V]) Add(k K, v V) bool {
	// ваш код
	return false
}

func (m *MultiMap[K, V]) Remove(k K, v V) bool {
	// ваш код
	return false
}

func (m *MultiMap[K, V]) RemoveAll(k K) []V {
	// ваш код
	return nil
}

func (m *MultiMap[K, V]) Get(k K) []V {
	// ваш код
	return nil
}

func (m *MultiMap[K, V]) Has(k K) bool {
	// ваш код
	return false
}

func (m *MultiMap[K, V]) Len() int {
	// ваш код
	return 0
}

func (m *MultiMap[K, V]) Size() int {
	// ваш код
	return 0
}
