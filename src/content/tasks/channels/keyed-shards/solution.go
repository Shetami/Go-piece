package main

import (
	"context"
	"hash/fnv"
	"sync"
)

// Task — задача с ключом (например, id заказа или пользователя).
type Task struct {
	Key string
	Seq int
}

// Dispatch раздаёт задачи из in по n воркерам (n >= 1) так, что задачи с
// одним ключом всегда попадают к одному воркеру и обрабатываются строго в
// порядке поступления, а задачи с разными ключами могут идти параллельно.
// Одновременно выполняется не больше n вызовов handle.
//
// Возвращает nil, когда in закрыт и все задачи обработаны. Если ctx
// отменён раньше — ждёт, пока вернутся уже начатые handle, и возвращает
// ctx.Err(); необработанные задачи после отмены можно пропустить.
func Dispatch(ctx context.Context, in <-chan Task, n int, handle func(Task)) error {
	// У каждого воркера свой канал: один ключ — один шард — один
	// последовательный читатель, отсюда и порядок внутри ключа.
	shards := make([]chan Task, n)
	var wg sync.WaitGroup
	for i := range shards {
		shards[i] = make(chan Task, 16)
		wg.Add(1)
		go func(ch <-chan Task) {
			defer wg.Done()
			for t := range ch {
				if ctx.Err() != nil {
					continue // после отмены только дочитываем, чтобы роутер не встал
				}
				handle(t)
			}
		}(shards[i])
	}

	route := func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case t, ok := <-in:
				if !ok {
					return nil
				}
				h := fnv.New32a()
				h.Write([]byte(t.Key))
				select {
				case shards[h.Sum32()%uint32(n)] <- t:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
	err := route()

	// Роутер — единственный писатель в шарды, он их и закрывает.
	for _, ch := range shards {
		close(ch)
	}
	wg.Wait()
	if err == nil {
		err = ctx.Err() // отмена, случившаяся во время дообработки хвоста
	}
	return err
}
