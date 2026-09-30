package main

import (
	"errors"
	"io"
)

var ErrTooLarge = errors.New("upload: too large")

// Upload возвращает Reader, который отдаёт данные r и одновременно пишет
// каждую отданную порцию в sink (например, в hash.Hash или файл-копию).
//   - Не больше max байт. Если в r ровно max байт — обычный io.EOF; если
//     больше — после max байт Read возвращает 0, ErrTooLarge (и дальше тоже).
//   - Байты сверх max не попадают ни к вызывающему, ни в sink.
//   - Из r читается не больше max+1 байт.
//   - Ошибка sink: Read возвращает n и эту ошибку (как io.TeeReader).
//   - Прочие ошибки r возвращаются как есть.
func Upload(r io.Reader, max int64, sink io.Writer) io.Reader {
	// ваш код
	return r
}
