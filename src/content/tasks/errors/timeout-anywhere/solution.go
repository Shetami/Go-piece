package main

import (
	"context"
	"errors"
)

// Kind — что случилось с запросом, с точки зрения метрик и алертов.
type Kind int

const (
	KindOK       Kind = iota // ошибки нет
	KindTimeout              // мы не уложились во время
	KindCanceled             // клиент ушёл сам
	KindFailure              // всё остальное
)

// IsTimeout сообщает, есть ли ГДЕ-НИБУДЬ в дереве err звено с методом
// Timeout() bool, вернувшим true. Обходится всё дерево — и цепочки
// Unwrap() error, и все ветки Unwrap() []error. Звено с Timeout() == false
// поиск не прекращает: ответ может быть глубже.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	// errors.As тут не подходит: он остановится на ПЕРВОМ звене с методом
	// Timeout, даже если оно ответило false.
	if t, ok := err.(interface{ Timeout() bool }); ok && t.Timeout() {
		return true
	}
	switch u := err.(type) {
	case interface{ Unwrap() error }:
		return IsTimeout(u.Unwrap())
	case interface{ Unwrap() []error }:
		for _, e := range u.Unwrap() {
			if IsTimeout(e) {
				return true
			}
		}
	}
	return false
}

// Classify: nil → KindOK; IsTimeout → KindTimeout; иначе
// errors.Is(err, context.Canceled) → KindCanceled; иначе KindFailure.
// Таймаут важнее отмены: если в дереве есть и то и другое — KindTimeout.
func Classify(err error) Kind {
	switch {
	case err == nil:
		return KindOK
	case IsTimeout(err):
		return KindTimeout
	case errors.Is(err, context.Canceled):
		return KindCanceled
	}
	return KindFailure
}
