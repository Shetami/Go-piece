package main

// Expr — узел арифметического выражения.
type Expr interface {
	Eval() float64
	String() string
}

// Реализуйте четыре типа узлов:
//
//	Num{V: 2}                → число;          String: "2"
//	Add{L: a, R: b}          → сумма;          String: "(a + b)"
//	Mul{L: a, R: b}          → произведение;   String: "(a * b)"
//	Neg{X: a}                → минус;          String: "-a"
//
// Числа печатаются через strconv.FormatFloat(v, 'g', -1, 64).

type Num struct{ V float64 }

type Add struct{ L, R Expr }

type Mul struct{ L, R Expr }

type Neg struct{ X Expr }
