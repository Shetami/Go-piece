package main

import (
	"context"
	"errors"
)

var ErrPoolClosed = errors.New("pool closed")

type Conn struct{ ID int }

// Pool — пул соединений. Свободные соединения лежат в буферизованном
// канале. Открыто (выдано + свободно) одновременно не больше max.
type Pool struct {
	// ваши поля
}

// NewPool создаёт пустой пул: соединения открываются лениво через dial
// и закрываются через closeConn.
func NewPool(max int, dial func(context.Context) (*Conn, error), closeConn func(*Conn)) *Pool {
	// ваш код
	return &Pool{}
}

// Get выдаёт соединение. Свободное берётся в первую очередь; если
// свободных нет, а открыто меньше max — открывается новое. Иначе Get ждёт
// возврата соединения, отмены ctx (ctx.Err()) или закрытия пула
// (ErrPoolClosed). Ошибка dial возвращается как есть и место не занимает.
func (p *Pool) Get(ctx context.Context) (*Conn, error) {
	// ваш код
	return nil, nil
}

// Put возвращает соединение. broken = true — соединение сломано: оно
// закрывается, и его место освобождается для нового. После Close любое
// возвращённое соединение закрывается.
func (p *Pool) Put(c *Conn, broken bool) {
	// ваш код
}

// Close закрывает пул: свободные соединения закрываются сразу, ждущие Get
// получают ErrPoolClosed, новые Get — тоже. Повторный Close безопасен.
func (p *Pool) Close() {
	// ваш код
}
