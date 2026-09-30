package main

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

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
	var b strings.Builder
	if err := encode(&b, v); err != nil {
		return "", err
	}
	return b.String(), nil
}

func encode(b *strings.Builder, v any) error {
	if v == nil {
		b.WriteString("null")
		return nil
	}
	// nil-указатель внутри интерфейса: v != nil, но вызвать метод на значении нельзя.
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Pointer && rv.IsNil() {
		b.WriteString("null")
		return nil
	}
	// Порядок case важен: тип может реализовывать несколько интерфейсов сразу.
	switch x := v.(type) {
	case Marshaler:
		s, err := x.MarshalText()
		if err != nil {
			return fmt.Errorf("marshal %T: %w", v, err)
		}
		b.WriteString(s)
	case error:
		b.WriteString(strconv.Quote(x.Error()))
	case fmt.Stringer:
		b.WriteString(strconv.Quote(x.String()))
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case int:
		b.WriteString(strconv.Itoa(x))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case float64:
		b.WriteString(strconv.FormatFloat(x, 'g', -1, 64))
	case string:
		b.WriteString(strconv.Quote(x))
	case []any:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := encode(b, e); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.Quote(k))
			b.WriteByte(':')
			if err := encode(b, x[k]); err != nil {
				return err
			}
		}
		b.WriteByte('}')
	default:
		return fmt.Errorf("%w: %T", ErrUnsupported, v)
	}
	return nil
}
