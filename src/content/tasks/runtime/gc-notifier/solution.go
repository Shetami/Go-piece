package main

import (
	"runtime"
	"sync/atomic"
)

type gcState struct {
	f       func()
	stopped atomic.Bool
}

// sentinel — пустышка, которую никто не держит: её «смерть» и есть
// сигнал, что прошла сборка мусора. Поле-указатель важно: объекты без
// указателей меньше 16 байт живут в tiny-аллокаторе, и их финализаторы
// могут не вызываться.
type sentinel struct {
	st *gcState
}

// OnGC вызывает f после каждого цикла сборки мусора, пока не вызвана stop —
// так метрики и адаптивные кэши узнают, что прошла сборка
// (например, чтобы подрезать кэш под давлением памяти).
//
// f вызывается в служебной горутине рантайма, не параллельно самой себе.
// Каждый цикл GC даёт не больше одного вызова f. После stop новые вызовы
// f не начинаются. Уведомители, созданные разными вызовами OnGC,
// независимы.
func OnGC(f func()) (stop func()) {
	st := &gcState{f: f}
	runtime.SetFinalizer(&sentinel{st: st}, onSentinel)
	return func() { st.stopped.Store(true) }
}

func onSentinel(s *sentinel) {
	if s.st.stopped.Load() {
		return // не перевзводим: объект и состояние соберутся
	}
	s.st.f()
	// Воскрешаем тот же объект и снова вешаем финализатор: после
	// возврата он опять недостижим и «умрёт» в следующем цикле GC.
	runtime.SetFinalizer(s, onSentinel)
}
