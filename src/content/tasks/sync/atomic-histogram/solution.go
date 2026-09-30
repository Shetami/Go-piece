package main

import (
	"math"
	"slices"
	"sort"
	"sync/atomic"
)

// Histogram — гистограмма латентностей в духе Prometheus, в которую
// пишут все обработчики одновременно. Observe не должен брать мьютекс.
type Histogram struct {
	bounds  []float64
	buckets []atomic.Uint64 // НЕ кумулятивно: ровно свой бакет, последний — +Inf
	count   atomic.Uint64
	sumBits atomic.Uint64 // float64 в виде битов: атомарного float в Go нет
}

// HistogramSnapshot — состояние гистограммы.
type HistogramSnapshot struct {
	Bounds []float64 // верхние границы бакетов (копия)
	// Counts[i] — сколько наблюдений <= Bounds[i] (кумулятивно, как le
	// в Prometheus); последний элемент, len(Bounds), — все наблюдения (+Inf).
	Counts []uint64
	Count  uint64  // всего наблюдений
	Sum    float64 // сумма всех наблюдений
}

// NewHistogram создаёт гистограмму. bounds — строго возрастающие верхние
// границы; вызывающий может потом менять свой слайс.
func NewHistogram(bounds []float64) *Histogram {
	return &Histogram{
		bounds:  slices.Clone(bounds),
		buckets: make([]atomic.Uint64, len(bounds)+1),
	}
}

// Observe учитывает одно наблюдение v: оно попадает в первый бакет,
// граница которого >= v, или в +Inf.
func (h *Histogram) Observe(v float64) {
	// SearchFloat64s — первый индекс с bounds[i] >= v: граница включительна.
	h.buckets[sort.SearchFloat64s(h.bounds, v)].Add(1)
	for {
		// Прибавить к float атомарно можно только циклом CAS над битами.
		old := h.sumBits.Load()
		next := math.Float64bits(math.Float64frombits(old) + v)
		if h.sumBits.CompareAndSwap(old, next) {
			break
		}
	}
	h.count.Add(1)
}

// Snapshot возвращает состояние. Во время конкурентной записи снимок
// может быть чуть несогласован между полями — это допустимо.
func (h *Histogram) Snapshot() HistogramSnapshot {
	s := HistogramSnapshot{
		Bounds: slices.Clone(h.bounds),
		Counts: make([]uint64, len(h.buckets)),
		Count:  h.count.Load(),
		Sum:    math.Float64frombits(h.sumBits.Load()),
	}
	var acc uint64
	for i := range h.buckets {
		acc += h.buckets[i].Load()
		s.Counts[i] = acc
	}
	return s
}
