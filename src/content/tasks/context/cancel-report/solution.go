package main

import (
	"context"
	"errors"
	"slices"
	"sync"
)

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
	// Все задания сразу в буферизованный канал: воркеры разбирают его до конца
	// и сами решают, делать задание или пропустить, — так ни одно не потеряется.
	jobs := make(chan int, len(ids))
	for _, id := range ids {
		jobs <- id
	}
	close(jobs)

	rep := Report{Failed: map[int]error{}}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				if ctx.Err() != nil { // отменено — не начинаем, но и не теряем
					mu.Lock()
					rep.Skipped = append(rep.Skipped, id)
					mu.Unlock()
					continue
				}
				err := handle(ctx, id)
				mu.Lock()
				switch {
				case err == nil:
					rep.Done = append(rep.Done, id)
				case ctx.Err() != nil && isCtxErr(err):
					rep.Skipped = append(rep.Skipped, id) // прервано нашей отменой
				default:
					rep.Failed[id] = err // в том числе чужой Canceled при живом ctx
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	slices.Sort(rep.Done)
	slices.Sort(rep.Skipped)
	if len(rep.Skipped) > 0 {
		return rep, context.Cause(ctx)
	}
	return rep, nil
}

func isCtxErr(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
