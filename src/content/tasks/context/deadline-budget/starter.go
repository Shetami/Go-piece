package main

import (
	"context"
	"errors"
	"time"
)

var ErrNoBudget = errors.New("бюджет времени исчерпан")

// Call — один вызов зависимости.
type Call struct {
	Name string
	Do   func(ctx context.Context) error
}

// CallSequence выполняет calls по очереди и делит между ними оставшееся
// время ctx: вызов получает контекст с таймаутом
// remaining / (сколько вызовов ещё не выполнено, включая этот),
// где remaining считается заново перед каждым вызовом.
// Если у ctx нет дедлайна — вызовы получают ctx без ограничения.
// Если доля вызова меньше minBudget — он не делается, а возвращается ошибка
// с ErrNoBudget и именем вызова. Ошибка вызова — "<имя>: <err>", дальше не идём.
// Контекст каждого вызова освобождается сразу после него.
func CallSequence(ctx context.Context, minBudget time.Duration, calls []Call) error {
	// ваш код
	for _, c := range calls {
		if err := c.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}
