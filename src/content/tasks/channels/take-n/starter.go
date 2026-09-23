package main

import "context"

// Take отдаёт не больше n первых значений из in и закрывает выход.
// Выход закрывается и раньше — если закрылся in или отменён ctx.
func Take(ctx context.Context, in <-chan int, n int) <-chan int {
	// ваш код
	return nil
}
