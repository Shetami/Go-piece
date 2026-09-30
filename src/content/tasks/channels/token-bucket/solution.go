package main

import (
	"context"
	"sync"
	"time"
)

// Bucket — ограничитель частоты «ведро токенов» на буферизованном канале.
// Вмещает до capacity токенов (capacity >= 1) и в начале полон. Каждый
// сигнал из refill добавляет один токен; если ведро полно, токен пропадает.
// refill приходит снаружи: в проде это time.NewTicker(...).C, в тестах —
// свой канал. Если refill закрыт, пополнение прекращается.
type Bucket struct {
	tokens chan struct{} // буфер канала — это и есть ведро
	stop   chan struct{}
	done   chan struct{}
	once   sync.Once
}

// NewBucket создаёт полное ведро и запускает горутину пополнения.
func NewBucket(capacity int, refill <-chan time.Time) *Bucket {
	b := &Bucket{
		tokens: make(chan struct{}, capacity),
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
	for range capacity {
		b.tokens <- struct{}{}
	}
	go func() {
		defer close(b.done)
		for {
			select {
			case _, ok := <-refill:
				if !ok {
					return // закрытый канал иначе крутил бы цикл вхолостую
				}
				select {
				case b.tokens <- struct{}{}:
				default: // ведро полно — токен пропадает, пополнение не встаёт
				}
			case <-b.stop:
				return
			}
		}
	}()
	return b
}

// Allow забирает токен без ожидания; false, если токенов нет.
func (b *Bucket) Allow() bool {
	select {
	case <-b.tokens:
		return true
	default:
		return false
	}
}

// Wait ждёт токен или отмены ctx (тогда ctx.Err()).
func (b *Bucket) Wait(ctx context.Context) error {
	select {
	case <-b.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop останавливает пополнение и ждёт завершения его горутины.
// Повторный Stop безопасен.
func (b *Bucket) Stop() {
	b.once.Do(func() { close(b.stop) })
	<-b.done
}
