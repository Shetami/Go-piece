package main

import (
	"slices"
	"sync"
)

type subscriber struct {
	id uint64 // функции несравнимы — отписываемся по id
	fn any    // на деле func(E) для своего E
}

// topic[E] — ключ мапы: у каждого E свой тип, значит, и свой ключ.
type topic[E any] struct{}

// Bus — шина событий. Нулевое значение готово к работе. Безопасна для
// использования из многих горутин.
type Bus struct {
	mu   sync.Mutex
	next uint64
	subs map[any][]subscriber
}

// Subscribe подписывает h на события ровно типа E. Возвращает функцию
// отписки; повторный вызов отписки ничего не делает.
// Методы не могут иметь своих параметров типа, поэтому это функция.
func Subscribe[E any](b *Bus, h func(E)) (unsubscribe func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subs == nil {
		b.subs = make(map[any][]subscriber)
	}
	b.next++
	id := b.next
	k := topic[E]{}
	b.subs[k] = append(b.subs[k], subscriber{id: id, fn: h})

	var once sync.Once
	return func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			b.subs[k] = slices.DeleteFunc(b.subs[k], func(s subscriber) bool { return s.id == id })
		})
	}
}

// Publish синхронно вызывает обработчики событий ровно типа E в порядке
// подписки и возвращает, сколько их было вызвано. Обработчик может сам
// подписываться, отписываться и публиковать; изменения подписок во
// время Publish действуют со следующего Publish.
func Publish[E any](b *Bus, e E) int {
	b.mu.Lock()
	// Снимок под замком: DeleteFunc и append меняют общий массив.
	hs := slices.Clone(b.subs[topic[E]{}])
	b.mu.Unlock()
	// Обработчики зовём без замка — иначе Publish изнутри обработчика
	// повиснет на b.mu.
	for _, s := range hs {
		s.fn.(func(E))(e)
	}
	return len(hs)
}
