package main

import "io"

// CopyAndClose копирует src в dst и закрывает оба — на любом пути,
// в том числе при панике внутри Read или Write (паника летит дальше).
// Порядок закрытия: сначала dst, потом src.
//
// n — сколько байт скопировано, даже если потом что-то не закрылось.
// err — errors.Join всех ошибок, каждая подписана:
// "copy: …", "close dst: …", "close src: …". Нет ошибок — nil.
func CopyAndClose(dst io.WriteCloser, src io.ReadCloser) (n int64, err error) {
	// ваш код
	return io.Copy(dst, src)
}
