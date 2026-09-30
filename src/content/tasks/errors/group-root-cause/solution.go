package main

import (
	"context"
	"errors"
	"sync"
)

// Group запускает задачи параллельно и сообщает о сбое так, чтобы в логе
// была первопричина, а не сотня «context canceled» от отменённых соседей.
type Group struct {
	cancel context.CancelCauseFunc
	wg     sync.WaitGroup

	mu    sync.Mutex
	first error
	rest  []error
}

// WithContext возвращает группу и производный от ctx контекст для задач.
// Контекст отменяется при первой ошибке любой задачи, и context.Cause(ctx)
// возвращает именно эту первую ошибку. После Wait контекст отменён в любом
// случае.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}

// Go запускает f в отдельной горутине.
func (g *Group) Go(f func() error) {
	g.wg.Add(1) // до go: иначе Wait может проскочить раньше Add
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.record(err)
		}
	}()
}

func (g *Group) record(err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.first == nil {
		g.first = err
		g.cancel(err) // причина отмены — сама ошибка, её увидит context.Cause
		return
	}
	// Всё, что пришло после, — либо эхо отмены, либо эхо первой ошибки
	// (задачи часто возвращают context.Cause(ctx)). Это не новые сбои.
	if errors.Is(err, context.Canceled) || errors.Is(err, g.first) {
		return
	}
	g.rest = append(g.rest, err)
}

// Wait ждёт завершения всех задач и возвращает:
//   - nil, если ни одна задача не вернула ошибку;
//   - иначе errors.Join(первая ошибка, остальные настоящие ошибки в порядке
//     поступления).
//
// Ошибки, пришедшие после первой, отбрасываются как следствия отмены, если
// они errors.Is(err, context.Canceled) или errors.Is(err, первая ошибка).
func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel(nil) // повторная отмена ничего не меняет: причина остаётся первой
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.first == nil {
		return nil
	}
	return errors.Join(append([]error{g.first}, g.rest...)...)
}
