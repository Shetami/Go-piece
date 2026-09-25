package main

// Must возвращает v, если err == nil, и паникует, если нет.
// Паника должна быть такой, чтобы Catch мог её узнать.
func Must[T any](v T, err error) T {
	// ваш код
	return v
}

// Catch выполняет f. Если f запаниковала из-за Must — возвращает ту самую
// ошибку, что была передана в Must (errors.Is по ней работает).
// Любую другую панику Catch не трогает — она летит дальше.
// Если паники не было — nil.
func Catch(f func()) (err error) {
	// ваш код
	f()
	return nil
}
