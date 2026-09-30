package main

type chkWClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkWClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

// Set ставит время: старт + ms миллисекунд.
func (c *chkWClock) Set(ms int) {
	c.mu.Lock()
	c.t = time.Unix(1_700_000_000, 0).Add(time.Duration(ms) * time.Millisecond)
	c.mu.Unlock()
}

func TestWindowSlides(t *testing.T) {
	clk := &chkWClock{}
	clk.Set(0)
	add, count := NewWindowCounter(10*time.Second, 10, clk.Now)
	clk.Set(500)
	add(1)
	clk.Set(3200)
	add(2)
	clk.Set(9900)
	add(3)
	for _, c := range []struct{ ms, want int }{{9900, 6}, {10000, 5}, {13100, 3}, {18500, 3}, {19950, 0}} {
		clk.Set(c.ms)
		if got := count(); got != c.want {
			t.Fatalf("count() в %.2f с = %d, ожидали %d (события: 1 в 0.5 с, 2 в 3.2 с, 3 в 9.9 с; окно 10 с по корзинам в 1 с)", float64(c.ms)/1000, got, c.want)
		}
	}
}

func TestWindowSlotReuse(t *testing.T) {
	clk := &chkWClock{}
	clk.Set(1000)
	add, count := NewWindowCounter(10*time.Second, 10, clk.Now)
	add(5)
	clk.Set(21000) // та же ячейка кольца, другая корзина
	add(1)
	if got := count(); got != 1 {
		t.Fatalf("после 20 с простоя: count() = %d, ожидали 1 — ячейка кольца теперь хранит новую корзину: старое из неё не считается, новое — считается", got)
	}
	clk.Set(35000)
	if got := count(); got != 0 {
		t.Fatalf("через 14 с после последнего события: count() = %d, ожидали 0", got)
	}
}

func TestWindowLateEvents(t *testing.T) {
	clk := &chkWClock{}
	clk.Set(22000)
	add, count := NewWindowCounter(10*time.Second, 10, clk.Now)
	add(1)
	clk.Set(20500) // опоздавшее, но ещё в окне
	add(4)
	clk.Set(12500) // та же ячейка, что у 22 с, но корзина старее
	add(7)
	clk.Set(22500)
	if got := count(); got != 5 {
		t.Fatalf("count() = %d, ожидали 5: опоздавшее в окне учитывается, вытесненное — нет и не портит новую корзину", got)
	}
}

func TestWindowConcurrent(t *testing.T) {
	clk := &chkWClock{}
	clk.Set(0)
	add, count := NewWindowCounter(time.Minute, 60, clk.Now)
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			add(2)
			count()
		}()
	}
	wg.Wait()
	if got := count(); got != 100 {
		t.Fatalf("50 горутин по add(2): count() = %d, ожидали 100", got)
	}
}
