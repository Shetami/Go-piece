package main

// Stats — сводка по добавленным значениям. Пока значений нет — все поля нулевые.
type Stats struct {
	Count          int
	Min, Max, Mean float64
}

// NewStats возвращает пару функций над общим состоянием:
// add добавляет значение (NaN игнорируется), snapshot возвращает текущую сводку.
// Каждый вызов NewStats — независимый счётчик.
func NewStats() (add func(x float64), snapshot func() Stats) {
	// ваш код
	return func(float64) {}, func() Stats { return Stats{} }
}
