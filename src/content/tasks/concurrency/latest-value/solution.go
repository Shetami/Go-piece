package main

import (
	"context"
	"sync"
)

// Latest хранит последнее значение и раздаёт его наблюдателям.
//
//   - Set никогда не блокируется, даже если наблюдатели не читают.
//   - Watch(ctx) возвращает канал. Если значение уже есть, наблюдатель
//     сразу получает текущее. Дальше — только самые свежие: медленный
//     читатель пропускает промежуточные значения, но значения приходят в
//     порядке Set, ни одно не приходит дважды, и последнее значение
//     обязательно дойдёт.
//   - Канал закрывается после отмены ctx; горутина наблюдателя завершается,
//     даже если из канала больше никто не читает.
type Latest[T any] struct {
	mu      sync.Mutex
	val     T
	ver     int           // 0 — значения ещё не было
	changed chan struct{} // закрывается и заменяется при каждом Set
}

func NewLatest[T any]() *Latest[T] {
	return &Latest[T]{changed: make(chan struct{})}
}

func (l *Latest[T]) Set(v T) {
	l.mu.Lock()
	l.val = v
	l.ver++
	close(l.changed) // будим всех наблюдателей разом
	l.changed = make(chan struct{})
	l.mu.Unlock()
}

func (l *Latest[T]) Watch(ctx context.Context) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		sent := 0 // версия, отправленная этому наблюдателю
		for {
			l.mu.Lock()
			val, ver, changed := l.val, l.ver, l.changed
			l.mu.Unlock()

			if ver == sent { // нового нет — ждём изменения
				select {
				case <-changed:
					continue
				case <-ctx.Done():
					return
				}
			}
			select {
			case out <- val:
				sent = ver
			case <-changed:
				// Пока ждали читателя, пришло новее — отправим его вместо этого.
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
