package main

import "context"

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
	// ваш код
	return ""
}

// Unwrap отдаёт Err и все ошибки компенсаций, чтобы errors.Is находил любую.
func (e *SagaError) Unwrap() []error {
	// ваш код
	return nil
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
	// ваш код
	for _, s := range steps {
		if err := s.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}
