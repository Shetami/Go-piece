package main

import (
	"context"
	"errors"
	"time"
)

var ErrStalled = errors.New("worker stalled")

// Supervise запускает work в отдельной горутине и передаёт ей beat —
// функцию-сердцебиение.
//
//   - Если между сердцебиениями (или от старта до первого) прошло больше
//     timeout, Supervise отменяет контекст work и возвращает ErrStalled —
//     что бы work ни вернула после отмены.
//   - Если work завершилась сама — её ошибка (или nil).
//   - Отмена внешнего ctx — ctx.Err().
//   - Во всех случаях Supervise возвращается только после завершения work.
//   - beat можно вызывать из любых горутин; он никогда не блокируется,
//     в том числе после возврата Supervise.
func Supervise(ctx context.Context, timeout time.Duration,
	work func(ctx context.Context, beat func()) error) error {
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	// Буфер 1 + неблокирующая отправка: частые beat схлопываются в один,
	// а после выхода Supervise beat просто ничего не делает.
	beats := make(chan struct{}, 1)
	beat := func() {
		select {
		case beats <- struct{}{}:
		default:
		}
	}

	done := make(chan error, 1)
	go func() { done <- work(ctx2, beat) }()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-beats:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(timeout)
		case err := <-done:
			return err
		case <-timer.C:
			cancel()
			<-done // дожидаемся work: она не должна пережить Supervise
			return ErrStalled
		case <-ctx.Done():
			cancel()
			<-done
			return ctx.Err()
		}
	}
}
