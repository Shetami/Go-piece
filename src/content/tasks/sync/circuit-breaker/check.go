package main

type chkClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkClock) Now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *chkClock) Add(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

var errChkDown = errors.New("зависимость лежит")

func chkFail() error { return errChkDown }
func chkOK() error   { return nil }

func TestBreakerOpensAfterConsecutive(t *testing.T) {
	clk := &chkClock{t: time.Unix(0, 0)}
	b := NewBreaker(3, 10*time.Second, clk.Now)
	b.Call(chkFail)
	b.Call(chkFail)
	b.Call(chkOK)
	b.Call(chkFail)
	b.Call(chkFail)
	if s := b.State(); s != "closed" {
		t.Fatalf("ошибки не подряд (успех между ними), а State() = %s, ожидали closed", s)
	}
	if err := b.Call(chkFail); !errors.Is(err, errChkDown) {
		t.Fatalf("третья ошибка подряд должна вернуться как есть, получили %v", err)
	}
	if s := b.State(); s != "open" {
		t.Fatalf("после 3 ошибок подряд State() = %s, ожидали open", s)
	}
	called := false
	if err := b.Call(func() error { called = true; return nil }); !errors.Is(err, ErrOpen) || called {
		t.Fatalf("разомкнутый выключатель: Call = %v, fn вызвана = %v; ожидали ErrOpen без вызова", err, called)
	}
	clk.Add(9 * time.Second)
	if err := b.Call(chkOK); !errors.Is(err, ErrOpen) {
		t.Fatalf("cooldown не прошёл, а Call = %v", err)
	}
}

func TestBreakerProbe(t *testing.T) {
	clk := &chkClock{t: time.Unix(0, 0)}
	b := NewBreaker(1, 10*time.Second, clk.Now)
	b.Call(chkFail)
	clk.Add(10 * time.Second)
	if err := b.Call(chkFail); !errors.Is(err, errChkDown) {
		t.Fatalf("после cooldown пробный вызов должен выполниться, Call = %v", err)
	}
	if s := b.State(); s != "open" {
		t.Fatalf("проба упала, State() = %s, ожидали open", s)
	}
	clk.Add(9 * time.Second)
	if err := b.Call(chkOK); !errors.Is(err, ErrOpen) {
		t.Fatalf("после неудачной пробы cooldown отсчитывается заново, а Call = %v", err)
	}
	clk.Add(time.Second)
	if err := b.Call(chkOK); err != nil {
		t.Fatalf("вторая проба: %v", err)
	}
	if s := b.State(); s != "closed" {
		t.Fatalf("после удачной пробы State() = %s, ожидали closed", s)
	}
}

func TestBreakerSingleProbe(t *testing.T) {
	clk := &chkClock{t: time.Unix(0, 0)}
	b := NewBreaker(1, time.Second, clk.Now)
	b.Call(chkFail)
	clk.Add(time.Second)
	release := make(chan struct{})
	var calls atomic.Int32
	probeDone := make(chan error, 1)
	go func() {
		probeDone <- b.Call(func() error { calls.Add(1); <-release; return nil })
	}()
	for calls.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	if s := b.State(); s != "half-open" {
		t.Fatalf("во время пробы State() = %s, ожидали half-open", s)
	}
	var wg sync.WaitGroup
	var rejected atomic.Int32
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.Call(func() error { calls.Add(1); return nil }); errors.Is(err, ErrOpen) {
				rejected.Add(1)
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		close(release)
		t.Fatal("вызовы во время пробы ждут её окончания — fn выполняется под мьютексом?")
	}
	close(release)
	<-probeDone
	if calls.Load() != 1 || rejected.Load() != 20 {
		t.Fatalf("во время пробы fn вызвана %d раз, отклонено %d из 20; ожидали ровно одну пробу", calls.Load(), rejected.Load())
	}
	if s := b.State(); s != "closed" {
		t.Fatalf("после удачной пробы State() = %s", s)
	}
}
