package main

import "context"

// Report — итог обработки: каждый id попадает ровно в одно из трёх мест.
type Report struct {
	Done    []int         // обработаны успешно, по возрастанию
	Failed  map[int]error // handle вернула свою ошибку (не из-за отмены ctx); не nil
	Skipped []int         // не начаты или прерваны отменой ctx, по возрастанию
}

// ProcessAll обрабатывает ids в n горутинах (n >= 1): не больше n вызовов
// handle одновременно. После отмены ctx новые задания не начинаются.
// Задание считается прерванным (Skipped), если handle вернула ошибку,
// для которой errors.Is(err, context.Canceled) или DeadlineExceeded,
// И при этом сам ctx уже отменён. Иначе ошибка — Failed.
// Возвращает отчёт и context.Cause(ctx), если хоть одно задание пропущено,
// иначе nil. Возвращается только после выхода всех своих горутин.
func ProcessAll(ctx context.Context, ids []int, n int, handle func(context.Context, int) error) (Report, error) {
	// ваш код
	return Report{}, nil
}
