package main

// Barrier — точка сбора n горутин. Многоразовый: после того как собрались
// n участников и все прошли, барьер готов к следующему кругу.
type Barrier struct {
	// ваши поля
}

func NewBarrier(n int) *Barrier {
	return &Barrier{}
}

// Wait блокирует горутину, пока Wait не вызовут n горутин,
// после чего пропускает всех n.
func (b *Barrier) Wait() {
	// ваш код
}
