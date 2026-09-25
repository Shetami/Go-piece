package main

// OrderedMap применяет f к значениям из in параллельно — одновременно
// работает не больше workers вызовов f, — но отдаёт результаты
// в том же порядке, в каком приходили значения. Выходной канал
// закрывается, когда in закрыт и все результаты отданы.
func OrderedMap(in <-chan int, workers int, f func(int) int) <-chan int {
	// ваш код
	out := make(chan int)
	close(out)
	return out
}
