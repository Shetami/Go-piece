package main

import "time"

// FakeClock — часы для тестов: время идёт только в Advance.
// Безопасен для горутин.
type FakeClock struct {
	// ваши поля
}

func NewFakeClock(start time.Time) *FakeClock {
	// ваш код
	return &FakeClock{}
}

// Now возвращает текущее время часов.
func (c *FakeClock) Now() time.Time {
	// ваш код
	return time.Time{}
}

// AfterFunc планирует f на момент Now()+d (d <= 0 — на текущий момент:
// f сработает при ближайшем Advance, даже Advance(0)).
// stop отменяет таймер и возвращает true, если f ещё не вызывалась и не
// была отменена; иначе false.
func (c *FakeClock) AfterFunc(d time.Duration, f func()) (stop func() bool) {
	// ваш код
	return func() bool { return false }
}

// Advance двигает время на d и вызывает созревшие f в порядке их времени,
// при равном времени — в порядке планирования. Пока выполняется f, Now()
// возвращает момент срабатывания её таймера. f может планировать и
// отменять таймеры; новые, созревшие до конца Advance, срабатывают в нём
// же. f вызывается не под блокировкой. После Advance Now() == старое + d.
func (c *FakeClock) Advance(d time.Duration) {
	// ваш код
}
