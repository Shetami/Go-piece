package main

import (
	"runtime"
	"strings"
)

// Location — место в коде.
type Location struct {
	Func string // полное имя функции, как в runtime.Frame.Function
	File string
	Line int
}

// PanicSite вызывает fn. Если fn запаниковала, возвращает значение паники,
// panicked=true и место паники: функцию, файл и строку, где вызван panic
// или где случилась паника рантайма (запись в nil-мапу, деление на ноль…).
// Кадры самого рантайма (runtime.*) местом паники не считаются.
// Если паник было несколько (паника внутри отложенного вызова во время
// другой паники) — место последней, той, что вернул recover.
// Без паники — нулевые значения и panicked=false.
func PanicSite(fn func()) (loc Location, value any, panicked bool) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		value, panicked = r, true
		loc = panicLocation()
	}()
	fn()
	return Location{}, nil, false
}

// panicLocation ищет в текущем стеке самый верхний runtime.gopanic
// (последнюю панику) и берёт первый кадр под ним, не принадлежащий
// рантайму. Работает, пока мы внутри отложенного вызова: кадры
// паникующей функции ещё не сняты.
func panicLocation() Location {
	pcs := make([]uintptr, 128)
	n := runtime.Callers(0, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	afterPanic := false
	for {
		f, more := frames.Next()
		if f.Function == "runtime.gopanic" {
			afterPanic = true
		} else if afterPanic && !strings.HasPrefix(f.Function, "runtime.") {
			// Пропускаем runtime.panicmem, runtime.mapassign, runtime.panicdivide…
			return Location{Func: f.Function, File: f.File, Line: f.Line}
		}
		if !more {
			return Location{}
		}
	}
}
