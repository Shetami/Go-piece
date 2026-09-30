package main

var chkT0 = time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

func chkAt(sec float64) time.Time { return chkT0.Add(time.Duration(sec * float64(time.Second))) }

func TestLimiterWindow(t *testing.T) {
	l := NewLimiter(3, 10*time.Second)
	steps := []struct {
		at   float64
		want bool
	}{
		{0, true}, {1, true}, {2, true}, {3, false}, {9.9, false},
		{10, true},  // событие в 0 ровно 10 с назад — уже вне окна
		{10.5, false}, // в окне 1, 2, 10
		{11, true},  // 1 выпало
		{12, true},  // 2 выпало
		{12, false},
		{100, true}, {100, true}, {100, true}, {100, false},
	}
	for _, s := range steps {
		if got := l.Allow(chkAt(s.at)); got != s.want {
			t.Fatalf("Allow(t=%v) = %v, ожидали %v", s.at, got, s.want)
		}
		if len(l.times) > 3 {
			t.Fatalf("len(times) = %d больше лимита 3", len(l.times))
		}
	}
}

func TestLimiterDeniedNotRecorded(t *testing.T) {
	l := NewLimiter(1, 10*time.Second)
	l.Allow(chkAt(0))
	for s := 1.0; s < 10; s++ {
		l.Allow(chkAt(s)) // отказы не должны продлевать блокировку
	}
	if !l.Allow(chkAt(10)) {
		t.Fatalf("в t=10 должно пропустить: отказанные события не считаются")
	}
}

func TestLimiterNoAllocSteadyState(t *testing.T) {
	l := NewLimiter(100, time.Second)
	now := chkT0
	for range 1000 { // разогрев
		now = now.Add(7 * time.Millisecond)
		l.Allow(now)
	}
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	for range 20000 {
		now = now.Add(7 * time.Millisecond)
		l.Allow(now)
	}
	runtime.ReadMemStats(&m1)
	if d := m1.Mallocs - m0.Mallocs; d > 5 {
		t.Fatalf("20000 вызовов Allow сделали %d аллокаций — массив times перевыделяется", d)
	}
	if cap(l.times) > 200 {
		t.Fatalf("cap(times) = %d при лимите 100 — массив разрастается", cap(l.times))
	}
}

func TestLimiterPanics(t *testing.T) {
	for _, c := range []struct {
		n int
		w time.Duration
	}{{0, time.Second}, {1, 0}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("NewLimiter(%d, %v): ожидали панику", c.n, c.w)
				}
			}()
			NewLimiter(c.n, c.w)
		}()
	}
}
