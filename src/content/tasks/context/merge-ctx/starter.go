package main

import "context"

// Merge возвращает контекст, который отменяется, когда отменён a ИЛИ b,
// или когда вызвана возвращённая cancel. Значения и дедлайн — от a.
// context.Cause результата — причина того, кто отменился первым.
// После cancel никакие горутины и регистрации Merge не остаются.
func Merge(a, b context.Context) (context.Context, context.CancelFunc) {
	// ваш код
	return context.WithCancel(a)
}
