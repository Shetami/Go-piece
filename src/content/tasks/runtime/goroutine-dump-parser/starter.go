package main

import "time"

// Goroutine — одна горутина из дампа runtime.Stack(buf, true) или из
// вывода паники/SIGQUIT.
type Goroutine struct {
	ID        int
	State     string        // "chan receive", "select", "running", "chan receive (nil chan)"…
	Wait      time.Duration // "…, 5 minutes]" → 5*time.Minute; нет — 0
	Top       string        // функция верхнего кадра без аргументов: "main.(*Server).handle"
	CreatedBy string        // "main.main" из "created by main.main in goroutine 1"; нет — ""
}

// ParseStacks разбирает дамп. Блоки горутин разделены пустой строкой;
// заголовок блока: "goroutine <id> [<состояние>[, <N> minutes][, locked to thread]]:".
// Некорректный заголовок — ошибка.
func ParseStacks(dump string) ([]Goroutine, error) {
	// ваш код
	return nil, nil
}

// Group — горутины с одинаковыми State и Top.
type Group struct {
	State, Top string
	IDs        []int // по возрастанию
}

// GroupStacks группирует горутины. Порядок групп: по убыванию числа
// горутин, затем по State, затем по Top.
func GroupStacks(gs []Goroutine) []Group {
	// ваш код
	return nil
}
