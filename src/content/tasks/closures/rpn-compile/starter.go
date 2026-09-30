package main

import "errors"

var (
	ErrSyntax  = errors.New("syntax error")
	ErrUnknown = errors.New("unknown variable")
	ErrDivZero = errors.New("division by zero")
)

// Expr вычисляет скомпилированное выражение на наборе переменных.
type Expr func(vars map[string]float64) (float64, error)

// Compile разбирает выражение в обратной польской записи: токены через
// пробелы — числа (как в strconv.ParseFloat), имена переменных (буквы,
// цифры, _; начинаются не с цифры) и операторы + - * /.
// "a b -" — это a - b.
//
// Весь разбор и все проверки структуры — в Compile: пустое выражение,
// нехватка операндов, лишние операнды, неизвестный токен → ошибка с
// ErrSyntax. Возвращённая Expr только вычисляет: ей нельзя заново резать
// строку. Ошибки вычисления: переменной нет в vars → ErrUnknown (в тексте
// имя), деление на ноль → ErrDivZero.
func Compile(src string) (Expr, error) {
	// ваш код
	return func(map[string]float64) (float64, error) { return 0, nil }, nil
}
