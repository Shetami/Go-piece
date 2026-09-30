package main

import (
	"context"
	"time"
)

// Bucket — ограничитель частоты «ведро токенов» на буферизованном канале.
// Вмещает до capacity токенов (capacity >= 1) и в начале полон. Каждый
// сигнал из refill добавляет один токен; если ведро полно, токен пропадает.
// refill приходит снаружи: в проде это time.NewTicker(...).C, в тестах —
// свой канал. Если refill закрыт, пополнение прекращается.
type Bucket struct {
	// ваши поля
}

// NewBucket создаёт полное ведро и запускает горутину пополнения.
func NewBucket(capacity int, refill <-chan time.Time) *Bucket {
	// ваш код
	return &Bucket{}
}

// Allow забирает токен без ожидания; false, если токенов нет.
func (b *Bucket) Allow() bool {
	// ваш код
	return true
}

// Wait ждёт токен или отмены ctx (тогда ctx.Err()).
func (b *Bucket) Wait(ctx context.Context) error {
	// ваш код
	return nil
}

// Stop останавливает пополнение и ждёт завершения его горутины.
// Повторный Stop безопасен.
func (b *Bucket) Stop() {
	// ваш код
}
