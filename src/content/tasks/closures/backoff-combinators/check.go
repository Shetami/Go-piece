package main

func chkConst(r float64) func() float64 { return func() float64 { return r } }

func TestExponential(t *testing.T) {
	b := Exponential(100 * time.Millisecond)
	want := map[int]time.Duration{0: 100 * time.Millisecond, 1: 100 * time.Millisecond, 2: 200 * time.Millisecond, 4: 800 * time.Millisecond}
	for a, w := range want {
		if got := b(a); got != w {
			t.Fatalf("Exponential(100ms)(%d) = %v, ожидали %v", a, got, w)
		}
	}
}

func TestExponentialSaturates(t *testing.T) {
	b := Exponential(time.Second)
	prev := time.Duration(0)
	for a := 1; a <= 300; a++ {
		d := b(a)
		if d < prev {
			t.Fatalf("Exponential(1s)(%d) = %v меньше, чем на попытке %d (%v) — переполнение", a, d, a-1, prev)
		}
		prev = d
	}
	if prev != math.MaxInt64 {
		t.Fatalf("Exponential(1s)(300) = %v, ожидали насыщение до math.MaxInt64", prev)
	}
}

func TestCapped(t *testing.T) {
	b := Capped(Exponential(100*time.Millisecond), 5*time.Second)
	for a, w := range map[int]time.Duration{3: 400 * time.Millisecond, 7: 5 * time.Second, 70: 5 * time.Second, 1000: 5 * time.Second} {
		if got := b(a); got != w {
			t.Fatalf("Capped(Exponential(100ms), 5s)(%d) = %v, ожидали %v", a, got, w)
		}
	}
}

func TestJitter(t *testing.T) {
	base := Exponential(time.Second)
	if got := Jitter(base, 0.5, chkConst(0))(2); got != 2*time.Second {
		t.Fatalf("Jitter при rnd()=0: %v, ожидали 2s", got)
	}
	if got := Jitter(base, 0.5, chkConst(0.5))(2); got != 1500*time.Millisecond {
		t.Fatalf("Jitter(frac=0.5) при rnd()=0.5: %v, ожидали 1.5s", got)
	}
	if got := Jitter(base, 3, chkConst(0.25))(1); got != 750*time.Millisecond {
		t.Fatalf("frac=3 должен прижаться к 1: %v, ожидали 750ms", got)
	}
	for _, r := range []float64{0, 0.001, 0.5, 0.999} {
		if got := Jitter(base, 0.5, chkConst(r))(500); got <= 0 {
			t.Fatalf("Jitter от насыщенной паузы при rnd()=%v: %v — переполнение при переводе из float64", r, got)
		}
	}
}

func TestMaxAttemptsAndStop(t *testing.T) {
	b := Jitter(Capped(MaxAttempts(Exponential(time.Second), 3), time.Minute), 0.5, chkConst(0.5))
	if got := b(3); got != 3*time.Second {
		t.Fatalf("третья попытка: 4s с джиттером 25%%: %v, ожидали 3s", got)
	}
	if got := b(4); got != Stop {
		t.Fatalf("четвёртая попытка при MaxAttempts 3: %v, ожидали Stop — Stop не должен искажаться обёртками", got)
	}
}
