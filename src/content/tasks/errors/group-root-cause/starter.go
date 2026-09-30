package main

import "context"

// Group запускает задачи параллельно и сообщает о сбое так, чтобы в логе
// была первопричина, а не сотня «context canceled» от отменённых соседей.
type Group struct {
	// ваш код
}

// WithContext возвращает группу и производный от ctx контекст для задач.
// Контекст отменяется при первой ошибке любой задачи, и context.Cause(ctx)
// возвращает именно эту первую ошибку. После Wait контекст отменён в любом
// случае.
func WithContext(ctx context.Context) (*Group, context.Context) {
	// ваш код
	return &Group{}, ctx
}

// Go запускает f в отдельной горутине.
func (g *Group) Go(f func() error) {
	// ваш код
}

// Wait ждёт завершения всех задач и возвращает:
//   - nil, если ни одна задача не вернула ошибку;
//   - иначе errors.Join(первая ошибка, остальные настоящие ошибки в порядке
//     поступления).
//
// Ошибки, пришедшие после первой, отбрасываются как следствия отмены, если
// они errors.Is(err, context.Canceled) или errors.Is(err, первая ошибка).
func (g *Group) Wait() error {
	// ваш код
	return nil
}
