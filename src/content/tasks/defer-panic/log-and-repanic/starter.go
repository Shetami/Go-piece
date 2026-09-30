package main

// LogPanic вызывает fn. Если fn паникует — один раз вызывает logf с
// сообщением, в котором есть значение паники и стек (runtime/debug.Stack),
// а потом паникует дальше тем же самым значением — не строкой, не обёрткой.
// Без паники logf не вызывается.
func LogPanic(logf func(msg string), fn func()) {
	// ваш код
	fn()
}
