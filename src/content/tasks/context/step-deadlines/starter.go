package main

import (
	"context"
	"errors"
	"time"
)

var (
	ErrStepTimeout  = errors.New("шаг не уложился в свой таймаут")
	ErrTotalTimeout = errors.New("пайплайн не уложился в общий таймаут")
)

// Step — один шаг пайплайна. Run обязан слушать свой ctx.
type Step struct {
	Name string
	Run  func(ctx context.Context) error
}

// RunPipeline выполняет шаги по очереди. На весь пайплайн даётся total,
// на каждый шаг — perStep, но не больше, чем осталось от total.
// Первая ошибка останавливает пайплайн; она возвращается обёрнутой и
// содержит имя шага. Если шаг упал из-за истёкшего времени, ошибка
// отвечает errors.Is и на context.DeadlineExceeded, и на ErrStepTimeout
// или ErrTotalTimeout — смотря что истекло. Если отменили родительский
// ctx — errors.Is и на Canceled, и на context.Cause(ctx).
// Контексты шагов освобождаются сразу после шага.
func RunPipeline(ctx context.Context, total, perStep time.Duration, steps []Step) error {
	// ваш код
	for _, s := range steps {
		if err := s.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
