package main

// Number — типы, которые можно складывать и вычитать, включая свои типы
// на их основе (type Cents int64).
type Number interface {
	~int | ~int32 | ~int64 | ~float64
}

// Fenwick — дерево Фенвика: изменение элемента и сумма префикса за
// O(log n). Индексы снаружи — с нуля.
type Fenwick[T Number] struct {
	// ваши поля
}

// NewFenwick создаёт дерево из n нулей.
func NewFenwick[T Number](n int) *Fenwick[T] {
	// ваш код
	return &Fenwick[T]{}
}

// Add прибавляет delta к элементу i.
func (f *Fenwick[T]) Add(i int, delta T) {
	// ваш код
}

// Set делает элемент i равным v.
func (f *Fenwick[T]) Set(i int, v T) {
	// ваш код
}

// Sum возвращает сумму элементов [0, i).
func (f *Fenwick[T]) Sum(i int) T {
	// ваш код
	return 0
}

// RangeSum возвращает сумму элементов [l, r); при l >= r — ноль.
func (f *Fenwick[T]) RangeSum(l, r int) T {
	// ваш код
	return 0
}
