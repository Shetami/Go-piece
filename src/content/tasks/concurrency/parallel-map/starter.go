package main

// ParallelMap применяет f к каждому элементу и возвращает результаты
// В ТОМ ЖЕ ПОРЯДКЕ, что и вход.
//
// Одновременно должно работать не больше workers горутин: f может быть
// тяжёлой, и запускать её сразу на всём слайсе нельзя.
func ParallelMap(items []int, workers int, f func(int) int) []int {
	// ваш код
	return nil
}
