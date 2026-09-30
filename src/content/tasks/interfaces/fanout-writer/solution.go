package main

import (
	"errors"
	"fmt"
	"io"
)

// ErrNoWriters — живых получателей не осталось ещё до вызова Write.
var ErrNoWriters = errors.New("fanout: no writers left")

type target struct {
	idx int // индекс в исходном списке — для текста ошибки
	w   io.Writer
}

// Fanout рассылает каждую запись во все живые writer'ы.
type Fanout struct {
	live []target
}

// NewFanout создаёт рассыльщик. nil в списке пропускаются, но индексы
// остальных считаются по исходному списку.
func NewFanout(ws ...io.Writer) *Fanout {
	f := &Fanout{}
	for i, w := range ws {
		if w != nil {
			f.live = append(f.live, target{i, w})
		}
	}
	return f
}

// Write пишет p во все живые writer'ы по порядку.
//   - Writer, вернувший ошибку или записавший меньше len(p) (тогда ошибка —
//     io.ErrShortWrite), выбывает навсегда.
//   - Ошибки вызова собираются через errors.Join, каждая обёрнута как
//     fmt.Errorf("writer %d: %w", i, err), i — индекс в списке NewFanout.
//   - Если хотя бы один writer принял данные — n = len(p), иначе n = 0.
//   - Если живых нет ещё до вызова — 0, ErrNoWriters.
func (f *Fanout) Write(p []byte) (int, error) {
	if len(f.live) == 0 {
		return 0, ErrNoWriters
	}
	var errs []error
	kept := f.live[:0] // фильтрация на месте: выбывшие просто не копируются
	for _, t := range f.live {
		n, err := t.w.Write(p)
		if err == nil && n < len(p) {
			err = io.ErrShortWrite // недописал молча — всё равно нарушение контракта
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("writer %d: %w", t.idx, err))
			continue
		}
		kept = append(kept, t)
	}
	clear(f.live[len(kept):]) // не держим ссылки на выбывших
	f.live = kept
	if len(kept) == 0 {
		return 0, errors.Join(errs...)
	}
	return len(p), errors.Join(errs...)
}

// Alive — сколько writer'ов ещё получают данные.
func (f *Fanout) Alive() int { return len(f.live) }
