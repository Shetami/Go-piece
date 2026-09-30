package main

import (
	"context"
	"sync"
)

// Conn — соединение из пула.
type Conn struct {
	ID int
}

// Pool — пул соединений не больше max штук одновременно (открытых:
// выданных плюс простаивающих).
type Pool struct {
	dial func() (*Conn, error)
	max  int

	mu      sync.Mutex
	idle    []*Conn
	open    int
	waiters []chan struct{} // очередь ждущих; сигнал — «что-то изменилось, проверь»
}

// NewPool создаёт пул. dial открывает новое соединение; он медленный
// (сеть), поэтому пул не держит свою блокировку, пока dial работает.
func NewPool(max int, dial func() (*Conn, error)) *Pool {
	return &Pool{max: max, dial: dial}
}

// Get возвращает простаивающее соединение (последнее возвращённое —
// самое «тёплое»), а если их нет и открыто меньше max — открывает новое.
// Иначе ждёт, пока соединение вернут или освободится место. При отмене
// ctx возвращает ctx.Err(). Ошибка dial возвращается и место не занимает.
func (p *Pool) Get(ctx context.Context) (*Conn, error) {
	for {
		p.mu.Lock()
		if n := len(p.idle); n > 0 {
			c := p.idle[n-1] // LIFO
			p.idle = p.idle[:n-1]
			p.mu.Unlock()
			return c, nil
		}
		if p.open < p.max {
			// Место бронируем под мьютексом, а dial — без него.
			p.open++
			p.mu.Unlock()
			c, err := p.dial()
			if err != nil {
				p.release() // бронь не пригодилась — отдаём место ждущему
				return nil, err
			}
			return c, nil
		}
		ch := make(chan struct{}, 1)
		p.waiters = append(p.waiters, ch)
		p.mu.Unlock()

		select {
		case <-ch:
			// Проверяем заново: соединение могли перехватить.
		case <-ctx.Done():
			p.mu.Lock()
			removed := p.removeWaiter(ch)
			p.mu.Unlock()
			if !removed {
				// Сигнал уже был отправлен нам — передаём его следующему,
				// иначе он пропадёт вместе с нами.
				p.wakeOne()
			}
			return nil, ctx.Err()
		}
	}
}

// Put возвращает исправное соединение в пул.
func (p *Pool) Put(c *Conn) {
	p.mu.Lock()
	p.idle = append(p.idle, c)
	p.wakeLocked()
	p.mu.Unlock()
}

// Discard сообщает, что выданное соединение сломано и закрыто: оно
// больше не считается открытым, и его место можно занять новым.
func (p *Pool) Discard(c *Conn) {
	p.release()
}

func (p *Pool) release() {
	p.mu.Lock()
	p.open--
	p.wakeLocked() // освободилось место — ждущий может открыть новое
	p.mu.Unlock()
}

// Open — сколько соединений сейчас открыто (выданных плюс простаивающих).
func (p *Pool) Open() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.open
}

func (p *Pool) wakeOne() {
	p.mu.Lock()
	p.wakeLocked()
	p.mu.Unlock()
}

// wakeLocked будит самого давнего ждущего (FIFO).
func (p *Pool) wakeLocked() {
	if len(p.waiters) == 0 {
		return
	}
	ch := p.waiters[0]
	p.waiters = p.waiters[1:]
	ch <- struct{}{} // буфер 1 — не блокируется
}

func (p *Pool) removeWaiter(ch chan struct{}) bool {
	for i, w := range p.waiters {
		if w == ch {
			p.waiters = append(p.waiters[:i], p.waiters[i+1:]...)
			return true
		}
	}
	return false
}
