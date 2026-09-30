package main

import "errors"

// ErrClosed — шлюз закрыт, новые операции не принимаются.
var ErrClosed = errors.New("gate closed")

// Gate пропускает операции к ресурсу (соединению, файлу, продюсеру Kafka)
// и умеет закрываться аккуратно: сначала перестаёт пускать новые операции,
// дожидается уже начатых и только потом освобождает ресурс.
type Gate struct {
	// ваши поля
}

// NewGate создаёт шлюз. onClose освобождает ресурс.
func NewGate(onClose func() error) *Gate {
	// ваш код
	return &Gate{}
}

// Do выполняет fn и возвращает её ошибку, если шлюз не закрывается;
// иначе сразу возвращает ErrClosed, не вызывая fn.
func (g *Gate) Do(fn func() error) error {
	// ваш код
	return fn()
}

// Close запрещает новые Do, ждёт завершения всех начатых, затем вызывает
// onClose — ровно один раз за жизнь шлюза. Повторные и одновременные
// вызовы Close тоже дожидаются окончания onClose и возвращают её ошибку.
func (g *Gate) Close() error {
	// ваш код
	return nil
}
