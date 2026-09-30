package main

// MultiError — плоский список ошибок.
type MultiError struct {
	Errs []error
}

// Error: для одной ошибки — её текст, для нескольких —
// "ошибок: N: a; b; c".
func (m *MultiError) Error() string {
	// ваш код
	return ""
}

// Unwrap открывает элементы для errors.Is и errors.As.
func (m *MultiError) Unwrap() []error {
	// ваш код
	return nil
}

// Append добавляет errs к err:
//   - nil-ошибки пропускаются; если в итоге ошибок нет — вернуть nil
//     (настоящий nil интерфейса error);
//   - если err или любой из errs — *MultiError, его элементы вливаются в
//     плоский список (без вложенности); errors.Join и прочие — не трогать;
//   - если ошибка ровно одна — вернуть её саму, без обёртки;
//   - аргументы не меняются: Append(m, a) и Append(m, b) независимы.
func Append(err error, errs ...error) error {
	// ваш код
	return err
}
