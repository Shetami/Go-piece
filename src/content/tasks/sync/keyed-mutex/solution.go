package main

import "sync"

type keyLock struct {
	mu   sync.Mutex
	refs int // держатель + ждущие; меняется только под KeyedMutex.mu
}

// KeyedMutex — блокировка по ключу: операции над одним заказом (или
// пользователем, или файлом) идут строго по одной, над разными — параллельно.
// Ключей бесконечно много, поэтому запись о ключе хранится, только пока
// кто-то держит его блокировку или ждёт её. Нулевое значение готово к работе.
type KeyedMutex struct {
	mu    sync.Mutex
	locks map[string]*keyLock
}

// Lock захватывает блокировку ключа и возвращает функцию разблокировки.
// Повторный вызов той же unlock ничего не делает.
func (m *KeyedMutex) Lock(key string) (unlock func()) {
	m.mu.Lock()
	if m.locks == nil {
		m.locks = make(map[string]*keyLock)
	}
	l, ok := m.locks[key]
	if !ok {
		l = &keyLock{}
		m.locks[key] = l
	}
	// Записываемся в ждущие ДО того, как отпустить общий мьютекс:
	// пока refs > 0, запись никто не удалит, и мы не встанем в очередь
	// к мьютексу, которого уже нет в мапе.
	l.refs++
	m.mu.Unlock()

	l.mu.Lock() // ждём ключ без общего мьютекса — другие ключи свободны

	return sync.OnceFunc(func() {
		l.mu.Unlock()
		m.mu.Lock()
		l.refs--
		if l.refs == 0 {
			delete(m.locks, key)
		}
		m.mu.Unlock()
	})
}

// Len — по скольким ключам сейчас есть держатель или ждущие.
func (m *KeyedMutex) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.locks)
}
