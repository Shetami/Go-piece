package main

import (
	"context"
	"errors"
	"fmt"
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
	for i, c := range calls {
		if err := callOne(ctx, minBudget, len(calls)-i, c); err != nil {
			return err
		}
	}
	return nil
}

func callOne(ctx context.Context, minBudget time.Duration, left int, c Call) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%s: %w", c.Name, context.Cause(ctx))
	}
	callCtx := ctx
	if dl, ok := ctx.Deadline(); ok {
		// Делим то, что осталось сейчас: быстрые вызовы отдают сэкономленное следующим.
		share := time.Until(dl) / time.Duration(left)
		if share < minBudget {
			return fmt.Errorf("%s: доля %v меньше %v: %w", c.Name, share, minBudget, ErrNoBudget)
		}
		var cancel context.CancelFunc
		callCtx, cancel = context.WithTimeout(ctx, share)
		defer cancel()
	}
	if err := c.Do(callCtx); err != nil {
		return fmt.Errorf("%s: %w", c.Name, err)
	}
	return nil
}
