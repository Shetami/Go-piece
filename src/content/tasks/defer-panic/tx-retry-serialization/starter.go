package main

import (
	"context"
	"errors"
)

// ErrSerialization — конфликт сериализации: транзакцию можно повторить целиком.
var ErrSerialization = errors.New("serialization failure")

type Tx interface {
	Commit() error
	Rollback() error
}

type Beginner interface {
	Begin(ctx context.Context) (Tx, error)
}

// RunInTx выполняет fn в транзакции и повторяет транзакцию целиком при
// конфликте сериализации.
//
// Одна попытка: Begin → fn(ctx, tx) → Commit.
//   - Ошибка Begin — вернуть её (с повтором, только если это конфликт).
//   - fn вернула ошибку — Rollback; если Rollback тоже упал, его ошибка
//     добавляется через errors.Join.
//   - fn запаниковала — Rollback, и паника летит дальше. Никаких повторов.
//   - После Commit (успешного или нет) Rollback не вызывается.
//
// Если ошибка попытки (из fn или Commit) — errors.Is(err, ErrSerialization),
// делается новая попытка, всего не больше maxAttempts (меньше 1 — одна).
// Перед каждой попыткой проверяется ctx: если он отменён, вернуть
// errors.Join(ошибка прошлой попытки, ctx.Err()). Другие ошибки не повторяются.
func RunInTx(ctx context.Context, db Beginner, maxAttempts int, fn func(context.Context, Tx) error) error {
	// ваш код
	return nil
}
