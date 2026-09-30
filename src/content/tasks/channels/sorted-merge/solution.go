package main

import "context"

// MergeSorted сливает каналы, каждый из которых отдаёт значения по
// неубыванию (по cmp), в один канал, тоже по неубыванию. При равенстве
// первым идёт значение из канала с меньшим индексом в ins.
//
// Выход закрывается, когда все входы закрыты и вычитаны или когда отменён
// ctx. После отмены горутины MergeSorted не остаются висеть.
func MergeSorted[T any](ctx context.Context, cmp func(a, b T) int, ins ...<-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		// head[i] — очередное значение канала i; has[i] — оно есть.
		// Отдельный флаг, а не «нулевое значение»: ноль — законные данные.
		heads := make([]T, len(ins))
		has := make([]bool, len(ins))

		// fill дочитывает голову канала i. Пока у каждого живого канала
		// нет головы, минимум выбирать нельзя: следующее значение
		// медленного канала может оказаться меньше всех.
		fill := func(i int) bool {
			select {
			case v, ok := <-ins[i]:
				if ok {
					heads[i], has[i] = v, true
				}
				return true
			case <-ctx.Done():
				return false
			}
		}
		for i := range ins {
			if !fill(i) {
				return
			}
		}
		for {
			best := -1
			for i := range ins {
				// Строгое «меньше» — при равенстве остаётся меньший индекс.
				if has[i] && (best < 0 || cmp(heads[i], heads[best]) < 0) {
					best = i
				}
			}
			if best < 0 {
				return // все входы кончились
			}
			select {
			case out <- heads[best]:
			case <-ctx.Done():
				return
			}
			has[best] = false
			if !fill(best) {
				return
			}
		}
	}()
	return out
}
