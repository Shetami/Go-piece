package main

import "context"

// OrDone пересылает значения из in, пока in не закрыт и ctx не отменён.
// В обоих случаях выход закрывается.
func OrDone(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				// Потребитель мог уйти, пока мы держим значение, — снова слушаем отмену.
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
