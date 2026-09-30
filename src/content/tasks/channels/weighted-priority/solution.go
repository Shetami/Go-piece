package main

import "context"

// Prioritize сливает high и low в один канал. Готовые значения из high
// идут первыми, но low не должен голодать: если из high подряд ушло burst
// значений (burst >= 1), а в low есть готовое значение, следующим уходит
// одно значение из low, и счёт начинается заново. Если high пуст, а low
// нет — значения low идут без задержки.
//
// Закрытый вход больше не читается. Выход закрывается, когда закрыты оба
// входа или отменён ctx; после отмены горутины не остаются висеть.
func Prioritize[T any](ctx context.Context, high, low <-chan T, burst int) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		streak := 0 // сколько значений из high ушло подряд

		// Закрытый вход обнуляем: чтение из nil-канала в select
		// никогда не готово, ветка выключается.
		for high != nil || low != nil {
			var v T
			got := false

			// 1. Квота high исчерпана — даём шанс low, но не ждём его.
			if streak >= burst && low != nil {
				select {
				case x, ok := <-low:
					if !ok {
						low = nil
						continue
					}
					v, got, streak = x, true, 0
				default:
				}
			}
			// 2. Приоритет: готовое значение high, тоже без ожидания.
			if !got && high != nil {
				select {
				case x, ok := <-high:
					if !ok {
						high = nil
						continue
					}
					v, got = x, true
					streak++
				default:
				}
			}
			// 3. Ничего не готово — ждём любой вход.
			if !got {
				select {
				case x, ok := <-high:
					if !ok {
						high = nil
						continue
					}
					v = x
					streak++
				case x, ok := <-low:
					if !ok {
						low = nil
						continue
					}
					v, streak = x, 0
				case <-ctx.Done():
					return
				}
			}

			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
