package main

import (
	"errors"
	"time"
)

// ErrOpen возвращается, когда предохранитель разомкнут и call не вызывается.
var ErrOpen = errors.New("circuit open")

// Breaker оборачивает call предохранителем.
//   - Замкнут: вызовы идут в call. После threshold ошибок ПОДРЯД —
//     размыкается; успех сбрасывает счётчик.
//   - Разомкнут: в течение cooldown с момента размыкания вызовы сразу
//     возвращают ErrOpen, call не трогают.
//   - После cooldown — полуоткрыт: ровно один пробный вызов идёт в call,
//     остальные, пока проба не закончилась, получают ErrOpen. Успех пробы
//     замыкает предохранитель, ошибка — снова размыкает на cooldown от now().
//
// Ошибки call возвращаются как есть. Безопасно для горутин; call не
// вызывается под блокировкой.
func Breaker(call func() error, threshold int, cooldown time.Duration, now func() time.Time) func() error {
	// ваш код
	return call
}
