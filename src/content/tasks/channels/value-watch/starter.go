package main

import "context"

// Var хранит значение (например, текущий конфиг или лидера кластера)
// и оповещает об изменениях любое число читателей без опроса.
type Var[T any] struct {
	// ваши поля
}

func NewVar[T any](v T) *Var[T] {
	// ваш код
	return &Var[T]{}
}

// Get возвращает текущее значение и канал, который закроется при
// следующем Set. Значение и канал согласованы: если Set случился после
// Get, канал уже закрыт или закроется.
func (x *Var[T]) Get() (T, <-chan struct{}) {
	// ваш код
	var zero T
	return zero, nil
}

// Set меняет значение и будит всех, кто ждёт на канале из Get.
func (x *Var[T]) Set(v T) {
	// ваш код
}

// WaitFor ждёт, пока значение станет удовлетворять pred (текущее тоже
// проверяется), и возвращает его. Если ctx отменён раньше — текущее
// значение и ctx.Err().
func (x *Var[T]) WaitFor(ctx context.Context, pred func(T) bool) (T, error) {
	// ваш код
	var zero T
	return zero, nil
}
