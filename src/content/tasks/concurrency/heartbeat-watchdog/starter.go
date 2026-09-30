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
	// ваш код
	return nil
}
