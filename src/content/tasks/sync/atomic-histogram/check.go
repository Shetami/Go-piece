package main

func TestHistogramBuckets(t *testing.T) {
	bounds := []float64{0.1, 0.5, 1}
	h := NewHistogram(bounds)
	bounds[0] = 100 // вызывающий испортил свой слайс
	for _, v := range []float64{0.05, 0.1, 0.3, 0.5, 0.7, 1, 2, 30} {
		h.Observe(v)
	}
	s := h.Snapshot()
	if !reflect.DeepEqual(s.Bounds, []float64{0.1, 0.5, 1}) {
		t.Fatalf("Bounds = %v, ожидали [0.1 0.5 1] — конструктор должен скопировать слайс", s.Bounds)
	}
	want := []uint64{2, 4, 6, 8}
	if !reflect.DeepEqual(s.Counts, want) {
		t.Fatalf("Counts = %v, ожидали %v (кумулятивно; значение на границе попадает в её бакет; последний — +Inf)", s.Counts, want)
	}
	if s.Count != 8 || math.Abs(s.Sum-34.65) > 1e-9 {
		t.Fatalf("Count, Sum = %d, %v, ожидали 8, 34.65", s.Count, s.Sum)
	}
	s.Bounds[0] = -1
	s.Counts[0] = 999
	if s2 := h.Snapshot(); s2.Bounds[0] != 0.1 || s2.Counts[0] != 2 {
		t.Fatalf("изменение снимка повлияло на гистограмму: %v %v", s2.Bounds, s2.Counts)
	}
}

func TestHistogramEmpty(t *testing.T) {
	s := NewHistogram([]float64{1, 2}).Snapshot()
	if len(s.Counts) != 3 || s.Count != 0 || s.Sum != 0 {
		t.Fatalf("пустая гистограмма: %+v, ожидали Counts [0 0 0], Count 0, Sum 0", s)
	}
}

func TestHistogramConcurrent(t *testing.T) {
	h := NewHistogram([]float64{0.25, 1})
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 1000 {
				if (i+g)%2 == 0 {
					h.Observe(0.25)
				} else {
					h.Observe(1.5)
				}
			}
		}()
	}
	wg.Wait()
	s := h.Snapshot()
	if s.Count != 8000 || s.Counts[2] != 8000 || s.Counts[0] != 4000 || s.Counts[1] != 4000 {
		t.Fatalf("после 8000 конкурентных Observe: Count %d, Counts %v, ожидали 8000 и [4000 4000 8000]", s.Count, s.Counts)
	}
	if s.Sum != 7000 {
		t.Fatalf("Sum = %v, ожидали 7000 — конкурентные прибавления к сумме теряются", s.Sum)
	}
}
