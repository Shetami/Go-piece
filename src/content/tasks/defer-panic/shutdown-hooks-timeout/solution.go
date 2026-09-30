package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type hook struct {
	name string
	fn   func(ctx context.Context) error
}

// Shutdown — список хуков остановки сервиса.
type Shutdown struct {
	hooks []hook
}

// Add регистрирует хук. Хуки выполняются в порядке, обратном добавлению.
func (s *Shutdown) Add(name string, fn func(ctx context.Context) error) {
	s.hooks = append(s.hooks, hook{name, fn})
}

// Run выполняет хуки по одному, в обратном порядке.
//   - Каждый хук получает контекст-наследник ctx с таймаутом perHook.
//     Если хук не вернулся за perHook (даже если он игнорирует контекст),
//     Run его больше не ждёт: ошибка хука — context.DeadlineExceeded,
//     и Run переходит к следующему.
//   - Паника хука становится его ошибкой с текстом паники; остальные
//     хуки выполняются.
//   - Ошибки подписываются именем хука: "<name>: <ошибка>".
//   - Если ctx отменён, следующие хуки не запускаются, а в результат
//     добавляется ctx.Err().
//   - Результат — errors.Join всех ошибок. Повторный Run ничего не делает.
func (s *Shutdown) Run(ctx context.Context, perHook time.Duration) error {
	hooks := s.hooks
	s.hooks = nil
	var errs []error
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			errs = append(errs, fmt.Errorf("остановка прервана: %w", err))
			break
		}
		if err := runHook(ctx, hooks[i].fn, perHook); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", hooks[i].name, err))
		}
	}
	return errors.Join(errs...)
}

// runHook запускает хук в отдельной горутине, чтобы можно было перестать
// его ждать, если он завис и не смотрит на контекст.
func runHook(ctx context.Context, fn func(context.Context) error, d time.Duration) error {
	hctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	// Буфер 1: зависший хук, если когда-нибудь вернётся, запишет результат
	// и завершится, а не повиснет навсегда на отправке.
	done := make(chan error, 1)
	go func() {
		// recover — в той же горутине, где выполняется хук: иначе его
		// паника уронит весь процесс.
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("паника: %v", r)
			}
		}()
		done <- fn(hctx)
	}()

	select {
	case err := <-done:
		return err
	case <-hctx.Done():
		return hctx.Err()
	}
}
