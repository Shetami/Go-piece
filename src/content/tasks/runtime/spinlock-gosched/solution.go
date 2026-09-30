package main

import (
	"runtime"
	"sync/atomic"
)

// SpinLock — спин-блокировка на атомике: реализует sync.Locker.
// Нулевое значение — открытый замок.
//
// Требования:
//   - взаимное исключение;
//   - Lock не должен жечь квант планировщика, когда владелец замка не
//     выполняется: программа с GOMAXPROCS=1, где владелец уступает
//     процессор внутри критической секции, должна работать быстро;
//   - Unlock незапертого замка паникует.
type SpinLock struct {
	state atomic.Int32 // 0 — открыт, 1 — заперт
}

func (l *SpinLock) Lock() {
	for !l.state.CompareAndSwap(0, 1) {
		// Уступаем процессор: владелец, возможно, ждёт своей очереди
		// на этом же P. Без Gosched горутина крутилась бы до
		// принудительного вытеснения — ~10 мс на каждый круг.
		runtime.Gosched()
	}
}

// TryLock захватывает замок, если он свободен, и сообщает, удалось ли.
func (l *SpinLock) TryLock() bool {
	return l.state.CompareAndSwap(0, 1)
}

func (l *SpinLock) Unlock() {
	if !l.state.CompareAndSwap(1, 0) {
		panic("SpinLock: Unlock незапертого замка")
	}
}
