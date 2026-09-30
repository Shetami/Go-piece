package main

// SpinLock — учебный спинлок на CompareAndSwap. Реализует sync.Locker,
// поэтому годится везде, где ждут Locker, — например, для sync.NewCond.
// Нулевое значение — открытый замок.
type SpinLock struct {
	// ваши поля
}

// Lock крутится, пока не захватит замок. Пока ждёт, уступает процессор
// другим горутинам.
func (l *SpinLock) Lock() {
	// ваш код
}

// TryLock захватывает замок, если он свободен, и сообщает, удалось ли.
// Никогда не ждёт.
func (l *SpinLock) TryLock() bool {
	// ваш код
	return true
}

// Unlock открывает замок. Unlock открытого замка — паника
// с сообщением "unlock of unlocked SpinLock".
func (l *SpinLock) Unlock() {
	// ваш код
}
