package main

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen — выключатель разомкнут, вызов не выполнялся.
var ErrOpen = errors.New("circuit open")

type state int

const (
	closed state = iota
	open
	halfOpen
)

// Breaker — circuit breaker перед нестабильной зависимостью.
//
//   - closed: вызовы проходят; после threshold ошибок подряд (успех
//     обнуляет счёт) выключатель размыкается.
//   - open: вызовы сразу получают ErrOpen, fn не вызывается. Через
//     cooldown после размыкания следующий вызов становится пробным.
//   - half-open: идёт ровно один пробный вызов; все остальные в это время
//     получают ErrOpen. Успех пробы замыкает выключатель, ошибка —
//     снова размыкает на cooldown, отсчитанный от момента ошибки.
//
// fn выполняется без удержания внутренней блокировки: вызовы в
// состоянии closed идут параллельно.
type Breaker struct {
	threshold int
	cooldown  time.Duration
	now       func() time.Time

	mu       sync.Mutex
	state    state
	failures int
	openedAt time.Time
}

func NewBreaker(threshold int, cooldown time.Duration, now func() time.Time) *Breaker {
	return &Breaker{threshold: threshold, cooldown: cooldown, now: now}
}

// Call выполняет fn через выключатель и возвращает её ошибку или ErrOpen.
func (b *Breaker) Call(fn func() error) error {
	probe, err := b.before()
	if err != nil {
		return err
	}
	err = fn() // без мьютекса: зависимость может отвечать секундами
	b.after(probe, err)
	return err
}

// before решает, пускать ли вызов, и не пробный ли он.
func (b *Breaker) before() (probe bool, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case closed:
		return false, nil
	case open:
		if b.now().Sub(b.openedAt) < b.cooldown {
			return false, ErrOpen
		}
		// Проверка и переход — под одной блокировкой: пробным станет
		// ровно один вызов, остальные увидят half-open.
		b.state = halfOpen
		return true, nil
	default: // halfOpen: проба уже идёт
		return false, ErrOpen
	}
}

func (b *Breaker) after(probe bool, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if probe {
		if err != nil {
			b.trip()
		} else {
			b.state = closed
			b.failures = 0
		}
		return
	}
	// Обычный вызов мог завершиться уже после того, как выключатель
	// разомкнули другие, — тогда его результат состояние не меняет.
	if b.state != closed {
		return
	}
	if err == nil {
		b.failures = 0
		return
	}
	b.failures++
	if b.failures >= b.threshold {
		b.trip()
	}
}

func (b *Breaker) trip() {
	b.state = open
	b.openedAt = b.now()
	b.failures = 0
}

// State возвращает "closed", "open" или "half-open" (пока идёт проба).
func (b *Breaker) State() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return [...]string{"closed", "open", "half-open"}[b.state]
}
