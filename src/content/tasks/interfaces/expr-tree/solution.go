package main

import "strconv"

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

// Методы на значениях: так и Num{…}, и &Num{…} реализуют Expr.

func (n Num) Eval() float64  { return n.V }
func (n Num) String() string { return strconv.FormatFloat(n.V, 'g', -1, 64) }
func (a Add) Eval() float64  { return a.L.Eval() + a.R.Eval() }
func (a Add) String() string { return "(" + a.L.String() + " + " + a.R.String() + ")" }
func (m Mul) Eval() float64  { return m.L.Eval() * m.R.Eval() }
func (m Mul) String() string { return "(" + m.L.String() + " * " + m.R.String() + ")" }
func (n Neg) Eval() float64  { return -n.X.Eval() }
func (n Neg) String() string { return "-" + n.X.String() }
