package main

import (
	"fmt"
	"sync"
)

// Conn — соединение, которое закрывают из нескольких мест: из defer
// обработчика, из обработчика сигнала, из watchdog'а.
type Conn struct {
	closeFn func() error
	once    sync.Once
	err     error
	done    chan struct{}
}

func NewConn(closeFn func() error) *Conn {
	return &Conn{closeFn: closeFn, done: make(chan struct{})}
}

// Close закрывает соединение.
//   - closeFn вызывается ровно один раз, даже при одновременных Close.
//   - Все вызовы Close — одновременные и последующие — возвращают то же,
//     что вернул closeFn. Одновременные ждут, пока closeFn завершится.
//   - Если closeFn запаниковала — Close не паникует, а возвращает ошибку
//     с текстом паники, и все остальные вызовы возвращают ту же ошибку.
func (c *Conn) Close() error {
	// once.Do ждёт завершения первого вызова во всех горутинах.
	// Но если f запаникует, Once всё равно считается выполненным —
	// поэтому recover стоит внутри f, и err записывается на любом пути.
	c.once.Do(func() {
		defer close(c.done)
		defer func() {
			if r := recover(); r != nil {
				c.err = fmt.Errorf("паника при закрытии: %v", r)
			}
		}()
		c.err = c.closeFn()
	})
	return c.err
}

// Done возвращает канал, который закрывается, когда closeFn завершилась
// (успешно, с ошибкой или паникой).
func (c *Conn) Done() <-chan struct{} {
	return c.done
}
