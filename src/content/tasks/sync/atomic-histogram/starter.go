package main

// Histogram — гистограмма латентностей в духе Prometheus, в которую
// пишут все обработчики одновременно. Observe не должен брать мьютекс.
type Histogram struct {
	// ваши поля
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
	// ваш код
	return &Histogram{}
}

// Observe учитывает одно наблюдение v: оно попадает в первый бакет,
// граница которого >= v, или в +Inf.
func (h *Histogram) Observe(v float64) {
	// ваш код
}

// Snapshot возвращает состояние. Во время конкурентной записи снимок
// может быть чуть несогласован между полями — это допустимо.
func (h *Histogram) Snapshot() HistogramSnapshot {
	// ваш код
	return HistogramSnapshot{}
}
