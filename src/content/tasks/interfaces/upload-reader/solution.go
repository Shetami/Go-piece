package main

import (
	"errors"
	"io"
)

var ErrTooLarge = errors.New("upload: too large")

type upload struct {
	r    io.Reader
	left int64 // сколько байт ещё можно отдать
	sink io.Writer
	err  error // запомненная окончательная ошибка
}

// Upload возвращает Reader, который отдаёт данные r и одновременно пишет
// каждую отданную порцию в sink (например, в hash.Hash или файл-копию).
//   - Не больше max байт. Если в r ровно max байт — обычный io.EOF; если
//     больше — после max байт Read возвращает 0, ErrTooLarge (и дальше тоже).
//   - Байты сверх max не попадают ни к вызывающему, ни в sink.
//   - Из r читается не больше max+1 байт.
//   - Ошибка sink: Read возвращает n и эту ошибку (как io.TeeReader).
//   - Прочие ошибки r возвращаются как есть.
func Upload(r io.Reader, max int64, sink io.Writer) io.Reader {
	return &upload{r: r, left: max, sink: sink}
}

func (u *upload) Read(p []byte) (int, error) {
	if u.err != nil {
		return 0, u.err
	}
	if len(p) == 0 {
		return 0, nil
	}
	if u.left == 0 {
		// Лимит исчерпан: пробуем прочитать ровно один байт, чтобы понять,
		// кончился ли источник. Сам байт никуда не отдаём.
		var one [1]byte
		n, err := io.ReadFull(u.r, one[:])
		switch {
		case n == 1:
			u.err = ErrTooLarge
		case err == io.EOF:
			u.err = io.EOF
		default:
			return 0, err
		}
		return 0, u.err
	}
	if int64(len(p)) > u.left {
		p = p[:u.left]
	}
	n, err := u.r.Read(p)
	u.left -= int64(n)
	if n > 0 {
		if _, werr := u.sink.Write(p[:n]); werr != nil {
			return n, werr
		}
	}
	return n, err
}
