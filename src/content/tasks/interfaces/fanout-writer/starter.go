package main

import (
	"errors"
	"io"
)

// ErrNoWriters — живых получателей не осталось ещё до вызова Write.
var ErrNoWriters = errors.New("fanout: no writers left")

// Fanout рассылает каждую запись во все живые writer'ы.
type Fanout struct {
	// ваши поля
}

// NewFanout создаёт рассыльщик. nil в списке пропускаются, но индексы
// остальных считаются по исходному списку.
func NewFanout(ws ...io.Writer) *Fanout {
	// ваш код
	return &Fanout{}
}

// Write пишет p во все живые writer'ы по порядку.
//   - Writer, вернувший ошибку или записавший меньше len(p) (тогда ошибка —
//     io.ErrShortWrite), выбывает навсегда.
//   - Ошибки вызова собираются через errors.Join, каждая обёрнута как
//     fmt.Errorf("writer %d: %w", i, err), i — индекс в списке NewFanout.
//   - Если хотя бы один writer принял данные — n = len(p), иначе n = 0.
//   - Если живых нет ещё до вызова — 0, ErrNoWriters.
func (f *Fanout) Write(p []byte) (int, error) {
	// ваш код
	return 0, nil
}

// Alive — сколько writer'ов ещё получают данные.
func (f *Fanout) Alive() int {
	// ваш код
	return 0
}
