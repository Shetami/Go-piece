package main

import (
	"bytes"
	"io"
)

// IndentWriter — io.Writer, который добавляет Prefix в начало каждой
// строки вывода и пишет результат в нижележащий writer. Так печатают
// вложенные блоки: вывод подкоманды с отступом, YAML, дерево ошибок.
type IndentWriter struct {
	w      io.Writer
	prefix []byte
	bol    bool // «начало строки»: следующий байт — первый в строке
}

// NewIndentWriter создаёт IndentWriter поверх w.
//
//   - префикс пишется перед первым байтом каждой строки — лениво: если
//     вывод кончается на "\n", висящего префикса в конце нет;
//   - строка может прийти по кускам в разных вызовах Write, а один Write
//     может содержать много строк — префикс всё равно ровно один на строку;
//   - пустые строки остаются пустыми: перед "\n" префикс не пишется;
//   - IndentWriter можно вкладывать друг в друга — отступы складываются.
func NewIndentWriter(w io.Writer, prefix string) *IndentWriter {
	return &IndentWriter{w: w, prefix: []byte(prefix), bol: true}
}

// Write по контракту io.Writer возвращает число байт из p (префиксы не в
// счёт): n == len(p), если ошибки нет, и n < len(p) вместе с ошибкой
// нижележащего writer, если она случилась.
func (iw *IndentWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		if iw.bol && p[0] != '\n' {
			if _, err := iw.w.Write(iw.prefix); err != nil {
				return written, err
			}
			iw.bol = false
		}
		// Кусок до конца строки включительно или до конца p.
		end := len(p)
		if i := bytes.IndexByte(p, '\n'); i >= 0 {
			end = i + 1
		}
		n, err := iw.w.Write(p[:end])
		written += n
		if err != nil {
			return written, err
		}
		if n < end {
			return written, io.ErrShortWrite
		}
		iw.bol = p[end-1] == '\n'
		p = p[end:]
	}
	return written, nil
}
