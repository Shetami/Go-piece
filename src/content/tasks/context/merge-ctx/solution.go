package main

import "context"

// Merge возвращает контекст, который отменяется, когда отменён a ИЛИ b,
// или когда вызвана возвращённая cancel. Значения и дедлайн — от a.
// context.Cause результата — причина того, кто отменился первым.
// После cancel никакие горутины и регистрации Merge не остаются.
func Merge(a, b context.Context) (context.Context, context.CancelFunc) {
	// От a наследуем всё: значения, дедлайн и отмену.
	ctx, cancel := context.WithCancelCause(a)
	// От b — только отмену. AfterFunc не держит горутину, пока b жив.
	stop := context.AfterFunc(b, func() {
		cancel(context.Cause(b))
	})
	return ctx, func() {
		stop() // снимаем регистрацию на b: иначе она жила бы, пока жив b
		cancel(context.Canceled)
	}
}
