package main

import (
	"errors"
	"sync"
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
	var (
		mu       sync.Mutex
		fails    int       // ошибок подряд в замкнутом состоянии
		open     bool      // разомкнут (или полуоткрыт)
		openedAt time.Time // когда разомкнулся
		probing  bool      // пробный вызов уже идёт
	)
	return func() error {
		mu.Lock()
		if open {
			if probing || now().Sub(openedAt) < cooldown {
				mu.Unlock()
				return ErrOpen
			}
			probing = true // этот вызов — единственная проба
		}
		isProbe := probing
		mu.Unlock()

		err := call() // вне замка: медленный call не держит остальных

		mu.Lock()
		defer mu.Unlock()
		switch {
		case isProbe && err == nil:
			open, probing, fails = false, false, 0
		case isProbe:
			probing = false
			openedAt = now() // снова разомкнут, отсчёт заново
		case err == nil:
			fails = 0
		default:
			fails++
			if fails >= threshold && !open {
				open, openedAt, fails = true, now(), 0
			}
		}
		return err
	}
}
