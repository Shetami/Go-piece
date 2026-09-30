package main

var chkSink []byte
var chkSinkP *[64]byte

func TestMeasureCounts(t *testing.T) {
	if s := Measure(100, func() {}); s.Allocs != 0 || s.Bytes != 0 {
		t.Fatalf("пустая функция: %+v, ожидали 0 выделений и 0 байт", s)
	}
	s := Measure(100, func() {
		chkSink = make([]byte, 1000)
		chkSinkP = new([64]byte)
		chkSinkP = new([64]byte)
	})
	if s.Allocs != 3 {
		t.Fatalf("три выделения за запуск: Allocs = %d, ожидали 3", s.Allocs)
	}
	if s.Bytes < 1000+128 || s.Bytes > 2*(1000+128) {
		t.Fatalf("Bytes = %d, ожидали около 1150 (1000 + 2×64 с округлением до классов размеров)", s.Bytes)
	}
}

var chkLazy map[int][]byte

func TestMeasureWarmup(t *testing.T) {
	chkLazy = nil
	s := Measure(50, func() {
		if chkLazy == nil {
			chkLazy = make(map[int][]byte)
			for i := 0; i < 200; i++ {
				chkLazy[i] = make([]byte, 100)
			}
		}
	})
	if s.Allocs != 0 {
		t.Fatalf("ленивая инициализация при первом запуске попала в среднее: Allocs = %d, ожидали 0", s.Allocs)
	}
}

func TestMeasureRestoresProcs(t *testing.T) {
	prev := runtime.GOMAXPROCS(3)
	defer runtime.GOMAXPROCS(prev)
	var inside int
	Measure(5, func() { inside = runtime.GOMAXPROCS(0) })
	if inside != 1 {
		t.Fatalf("во время замера GOMAXPROCS = %d, ожидали 1 — иначе чужие горутины параллельно портят счётчики", inside)
	}
	if got := runtime.GOMAXPROCS(0); got != 3 {
		t.Fatalf("после Measure GOMAXPROCS = %d, ожидали прежнее 3", got)
	}
}

func TestMeasureMatchesTesting(t *testing.T) {
	f := func() {
		chkSink = []byte(strconv.Itoa(len(chkSink) + 1000000))
	}
	want := testing.AllocsPerRun(100, f)
	if got := Measure(100, f); float64(got.Allocs) != want {
		t.Fatalf("Measure = %d выделений, testing.AllocsPerRun = %.0f", got.Allocs, want)
	}
}
