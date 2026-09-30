package main

import (
	"context"
	"sync"
)

type call[V any] struct {
	done    chan struct{} // закрыт — val и err готовы
	val     V
	err     error
	waiters int
	cancel  context.CancelFunc
}

// Group объединяет одновременные вызовы Do с одинаковым ключом в один
// вызов fn. Нулевое значение готово к работе.
type Group[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]*call[V]
}

// Do возвращает результат fn для key; если такой вызов уже идёт —
// присоединяется к нему и не вызывает fn повторно.
//   - fn получает контекст, который видит значения первого вызывающего, но
//     НЕ отменяется, когда уходит кто-то один из ждущих;
//   - каждый вызывающий ждёт, пока жив его ctx; отменили — сразу получает
//     context.Cause(ctx), остальные продолжают ждать;
//   - когда ушли ВСЕ ждущие, контекст fn отменяется, а ключ освобождается:
//     следующий Do начнёт новый вызов, а не присоединится к брошенному;
//   - после завершения fn ключ освобождается (результат не кэшируется).
func (g *Group[K, V]) Do(ctx context.Context, key K, fn func(context.Context) (V, error)) (V, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[K]*call[V])
	}
	c, ok := g.m[key]
	if !ok {
		// Отмена первого вызывающего не должна убивать общий вызов:
		// отвязываемся от его отмены, но оставляем свою — по счётчику.
		fctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
		c = &call[V]{done: make(chan struct{}), cancel: cancel}
		g.m[key] = c
		go func() {
			v, err := fn(fctx)
			c.val, c.err = v, err // до close(done): читатели увидят их после <-done
			g.mu.Lock()
			if g.m[key] == c { // ключ могли уже освободить и занять новым вызовом
				delete(g.m, key)
			}
			g.mu.Unlock()
			cancel()
			close(c.done)
		}()
	}
	c.waiters++
	g.mu.Unlock()

	select {
	case <-c.done:
		return c.val, c.err
	case <-ctx.Done():
		g.mu.Lock()
		c.waiters--
		if c.waiters == 0 {
			if g.m[key] == c {
				delete(g.m, key) // брошенный вызов больше никому не отдаём
			}
			c.cancel()
		}
		g.mu.Unlock()
		var zero V
		return zero, context.Cause(ctx)
	}
}
