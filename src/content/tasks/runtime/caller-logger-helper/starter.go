package main

import "io"

// Logger пишет строки вида "file.go:42: сообщение\n", где file.go:42 —
// место в коде, откуда вызвали Log/Logf (только имя файла, без каталогов).
// Как в testing.T, функции можно пометить вспомогательными через Helper:
// тогда место вызова ищется выше по стеку — в первой непомеченной функции.
// Безопасен для конкурентного использования; каждая строка пишется в w
// одним вызовом Write.
type Logger struct {
	// ваши поля
}

func NewLogger(w io.Writer) *Logger {
	// ваш код
	return &Logger{}
}

// Log пишет msg с местом вызова.
func (l *Logger) Log(msg string) {
	// ваш код
}

// Logf — как Log, но с форматированием fmt.Sprintf. Место вызова —
// там, где вызвали Logf, а не внутри Logger.
func (l *Logger) Logf(format string, args ...any) {
	// ваш код
}

// Helper помечает функцию, из которой он вызван, как вспомогательную
// (для всех последующих вызовов Log/Logf из любых горутин).
func (l *Logger) Helper() {
	// ваш код
}
