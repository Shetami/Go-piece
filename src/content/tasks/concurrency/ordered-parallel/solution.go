package main

// OrderedMap применяет f к значениям из in параллельно — одновременно
// работает не больше workers вызовов f, — но отдаёт результаты
// в том же порядке, в каком приходили значения. Выходной канал
// закрывается, когда in закрыт и все результаты отданы.
func OrderedMap(in <-chan int, workers int, f func(int) int) <-chan int {
	out := make(chan int)
	// Очередь «обещаний» в порядке поступления. Её ёмкость и ограничивает
	// параллельность: workers-1 обещаний в буфере плюс одно, которого
	// сейчас ждёт второй этап. Горутина с f стартует только после того,
	// как обещание встало в очередь.
	pending := make(chan chan int, max(workers-1, 0))

	go func() {
		defer close(pending)
		for v := range in {
			res := make(chan int, 1) // буфер: воркер не ждёт, пока результат заберут
			pending <- res
			go func() { res <- f(v) }()
		}
	}()

	go func() {
		defer close(out)
		// Забираем результаты строго по очереди обещаний — отсюда порядок.
		for res := range pending {
			out <- <-res
		}
	}()
	return out
}
