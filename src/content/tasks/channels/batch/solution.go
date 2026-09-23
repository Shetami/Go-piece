package main

import "time"

// Batch собирает значения из in в пачки. Пачка отправляется, когда
// в ней size элементов или с первого элемента пачки прошло maxWait.
// Когда in закрылся, недобранная пачка отправляется, и выход закрывается.
// Пустые пачки не отправляются.
func Batch(in <-chan int, size int, maxWait time.Duration) <-chan []int {
	out := make(chan []int)
	go func() {
		defer close(out)
		var batch []int
		// Таймер заводится на первом элементе пачки. Пока пачка пуста,
		// канал таймера nil — его ветка в select выключена.
		var timer <-chan time.Time

		flush := func() {
			if len(batch) > 0 {
				out <- batch
				batch = nil // новая пачка — новый слайс: отданный уже чужой
			}
			timer = nil
		}

		for {
			select {
			case v, ok := <-in:
				if !ok {
					flush()
					return
				}
				if len(batch) == 0 {
					timer = time.After(maxWait)
				}
				batch = append(batch, v)
				if len(batch) == size {
					flush()
				}
			case <-timer:
				flush()
			}
		}
	}()
	return out
}
