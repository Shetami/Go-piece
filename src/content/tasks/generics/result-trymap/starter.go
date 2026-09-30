package main

// Result — значение или ошибка одного элемента.
type Result[T any] struct {
	Val T
	Err error
}

// ItemError — ошибка элемента с его индексом во входе.
type ItemError struct {
	Index int
	Err   error
}

// Error возвращает "элемент <Index>: <текст Err>".
func (e *ItemError) Error() string {
	// ваш код
	return ""
}

// Unwrap открывает Err для errors.Is и errors.As.
func (e *ItemError) Unwrap() error {
	// ваш код
	return nil
}

// Collect разбирает результаты: значения успешных — в исходном порядке,
// ошибки — каждая обёрнута в *ItemError с индексом в rs и все
// объединены errors.Join. Если ошибок нет, err == nil.
func Collect[T any](rs []Result[T]) ([]T, error) {
	// ваш код
	return nil, nil
}

// TryMap применяет f к каждому элементу in — даже после ошибок — и
// возвращает то же, что Collect от результатов.
func TryMap[T, U any](in []T, f func(T) (U, error)) ([]U, error) {
	// ваш код
	return nil, nil
}
