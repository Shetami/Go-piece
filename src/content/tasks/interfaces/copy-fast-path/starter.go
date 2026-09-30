package main

import (
	"errors"
	"io"
)

// ErrInvalidWrite — writer вернул n < 0 или n > len(p).
var ErrInvalidWrite = errors.New("invalid write result")

// Copy копирует src в dst до EOF и возвращает число скопированных байт.
// io.Copy, io.CopyBuffer и io.CopyN использовать нельзя. Порядок как у io.Copy:
//  1. src умеет io.WriterTo — вся работа ему: src.WriteTo(dst);
//  2. иначе dst умеет io.ReaderFrom — dst.ReadFrom(src);
//  3. иначе — цикл через буфер 32 КБ:
//     - данные, пришедшие вместе с ошибкой (включая EOF), сначала пишутся;
//     - io.EOF из src — нормальное завершение, err = nil;
//     - другая ошибка src возвращается как есть;
//     - Read вернул 0, nil — просто читать дальше;
//     - writer вернул n < 0 или n > len — ErrInvalidWrite;
//     - записал меньше и без ошибки — io.ErrShortWrite.
func Copy(dst io.Writer, src io.Reader) (int64, error) {
	// ваш код
	return 0, nil
}
