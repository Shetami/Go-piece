package main

import "math"

// Stats — сводка по добавленным значениям. Пока значений нет — все поля нулевые.
type Stats struct {
	Count          int
	Min, Max, Mean float64
}

// NewStats возвращает пару функций над общим состоянием:
// add добавляет значение (NaN игнорируется), snapshot возвращает текущую сводку.
// Каждый вызов NewStats — независимый счётчик.
func NewStats() (add func(x float64), snapshot func() Stats) {
	var (
		n      int
		sum    float64
		lo, hi float64
	)
	add = func(x float64) {
		if math.IsNaN(x) {
			return
		}
		// Первое значение задаёт и минимум, и максимум: ноль как старт
		// сломал бы Max для отрицательных и Min для положительных.
		if n == 0 || x < lo {
			lo = x
		}
		if n == 0 || x > hi {
			hi = x
		}
		n++
		sum += x
	}
	snapshot = func() Stats {
		if n == 0 {
			return Stats{}
		}
		return Stats{Count: n, Min: lo, Max: hi, Mean: sum / float64(n)}
	}
	return add, snapshot
}
