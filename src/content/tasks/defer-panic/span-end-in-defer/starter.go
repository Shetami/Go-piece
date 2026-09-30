package main

import (
	"context"
	"sync"
	"time"
)

// Span — завершённый участок трассы.
type Span struct {
	Name   string
	Parent string // имя родительского спана, "" для корня
	Start  time.Time
	End    time.Time
	Status string // "ok", "error" или "panic"
	Err    error  // ошибка функции или ошибка с текстом паники
}

// Tracer собирает завершённые спаны. Безопасен для конкурентного использования.
type Tracer struct {
	Now   func() time.Time
	mu    sync.Mutex
	spans []Span
}

// Spans возвращает завершённые спаны в порядке завершения.
func (t *Tracer) Spans() []Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]Span(nil), t.spans...)
}

// Start начинает спан name. Родитель — спан, лежащий в ctx (если есть).
// Возвращает ctx с новым спаном и функцию end, которую вызывают так:
//
//	func Do(ctx context.Context) (err error) {
//		ctx, end := tr.Start(ctx, "do")
//		defer end(&err)
//		...
//	}
//
// end фиксирует End = Now() и статус:
//   - функция паникует — "panic", Err содержит текст паники, а паника летит дальше;
//   - иначе *errp != nil — "error" и Err = *errp;
//   - иначе "ok". errp может быть nil (функция без ошибки).
func (t *Tracer) Start(ctx context.Context, name string) (context.Context, func(errp *error)) {
	// ваш код
	return ctx, func(*error) {}
}
