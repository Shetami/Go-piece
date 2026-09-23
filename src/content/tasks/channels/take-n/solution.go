package main

import "context"

// Take отдаёт не больше n первых значений из in и закрывает выход.
// Выход закрывается и раньше — если закрылся in или отменён ctx.
func Take(ctx context.Context, in <-chan int, n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for range n {
			// Ждём и значение, и отмену: иначе на пустом in горутина зависнет.
			var v int
			var ok bool
			select {
			case v, ok = <-in:
				if !ok {
					return
				}
			case <-ctx.Done():
				return
			}
			// Отправка тоже может ждать вечно, если потребитель ушёл.
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
