package main

import (
	"context"
	"time"
)

// Background запускает задачи, которые должны пережить запрос:
// запись аудита, отправку события, прогрев кэша.
type Background struct {
	// ваши поля
}

func NewBackground(timeout time.Duration, limit int) *Background {
	// ваш код
	return &Background{}
}

// Go запускает f в отдельной горутине. Контекст f:
//   - НЕ отменяется вместе с ctx запроса, но видит его значения;
//   - имеет свой дедлайн: timeout от момента вызова Go.
//
// Одновременно работает не больше limit задач. Если все места заняты,
// Go не ждёт, а возвращает false (задача отброшена). После Shutdown —
// тоже false. true — задача запущена.
func (b *Background) Go(ctx context.Context, f func(context.Context)) bool {
	// ваш код
	go f(ctx)
	return true
}

// Shutdown запрещает новые задачи и ждёт текущие, но не дольше ctx.
// Если ctx кончился раньше — возвращает ctx.Err(). Повторный вызов безопасен.
func (b *Background) Shutdown(ctx context.Context) error {
	// ваш код
	return nil
}
