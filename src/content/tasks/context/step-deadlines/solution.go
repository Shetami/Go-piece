package main

import (
	"context"
	"errors"
	"fmt"
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
	// Общий дедлайн — один на все шаги, шаги наследуют его от этого контекста.
	all, cancelAll := context.WithTimeoutCause(ctx, total, ErrTotalTimeout)
	defer cancelAll()

	for _, s := range steps {
		if all.Err() != nil {
			return fmt.Errorf("перед шагом %s: %w: %w", s.Name, context.Cause(all), all.Err())
		}
		if err := runStep(all, perStep, s); err != nil {
			return err
		}
	}
	return nil
}

func runStep(all context.Context, perStep time.Duration, s Step) error {
	// Если от total осталось меньше perStep, сработает дедлайн родителя,
	// и Cause у шага будет родительская — ErrTotalTimeout.
	ctx, cancel := context.WithTimeoutCause(all, perStep, ErrStepTimeout)
	defer cancel() // не держим таймер шага до его срабатывания

	err := s.Run(ctx)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil && (errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)) {
		// Err() говорит только «дедлайн», Cause — чей именно.
		return fmt.Errorf("шаг %s: %w: %w", s.Name, context.Cause(ctx), err)
	}
	return fmt.Errorf("шаг %s: %w", s.Name, err)
}
