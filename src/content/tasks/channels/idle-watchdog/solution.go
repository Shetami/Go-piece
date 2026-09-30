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
	// Буфер на одно значение: горутина отдаёт результат и уходит,
	// не дожидаясь читателя.
	res := make(chan error, 1)
	go func() {
		defer close(res)
		// Один таймер на всё время, перезаводимый на каждый kick,
		// а не time.After в цикле: тот создавал бы таймер на каждый kick.
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		for {
			select {
			case _, ok := <-kicks:
				if !ok {
					res <- nil
					return
				}
				// С Go 1.23 Reset безопасен без ручного вычерпывания
				// timer.C: старое срабатывание не «догонит» новый отсчёт.
				timer.Reset(timeout)
			case <-timer.C:
				res <- ErrIdle
				return
			case <-ctx.Done():
				res <- ctx.Err()
				return
			}
		}
	}()
	return res
}
