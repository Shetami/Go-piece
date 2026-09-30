package main

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

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
	var zero R
	switch x := n.(type) {
	case nil:
		return zero, ErrNilNode
	case Num:
		return v.Num(x)
	case *Num:
		if x == nil { // nil-указатель в интерфейсе — не nil-интерфейс
			return zero, ErrNilNode
		}
		return v.Num(*x)
	case Var:
		return v.Var(x)
	case *Var:
		if x == nil {
			return zero, ErrNilNode
		}
		return v.Var(*x)
	case Bin:
		return walkBin(x, v)
	case *Bin:
		if x == nil {
			return zero, ErrNilNode
		}
		return walkBin(*x, v)
	default:
		return zero, fmt.Errorf("unknown node %T", n)
	}
}

func walkBin[R any](b Bin, v Visitor[R]) (R, error) {
	var zero R
	l, err := Walk(b.L, v)
	if err != nil {
		return zero, err
	}
	r, err := Walk(b.R, v)
	if err != nil {
		return zero, err
	}
	return v.Bin(b.Op, l, r)
}

// UnboundError — переменной нет в окружении.
type UnboundError struct{ Name string }

func (e *UnboundError) Error() string { return "unbound variable " + e.Name }

// Eval вычисляет выражение: переменные берутся из Env.
// Нет переменной — *UnboundError; деление на 0 — ErrDivByZero;
// неизвестная операция — ошибка.
type Eval struct{ Env map[string]float64 }

func (e Eval) Num(n Num) (float64, error) { return n.V, nil }

func (e Eval) Var(v Var) (float64, error) {
	x, ok := e.Env[v.Name]
	if !ok {
		return 0, &UnboundError{v.Name}
	}
	return x, nil
}

func (e Eval) Bin(op byte, l, r float64) (float64, error) {
	switch op {
	case '+':
		return l + r, nil
	case '-':
		return l - r, nil
	case '*':
		return l * r, nil
	case '/':
		if r == 0 {
			return 0, ErrDivByZero
		}
		return l / r, nil
	}
	return 0, fmt.Errorf("unknown op %q", op)
}

// Printer печатает выражение: числа — strconv.FormatFloat(v, 'g', -1, 64),
// переменные — по имени, Bin — "(l op r)".
type Printer struct{}

func (Printer) Num(n Num) (string, error) { return strconv.FormatFloat(n.V, 'g', -1, 64), nil }
func (Printer) Var(v Var) (string, error) { return v.Name, nil }
func (Printer) Bin(op byte, l, r string) (string, error) {
	return "(" + l + " " + string(op) + " " + r + ")", nil
}

// FreeVars собирает имена переменных: без повторов, по возрастанию.
type FreeVars struct{}

func (FreeVars) Num(Num) ([]string, error)   { return nil, nil }
func (FreeVars) Var(v Var) ([]string, error) { return []string{v.Name}, nil }
func (FreeVars) Bin(_ byte, l, r []string) ([]string, error) {
	// Новый срез: append(l, r...) мог бы писать в массив, общий с другой веткой.
	out := slices.Concat(l, r)
	slices.Sort(out)
	return slices.Compact(out), nil
}
