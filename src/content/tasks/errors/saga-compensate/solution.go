package main

import (
	"context"
	"fmt"
)

// Step — шаг саги и его компенсация.
type Step struct {
	Name       string
	Do         func(ctx context.Context) error
	Compensate func(ctx context.Context) error // может быть nil
}

// SagaError — сага не прошла.
type SagaError struct {
	Step         string  // имя шага, на котором всё остановилось
	Err          error   // ошибка шага или ctx.Err()
	Compensation []error // ошибки компенсаций в порядке выполнения,
	// каждая обёрнута как "компенсация <имя шага>: %w"
}

// Error: "сага: шаг <Step>: <Err>", и если компенсации падали —
// "; компенсаций не удалось: <N>".
func (e *SagaError) Error() string {
	s := fmt.Sprintf("сага: шаг %s: %v", e.Step, e.Err)
	if n := len(e.Compensation); n > 0 {
		s += fmt.Sprintf("; компенсаций не удалось: %d", n)
	}
	return s
}

// Unwrap отдаёт Err и все ошибки компенсаций, чтобы errors.Is находил любую.
func (e *SagaError) Unwrap() []error {
	return append([]error{e.Err}, e.Compensation...)
}

// Run выполняет шаги по порядку. Всё прошло → nil. Иначе:
//   - шаг вернул ошибку, или перед очередным шагом ctx уже отменён (тогда
//     Err = ctx.Err(), а сам шаг не выполняется);
//   - компенсировать все УСПЕШНО выполненные шаги в обратном порядке;
//     упавший шаг не компенсируется, шаг без Compensate пропускается;
//   - компенсации выполняются, даже если ctx отменён: им передаётся
//     контекст, который не отменяется вместе с ctx (но хранит его значения);
//   - ошибка одной компенсации не останавливает остальные;
//   - вернуть *SagaError.
func Run(ctx context.Context, steps []Step) error {
	for i, s := range steps {
		err := ctx.Err()
		if err == nil {
			err = s.Do(ctx)
		}
		if err != nil {
			return &SagaError{Step: s.Name, Err: err, Compensation: compensate(ctx, steps[:i])}
		}
	}
	return nil // литерал nil, а не (*SagaError)(nil)
}

// compensate откатывает выполненные шаги с конца.
func compensate(ctx context.Context, done []Step) []error {
	// Откат должен дойти до конца, даже если запрос уже отменён: иначе
	// деньги списаны, а заказа нет. WithoutCancel (Go 1.21) сохраняет
	// значения ctx (trace ID и т.п.), но не его отмену и дедлайн.
	cctx := context.WithoutCancel(ctx)
	var errs []error
	for i := len(done) - 1; i >= 0; i-- {
		s := done[i]
		if s.Compensate == nil {
			continue
		}
		if err := s.Compensate(cctx); err != nil {
			errs = append(errs, fmt.Errorf("компенсация %s: %w", s.Name, err))
		}
	}
	return errs
}
