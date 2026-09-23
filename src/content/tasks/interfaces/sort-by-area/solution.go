package main

type Shape interface {
	Area() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

type Circle struct{ R float64 }

func (c Circle) Area() float64 { return 3.14159 * c.R * c.R }

// ByArea реализует sort.Interface: фигуры по возрастанию площади.
// Методы на значении: слайс и так ссылается на общий массив,
// Swap меняет элементы именно в нём.
type ByArea []Shape

func (s ByArea) Len() int           { return len(s) }
func (s ByArea) Less(i, j int) bool { return s[i].Area() < s[j].Area() }
func (s ByArea) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
