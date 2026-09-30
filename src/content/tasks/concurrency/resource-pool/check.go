package main

type chkConn struct{ id int }

type chkFactory struct {
	created, closed atomic.Int32
	live, peak      atomic.Int32
}

func (f *chkFactory) make(ctx context.Context) (*chkConn, error) {
	n := f.live.Add(1)
	for {
		p := f.peak.Load()
		if n <= p || f.peak.CompareAndSwap(p, n) {
			break
		}
	}
	return &chkConn{id: int(f.created.Add(1))}, nil
}

func (f *chkFactory) close(*chkConn) { f.closed.Add(1); f.live.Add(-1) }

func chkGetAsync[R any](p *Pool[R], ctx context.Context) chan error {
	ch := make(chan error, 1)
	go func() { _, err := p.Get(ctx); ch <- err }()
	return ch
}

func TestPoolReuse(t *testing.T) {
	var f chkFactory
	p := NewPool(3, f.make, f.close)
	c1, _ := p.Get(context.Background())
	p.Put(c1)
	c2, _ := p.Get(context.Background())
	if c2 != c1 || f.created.Load() != 1 {
		t.Fatalf("после Put ожидали тот же ресурс без создания нового, создано %d", f.created.Load())
	}
}

func TestPoolLimitAndWait(t *testing.T) {
	var f chkFactory
	p := NewPool(2, f.make, f.close)
	ctx := context.Background()
	a, _ := p.Get(ctx)
	p.Get(ctx)
	tctx, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()
	if _, err := p.Get(tctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get при исчерпанном лимите с таймаутом = %v, ожидали DeadlineExceeded", err)
	}
	waiter := chkGetAsync(p, ctx)
	time.Sleep(20 * time.Millisecond)
	p.Put(a)
	select {
	case err := <-waiter:
		if err != nil {
			t.Fatalf("ждущий Get = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ждущий Get не получил ресурс после Put")
	}
	if n := f.created.Load(); n != 2 {
		t.Fatalf("создано %d ресурсов при max=2", n)
	}
}

func TestPoolFactoryErrorFreesSlot(t *testing.T) {
	fails := 2
	p := NewPool(1, func(ctx context.Context) (int, error) {
		if fails > 0 {
			fails--
			return 0, errors.New("connection refused")
		}
		return 7, nil
	}, func(int) {})
	for range 2 {
		if _, err := p.Get(context.Background()); err == nil {
			t.Fatal("factory вернула ошибку, а Get — nil")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if v, err := p.Get(ctx); err != nil || v != 7 {
		t.Fatalf("Get после ошибок factory = %d, %v; место должно освобождаться", v, err)
	}
}

func TestPoolSlowFactoryDoesNotBlock(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	var n atomic.Int32
	p := NewPool(3, func(ctx context.Context) (int, error) {
		if n.Add(1) == 2 {
			<-release // второе соединение устанавливается очень долго
		}
		return int(n.Load()), nil
	}, func(int) {})
	first, _ := p.Get(context.Background())
	go p.Get(context.Background()) // зависнет в factory
	time.Sleep(20 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		p.Put(first)
		p.Get(context.Background())
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Put/Get свободного ресурса ждут, пока медленная factory создаёт другой")
	}
}

func TestPoolCloseAndStress(t *testing.T) {
	var f chkFactory
	p := NewPool(4, f.make, f.close)
	var wg sync.WaitGroup
	for range 40 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				c, err := p.Get(context.Background())
				if err != nil {
					return
				}
				time.Sleep(50 * time.Microsecond)
				p.Put(c)
			}
		}()
	}
	wg.Wait()
	if pk := f.peak.Load(); pk > 4 {
		t.Fatalf("одновременно существовало %d ресурсов при max=4", pk)
	}
	held, _ := p.Get(context.Background())
	for range 3 { // занимаем всё, что осталось
		p.Get(context.Background())
	}
	waiter := chkGetAsync(p, context.Background())
	time.Sleep(20 * time.Millisecond)
	p.Close()
	p.Close()
	select {
	case err := <-waiter:
		if !errors.Is(err, ErrPoolClosed) {
			t.Fatalf("ждущий Get после Close = %v, ожидали ErrPoolClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ждущий Get не проснулся после Close")
	}
	before := f.closed.Load()
	p.Put(held)
	if f.closed.Load() != before+1 {
		t.Fatal("ресурс, возвращённый после Close, должен закрываться сразу")
	}
	if _, err := p.Get(context.Background()); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("Get после Close = %v, ожидали ErrPoolClosed", err)
	}
}
