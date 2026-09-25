package main

// OpError добавляет к ошибке название операции.
type OpError struct {
	Op  string
	Err error
}

// Error возвращает "Op: текст Err".
func (e *OpError) Error() string {
	// ваш код
	return ""
}

// Реализуйте и метод Unwrap, чтобы errors.Is и errors.As видели Err.

// Wrap оборачивает err в *OpError. Для err == nil возвращает nil —
// именно nil типа error, а не пустой *OpError.
func Wrap(op string, err error) error {
	// ваш код
	return nil
}
