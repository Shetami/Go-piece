package main

import (
	"sync"
	"time"
)

// Debouncer откладывает вызов fn, пока события не затихнут.
//
//   - fn вызывается через wait после ПОСЛЕДНЕГО Trigger: серия Trigger с
//     паузами меньше wait даёт один вызов fn со значением последнего Trigger.
//   - Вызовы fn никогда не перекрываются и идут в порядке значений:
//     более старое значение не может прийти в fn после более нового.
//   - Flush немедленно (синхронно) вызывает fn с отложенным значением, если
//     оно есть, и отменяет отложенный вызов.
//   - Stop отменяет отложенный вызов и дожидается выполняющегося fn;
//     после Stop Trigger и Flush ничего не делают.
type Debouncer[T any] struct {
	wait time.Duration
	fn   func(T)

	callMu sync.Mutex // сериализует вызовы fn

	mu      sync.Mutex // защищает поля ниже
	timer   *time.Timer
	pending bool
	val     T
	stopped bool
}

func NewDebouncer[T any](wait time.Duration, fn func(T)) *Debouncer[T] {
	return &Debouncer[T]{wait: wait, fn: fn}
}

func (d *Debouncer[T]) Trigger(v T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped {
		return
	}
	d.val, d.pending = v, true
	if d.timer == nil {
		d.timer = time.AfterFunc(d.wait, d.fire)
	} else {
		d.timer.Reset(d.wait) // отсчёт заново от последнего события
	}
}

// fire забирает отложенное значение и вызывает fn.
func (d *Debouncer[T]) fire() {
	// callMu берём ДО чтения значения: иначе два fire могли бы прочитать
	// старое и новое значения, а вызвать fn в обратном порядке.
	d.callMu.Lock()
	defer d.callMu.Unlock()
	d.mu.Lock()
	if !d.pending || d.stopped {
		d.mu.Unlock()
		return
	}
	v := d.val
	d.pending = false
	d.mu.Unlock()
	d.fn(v) // без d.mu: fn может сама вызвать Trigger
}

func (d *Debouncer[T]) Flush() {
	d.mu.Lock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.mu.Unlock()
	d.fire()
}

func (d *Debouncer[T]) Stop() {
	d.mu.Lock()
	d.stopped = true
	d.pending = false
	if d.timer != nil {
		d.timer.Stop()
	}
	d.mu.Unlock()
	d.callMu.Lock() // ждём выполняющийся fn
	d.callMu.Unlock()
}
