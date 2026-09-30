package main

import (
	"context"
	"fmt"
)

// Key — типизированный ключ контекста. Каждый NewKey даёт отдельный ключ,
// даже при одинаковом имени и типе: значения разных ключей не пересекаются.
type Key[T any] struct {
	name string
}

// box отличает «положили нулевое значение (например, nil-ошибку)»
// от «ничего не положили».
type box[T any] struct{ v T }

func NewKey[T any](name string) *Key[T] {
	return &Key[T]{name: name}
}

// With возвращает контекст со значением v по этому ключу.
func (k *Key[T]) With(ctx context.Context, v T) context.Context {
	// Ключ — указатель: он уникален и сравним, а чужой код его не подделает.
	return context.WithValue(ctx, k, box[T]{v})
}

// Get достаёт значение; ok == false, если по этому ключу ничего не клали.
// Положенное nil-значение (T — интерфейс, указатель) даёт (nil, true).
func (k *Key[T]) Get(ctx context.Context) (T, bool) {
	b, ok := ctx.Value(k).(box[T])
	return b.v, ok
}

// MustGet — как Get, но паникует с сообщением, содержащим имя ключа.
func (k *Key[T]) MustGet(ctx context.Context) T {
	v, ok := k.Get(ctx)
	if !ok {
		panic(fmt.Sprintf("в контексте нет значения %q", k.name))
	}
	return v
}

// String — имя ключа (удобно в логах и в выводе context.WithValue).
func (k *Key[T]) String() string { return k.name }
