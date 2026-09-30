package main

import (
	"context"
	"errors"
	"fmt"
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
	var lastErr error
	for range max(maxAttempts, 1) {
		if err := ctx.Err(); err != nil {
			return errors.Join(lastErr, err)
		}
		// Паника из attempt пролетит сквозь цикл — повтора не будет.
		lastErr = attempt(ctx, db, fn)
		if !errors.Is(lastErr, ErrSerialization) {
			return lastErr // nil или неповторяемая ошибка
		}
	}
	return lastErr
}

// attempt — одна транзакция. Отдельная функция, чтобы defer срабатывал
// в конце каждой попытки, а не всего цикла.
func attempt(ctx context.Context, db Beginner, fn func(context.Context, Tx) error) (err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		// Сюда попадаем и при ошибке fn, и при панике. recover не нужен:
		// откатили — и паника полетит дальше сама, с исходным стеком.
		if rbErr := tx.Rollback(); rbErr != nil {
			err = errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
		}
	}()

	if err := fn(ctx, tx); err != nil {
		return err
	}
	committed = true
	return tx.Commit()
}
