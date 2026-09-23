package main

import "context"

// Group запускает задачи в горутинах и собирает первую ошибку.
type Group struct {
	// ваши поля
}

// WithContext возвращает группу и контекст, который отменяется,
// как только одна из задач вернула ошибку (или после Wait).
func WithContext(ctx context.Context) (*Group, context.Context) {
	return &Group{}, ctx
}

// Go запускает f в новой горутине.
func (g *Group) Go(f func() error) {
	// ваш код
}

// Wait ждёт все задачи и возвращает первую ошибку (или nil).
func (g *Group) Wait() error {
	// ваш код
	return nil
}
