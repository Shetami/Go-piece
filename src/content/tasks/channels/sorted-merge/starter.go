package main

import "context"

// MergeSorted сливает каналы, каждый из которых отдаёт значения по
// неубыванию (по cmp), в один канал, тоже по неубыванию. При равенстве
// первым идёт значение из канала с меньшим индексом в ins.
//
// Выход закрывается, когда все входы закрыты и вычитаны или когда отменён
// ctx. После отмены горутины MergeSorted не остаются висеть.
func MergeSorted[T any](ctx context.Context, cmp func(a, b T) int, ins ...<-chan T) <-chan T {
	// ваш код
	return nil
}
