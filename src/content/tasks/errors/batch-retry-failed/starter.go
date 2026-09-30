package main

import (
	"context"
	"fmt"
)

type Item struct {
	ID      string
	Payload string
}

// BatchError — пачка принята частично. Ключи Failed — индексы в ТОЙ
// пачке, которую передали в send, а не в исходном списке.
type BatchError struct {
	Failed map[int]error
}

func (e *BatchError) Error() string {
	return fmt.Sprintf("не принято элементов: %d", len(e.Failed))
}

// ItemError — окончательная неудача одного элемента.
type ItemError struct {
	ID       string
	Attempts int // сколько раз элемент отправляли
	Err      error
}

func (e *ItemError) Error() string {
	return fmt.Sprintf("%s (попыток: %d): %v", e.ID, e.Attempts, e.Err)
}

func (e *ItemError) Unwrap() error { return e.Err }

// SendAll отправляет items через send, повторяя только то, что не принято.
//
// Ответ send(ctx, batch):
//   - nil — принята вся пачка;
//   - *BatchError в цепочке (errors.As) — не приняты элементы из Failed,
//     остальные приняты;
//   - любая другая ошибка — не принята вся пачка, у каждого элемента эта
//     ошибка.
//
// Непринятый элемент попадает в следующую пачку (в исходном порядке), если
// его ошибка временная — в цепочке есть звено с Temporary() bool == true
// (errors.As), — и его отправляли меньше maxAttempts раз. Иначе это
// окончательная неудача.
//
// Перед каждой отправкой проверяется ctx: если он отменён, все ещё не
// принятые элементы становятся окончательными неудачами с Err = ctx.Err().
//
// Результат: nil, если всё принято; иначе errors.Join из *ItemError в
// исходном порядке элементов. send с пустой пачкой не вызывается.
func SendAll(ctx context.Context, items []Item, maxAttempts int, send func(context.Context, []Item) error) error {
	// ваш код
	return send(ctx, items)
}
