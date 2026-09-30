package main

import (
	"errors"
	"fmt"
	"io"
)

// CopyAndClose копирует src в dst и закрывает оба — на любом пути,
// в том числе при панике внутри Read или Write (паника летит дальше).
// Порядок закрытия: сначала dst, потом src.
//
// n — сколько байт скопировано, даже если потом что-то не закрылось.
// err — errors.Join всех ошибок, каждая подписана:
// "copy: …", "close dst: …", "close src: …". Нет ошибок — nil.
func CopyAndClose(dst io.WriteCloser, src io.ReadCloser) (n int64, err error) {
	// defer выполняются в обратном порядке: src откладываем первым,
	// чтобы dst закрылся раньше него.
	defer func() {
		if cerr := src.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("close src: %w", cerr))
		}
	}()
	defer func() {
		// Close у писателя — часто последний Flush: его ошибку терять нельзя.
		if cerr := dst.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("close dst: %w", cerr))
		}
	}()

	n, err = io.Copy(dst, src)
	if err != nil {
		err = fmt.Errorf("copy: %w", err)
	}
	return n, err
}
