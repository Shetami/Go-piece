package main

func chkAcquireAsync(s *Weighted, ctx context.Context, n int64) chan error {
	ch := make(chan error, 1)
	go func() { ch <- s.Acquire(ctx, n) }()
	time.Sleep(20 * time.Millisecond) // встаёт в очередь в порядке запуска
	return ch
}

func chkPending(ch chan error) bool {
	select {
	case <-ch:
		return false
	default:
		return true
	}
}

func TestWeightedBasic(t *testing.T) {
	s := NewWeighted(10)
	ctx := context.Background()
	if err := s.Acquire(ctx, 7); err != nil {
		t.Fatalf("Acquire(7) = %v", err)
	}
	if s.TryAcquire(4) {
		t.Fatal("TryAcquire(4) при 3 свободных должен вернуть false")
	}
	if !s.TryAcquire(3) {
		t.Fatal("TryAcquire(3) при 3 свободных должен вернуть true")
	}
	s.Release(10)
	if err := s.Acquire(ctx, 11); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Acquire(11) при size=10 = %v, ожидали ErrTooLarge сразу", err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("Release больше занятого должен паниковать")
			}
		}()
		s.Release(1)
	}()
}

func TestWeightedFIFONoStarvation(t *testing.T) {
	s := NewWeighted(10)
	ctx := context.Background()
	s.Acquire(ctx, 8)
	big := chkAcquireAsync(s, ctx, 5)   // ждёт: свободно 2
	small := chkAcquireAsync(s, ctx, 1) // хватило бы, но стоит за big
	if !chkPending(small) {
		t.Fatal("Acquire(1) обогнал ждущий Acquire(5) — большой запрос будет голодать")
	}
	if s.TryAcquire(1) {
		t.Fatal("TryAcquire прошёл без очереди, хотя в очереди есть ждущие")
	}
	s.Release(8)
	for name, ch := range map[string]chan error{"Acquire(5)": big, "Acquire(1)": small} {
		select {
		case err := <-ch:
			if err != nil {
				t.Fatalf("%s = %v", name, err)
			}
		case <-time.After(time.Second):
			t.Fatalf("после Release %s так и не получил единицы", name)
		}
	}
}

func TestWeightedCancelHeadWakesNext(t *testing.T) {
	s := NewWeighted(10)
	s.Acquire(context.Background(), 8)
	ctx, cancel := context.WithCancel(context.Background())
	big := chkAcquireAsync(s, ctx, 5)
	small := chkAcquireAsync(s, context.Background(), 2)
	cancel()
	if err := <-big; !errors.Is(err, context.Canceled) {
		t.Fatalf("отменённый Acquire вернул %v, ожидали context.Canceled", err)
	}
	select {
	case err := <-small:
		if err != nil {
			t.Fatalf("Acquire(2) = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("голова очереди ушла по отмене, а следующий, которому хватает 2 свободных, не проснулся")
	}
	s.Release(10)
	if !s.TryAcquire(10) {
		t.Fatal("после всех Release семафор должен быть полностью свободен — отменённый запрос что-то занял")
	}
}

func TestWeightedStress(t *testing.T) {
	const size = 10
	s := NewWeighted(size)
	var held, peak atomic.Int64
	var wg sync.WaitGroup
	for i := range 60 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := int64(i%4 + 1)
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(i%3+1)*15*time.Millisecond)
			defer cancel()
			if s.Acquire(ctx, n) != nil {
				return
			}
			h := held.Add(n)
			for {
				p := peak.Load()
				if h <= p || peak.CompareAndSwap(p, h) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			held.Add(-n)
			s.Release(n)
		}()
	}
	wg.Wait()
	if p := peak.Load(); p > size {
		t.Fatalf("одновременно занято %d единиц при size=%d", p, size)
	}
	if !s.TryAcquire(size) {
		t.Fatal("после стресс-теста не все единицы вернулись — где-то утечка при отмене")
	}
}
