package main

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

// Logger пишет строки вида "file.go:42: сообщение\n", где file.go:42 —
// место в коде, откуда вызвали Log/Logf (только имя файла, без каталогов).
// Как в testing.T, функции можно пометить вспомогательными через Helper:
// тогда место вызова ищется выше по стеку — в первой непомеченной функции.
// Безопасен для конкурентного использования; каждая строка пишется в w
// одним вызовом Write.
type Logger struct {
	mu      sync.Mutex
	w       io.Writer
	helpers map[string]struct{} // полные имена функций, например main.assertOK
}

func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w, helpers: make(map[string]struct{})}
}

// Log пишет msg с местом вызова.
func (l *Logger) Log(msg string) { l.output(msg) }

// Logf — как Log, но с форматированием fmt.Sprintf. Место вызова —
// там, где вызвали Logf, а не внутри Logger.
func (l *Logger) Logf(format string, args ...any) { l.output(fmt.Sprintf(format, args...)) }

// Helper помечает функцию, из которой он вызван, как вспомогательную
// (для всех последующих вызовов Log/Logf из любых горутин).
func (l *Logger) Helper() {
	var pc [1]uintptr
	// 0 — сам runtime.Callers, 1 — Helper, 2 — функция, которую помечаем.
	if runtime.Callers(2, pc[:]) == 0 {
		return
	}
	// CallersFrames правильно раскрывает встроенные (inlined) функции.
	f, _ := runtime.CallersFrames(pc[:]).Next()
	l.mu.Lock()
	l.helpers[f.Function] = struct{}{}
	l.mu.Unlock()
}

// output вызывается ровно из Log или Logf: пропускаем Callers, output и Log/Logf.
func (l *Logger) output(msg string) {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	l.mu.Lock()
	defer l.mu.Unlock()
	file, line := "???", 0
	for {
		f, more := frames.Next()
		if _, ok := l.helpers[f.Function]; !ok || !more {
			file, line = filepath.Base(f.File), f.Line
			break
		}
	}
	// Собираем строку целиком и пишем одним Write под мьютексом —
	// строки из разных горутин не перемешаются.
	b := make([]byte, 0, len(file)+len(msg)+16)
	b = append(b, file...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(line), 10)
	b = append(b, ": "...)
	b = append(b, msg...)
	b = append(b, '\n')
	l.w.Write(b)
}
