package main

import "runtime"

// AllocStats — средние показатели на один запуск.
type AllocStats struct {
	Allocs uint64 // выделений памяти (целая часть среднего)
	Bytes  uint64 // выделенных байт (целая часть среднего)
}

// Measure запускает f runs раз (runs >= 1) и считает, сколько в среднем
// выделений памяти и байт приходится на один запуск — как
// testing.AllocsPerRun, но в любом коде (бенчмарк-хелпер, самодиагностика).
//
// Разовая инициализация внутри f (ленивый кэш, sync.Once) в результат
// попадать не должна. На время замера другие горутины программы не
// должны сильно искажать счётчики. После возврата настройки рантайма
// такие же, как до вызова.
func Measure(runs int, f func()) AllocStats {
	// Один P: пока мы меряем, остальные горутины не выполняются параллельно
	// и не добавляют свои аллокации в общие счётчики. Обязательно вернуть.
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))

	// Прогрев: разовая инициализация внутри f не должна попасть в замер.
	f()

	// MemStats — большая структура; объявляем заранее, чтобы
	// сам замер не добавлял аллокаций между двумя чтениями.
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for range runs {
		f()
	}
	runtime.ReadMemStats(&after)

	// Mallocs и TotalAlloc — монотонные счётчики (в отличие от HeapAlloc,
	// который уменьшается после сборки мусора).
	return AllocStats{
		Allocs: (after.Mallocs - before.Mallocs) / uint64(runs),
		Bytes:  (after.TotalAlloc - before.TotalAlloc) / uint64(runs),
	}
}
