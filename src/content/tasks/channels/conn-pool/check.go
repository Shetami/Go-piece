package main

type chkDialer struct {
	mu            sync.Mutex
	dials, closes int
	failNext      bool
}

func (d *chkDialer) dial(ctx context.Context) (*Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.failNext {
		d.failNext = false
		return nil, errors.New("connection refused")
	}
	d.dials++
	return &Conn{ID: d.dials}, nil
}

func (d *chkDialer) close(*Conn) {
	d.mu.Lock()
	d.closes++
	d.mu.Unlock()
}

func (d *chkDialer) open() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dials - d.closes
}

func chkGet(t *testing.T, p *Pool, ms int) (*Conn, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ms)*time.Millisecond)
	defer cancel()
	return p.Get(ctx)
}

func TestPoolReusesIdle(t *testing.T) {
	d := &chkDialer{}
	p := NewPool(3, d.dial, d.close)
	for i := range 100 {
		c, err := chkGet(t, p, 1000)
		if err != nil || c == nil {
			t.Fatalf("Get = %v, %v", c, err)
		}
		p.Put(c, false)
		if d.dials != 1 {
			t.Fatalf("итерация %d: открыто %d соединений, хотя свободное было — Get должен брать его первым", i, d.dials)
		}
	}
}

func TestPoolLimitAndWait(t *testing.T) {
	d := &chkDialer{}
	p := NewPool(2, d.dial, d.close)
	a, _ := chkGet(t, p, 1000)
	b, _ := chkGet(t, p, 1000)
	if a == nil || b == nil {
		t.Fatal("не удалось взять два соединения из пула на 2")
	}
	if _, err := chkGet(t, p, 20); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("третий Get при max = 2 = %v, ожидали DeadlineExceeded", err)
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		p.Put(a, true) // сломано: место освобождается
	}()
	c, err := chkGet(t, p, 1000)
	if err != nil || c == nil || c == a {
		t.Fatalf("после Put(broken) Get = %v, %v; ожидали новое соединение", c, err)
	}
	if n := d.open(); n != 2 {
		t.Fatalf("открыто %d соединений, ожидали 2 (сломанное закрыто)", n)
	}
}

func TestPoolDialErrorFreesSlot(t *testing.T) {
	d := &chkDialer{failNext: true}
	p := NewPool(1, d.dial, d.close)
	if _, err := chkGet(t, p, 1000); err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get при ошибке dial = %v, ожидали ошибку dial", err)
	}
	if c, err := chkGet(t, p, 200); err != nil || c == nil {
		t.Fatalf("после неудачного dial место должно освободиться, а Get = %v, %v", c, err)
	}
}

func TestPoolConcurrent(t *testing.T) {
	d := &chkDialer{}
	p := NewPool(3, d.dial, d.close)
	var peak atomic.Int64
	var wg sync.WaitGroup
	for g := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 20 {
				c, err := chkGet(t, p, 2000)
				if err != nil {
					t.Errorf("Get = %v", err)
					return
				}
				if n := int64(d.open()); n > peak.Load() {
					peak.Store(n)
				}
				time.Sleep(20 * time.Microsecond)
				p.Put(c, (g+i)%7 == 0)
			}
		}()
	}
	wg.Wait()
	if pk := peak.Load(); pk > 3 {
		t.Fatalf("одновременно было открыто %d соединений при max = 3", pk)
	}
}

func TestPoolClose(t *testing.T) {
	d := &chkDialer{}
	p := NewPool(2, d.dial, d.close)
	a, _ := chkGet(t, p, 1000)
	b, _ := chkGet(t, p, 1000)
	p.Put(a, false)
	a2, _ := chkGet(t, p, 1000)
	waiting := make(chan error, 1)
	go func() { _, err := chkGet(t, p, 2000); waiting <- err }()
	time.Sleep(10 * time.Millisecond)
	p.Close()
	p.Close()
	select {
	case err := <-waiting:
		if !errors.Is(err, ErrPoolClosed) {
			t.Fatalf("ждущий Get после Close = %v, ожидали ErrPoolClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ждущий Get не проснулся после Close")
	}
	if _, err := chkGet(t, p, 1000); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("Get после Close = %v, ожидали ErrPoolClosed", err)
	}
	p.Put(a2, false)
	p.Put(b, false)
	if n := d.open(); n != 0 {
		t.Fatalf("после Close и возврата всех соединений открыто %d, ожидали 0", n)
	}
}
