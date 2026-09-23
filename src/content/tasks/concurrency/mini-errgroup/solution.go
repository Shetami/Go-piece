package main

import (
	"context"
	"sync"
)

// Group запускает задачи в горутинах и собирает первую ошибку.
type Group struct {
	wg     sync.WaitGroup
	once   sync.Once
	err    error
	cancel context.CancelFunc
}

// WithContext возвращает группу и контекст, который отменяется,
// как только одна из задач вернула ошибку (или после Wait).
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

// Go запускает f в новой горутине.
func (g *Group) Go(f func() error) {
	g.wg.Go(func() {
		if err := f(); err != nil {
			// Once: запоминаем только первую ошибку, и отмена — один раз.
			g.once.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	})
}

// Wait ждёт все задачи и возвращает первую ошибку (или nil).
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
