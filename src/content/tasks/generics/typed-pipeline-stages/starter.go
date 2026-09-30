package main

import "context"

// Stage — этап конвейера. Он запускает свою горутину, читает in, пишет
// результаты в возвращаемый канал и закрывает его, когда in закрыт или
// ctx отменён. При отмене этап завершается, даже если его выход никто
// не читает.
type Stage[T, U any] func(ctx context.Context, in <-chan T) <-chan U

// MapStage применяет f к каждому элементу.
func MapStage[T, U any](f func(T) U) Stage[T, U] {
	// ваш код
	return func(ctx context.Context, in <-chan T) <-chan U { return nil }
}

// FilterStage пропускает элементы, для которых keep вернул true.
func FilterStage[T any](keep func(T) bool) Stage[T, T] {
	// ваш код
	return func(ctx context.Context, in <-chan T) <-chan T { return nil }
}

// BatchStage собирает элементы в пачки по n (n >= 1). Остаток, если он
// есть, уходит последней пачкой после закрытия in. Каждая пачка —
// отдельный слайс: получатель может хранить и менять её.
func BatchStage[T any](n int) Stage[T, []T] {
	// ваш код
	return func(ctx context.Context, in <-chan T) <-chan []T { return nil }
}

// Then соединяет два этапа в один: выход first идёт на вход second.
func Then[A, B, C any](first Stage[A, B], second Stage[B, C]) Stage[A, C] {
	// ваш код
	return func(ctx context.Context, in <-chan A) <-chan C { return nil }
}
