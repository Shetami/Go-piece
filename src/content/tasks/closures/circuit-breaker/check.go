package main

type chkBrClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkBrClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *chkBrClock) Add(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

var chkDown = errors.New("upstream down")

// chkUpstream — зависимость, у которой можно переключать ответ и считать вызовы.
type chkUpstream struct {
	fail  atomic.Bool
	calls atomic.Int32
}

func (u *chkUpstream) Call() error {
	u.calls.Add(1)
	if u.fail.Load() {
		return chkDown
	}
	return nil
}

func TestBreakerConsecutive(t *testing.T) {
	clk := &chkBrClock{t: time.Unix(1_700_000_000, 0)}
	u := &chkUpstream{}
	b := Breaker(u.Call, 3, time.Minute, clk.Now)
	for _, fail := range []bool{true, true, false, true, true} {
		u.fail.Store(fail)
		b()
	}
	u.fail.Store(false)
	if err := b(); err != nil {
		t.Fatalf("ошибки были не подряд (успех посередине) — предохранитель не должен размыкаться: %v", err)
	}
	u.fail.Store(true)
	for i := range 3 {
		if err := b(); !errors.Is(err, chkDown) {
			t.Fatalf("ошибка %d: ожидали ошибку call как есть, получили %v", i+1, err)
		}
	}
	before := u.calls.Load()
	if err := b(); !errors.Is(err, ErrOpen) {
		t.Fatalf("после 3 ошибок подряд ожидали ErrOpen, получили %v", err)
	}
	if u.calls.Load() != before {
		t.Fatal("разомкнутый предохранитель не должен вызывать call")
	}
}

func TestBreakerCooldownAndProbe(t *testing.T) {
	clk := &chkBrClock{t: time.Unix(1_700_000_000, 0)}
	u := &chkUpstream{}
	u.fail.Store(true)
	b := Breaker(u.Call, 2, time.Minute, clk.Now)
	b()
	b()
	clk.Add(59 * time.Second)
	if err := b(); !errors.Is(err, ErrOpen) {
		t.Fatalf("через 59 с при cooldown 1 мин ожидали ErrOpen, получили %v", err)
	}
	clk.Add(time.Second)
	if err := b(); !errors.Is(err, chkDown) {
		t.Fatalf("через cooldown пробный вызов должен дойти до call: %v", err)
	}
	clk.Add(30 * time.Second)
	if err := b(); !errors.Is(err, ErrOpen) {
		t.Fatalf("неудачная проба размыкает заново на полный cooldown: через 30 с ожидали ErrOpen, получили %v", err)
	}
	clk.Add(30 * time.Second)
	u.fail.Store(false)
	if err := b(); err != nil {
		t.Fatalf("удачная проба: %v", err)
	}
	u.fail.Store(true)
	if err := b(); !errors.Is(err, chkDown) {
		t.Fatalf("после удачной пробы предохранитель замкнут, счётчик с нуля — одна ошибка не размыкает: %v", err)
	}
	if err := b(); !errors.Is(err, chkDown) {
		t.Fatalf("вторая ошибка подряд возвращается как есть: %v", err)
	}
	if err := b(); !errors.Is(err, ErrOpen) {
		t.Fatalf("после двух ошибок подряд снова разомкнут: %v", err)
	}
}

func TestBreakerSingleProbe(t *testing.T) {
	clk := &chkBrClock{t: time.Unix(1_700_000_000, 0)}
	var fail atomic.Bool
	fail.Store(true)
	var calls atomic.Int32
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	b := Breaker(func() error {
		calls.Add(1)
		if fail.Load() {
			return chkDown
		}
		entered <- struct{}{}
		<-release
		return nil
	}, 1, time.Second, clk.Now)
	b()
	clk.Add(time.Second)
	fail.Store(false)
	probeDone := make(chan error, 1)
	go func() { probeDone <- b() }()
	<-entered
	others := make(chan error, 1)
	go func() { others <- b() }()
	select {
	case err := <-others:
		if !errors.Is(err, ErrOpen) {
			t.Fatalf("пока идёт проба, остальные должны сразу получать ErrOpen, получили %v", err)
		}
	case <-time.After(2 * time.Second):
		close(release)
		t.Fatal("вызов во время пробы повис — call вызывается под блокировкой")
	}
	close(release)
	if err := <-probeDone; err != nil {
		t.Fatalf("проба: %v", err)
	}
	if n := calls.Load(); n != 2 {
		t.Fatalf("call вызвана %d раз, ожидали 2 (ошибка + одна проба)", n)
	}
}
