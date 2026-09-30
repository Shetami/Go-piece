package main

func TestStatsClasses(t *testing.T) {
	var s Stats
	for _, st := range []int{200, 204, 301, 399, 400, 404, 499, 500, 503, 599} {
		s.Record(st, 10)
	}
	got := s.Snapshot()
	want := Snapshot{Requests: 10, ClientErrors: 3, ServerErrors: 3, Bytes: 100}
	if got != want {
		t.Fatalf("Snapshot() = %+v, ожидали %+v (границы 399/400, 499/500, 599)", got, want)
	}
	if r := s.ErrorRate(); r != 0.3 {
		t.Fatalf("ErrorRate() = %v, ожидали 0.3", r)
	}
}

func TestStatsEmpty(t *testing.T) {
	var s Stats
	if r := s.ErrorRate(); r != 0 {
		t.Fatalf("ErrorRate() без запросов = %v, ожидали 0 (а не NaN)", r)
	}
	if got := s.Snapshot(); got != (Snapshot{}) {
		t.Fatalf("пустой Snapshot() = %+v", got)
	}
}

func TestStatsConcurrent(t *testing.T) {
	var s Stats
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 2000 {
				st := 200
				if (i+g)%4 == 0 {
					st = 502
				}
				s.Record(st, 3)
				if r := s.ErrorRate(); r < 0 || r > 1 {
					t.Errorf("ErrorRate() = %v во время записи — доля вне [0, 1]", r)
					return
				}
			}
		}()
	}
	wg.Wait()
	got := s.Snapshot()
	want := Snapshot{Requests: 16000, ServerErrors: 4000, Bytes: 48000}
	if got != want {
		t.Fatalf("после 8×2000 конкурентных Record: %+v, ожидали %+v — потерянные обновления", got, want)
	}
}
