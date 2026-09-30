package main

import "context"

// Key — типизированный ключ контекста. Каждый NewKey даёт отдельный ключ,
// даже при одинаковом имени и типе: значения разных ключей не пересекаются.
type Key[T any] struct {
	name string
}

func NewKey[T any](name string) *Key[T] {
	return &Key[T]{name: name}
}

// With возвращает контекст со значением v по этому ключу.
func (k *Key[T]) With(ctx context.Context, v T) context.Context {
	// ваш код
	return ctx
}

// Get достаёт значение; ok == false, если по этому ключу ничего не клали.
// Положенное nil-значение (T — интерфейс, указатель) даёт (nil, true).
func (k *Key[T]) Get(ctx context.Context) (T, bool) {
	// ваш код
	var zero T
	return zero, false
}

// MustGet — как Get, но паникует с сообщением, содержащим имя ключа.
func (k *Key[T]) MustGet(ctx context.Context) T {
	// ваш код
	var zero T
	return zero
}

// String — имя ключа (удобно в логах и в выводе context.WithValue).
func (k *Key[T]) String() string { return k.name }
