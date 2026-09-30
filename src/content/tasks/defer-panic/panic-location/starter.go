package main

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
	// ваш код
	fn()
	return Location{}, nil, false
}
