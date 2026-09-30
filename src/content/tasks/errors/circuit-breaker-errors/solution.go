package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrOpen — автомат разомкнут, вызов не выполнялся.
	ErrOpen = errors.New("цепь разомкнута")
	// ErrClient помечает ошибки клиента (невалидный запрос и т.п.):
	// о здоровье зависимости они ничего не говорят.
	ErrClient = errors.New("ошибка клиента")
)

// OpenError возвращается вместо вызова, пока автомат разомкнут.
// Допишите методы так, чтобы errors.Is(err, ErrOpen) и
// errors.Is(err, LastErr) были истинны.
type OpenError struct {
	Until   time.Time // когда пропустят пробный вызов
	LastErr error     // сбой, который разомкнул цепь
}

func (e *OpenError) Error() string { return ErrOpen.Error() + ": " + e.LastErr.Error() }

// Is делает OpenError «разновидностью» ErrOpen, а Unwrap остаётся свободным
// для настоящей причины.
func (e *OpenError) Is(target error) bool { return target == ErrOpen }

func (e *OpenError) Unwrap() error { return e.LastErr }

// Breaker — автомат защиты для вызовов зависимости.
type Breaker struct {
	threshold int
	cooldown  time.Duration
	now       func() time.Time

	mu      sync.Mutex
	fails   int
	until   time.Time // нулевое — замкнут
	lastErr error
	probing bool
}

// NewBreaker: threshold сбоев подряд размыкают цепь на cooldown.
// now — часы (тесты подставляют свои).
func NewBreaker(threshold int, cooldown time.Duration, now func() time.Time) *Breaker {
	return &Breaker{threshold: threshold, cooldown: cooldown, now: now}
}

// Call выполняет f(ctx) через автомат и возвращает её ошибку как есть.
//   - ctx уже отменён → вернуть ctx.Err(), не вызывая f и не меняя состояние.
//   - Замкнут: вызвать f. Сбой — любая ошибка, КРОМЕ context.Canceled и
//     ошибок с ErrClient в цепочке: те счётчик не трогают (ни прибавляют,
//     ни сбрасывают). Успех сбрасывает счётчик. На threshold-м сбое подряд
//     цепь размыкается до now()+cooldown.
//   - Разомкнут и now() раньше Until: f не вызывать, вернуть *OpenError.
//   - Полуоткрыт (Until наступил): пропустить РОВНО ОДИН пробный вызов,
//     остальные одновременные получают *OpenError. Успех пробы замыкает
//     цепь, сбой — снова размыкает на cooldown. Проба с «не нашей» ошибкой
//     ничего не решает: следующий вызов снова будет пробным.
//
// Безопасен для одновременного использования; f вызывается без блокировки.
func (b *Breaker) Call(ctx context.Context, f func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	probe := false
	if !b.until.IsZero() {
		if b.probing || b.now().Before(b.until) {
			err := &OpenError{Until: b.until, LastErr: b.lastErr}
			b.mu.Unlock()
			return err
		}
		probe = true
		b.probing = true
	}
	b.mu.Unlock()

	err := f(ctx) // без мьютекса: медленная зависимость не должна блокировать всех

	b.mu.Lock()
	defer b.mu.Unlock()
	if probe {
		b.probing = false
	}
	switch {
	case err == nil:
		b.fails, b.until, b.lastErr = 0, time.Time{}, nil
	case errors.Is(err, context.Canceled) || errors.Is(err, ErrClient):
		// Ни о чём не говорит: клиент ушёл или сам прислал чушь.
	default:
		b.fails++
		b.lastErr = err
		if probe || b.fails >= b.threshold {
			b.until = b.now().Add(b.cooldown)
		}
	}
	return err
}
