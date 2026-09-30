package main

import (
	"fmt"
	"runtime/debug"
)

// LogPanic вызывает fn. Если fn паникует — один раз вызывает logf с
// сообщением, в котором есть значение паники и стек (runtime/debug.Stack),
// а потом паникует дальше тем же самым значением — не строкой, не обёрткой.
// Без паники logf не вызывается.
func LogPanic(logf func(msg string), fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// Стек снимаем здесь: кадры паникующей функции ещё на месте.
			logf(fmt.Sprintf("паника: %v\n%s", r, debug.Stack()))
			panic(r) // то же значение — вызывающий может проверить его тип
		}
	}()
	fn()
}
