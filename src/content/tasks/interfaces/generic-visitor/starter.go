package main

import "errors"

// Node — узел выражения. Интерфейс «запечатан»: неэкспортируемый метод
// не даёт реализовать его за пределами пакета.
type Node interface{ node() }

type Num struct{ V float64 }
type Var struct{ Name string }
type Bin struct {
	Op   byte // '+', '-', '*', '/'
	L, R Node
}

func (Num) node() {}
func (Var) node() {}
func (Bin) node() {}

// Visitor[R] сворачивает дерево в значение типа R. Bin получает уже
// посчитанные результаты детей (обход снизу вверх).
type Visitor[R any] interface {
	Num(n Num) (R, error)
	Var(v Var) (R, error)
	Bin(op byte, l, r R) (R, error)
}

var (
	ErrNilNode   = errors.New("nil node")
	ErrDivByZero = errors.New("division by zero")
)

// Walk обходит дерево n visitor'ом v.
//   - Узлы приходят и по значению, и по указателю (Num и *Num); nil — ErrNilNode
//     (в том числе nil-указатель и nil-потомок).
//   - Сначала левый потомок, потом правый; первая ошибка останавливает обход —
//     правый потомок после ошибки в левом не посещается.
//   - Неизвестный тип узла — ошибка.
func Walk[R any](n Node, v Visitor[R]) (R, error) {
	// ваш код
	var zero R
	return zero, nil
}

// UnboundError — переменной нет в окружении.
type UnboundError struct{ Name string }

func (e *UnboundError) Error() string { return "unbound variable " + e.Name }

// Eval вычисляет выражение: переменные берутся из Env.
// Нет переменной — *UnboundError; деление на 0 — ErrDivByZero;
// неизвестная операция — ошибка.
type Eval struct{ Env map[string]float64 }

func (e Eval) Num(n Num) (float64, error) {
	// ваш код
	return 0, nil
}

func (e Eval) Var(v Var) (float64, error) {
	// ваш код
	return 0, nil
}

func (e Eval) Bin(op byte, l, r float64) (float64, error) {
	// ваш код
	return 0, nil
}

// Printer печатает выражение: числа — strconv.FormatFloat(v, 'g', -1, 64),
// переменные — по имени, Bin — "(l op r)".
type Printer struct{}

func (Printer) Num(n Num) (string, error) {
	// ваш код
	return "", nil
}

func (Printer) Var(v Var) (string, error) {
	// ваш код
	return "", nil
}

func (Printer) Bin(op byte, l, r string) (string, error) {
	// ваш код
	return "", nil
}

// FreeVars собирает имена переменных: без повторов, по возрастанию.
type FreeVars struct{}

func (FreeVars) Num(Num) ([]string, error) {
	// ваш код
	return nil, nil
}

func (FreeVars) Var(v Var) ([]string, error) {
	// ваш код
	return nil, nil
}

func (FreeVars) Bin(_ byte, l, r []string) ([]string, error) {
	// ваш код
	return nil, nil
}
