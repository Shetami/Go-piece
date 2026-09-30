package main

func chkWait(t *testing.T, g *Group) (err error, rec any) {
	t.Helper()
	type out struct {
		err error
		rec any
	}
	ch := make(chan out, 1)
	go func() {
		var o out
		defer func() { o.rec = recover(); ch <- o }()
		o.err = g.Wait()
	}()
	select {
	case o := <-ch:
		return o.err, o.rec
	case <-time.After(3 * time.Second):
		t.Fatal("Wait не вернулся за 3 секунды")
		return nil, nil
	}
}

func TestGroupSuccessCancelsAfterWait(t *testing.T) {
	g, ctx := WithContext(context.Background())
	var n atomic.Int32
	for range 3 {
		g.Go(func() error { time.Sleep(time.Millisecond); n.Add(1); return nil })
	}
	err, rec := chkWait(t, g)
	if err != nil || rec != nil || n.Load() != 3 {
		t.Fatalf("Wait = %v, паника %v, выполнено %d; ожидали nil, nil, 3", err, rec, n.Load())
	}
	if ctx.Err() == nil {
		t.Fatal("после Wait контекст группы должен быть отменён")
	}
}

func TestGroupFirstErrorCancels(t *testing.T) {
	g, ctx := WithContext(context.Background())
	first := errors.New("реплика недоступна")
	g.Go(func() error { return first })
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return errors.New("контекст не отменили после ошибки")
		}
	})
	err, rec := chkWait(t, g)
	if rec != nil || err != first {
		t.Fatalf("Wait = %v (паника %v), ожидали именно первую ошибку %q", err, rec, first)
	}
}

func TestGroupPanicMovesToWait(t *testing.T) {
	g, ctx := WithContext(context.Background())
	var finished atomic.Bool
	g.Go(func() error { panic("сломался парсер") })
	g.Go(func() error {
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
			return errors.New("паника должна отменять контекст")
		}
		time.Sleep(10 * time.Millisecond)
		finished.Store(true)
		return ctx.Err()
	})
	_, rec := chkWait(t, g)
	if rec != "сломался парсер" {
		t.Fatalf("Wait должен запаниковать тем же значением, recover() = %v", rec)
	}
	if !finished.Load() {
		t.Fatal("Wait запаниковал, не дождавшись остальных горутин")
	}
}

func TestGroupPanicBeatsError(t *testing.T) {
	g, _ := WithContext(context.Background())
	boom := errors.New("boom")
	g.Go(func() error { return boom })
	g.Go(func() error { time.Sleep(5 * time.Millisecond); panic(boom) })
	err, rec := chkWait(t, g)
	if rec != boom {
		t.Fatalf("при панике Wait должен паниковать, а не возвращать ошибку: err=%v, recover()=%v", err, rec)
	}
}

func TestGroupRuntimePanic(t *testing.T) {
	g, _ := WithContext(context.Background())
	g.Go(func() error { var p *struct{ x int }; return errors.New(strconv.Itoa(p.x)) })
	_, rec := chkWait(t, g)
	if _, ok := rec.(runtime.Error); !ok {
		t.Fatalf("паника рантайма должна дойти до Wait как runtime.Error, получили %T %v", rec, rec)
	}
}
