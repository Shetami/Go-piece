package main

func chkPattern(s *Sampler, k int) string {
	var b strings.Builder
	for i := 0; i < k; i++ {
		if s.Allow() {
			b.WriteByte('+')
		} else {
			b.WriteByte('.')
		}
	}
	return b.String()
}

func TestSamplerLoop(t *testing.T) {
	s := NewSampler(3)
	if got, want := chkPattern(s, 10), "+..+..+..+"; got != want {
		t.Fatalf("n=3, 10 вызовов из одного места: %s, ожидали %s", got, want)
	}
	if got := chkPattern(NewSampler(1), 5); got != "+++++" {
		t.Fatalf("n=1: %s, ожидали +++++", got)
	}
	if got := chkPattern(NewSampler(0), 3); got != "+++" {
		t.Fatalf("n=0: %s, ожидали +++", got)
	}
}

func TestSamplerSites(t *testing.T) {
	s := NewSampler(5)
	a1 := s.Allow()
	b1 := s.Allow()
	if !a1 || !b1 {
		t.Fatalf("первые вызовы из двух разных строк одной функции: %v %v, ожидали true true", a1, b1)
	}
	x, y := s.Allow(), s.Allow()
	if !x || !y {
		t.Fatalf("два вызова в одной строке — разные места, первый вызов каждого должен пройти: %v %v", x, y)
	}
	if got, want := chkPattern(s, 6), "+....+"; got != want {
		t.Fatalf("место внутри другой функции: %s, ожидали %s", got, want)
	}
}

func TestSamplerConcurrent(t *testing.T) {
	s := NewSampler(10)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if s.Allow() {
					allowed.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	if got := allowed.Load(); got != 80 {
		t.Fatalf("800 вызовов из одного места при n=10 пропустили %d, ожидали ровно 80", got)
	}
}
