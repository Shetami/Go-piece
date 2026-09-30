package main

import (
	"errors"
	"io"
	"sync"
)

// ErrBodyClosed — Read после Close.
var ErrBodyClosed = errors.New("body: read after close")

type body struct {
	rc     io.ReadCloser
	onDone func(n int64, err error)
	n      int64
	closed bool

	doneOnce  sync.Once
	closeOnce sync.Once
}

// NewBody оборачивает rc (например, тело HTTP-ответа) для метрик.
//   - Считает байты, отданные наружу через Read.
//   - Вызывает onDone(n, err) ровно один раз — при первом из событий:
//     Read вернул io.EOF (err = nil), Read вернул другую ошибку (err = она),
//     вызван Close (err = nil). n — сколько байт отдано к этому моменту.
//     onDone может быть nil.
//   - Close закрывает rc ровно один раз и возвращает его ошибку;
//     повторные Close возвращают nil. Close можно звать параллельно из
//     нескольких горутин (Read и Close параллельно не вызываются).
//   - Read после Close возвращает 0, ErrBodyClosed и не трогает rc.
func NewBody(rc io.ReadCloser, onDone func(n int64, err error)) io.ReadCloser {
	return &body{rc: rc, onDone: onDone}
}

func (b *body) done(err error) {
	b.doneOnce.Do(func() {
		if b.onDone != nil {
			b.onDone(b.n, err)
		}
	})
}

func (b *body) Read(p []byte) (int, error) {
	if b.closed {
		return 0, ErrBodyClosed
	}
	n, err := b.rc.Read(p)
	b.n += int64(n) // сначала учесть данные: n > 0 может прийти вместе с EOF
	switch {
	case err == io.EOF:
		b.done(nil)
	case err != nil:
		b.done(err)
	}
	return n, err
}

func (b *body) Close() error {
	var err error // локальная: ошибку увидит только тот, кто реально закрыл
	b.closeOnce.Do(func() {
		b.closed = true
		b.done(nil) // если EOF или ошибка уже были — Once ничего не сделает
		err = b.rc.Close()
	})
	return err
}
