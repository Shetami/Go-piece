package main

import (
	"errors"
	"fmt"
	"io"
)

// Save пишет data в w и закрывает его. Ошибка Close важна: у
// буферизованных писателей и файлов на сетевых дисках данные доходят до
// места только при закрытии.
//   - Close вызывается всегда и ровно один раз, даже если Write упал.
//   - Упал только Write → его ошибка, обёрнутая как "запись: %w".
//   - Упал только Close → "закрытие: %w".
//   - Упали оба → errors.Join(ошибка записи, ошибка закрытия) — с теми же
//     обёртками, сначала запись.
//   - Write записал меньше len(data) без ошибки — это ошибка записи
//     io.ErrShortWrite.
func Save(w io.WriteCloser, data []byte) (err error) {
	// defer с именованным результатом: Close выполнится на любом пути
	// выхода, а его ошибка дополнит, а не затрёт ошибку записи.
	defer func() {
		if cerr := w.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("закрытие: %w", cerr))
		}
	}()
	n, werr := w.Write(data)
	if werr == nil && n < len(data) {
		werr = io.ErrShortWrite
	}
	if werr != nil {
		return fmt.Errorf("запись: %w", werr)
	}
	return nil
}
