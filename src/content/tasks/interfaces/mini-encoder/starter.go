package main

import "errors"

// Marshaler — тип сам знает своё представление.
type Marshaler interface {
	MarshalText() (string, error)
}

var ErrUnsupported = errors.New("unsupported type")

// Encode печатает значение в JSON-подобном виде. Проверки — строго в таком порядке:
//  1. nil и nil-указатель любого типа (даже реализующего интерфейсы) → null
//  2. Marshaler → результат MarshalText как есть; его ошибка — обёрнута через %w
//  3. error → текст ошибки в кавычках (strconv.Quote)
//  4. fmt.Stringer → String() в кавычках
//  5. bool, int, int64, float64 → strconv (float — 'g', -1); string → strconv.Quote
//  6. []any → [a,b] без пробелов; map[string]any → {"k":v}, ключи по возрастанию
//  7. прочее → ошибка с ErrUnsupported и именем типа (%T) в тексте
//
// Ошибка во вложенном элементе прерывает кодирование и возвращается наружу.
func Encode(v any) (string, error) {
	// ваш код
	return "", nil
}
