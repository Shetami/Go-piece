package main

// PanicError — паника горутины, превращённая в ошибку.
type PanicError struct {
	Value any    // то, что вернул recover
	Stack []byte // стек горутины в момент паники (runtime/debug.Stack)
}

// Error содержит значение паники.
func (e *PanicError) Error() string {
	// ваш код
	return ""
}

// Unwrap возвращает Value, если это error, иначе nil, — чтобы errors.Is/As
// находили исходную ошибку, с которой паниковали.
func (e *PanicError) Unwrap() error {
	// ваш код
	return nil
}

// Group запускает функции в отдельных горутинах и собирает их ошибки.
// Нулевое значение готово к работе.
type Group struct {
	// ваши поля
}

// Go запускает fn в новой горутине. Паника внутри fn не роняет программу,
// а становится *PanicError.
func (g *Group) Go(fn func() error) {
	// ваш код
}

// Wait ждёт завершения всех запущенных функций и возвращает все их ошибки
// и паники, объединённые через errors.Join. Если ошибок нет — nil.
func (g *Group) Wait() error {
	// ваш код
	return nil
}
