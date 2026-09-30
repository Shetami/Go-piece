package main

import (
	"errors"
	"io"
)

// ErrBodyClosed — Read после Close.
var ErrBodyClosed = errors.New("body: read after close")

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
	// ваш код
	return rc
}
