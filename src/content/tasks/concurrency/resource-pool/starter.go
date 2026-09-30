package main

import (
	"context"
	"errors"
)

var ErrPoolClosed = errors.New("pool closed")

// Pool — пул переиспользуемых ресурсов (как пул соединений к базе).
//
//   - Одновременно существует не больше max ресурсов (выданных + свободных).
//   - Get(ctx): вернуть свободный ресурс, если он есть; иначе создать новый
//     через factory, если лимит не исчерпан; иначе ждать Put, отмены ctx
//     (→ ctx.Err()) или Close (→ ErrPoolClosed).
//   - factory может быть медленной: пока она работает, остальные Get и Put
//     не блокируются. Если factory вернула ошибку, место под ресурс
//     освобождается.
//   - Put(r) возвращает ресурс в пул.
//   - Close закрывает свободные ресурсы через closeFn; ждущие и новые Get
//     получают ErrPoolClosed; ресурсы, возвращённые Put после Close,
//     закрываются сразу. Повторный Close безопасен.
type Pool[R any] struct {
	factory func(ctx context.Context) (R, error)
	closeFn func(R)
	// ваши поля
}

func NewPool[R any](max int, factory func(ctx context.Context) (R, error), closeFn func(R)) *Pool[R] {
	return &Pool[R]{factory: factory, closeFn: closeFn}
}

func (p *Pool[R]) Get(ctx context.Context) (R, error) {
	// ваш код
	return p.factory(ctx)
}

func (p *Pool[R]) Put(r R) {
	// ваш код
}

func (p *Pool[R]) Close() {
	// ваш код
}
