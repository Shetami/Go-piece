package main

func TestStatsBasic(t *testing.T) {
	add, snap := NewStats()
	for _, x := range []float64{4, 1, 7} {
		add(x)
	}
	want := Stats{Count: 3, Min: 1, Max: 7, Mean: 4}
	if got := snap(); got != want {
		t.Fatalf("после 4, 1, 7: %+v, ожидали %+v", got, want)
	}
}

func TestStatsNegativeAndPositive(t *testing.T) {
	add, snap := NewStats()
	add(-5)
	add(-2)
	if got := snap(); got.Max != -2 || got.Min != -5 {
		t.Fatalf("после -5, -2: Min=%v Max=%v, ожидали -5 и -2", got.Min, got.Max)
	}
	add2, snap2 := NewStats()
	add2(3)
	add2(8)
	if got := snap2(); got.Min != 3 {
		t.Fatalf("после 3, 8: Min=%v, ожидали 3", got.Min)
	}
}

func TestStatsEmptyAndNaN(t *testing.T) {
	add, snap := NewStats()
	if got := snap(); got != (Stats{}) {
		t.Fatalf("пустая сводка: %+v, ожидали нули (и никакого NaN в Mean)", got)
	}
	add(math.NaN())
	add(2)
	add(math.NaN())
	if got := snap(); got != (Stats{Count: 1, Min: 2, Max: 2, Mean: 2}) {
		t.Fatalf("NaN должны игнорироваться: %+v", got)
	}
}

func TestStatsSnapshotAndIndependence(t *testing.T) {
	addA, snapA := NewStats()
	addB, snapB := NewStats()
	addA(10)
	before := snapA()
	addA(20)
	if before.Count != 1 || before.Max != 10 {
		t.Fatalf("старый снимок изменился после add: %+v", before)
	}
	addB(1)
	if a, b := snapA(), snapB(); a.Count != 2 || b.Count != 1 || b.Max != 1 {
		t.Fatalf("счётчики должны быть независимы: A=%+v B=%+v", a, b)
	}
}
