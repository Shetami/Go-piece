package main

// mustPanic — своя обёртка, по которой Catch отличает «свои» паники от чужих.
type mustPanic struct{ err error }

// Must возвращает v, если err == nil, и паникует, если нет.
// Паника должна быть такой, чтобы Catch мог её узнать.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(mustPanic{err})
	}
	return v
}

// Catch выполняет f. Если f запаниковала из-за Must — возвращает ту самую
// ошибку, что была передана в Must (errors.Is по ней работает).
// Любую другую панику Catch не трогает — она летит дальше.
// Если паники не было — nil.
func Catch(f func()) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		mp, ok := r.(mustPanic)
		if !ok {
			panic(r) // чужая паника — отпускаем дальше как есть
		}
		err = mp.err // именованный результат: только так defer может его задать
	}()
	f()
	return nil
}
