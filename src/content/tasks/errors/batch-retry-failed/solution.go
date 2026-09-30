package main

import (
	"context"
	"errors"
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
	final := make([]error, len(items)) // по индексу исходного списка
	attempts := make([]int, len(items))
	pending := make([]int, len(items)) // индексы исходного списка, ждущие отправки
	for i := range pending {
		pending[i] = i
	}

	for len(pending) > 0 {
		if err := ctx.Err(); err != nil {
			for _, i := range pending {
				final[i] = &ItemError{ID: items[i].ID, Attempts: attempts[i], Err: err}
			}
			break
		}
		batch := make([]Item, len(pending))
		for j, i := range pending {
			batch[j] = items[i]
			attempts[i]++
		}

		err := send(ctx, batch)
		var be *BatchError
		isBatch := errors.As(err, &be)

		var next []int
		for j, i := range pending { // j — индекс в пачке, i — в исходном списке
			var ierr error
			switch {
			case err == nil:
			case isBatch:
				ierr = be.Failed[j]
			default:
				ierr = err
			}
			if ierr == nil {
				continue // принят
			}
			if isTemporary(ierr) && attempts[i] < maxAttempts {
				next = append(next, i)
			} else {
				final[i] = &ItemError{ID: items[i].ID, Attempts: attempts[i], Err: ierr}
			}
		}
		pending = next
	}
	return errors.Join(final...)
}

func isTemporary(err error) bool {
	var t interface{ Temporary() bool }
	return errors.As(err, &t) && t.Temporary()
}
