package main

import "sync"

// OnceMap лениво создаёт значение для каждого ключа ровно один раз —
// например, клиент к каждому шарду базы или скомпилированный шаблон на
// каждое имя. Нулевое значение готово к работе.
type OnceMap[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]func() (V, error)
}

// Get возвращает значение для key. При первом обращении к ключу
// вызывает init(key); одновременные Get того же ключа ждут этот вызов,
// а init для одного ключа никогда не вызывается дважды — даже если он
// вернул ошибку (её получают все последующие Get этого ключа).
// Медленный init одного ключа не задерживает Get других ключей.
func (m *OnceMap[K, V]) Get(key K, init func(K) (V, error)) (V, error) {
	m.mu.Lock()
	if m.m == nil {
		m.m = make(map[K]func() (V, error))
	}
	get, ok := m.m[key]
	if !ok {
		// Под мьютексом только регистрируем «обещание» — дешёвый OnceValues.
		get = sync.OnceValues(func() (V, error) { return init(key) })
		m.m[key] = get
	}
	m.mu.Unlock()
	// Сам init выполняется вне общего мьютекса: ждут только
	// вызывающие с тем же ключом, на своём OnceValues.
	return get()
}

// Len — сколько ключей уже инициализировано или инициализируется.
func (m *OnceMap[K, V]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.m)
}
