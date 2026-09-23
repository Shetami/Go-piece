package main

import "context"

// OrDone пересылает значения из in, пока in не закрыт и ctx не отменён.
// В обоих случаях выход закрывается.
func OrDone(ctx context.Context, in <-chan int) <-chan int {
	// ваш код
	return in
}
