package main

import (
	"context"
	"errors"
	"fmt"
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
	for i, c := range a.Components {
		if err := c.Start(ctx); err != nil {
			startErr := fmt.Errorf("start %s: %w", c.Name, err)
			a.started = i
			// ctx старта мог быть уже отменён — именно поэтому и упали.
			// Откат обязан получить живой контекст, иначе ресурсы утекут.
			return errors.Join(startErr, a.Stop(context.WithoutCancel(ctx)))
		}
	}
	a.started = len(a.Components)
	return nil
}

// Stop останавливает запущенные компоненты в обратном порядке. Каждый Stop
// получает свой контекст: StopTimeout от момента вызова, но не дольше ctx.
// Ошибка одного не мешает остановить остальные; все ошибки — errors.Join,
// каждая как "stop <имя>: <err>". Повторный Stop ничего не делает.
func (a *App) Stop(ctx context.Context) error {
	var errs []error
	for i := a.started - 1; i >= 0; i-- {
		c := a.Components[i]
		if err := a.stopOne(ctx, c); err != nil {
			errs = append(errs, fmt.Errorf("stop %s: %w", c.Name, err))
		}
	}
	a.started = 0
	return errors.Join(errs...)
}

func (a *App) stopOne(ctx context.Context, c Component) error {
	// Свой таймаут на каждый: зависший компонент не съедает время остальных.
	sctx, cancel := context.WithTimeout(ctx, a.StopTimeout)
	defer cancel()
	return c.Stop(sctx)
}
