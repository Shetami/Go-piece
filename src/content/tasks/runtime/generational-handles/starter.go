package main

import "iter"

// Handle — ссылка на объект в Pool: индекс слота и поколение. Сравнимый,
// его можно хранить в картах и передавать по значению. Нулевой Handle{}
// никогда не ссылается на живой объект.
type Handle struct {
	index uint32
	gen   uint32
}

// Pool хранит объекты в слайсе слотов и переиспользует освободившиеся слоты.
// Устаревший Handle (объект удалён, даже если слот уже занят новым) не
// даёт доступа ни к чему. Удалённый объект пулом не удерживается.
// Нулевое значение готово к работе. Не обязан быть потокобезопасным.
type Pool[T any] struct {
	// ваши поля
}

// Insert кладёт v в свободный слот (или в новый, если свободных нет).
func (p *Pool[T]) Insert(v T) Handle {
	// ваш код
	return Handle{}
}

// Get возвращает указатель на живой объект или nil, false для устаревшего Handle.
func (p *Pool[T]) Get(h Handle) (*T, bool) {
	// ваш код
	return nil, false
}

// Remove удаляет объект; false, если Handle уже устарел (повторный Remove).
func (p *Pool[T]) Remove(h Handle) bool {
	// ваш код
	return false
}

// Len — число живых объектов; Slots — сколько слотов выделено всего.
func (p *Pool[T]) Len() int {
	// ваш код
	return 0
}

func (p *Pool[T]) Slots() int {
	// ваш код
	return 0
}

// All перебирает живые объекты в порядке индексов слотов.
func (p *Pool[T]) All() iter.Seq2[Handle, *T] {
	// ваш код
	return func(func(Handle, *T) bool) {}
}
