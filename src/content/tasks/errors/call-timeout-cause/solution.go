package main

import (
	"context"
	"fmt"
	"time"
)

// TimeoutError — сработал НАШ таймаут вызова, а не дедлайн вызывающего.
type TimeoutError struct {
	Op    string
	After time.Duration
	Err   error // что вернула f (может быть nil)
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("%s: не уложились в %v", e.Op, e.After)
}

// Unwrap открывает то, что вернула f.
func (e *TimeoutError) Unwrap() error { return e.Err }

// Is делает наш таймаут разновидностью context.DeadlineExceeded, чтобы
// общие проверки «это таймаут?» продолжали работать.
func (e *TimeoutError) Is(target error) bool { return target == context.DeadlineExceeded }

// Call выполняет f с собственным таймаутом d поверх ctx и возвращает
// ошибку, по которой видно, ЧЕЙ это таймаут или отмена:
//   - f вернула nil → nil;
//   - к моменту возврата f родительский ctx уже завершён (отменён или истёк
//     его дедлайн) → ошибка с op в тексте, для которой errors.Is находит и
//     ctx.Err(), и context.Cause(ctx);
//   - иначе, если истёк наш таймаут d → *TimeoutError{op, d, ошибка f};
//   - иначе → ошибка f, обёрнутая как "<op>: %w".
//
// Контекст, который получила f, после возврата Call должен быть отменён.
func Call(ctx context.Context, op string, d time.Duration, f func(context.Context) error) error {
	cctx, cancel := context.WithTimeout(ctx, d)
	defer cancel() // иначе таймер и контекст живут до срабатывания d

	err := f(cctx)
	if err == nil {
		return nil
	}
	// Сначала родитель: если истёк и его дедлайн, и наш, виноват не наш
	// таймаут — вызывающий всё равно уже не ждёт.
	if perr := ctx.Err(); perr != nil {
		cause := context.Cause(ctx)
		if cause == perr {
			return fmt.Errorf("%s: %w", op, perr)
		}
		return fmt.Errorf("%s: %w: %w", op, perr, cause)
	}
	if cctx.Err() != nil {
		return &TimeoutError{Op: op, After: d, Err: err}
	}
	return fmt.Errorf("%s: %w", op, err)
}
