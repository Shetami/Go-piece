package main

import "sync"

// WithLock выполняет f под мьютексом mu.
// Мьютекс должен освободиться, даже если f паникует;
// сама паника при этом не глотается, а летит дальше.
func WithLock(mu *sync.Mutex, f func()) {
	mu.Lock()
	// defer выполняется и при панике — без recover она продолжит раскручивать стек.
	defer mu.Unlock()
	f()
}
