package main

import (
	"log/slog"
	"maps"
	"slices"
)

// fieldsError — обёртка, которая только хранит поля.
type fieldsError struct {
	err error
	kv  []any
}

// Текст — ровно текст исходной ошибки: поля идут в лог отдельно.
func (e *fieldsError) Error() string { return e.err.Error() }
func (e *fieldsError) Unwrap() error { return e.err }

// WithFields прикрепляет к ошибке поля для структурного лога, не меняя её
// текст и не мешая errors.Is/As. kv — чередование ключ (string), значение.
// nil → nil.
func WithFields(err error, kv ...any) error {
	if err == nil {
		return nil
	}
	return &fieldsError{err: err, kv: slices.Clone(kv)}
}

// Fields собирает поля со всего дерева err (цепочки Unwrap() error и все
// ветки Unwrap() []error). При совпадении ключей побеждает поле,
// прикреплённое ГЛУБЖЕ (ближе к источнику ошибки), а между ветками одного
// Join — более ранняя ветка. Нет полей — пустая, но не nil мапа.
func Fields(err error) map[string]any {
	out := make(map[string]any)
	collect(err, out)
	return out
}

// Attrs — поля для slog: отсортированы по ключу, последним добавлен
// атрибут "err" с текстом ошибки. nil → nil.
func Attrs(err error) []slog.Attr {
	if err == nil {
		return nil
	}
	fs := Fields(err)
	attrs := make([]slog.Attr, 0, len(fs)+1)
	for _, k := range slices.Sorted(maps.Keys(fs)) {
		attrs = append(attrs, slog.Any(k, fs[k]))
	}
	return append(attrs, slog.String("err", err.Error()))
}

// collect заполняет out так, чтобы глубокие поля перекрывали внешние:
// сначала поля текущего звена, потом — поверх — поля детей. Дети Join
// обходятся с конца, чтобы ранняя ветка записалась последней и победила.
func collect(err error, out map[string]any) {
	if err == nil {
		return
	}
	if fe, ok := err.(*fieldsError); ok {
		for i := 0; i+1 < len(fe.kv); i += 2 {
			if k, ok := fe.kv[i].(string); ok {
				out[k] = fe.kv[i+1]
			}
		}
	}
	switch u := err.(type) {
	case interface{ Unwrap() error }:
		collect(u.Unwrap(), out)
	case interface{ Unwrap() []error }:
		errs := u.Unwrap()
		for i := len(errs) - 1; i >= 0; i-- {
			collect(errs[i], out)
		}
	}
}
