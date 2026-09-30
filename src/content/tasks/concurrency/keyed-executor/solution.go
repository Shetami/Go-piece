package main

import "sync"

// KeyedExecutor выполняет задачи в фоне по правилам:
//
//   - задачи с одинаковым ключом — строго последовательно, в порядке Submit;
//   - задачи с разными ключами — параллельно;
//   - Submit не блокируется на выполнении задач (очередь ключа не ограничена);
//   - для ключа без незавершённых задач не держится ни горутины, ни записи
//     в карте: миллион разовых ключей не должен копиться.
type KeyedExecutor struct {
	mu     sync.Mutex
	queues map[string][]func() // есть ключ — по нему работает горутина
	wg     sync.WaitGroup
}

func NewKeyedExecutor() *KeyedExecutor {
	return &KeyedExecutor{queues: make(map[string][]func())}
}

// Submit ставит задачу в очередь ключа.
func (e *KeyedExecutor) Submit(key string, task func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.wg.Add(1)
	if q, busy := e.queues[key]; busy {
		e.queues[key] = append(q, task) // горутина ключа заберёт сама
		return
	}
	e.queues[key] = nil // ключ занят, очередь пуста
	go e.run(key, task)
}

// run выполняет задачи ключа, пока его очередь не опустеет.
func (e *KeyedExecutor) run(key string, task func()) {
	for {
		task()
		e.mu.Lock()
		q := e.queues[key]
		if len(q) == 0 {
			// Проверка «пусто» и удаление — под одним мьютексом с Submit:
			// иначе задача, пришедшая между ними, осталась бы без горутины.
			delete(e.queues, key)
			e.mu.Unlock()
			// Done — после удаления: иначе Wait вернётся, пока запись ключа
			// ещё в карте, и Keys() сразу после Wait увидит «живой» ключ.
			e.wg.Done()
			return
		}
		task = q[0]
		q[0] = nil // не держим ссылку на выполненную задачу
		e.queues[key] = q[1:]
		e.mu.Unlock()
		e.wg.Done()
	}
}

// Wait дожидается выполнения всех задач, отправленных до вызова Wait.
func (e *KeyedExecutor) Wait() { e.wg.Wait() }

// Keys возвращает число ключей, у которых есть незавершённые задачи.
func (e *KeyedExecutor) Keys() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.queues)
}
