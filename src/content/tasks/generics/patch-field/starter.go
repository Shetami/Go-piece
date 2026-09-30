package main

// Field — поле PATCH-запроса с тремя состояниями:
// ключа нет в JSON (Present == false), пришёл null (Present && Null),
// пришло значение (Present && !Null, значение в Value).
type Field[T any] struct {
	Present bool
	Null    bool
	Value   T
}

// UnmarshalJSON вызывается пакетом encoding/json, только если ключ есть.
func (f *Field[T]) UnmarshalJSON(data []byte) error {
	// ваш код
	return nil
}

// Get возвращает значение и true, только если пришло не-null значение.
func (f Field[T]) Get() (T, bool) {
	// ваш код
	var zero T
	return zero, false
}

// Apply переносит поле в nullable-колонку, представленную указателем:
// ключа не было — *dst не меняется; null — *dst = nil; значение —
// *dst указывает на новую копию значения.
func (f Field[T]) Apply(dst **T) {
	// ваш код
}
