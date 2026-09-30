package main

func chkDialer() (func() (*Conn, error), *atomic.Int32) {
	var n atomic.Int32
	return func() (*Conn, error) { return &Conn{ID: int(n.Add(1))}, nil }, &n
}

func TestPoolReuseLIFO(t *testing.T) {
	dial, dials := chkDialer()
	p := NewPool(3, dial)
	ctx := context.Background()
	a, _ := p.Get(ctx)
	b, _ := p.Get(ctx)
	p.Put(a)
	p.Put(b)
	c, err := p.Get(ctx)
	if err != nil || c != b {
		t.Fatalf("Get после Put(a), Put(b) вернул %v, %v, ожидали последнее возвращённое соединение b", c, err)
	}
	if dials.Load() != 2 || p.Open() != 2 {
		t.Fatalf("dial вызван %d раз, Open() = %d, ожидали 2 и 2 — простаивающее соединение не переиспользовано", dials.Load(), p.Open())
	}
}

func TestPoolLimitAndWait(t *testing.T) {
	dial, _ := chkDialer()
	p := NewPool(2, dial)
	ctx := context.Background()
	a, _ := p.Get(ctx)
	p.Get(ctx)
	short, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()
	if c, err := p.Get(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get при исчерпанном пуле = %v, %v, ожидали DeadlineExceeded", c, err)
	}
	got := make(chan *Conn, 1)
	go func() { c, _ := p.Get(ctx); got <- c }()
	time.Sleep(20 * time.Millisecond)
	p.Put(a)
	select {
	case c := <-got:
		if c != a {
			t.Fatalf("ждущий получил %v, ожидали возвращённое соединение", c)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Put не разбудил ждущего Get")
	}
	if p.Open() != 2 {
		t.Fatalf("Open() = %d, ожидали 2", p.Open())
	}
}

func TestPoolDiscardWakesWaiter(t *testing.T) {
	dial, dials := chkDialer()
	p := NewPool(1, dial)
	ctx := context.Background()
	a, _ := p.Get(ctx)
	got := make(chan *Conn, 1)
	go func() { c, _ := p.Get(ctx); got <- c }()
	time.Sleep(20 * time.Millisecond)
	p.Discard(a)
	select {
	case c := <-got:
		if c == nil || c == a || dials.Load() != 2 {
			t.Fatalf("после Discard ждущий получил %v (dial вызван %d раз), ожидали новое соединение", c, dials.Load())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Discard освободил место, но ждущий Get так и не проснулся")
	}
}

func TestPoolDialError(t *testing.T) {
	boom := errors.New("connection refused")
	fail := atomic.Bool{}
	fail.Store(true)
	p := NewPool(1, func() (*Conn, error) {
		if fail.Load() {
			return nil, boom
		}
		return &Conn{ID: 1}, nil
	})
	for range 3 {
		if _, err := p.Get(context.Background()); !errors.Is(err, boom) {
			t.Fatalf("Get при ошибке dial = %v, ожидали её", err)
		}
	}
	fail.Store(false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if c, err := p.Get(ctx); err != nil || c == nil {
		t.Fatalf("после трёх ошибок dial Get = %v, %v — неудачный dial занял место в пуле навсегда", c, err)
	}
}

func TestPoolStress(t *testing.T) {
	dial, _ := chkDialer()
	p := NewPool(3, dial)
	var inUse, maxInUse atomic.Int32
	var wg sync.WaitGroup
	for g := range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1+(g+i)%3)*time.Millisecond)
				c, err := p.Get(ctx)
				cancel()
				if err != nil {
					continue
				}
				n := inUse.Add(1)
				for {
					m := maxInUse.Load()
					if n <= m || maxInUse.CompareAndSwap(m, n) {
						break
					}
				}
				runtime.Gosched()
				inUse.Add(-1)
				if i%7 == 0 {
					p.Discard(c)
				} else {
					p.Put(c)
				}
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("12 горутин с таймаутами и Discard зависли на пуле из 3 соединений — потерянная побудка?")
	}
	if m := maxInUse.Load(); m > 3 {
		t.Fatalf("одновременно выдано %d соединений при max 3", m)
	}
	if o := p.Open(); o < 0 || o > 3 {
		t.Fatalf("Open() = %d после нагрузки, ожидали от 0 до 3", o)
	}
}
