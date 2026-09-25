package main

import "sync"

// WithLock выполняет f под мьютексом mu.
// Мьютекс должен освободиться, даже если f паникует;
// сама паника при этом не глотается, а летит дальше.
func WithLock(mu *sync.Mutex, f func()) {
	// ваш код
}
