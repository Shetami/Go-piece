package main

import (
	"context"
	"errors"
	"sync"
)

var ErrPoolClosed = errors.New("pool closed")

type Conn struct{ ID int }

// Pool — пул соединений. Свободные соединения лежат в буферизованном
// канале. Открыто (выдано + свободно) одновременно не больше max.
type Pool struct {
	dial      func(context.Context) (*Conn, error)
	closeConn func(*Conn)

	idle  chan *Conn    // свободные соединения
	slots chan struct{} // семафор: один элемент — одно открытое соединение

	mu       sync.Mutex // согласует Put и Close: иначе Put положит в idle после чистки
	closed   bool
	closedCh chan struct{}
}

// NewPool создаёт пустой пул: соединения открываются лениво через dial
// и закрываются через closeConn.
func NewPool(max int, dial func(context.Context) (*Conn, error), closeConn func(*Conn)) *Pool {
	return &Pool{
		dial:      dial,
		closeConn: closeConn,
		idle:      make(chan *Conn, max),
		slots:     make(chan struct{}, max),
		closedCh:  make(chan struct{}),
	}
}

// Get выдаёт соединение. Свободное берётся в первую очередь; если
// свободных нет, а открыто меньше max — открывается новое. Иначе Get ждёт
// возврата соединения, отмены ctx (ctx.Err()) или закрытия пула
// (ErrPoolClosed). Ошибка dial возвращается как есть и место не занимает.
func (p *Pool) Get(ctx context.Context) (*Conn, error) {
	select {
	case <-p.closedCh:
		return nil, ErrPoolClosed
	default:
	}
	// Сначала — только свободное: в общем select ветка «открыть новое»
	// выигрывала бы наравне, и пул плодил бы соединения при свободных.
	select {
	case c := <-p.idle:
		return c, nil
	default:
	}
	select {
	case c := <-p.idle:
		return c, nil
	case p.slots <- struct{}{}:
		c, err := p.dial(ctx)
		if err != nil {
			<-p.slots // неудачный dial не должен съесть место навсегда
			return nil, err
		}
		return c, nil
	case <-p.closedCh:
		return nil, ErrPoolClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Put возвращает соединение. broken = true — соединение сломано: оно
// закрывается, и его место освобождается для нового. После Close любое
// возвращённое соединение закрывается.
func (p *Pool) Put(c *Conn, broken bool) {
	p.mu.Lock()
	if broken || p.closed {
		p.mu.Unlock()
		p.closeConn(c)
		<-p.slots // освобождаем место — ждущий Get сможет открыть новое
		return
	}
	p.idle <- c // не блокируется: в idle не больше max, а буфер — max
	p.mu.Unlock()
}

// Close закрывает пул: свободные соединения закрываются сразу, ждущие Get
// получают ErrPoolClosed, новые Get — тоже. Повторный Close безопасен.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	p.closed = true
	close(p.closedCh)
	for {
		select {
		case c := <-p.idle:
			p.closeConn(c)
			<-p.slots
		default:
			return
		}
	}
}
