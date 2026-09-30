package main

import (
	"context"
	"sync"
)

// Group — группа горутин с общей отменой, как errgroup.WithContext.
type Group struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu       sync.Mutex
	err      error
	panicked bool
	pval     any
}

// WithContext возвращает группу и производный контекст. Контекст
// отменяется при первой ошибке или панике в группе, а также после Wait.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

// Go запускает fn в новой горутине. Паника в fn не роняет процесс сразу:
// она запоминается и отменяет контекст группы.
func (g *Group) Go(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				g.mu.Lock()
				if !g.panicked { // запоминаем только первую
					g.panicked, g.pval = true, r
				}
				g.mu.Unlock()
				g.cancel() // остальные должны узнать об аварии как можно раньше
			}
		}()
		if err := fn(); err != nil {
			g.mu.Lock()
			if g.err == nil {
				g.err = err
			}
			g.mu.Unlock()
			g.cancel()
		}
	}()
}

// Wait ждёт завершения всех горутин группы.
//   - Если хоть одна запаниковала — Wait паникует в горутине вызывающего
//     значением первой паники (тем же самым значением), уже после того,
//     как дождался всех.
//   - Иначе возвращает первую ошибку (или nil).
func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel()
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.panicked {
		// Паника переезжает в горутину вызывающего: там её можно
		// перехватить, а стек покажет, кто ждал группу.
		panic(g.pval)
	}
	return g.err
}
