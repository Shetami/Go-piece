package main

import (
	"bytes"
	"encoding/json"
)

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
	f.Present = true // сам факт вызова значит, что ключ пришёл
	if bytes.Equal(data, []byte("null")) {
		f.Null = true
		var zero T
		f.Value = zero
		return nil
	}
	f.Null = false
	// Разбираем в Value, а не в f: иначе json снова вызовет этот же
	// метод, и будет бесконечная рекурсия.
	return json.Unmarshal(data, &f.Value)
}

// Get возвращает значение и true, только если пришло не-null значение.
func (f Field[T]) Get() (T, bool) {
	if !f.Present || f.Null {
		var zero T
		return zero, false
	}
	return f.Value, true
}

// Apply переносит поле в nullable-колонку, представленную указателем:
// ключа не было — *dst не меняется; null — *dst = nil; значение —
// *dst указывает на новую копию значения.
func (f Field[T]) Apply(dst **T) {
	switch {
	case !f.Present:
	case f.Null:
		*dst = nil
	default:
		v := f.Value // f — копия, но свою переменную заводим явно
		*dst = &v
	}
}
