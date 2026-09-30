package main

import "iter"

// Handle — ссылка на объект в Pool: индекс слота и поколение. Сравнимый,
// его можно хранить в картах и передавать по значению. Нулевой Handle{}
// никогда не ссылается на живой объект.
type Handle struct {
	index uint32
	gen   uint32
}

type slot[T any] struct {
	val T
	gen uint32 // нечётное — слот занят, чётное — свободен
}

// Pool хранит объекты в слайсе слотов и переиспользует освободившиеся слоты.
// Устаревший Handle (объект удалён, даже если слот уже занят новым) не
// даёт доступа ни к чему. Удалённый объект пулом не удерживается.
// Нулевое значение готово к работе. Не обязан быть потокобезопасным.
type Pool[T any] struct {
	slots []slot[T]
	free  []uint32 // стек индексов свободных слотов
	live  int
}

// Insert кладёт v в свободный слот (или в новый, если свободных нет).
func (p *Pool[T]) Insert(v T) Handle {
	var i uint32
	if n := len(p.free); n > 0 {
		i = p.free[n-1]
		p.free = p.free[:n-1]
	} else {
		p.slots = append(p.slots, slot[T]{})
		i = uint32(len(p.slots) - 1)
	}
	s := &p.slots[i]
	// Поколение живого слота всегда нечётное, поэтому Handle{} с gen 0
	// не совпадёт ни с одним живым объектом.
	s.gen++
	s.val = v
	p.live++
	return Handle{index: i, gen: s.gen}
}

func (p *Pool[T]) slotOf(h Handle) *slot[T] {
	if int(h.index) >= len(p.slots) {
		return nil
	}
	s := &p.slots[h.index]
	if s.gen != h.gen || s.gen%2 == 0 {
		return nil // объект удалён или слот уже отдан другому
	}
	return s
}

// Get возвращает указатель на живой объект или nil, false для устаревшего Handle.
func (p *Pool[T]) Get(h Handle) (*T, bool) {
	if s := p.slotOf(h); s != nil {
		return &s.val, true
	}
	return nil, false
}

// Remove удаляет объект; false, если Handle уже устарел (повторный Remove).
func (p *Pool[T]) Remove(h Handle) bool {
	s := p.slotOf(h)
	if s == nil {
		// Без этой проверки повторный Remove положил бы индекс в free дважды,
		// и два следующих Insert получили бы один и тот же слот.
		return false
	}
	var zero T
	s.val = zero // не держим удалённый объект: GC должен его собрать
	s.gen++      // старые Handle устаревают навсегда
	p.free = append(p.free, h.index)
	p.live--
	return true
}

// Len — число живых объектов; Slots — сколько слотов выделено всего.
func (p *Pool[T]) Len() int { return p.live }

func (p *Pool[T]) Slots() int { return len(p.slots) }

// All перебирает живые объекты в порядке индексов слотов.
func (p *Pool[T]) All() iter.Seq2[Handle, *T] {
	return func(yield func(Handle, *T) bool) {
		for i := range p.slots {
			s := &p.slots[i]
			if s.gen%2 == 1 && !yield(Handle{index: uint32(i), gen: s.gen}, &s.val) {
				return
			}
		}
	}
}
