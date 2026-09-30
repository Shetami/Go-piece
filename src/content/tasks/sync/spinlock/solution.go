package main

import (
	"runtime"
	"sync/atomic"
)

// SpinLock — учебный спинлок на CompareAndSwap. Реализует sync.Locker,
// поэтому годится везде, где ждут Locker, — например, для sync.NewCond.
// Нулевое значение — открытый замок.
type SpinLock struct {
	locked atomic.Bool
}

// Lock крутится, пока не захватит замок. Пока ждёт, уступает процессор
// другим горутинам.
func (l *SpinLock) Lock() {
	for !l.locked.CompareAndSwap(false, true) {
		// Держатель замка мог быть вытеснен с процессора — без Gosched
		// мы жгли бы свой квант впустую, мешая ему дойти до Unlock.
		runtime.Gosched()
	}
}

// TryLock захватывает замок, если он свободен, и сообщает, удалось ли.
// Никогда не ждёт.
func (l *SpinLock) TryLock() bool {
	return l.locked.CompareAndSwap(false, true)
}

// Unlock открывает замок. Unlock открытого замка — паника
// с сообщением "unlock of unlocked SpinLock".
func (l *SpinLock) Unlock() {
	// Swap, а не Store: так мы узнаём, был ли замок закрыт, одной операцией.
	if !l.locked.Swap(false) {
		panic("unlock of unlocked SpinLock")
	}
}
