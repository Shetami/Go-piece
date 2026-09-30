package main

import (
	"errors"
	"sync"
)

// ErrClosed — шлюз закрыт, новые операции не принимаются.
var ErrClosed = errors.New("gate closed")

// Gate пропускает операции к ресурсу (соединению, файлу, продюсеру Kafka)
// и умеет закрываться аккуратно: сначала перестаёт пускать новые операции,
// дожидается уже начатых и только потом освобождает ресурс.
type Gate struct {
	mu       sync.Mutex
	closed   bool
	inflight sync.WaitGroup

	onClose  func() error
	once     sync.Once
	closeErr error
}

// NewGate создаёт шлюз. onClose освобождает ресурс.
func NewGate(onClose func() error) *Gate {
	return &Gate{onClose: onClose}
}

// Do выполняет fn и возвращает её ошибку, если шлюз не закрывается;
// иначе сразу возвращает ErrClosed, не вызывая fn.
func (g *Gate) Do(fn func() error) error {
	// Проверка флага и Add — под одним мьютексом с установкой флага в
	// Close. Иначе Do может проверить флаг, Close — выставить его и
	// пройти Wait, и только потом Do сделает Add и полезет в закрытый ресурс.
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return ErrClosed
	}
	g.inflight.Add(1)
	g.mu.Unlock()

	defer g.inflight.Done() // defer: даже если fn паникует
	return fn()
}

// Close запрещает новые Do, ждёт завершения всех начатых, затем вызывает
// onClose — ровно один раз за жизнь шлюза. Повторные и одновременные
// вызовы Close тоже дожидаются окончания onClose и возвращают её ошибку.
func (g *Gate) Close() error {
	// Once.Do блокирует одновременных вызывающих, пока первый не закончит, —
	// поэтому все они возвращают уже записанную closeErr.
	g.once.Do(func() {
		g.mu.Lock()
		g.closed = true
		g.mu.Unlock()
		g.inflight.Wait() // после closed новых Add не будет
		g.closeErr = g.onClose()
	})
	return g.closeErr
}
