package main

// KeyedExecutor выполняет задачи в фоне по правилам:
//
//   - задачи с одинаковым ключом — строго последовательно, в порядке Submit;
//   - задачи с разными ключами — параллельно;
//   - Submit не блокируется на выполнении задач (очередь ключа не ограничена);
//   - для ключа без незавершённых задач не держится ни горутины, ни записи
//     в карте: миллион разовых ключей не должен копиться.
type KeyedExecutor struct {
	// ваши поля
}

func NewKeyedExecutor() *KeyedExecutor {
	return &KeyedExecutor{}
}

// Submit ставит задачу в очередь ключа.
func (e *KeyedExecutor) Submit(key string, task func()) {
	// ваш код
}

// Wait дожидается выполнения всех задач, отправленных до вызова Wait.
func (e *KeyedExecutor) Wait() {
	// ваш код
}

// Keys возвращает число ключей, у которых есть незавершённые задачи.
func (e *KeyedExecutor) Keys() int {
	// ваш код
	return 0
}
