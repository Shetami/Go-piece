package main

import "sync"

// Barrier — точка сбора n горутин. Многоразовый: после того как собрались
// n участников и все прошли, барьер готов к следующему кругу.
type Barrier struct {
	mu      sync.Mutex
	cond    *sync.Cond
	n       int
	arrived int
	round   int // номер круга: по нему ждущие понимают, что их круг завершён
}

func NewBarrier(n int) *Barrier {
	b := &Barrier{n: n}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Wait блокирует горутину, пока Wait не вызовут n горутин,
// после чего пропускает всех n.
func (b *Barrier) Wait() {
	b.mu.Lock()
	defer b.mu.Unlock()

	round := b.round
	b.arrived++
	if b.arrived == b.n {
		// Последний пришедший открывает барьер и готовит следующий круг.
		b.arrived = 0
		b.round++
		b.cond.Broadcast()
		return
	}
	// Ждём смены круга, а не «arrived == n»: к моменту пробуждения
	// счётчик уже обнулён, и быстрые участники могли начать новый круг.
	for round == b.round {
		b.cond.Wait()
	}
}
