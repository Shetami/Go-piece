package main

import "time"

// Debouncer откладывает вызов fn, пока события не затихнут.
//
//   - fn вызывается через wait после ПОСЛЕДНЕГО Trigger: серия Trigger с
//     паузами меньше wait даёт один вызов fn со значением последнего Trigger.
//   - Вызовы fn никогда не перекрываются и идут в порядке значений:
//     более старое значение не может прийти в fn после более нового.
//   - Flush немедленно (синхронно) вызывает fn с отложенным значением, если
//     оно есть, и отменяет отложенный вызов.
//   - Stop отменяет отложенный вызов и дожидается выполняющегося fn;
//     после Stop Trigger и Flush ничего не делают.
type Debouncer[T any] struct {
	wait time.Duration
	fn   func(T)
	// ваши поля
}

func NewDebouncer[T any](wait time.Duration, fn func(T)) *Debouncer[T] {
	return &Debouncer[T]{wait: wait, fn: fn}
}

func (d *Debouncer[T]) Trigger(v T) {
	// ваш код
}

func (d *Debouncer[T]) Flush() {
	// ваш код
}

func (d *Debouncer[T]) Stop() {
	// ваш код
}
