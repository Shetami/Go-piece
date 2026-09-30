package main

func chkClock() func() time.Time {
	var mu sync.Mutex
	cur := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		cur = cur.Add(time.Second)
		return cur
	}
}

var chkErrNoRows = errors.New("no rows")

func chkLoadUser(ctx context.Context, tr *Tracer, fail bool) (err error) {
	ctx, end := tr.Start(ctx, "load-user")
	defer end(&err)
	if fail {
		return fmt.Errorf("select: %w", chkErrNoRows)
	}
	return chkQuery(ctx, tr)
}

func chkQuery(ctx context.Context, tr *Tracer) (err error) {
	_, end := tr.Start(ctx, "query")
	defer end(&err)
	return nil
}

func TestSpanOKAndNesting(t *testing.T) {
	tr := &Tracer{Now: chkClock()}
	if err := chkLoadUser(context.Background(), tr, false); err != nil {
		t.Fatalf("chkLoadUser = %v", err)
	}
	sp := tr.Spans()
	if len(sp) != 2 {
		t.Fatalf("завершено %d спанов, ожидали 2", len(sp))
	}
	q, u := sp[0], sp[1]
	if q.Name != "query" || q.Parent != "load-user" || u.Name != "load-user" || u.Parent != "" {
		t.Fatalf("спаны %+v и %+v: ожидали query с родителем load-user, затем корневой load-user", q, u)
	}
	if q.Status != "ok" || u.Status != "ok" {
		t.Fatalf("статусы %q, %q, ожидали ok", q.Status, u.Status)
	}
	if !u.Start.Before(q.Start) || !q.End.Before(u.End) {
		t.Fatalf("load-user [%v, %v] должен охватывать query [%v, %v]: Start берётся в Start, End — в end",
			u.Start, u.End, q.Start, q.End)
	}
}

func TestSpanError(t *testing.T) {
	tr := &Tracer{Now: chkClock()}
	chkLoadUser(context.Background(), tr, true)
	sp := tr.Spans()
	if len(sp) != 1 || sp[0].Status != "error" || !errors.Is(sp[0].Err, chkErrNoRows) {
		t.Fatalf("спаны %+v, ожидали один со статусом error и ошибкой функции", sp)
	}
}

func TestSpanPanic(t *testing.T) {
	tr := &Tracer{Now: chkClock()}
	func() {
		defer func() {
			if r := recover(); r != "кэш сломан" {
				t.Fatalf("паника должна пролететь дальше, recover() = %v", r)
			}
		}()
		_, end := tr.Start(context.Background(), "render")
		defer end(nil)
		panic("кэш сломан")
	}()
	sp := tr.Spans()
	if len(sp) != 1 || sp[0].Status != "panic" || sp[0].Err == nil || !strings.Contains(sp[0].Err.Error(), "кэш сломан") {
		t.Fatalf("спаны %+v, ожидали один со статусом panic и текстом паники", sp)
	}
	if sp[0].End.IsZero() {
		t.Fatal("у спана с паникой должно быть время конца")
	}
}

func TestSpanConcurrent(t *testing.T) {
	tr := &Tracer{Now: chkClock()}
	root, end := tr.Start(context.Background(), "batch")
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chkQuery(root, tr)
		}()
	}
	wg.Wait()
	end(nil)
	sp := tr.Spans()
	if len(sp) != 21 || sp[20].Name != "batch" {
		t.Fatalf("спанов %d, ожидали 21, последним — batch", len(sp))
	}
	for _, s := range sp[:20] {
		if s.Parent != "batch" {
			t.Fatalf("у query родитель %q, ожидали batch", s.Parent)
		}
	}
}
