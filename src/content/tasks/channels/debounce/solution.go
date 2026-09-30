package main

import (
	"context"
	"time"
)

// Debounce пропускает из серии частых событий только последнее: событие
// уходит наружу, когда после него в in было тихо в течение wait.
//
// Если maxWait > 0, серия не может копиться бесконечно: последнее событие
// уходит не позже чем через maxWait после первого события серии, даже
// если тишины так и не было. После отправки начинается новая серия.
//
// Если in закрылся, пока событие ждёт, оно отправляется сразу, и выход
// закрывается. При отмене ctx выход закрывается, ожидающее событие теряется.
func Debounce[T any](ctx context.Context, in <-chan T, wait, maxWait time.Duration) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		var (
			last    T
			pending bool
			// nil-каналы: пока серии нет, ветки таймеров выключены.
			quiet    <-chan time.Time // тишина wait после последнего события
			deadline <-chan time.Time // maxWait от первого события серии
		)
		emit := func() bool {
			pending, quiet, deadline = false, nil, nil
			select {
			case out <- last:
				return true
			case <-ctx.Done():
				return false
			}
		}
		for {
			select {
			case v, ok := <-in:
				if !ok {
					if pending {
						emit() // хвост не теряем
					}
					return
				}
				last = v
				if !pending && maxWait > 0 {
					deadline = time.After(maxWait) // только на первом событии серии
				}
				pending = true
				quiet = time.After(wait) // каждое событие откладывает тишину
			case <-quiet:
				if !emit() {
					return
				}
			case <-deadline:
				if !emit() {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
