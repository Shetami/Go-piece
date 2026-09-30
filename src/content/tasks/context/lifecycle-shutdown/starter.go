package main

import (
	"context"
	"time"
)

// Component — часть приложения: база, кэш, консьюмер, HTTP-сервер.
type Component struct {
	Name  string
	Start func(ctx context.Context) error
	Stop  func(ctx context.Context) error
}

type App struct {
	Components  []Component
	StopTimeout time.Duration // на каждый Stop, не больше остатка общего ctx
	started     int           // сколько первых компонентов запущено
}

// Start запускает компоненты по порядку. Если Start компонента упал
// (в том числе из-за отмены ctx), уже запущенные останавливаются в обратном
// порядке — с контекстом, который не отменён вместе с ctx старта, но
// ограничен StopTimeout. Ошибка: "start <имя>: <err>", объединённая через
// errors.Join с ошибками этих остановок.
func (a *App) Start(ctx context.Context) error {
	// ваш код
	for _, c := range a.Components {
		if err := c.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Stop останавливает запущенные компоненты в обратном порядке. Каждый Stop
// получает свой контекст: StopTimeout от момента вызова, но не дольше ctx.
// Ошибка одного не мешает остановить остальные; все ошибки — errors.Join,
// каждая как "stop <имя>: <err>". Повторный Stop ничего не делает.
func (a *App) Stop(ctx context.Context) error {
	// ваш код
	return nil
}
