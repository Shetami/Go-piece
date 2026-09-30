package main

// Number — типы, которые можно складывать и вычитать, включая свои типы
// на их основе (type Cents int64).
type Number interface {
	~int | ~int32 | ~int64 | ~float64
}

// Fenwick — дерево Фенвика: изменение элемента и сумма префикса за
// O(log n). Индексы снаружи — с нуля.
type Fenwick[T Number] struct {
	tree []T // tree[i] (i с единицы) — сумма отрезка (i - lowbit(i), i]
	vals []T // текущие значения — нужны для Set
}

// NewFenwick создаёт дерево из n нулей.
func NewFenwick[T Number](n int) *Fenwick[T] {
	return &Fenwick[T]{tree: make([]T, n+1), vals: make([]T, n)}
}

// Add прибавляет delta к элементу i.
func (f *Fenwick[T]) Add(i int, delta T) {
	f.vals[i] += delta
	// Внутри индексы с единицы: у нуля нет младшего бита, i += i&-i
	// на нуле стоял бы на месте вечно.
	for i++; i < len(f.tree); i += i & -i {
		f.tree[i] += delta
	}
}

// Set делает элемент i равным v.
func (f *Fenwick[T]) Set(i int, v T) {
	f.Add(i, v-f.vals[i])
}

// Sum возвращает сумму элементов [0, i).
func (f *Fenwick[T]) Sum(i int) T {
	var s T // нулевое значение любого числового T — ноль
	for ; i > 0; i -= i & -i {
		s += f.tree[i]
	}
	return s
}

// RangeSum возвращает сумму элементов [l, r); при l >= r — ноль.
func (f *Fenwick[T]) RangeSum(l, r int) T {
	if l >= r {
		return 0 // нетипизированная константа 0 подходит любому T из Number
	}
	return f.Sum(r) - f.Sum(l)
}
