package main

import (
	"context"
	"time"
)

// Shutdown — список хуков остановки сервиса.
type Shutdown struct {
	// ваши поля
}

// Add регистрирует хук. Хуки выполняются в порядке, обратном добавлению.
func (s *Shutdown) Add(name string, hook func(ctx context.Context) error) {
	// ваш код
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
	// ваш код
	return nil
}
