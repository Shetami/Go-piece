package main

// OpError добавляет к ошибке название операции.
type OpError struct {
	Op  string
	Err error
}

// Error возвращает "Op: текст Err".
func (e *OpError) Error() string {
	return e.Op + ": " + e.Err.Error()
}

// Unwrap открывает Err для errors.Is и errors.As.
func (e *OpError) Unwrap() error { return e.Err }

// Wrap оборачивает err в *OpError. Для err == nil возвращает nil —
// именно nil типа error, а не пустой *OpError.
func Wrap(op string, err error) error {
	if err == nil {
		// Литерал nil, а не (*OpError)(nil): иначе интерфейс был бы не nil.
		return nil
	}
	return &OpError{Op: op, Err: err}
}
