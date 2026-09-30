package main

import (
	"context"
	"time"
)

// Work запускает воркер: он берёт задачи из jobs, считает f и отдаёт
// результаты в results в порядке задач. Параллельно раз в pulse воркер
// сигналит в heartbeat «я жив» — и когда ждёт задач, и когда ждёт, пока
// заберут результат. Сигнал не блокирует воркер: если heartbeat никто не
// читает, сигнал теряется.
//
// Когда jobs закрыт и все результаты отданы или когда отменён ctx,
// воркер закрывает оба канала и завершается.
func Work(ctx context.Context, pulse time.Duration, jobs <-chan int, f func(int) int) (heartbeat <-chan struct{}, results <-chan int) {
	// Буфер на один сигнал: читатель, пришедший между тиками, увидит последний.
	hb := make(chan struct{}, 1)
	out := make(chan int)
	go func() {
		defer close(hb)
		defer close(out)
		ticker := time.NewTicker(pulse)
		defer ticker.Stop()

		beat := func() {
			select {
			case hb <- struct{}{}:
			default: // никто не слушает — не ждём
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				beat()
			case j, ok := <-jobs:
				if !ok {
					return
				}
				r := f(j)
				// Отправка результата — тоже ожидание, и на нём пульс
				// не должен замирать.
				for sent := false; !sent; {
					select {
					case out <- r:
						sent = true
					case <-ticker.C:
						beat()
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return hb, out
}
