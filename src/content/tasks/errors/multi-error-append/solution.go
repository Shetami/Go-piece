package main

import (
	"fmt"
	"strings"
)

// MultiError — плоский список ошибок.
type MultiError struct {
	Errs []error
}

// Error: для одной ошибки — её текст, для нескольких —
// "ошибок: N: a; b; c".
func (m *MultiError) Error() string {
	if len(m.Errs) == 1 {
		return m.Errs[0].Error()
	}
	parts := make([]string, len(m.Errs))
	for i, e := range m.Errs {
		parts[i] = e.Error()
	}
	return fmt.Sprintf("ошибок: %d: %s", len(m.Errs), strings.Join(parts, "; "))
}

// Unwrap открывает элементы для errors.Is и errors.As.
func (m *MultiError) Unwrap() []error {
	return m.Errs
}

// Append добавляет errs к err:
//   - nil-ошибки пропускаются; если в итоге ошибок нет — вернуть nil
//     (настоящий nil интерфейса error);
//   - если err или любой из errs — *MultiError, его элементы вливаются в
//     плоский список (без вложенности); errors.Join и прочие — не трогать;
//   - если ошибка ровно одна — вернуть её саму, без обёртки;
//   - аргументы не меняются: Append(m, a) и Append(m, b) независимы.
func Append(err error, errs ...error) error {
	var out []error // новый слайс: чужой backing array не трогаем
	add := func(e error) {
		if m, ok := e.(*MultiError); ok {
			if m != nil { // (*MultiError)(nil) в интерфейсе — не nil, но пуст
				out = append(out, m.Errs...)
			}
		} else if e != nil {
			out = append(out, e)
		}
	}
	add(err)
	for _, e := range errs {
		add(e)
	}
	switch len(out) {
	case 0:
		return nil // не (*MultiError)(nil): такой интерфейс != nil
	case 1:
		return out[0]
	}
	return &MultiError{Errs: out}
}
