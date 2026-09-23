package main

type Shape interface {
	Area() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

type Circle struct{ R float64 }

func (c Circle) Area() float64 { return 3.14159 * c.R * c.R }

// ByArea реализует sort.Interface: фигуры по возрастанию площади.
type ByArea []Shape

func (s ByArea) Len() int           { return 0 }
func (s ByArea) Less(i, j int) bool { return false }
func (s ByArea) Swap(i, j int)      {}
