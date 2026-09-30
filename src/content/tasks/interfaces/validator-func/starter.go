package main

// Validator проверяет строку.
type Validator interface {
	Validate(s string) error
}

// ValidatorFunc превращает функцию в Validator.
type ValidatorFunc func(s string) error

// Validate вызывает f.
func (f ValidatorFunc) Validate(s string) error {
	// ваш код
	return nil
}

// MinLen: строка короче n символов (не байт!) — ошибка
// fmt.Errorf("length %d is less than %d", длина, n).
func MinLen(n int) Validator {
	// ваш код
	return ValidatorFunc(nil)
}

// NotBlank: строка пустая или из одних пробельных символов — ошибка "blank".
func NotBlank() Validator {
	// ваш код
	return ValidatorFunc(nil)
}

// All проверяет строку всеми валидаторами по порядку и возвращает
// errors.Join всех ошибок (nil, если ошибок нет). nil-валидаторы пропускает.
func All(vs ...Validator) Validator {
	// ваш код
	return ValidatorFunc(nil)
}
