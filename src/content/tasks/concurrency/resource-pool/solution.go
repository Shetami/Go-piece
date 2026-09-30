package main

import (
	"context"
	"errors"
	"sync"
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

	idle   chan R        // свободные ресурсы, cap = max
	slots  chan struct{} // один токен на существующий ресурс, cap = max
	closed chan struct{}

	mu       sync.Mutex // Put/Close: не положить в idle после его очистки
	isClosed bool
}

func NewPool[R any](max int, factory func(ctx context.Context) (R, error), closeFn func(R)) *Pool[R] {
	return &Pool[R]{
		factory: factory,
		closeFn: closeFn,
		idle:    make(chan R, max),
		slots:   make(chan struct{}, max),
		closed:  make(chan struct{}),
	}
}

func (p *Pool[R]) Get(ctx context.Context) (R, error) {
	var zero R
	select {
	case <-p.closed:
		return zero, ErrPoolClosed
	default:
	}
	// Сначала — свободный ресурс без ожидания: не создаём лишних.
	select {
	case r := <-p.idle:
		return r, nil
	default:
	}
	select {
	case r := <-p.idle:
		return r, nil
	case p.slots <- struct{}{}:
		// Место под новый ресурс заняли; создаём БЕЗ блокировок пула.
		r, err := p.factory(ctx)
		if err != nil {
			<-p.slots
			return zero, err
		}
		return r, nil
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-p.closed:
		return zero, ErrPoolClosed
	}
}

func (p *Pool[R]) Put(r R) {
	p.mu.Lock()
	if p.isClosed {
		p.mu.Unlock()
		p.closeFn(r)
		<-p.slots
		return
	}
	p.idle <- r // не блокируется: ресурсов не больше cap(idle)
	p.mu.Unlock()
}

func (p *Pool[R]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClosed {
		return
	}
	p.isClosed = true
	close(p.closed)
	for {
		select {
		case r := <-p.idle:
			p.closeFn(r)
			<-p.slots
		default:
			return
		}
	}
}
