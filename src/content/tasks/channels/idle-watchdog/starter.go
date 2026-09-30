package main

import (
	"context"
	"errors"
	"time"
)

var ErrIdle = errors.New("idle timeout")

// Watchdog следит за активностью: каждое значение из kicks означает
// «я жив». Возвращённый канал получает ровно одно значение и закрывается:
//   - ErrIdle — если kicks молчал дольше timeout (с момента старта или
//     с последнего kick);
//   - nil — если kicks закрыли (штатная остановка);
//   - ctx.Err() — если отменили ctx.
//
// Горутина Watchdog завершается сразу после этого, даже если результат
// никто не читает.
func Watchdog(ctx context.Context, timeout time.Duration, kicks <-chan struct{}) <-chan error {
	// ваш код
	return nil
}
