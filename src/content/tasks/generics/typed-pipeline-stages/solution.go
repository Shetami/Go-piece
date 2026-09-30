package main

import "context"

// Stage — этап конвейера. Он запускает свою горутину, читает in, пишет
// результаты в возвращаемый канал и закрывает его, когда in закрыт или
// ctx отменён. При отмене этап завершается, даже если его выход никто
// не читает.
type Stage[T, U any] func(ctx context.Context, in <-chan T) <-chan U

// send отправляет v или сдаётся при отмене ctx. Возвращает false, если
// этапу пора выходить.
func send[T any](ctx context.Context, out chan<- T, v T) bool {
	select {
	case out <- v:
		return true
	case <-ctx.Done():
		return false
	}
}

// MapStage применяет f к каждому элементу.
func MapStage[T, U any](f func(T) U) Stage[T, U] {
	return func(ctx context.Context, in <-chan T) <-chan U {
		out := make(chan U)
		go func() {
			defer close(out)
			for v := range in {
				if !send(ctx, out, f(v)) {
					return
				}
			}
		}()
		return out
	}
}

// FilterStage пропускает элементы, для которых keep вернул true.
func FilterStage[T any](keep func(T) bool) Stage[T, T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T)
		go func() {
			defer close(out)
			for v := range in {
				if keep(v) && !send(ctx, out, v) {
					return
				}
			}
		}()
		return out
	}
}

// BatchStage собирает элементы в пачки по n (n >= 1). Остаток, если он
// есть, уходит последней пачкой после закрытия in. Каждая пачка —
// отдельный слайс: получатель может хранить и менять её.
func BatchStage[T any](n int) Stage[T, []T] {
	return func(ctx context.Context, in <-chan T) <-chan []T {
		out := make(chan []T)
		go func() {
			defer close(out)
			batch := make([]T, 0, n)
			for v := range in {
				batch = append(batch, v)
				if len(batch) == n {
					if !send(ctx, out, batch) {
						return
					}
					// Новый слайс: отправленный теперь принадлежит получателю.
					batch = make([]T, 0, n)
				}
			}
			if len(batch) > 0 && ctx.Err() == nil {
				send(ctx, out, batch)
			}
		}()
		return out
	}
}

// Then соединяет два этапа в один: выход first идёт на вход second.
func Then[A, B, C any](first Stage[A, B], second Stage[B, C]) Stage[A, C] {
	return func(ctx context.Context, in <-chan A) <-chan C {
		return second(ctx, first(ctx, in))
	}
}
