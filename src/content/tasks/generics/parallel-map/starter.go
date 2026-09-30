package main

import "context"

// ParallelMap возвращает out, где out[i] = f(ctx, in[i]), запуская не
// больше limit вызовов f одновременно (limit <= 0 значит 1).
//
// Если какой-то f вернул ошибку, контекст, переданный в f, отменяется,
// новые вызовы не начинаются, ParallelMap дожидается уже запущенных и
// возвращает (nil, ошибку этого f). Если отменён внешний ctx и обход не
// закончен — (nil, ctx.Err()). После возврата не остаётся горутин.
func ParallelMap[T, U any](ctx context.Context, in []T, limit int, f func(context.Context, T) (U, error)) ([]U, error) {
	// ваш код
	return nil, nil
}
