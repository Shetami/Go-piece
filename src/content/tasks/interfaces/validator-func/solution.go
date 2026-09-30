package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Validator проверяет строку.
type Validator interface {
	Validate(s string) error
}

// ValidatorFunc превращает функцию в Validator.
type ValidatorFunc func(s string) error

// Validate вызывает f.
func (f ValidatorFunc) Validate(s string) error { return f(s) }

// MinLen: строка короче n символов (не байт!) — ошибка
// fmt.Errorf("length %d is less than %d", длина, n).
func MinLen(n int) Validator {
	return ValidatorFunc(func(s string) error {
		if l := utf8.RuneCountInString(s); l < n {
			return fmt.Errorf("length %d is less than %d", l, n)
		}
		return nil
	})
}

// NotBlank: строка пустая или из одних пробельных символов — ошибка "blank".
func NotBlank() Validator {
	return ValidatorFunc(func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("blank")
		}
		return nil
	})
}

// All проверяет строку всеми валидаторами по порядку и возвращает
// errors.Join всех ошибок (nil, если ошибок нет). nil-валидаторы пропускает.
func All(vs ...Validator) Validator {
	return ValidatorFunc(func(s string) error {
		var errs []error
		for _, v := range vs {
			if v == nil {
				continue
			}
			if err := v.Validate(s); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...) // для пустого списка — nil
	})
}
