package main

func chkOpenAndDrop(t *Tracker, names ...string) {
	for _, n := range names {
		t.Open(n)
	}
}

func chkOpenAndClose(tr *Tracker, t *testing.T, names ...string) {
	for _, n := range names {
		r := tr.Open(n)
		if err := r.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
		if err := r.Close(); !errors.Is(err, ErrClosed) {
			t.Fatalf("повторный Close вернул %v, ожидали ErrClosed", err)
		}
	}
}

func chkWaitLeaked(tr *Tracker, want int) []string {
	for i := 0; i < 100; i++ {
		runtime.GC()
		if l := tr.Leaked(); len(l) >= want {
			return l
		}
		time.Sleep(time.Millisecond)
	}
	return tr.Leaked()
}

func TestTrackerLeaks(t *testing.T) {
	tr := &Tracker{}
	chkOpenAndDrop(tr, "db-3", "db-1", "cache", "db-2", "queue")
	chkOpenAndClose(tr, t, "closed-1", "closed-2", "closed-3")
	if n := tr.OpenCount(); n != 5 {
		t.Fatalf("OpenCount = %d, ожидали 5 (8 открыто, 3 закрыто)", n)
	}
	got := chkWaitLeaked(tr, 5)
	want := []string{"cache", "db-1", "db-2", "db-3", "queue"}
	if !slices.Equal(got, want) {
		t.Fatalf("Leaked = %v, ожидали %v — ресурсы без Close должны обнаруживаться после GC", got, want)
	}
	for i := 0; i < 5; i++ {
		runtime.GC()
		time.Sleep(time.Millisecond)
	}
	if got := tr.Leaked(); len(got) != 5 {
		t.Fatalf("после нескольких GC Leaked = %v — закрытые ресурсы не утечка", got)
	}
	if n := tr.OpenCount(); n != 0 {
		t.Fatalf("OpenCount после сбора утёкших = %d, ожидали 0", n)
	}
}

func TestTrackerAliveNotLeaked(t *testing.T) {
	tr := &Tracker{}
	keep := tr.Open("alive")
	chkOpenAndDrop(tr, "lost")
	got := chkWaitLeaked(tr, 1)
	if !slices.Equal(got, []string{"lost"}) {
		t.Fatalf("Leaked = %v, ожидали [lost]", got)
	}
	if tr.OpenCount() != 1 {
		t.Fatalf("OpenCount = %d, ожидали 1", tr.OpenCount())
	}
	runtime.KeepAlive(keep)
}

func TestTrackerConcurrent(t *testing.T) {
	tr := &Tracker{}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				r := tr.Open(fmt.Sprintf("g%d-%d", g, i))
				if i%2 == 0 {
					r.Close()
				}
				_ = tr.Leaked()
			}
		}(g)
	}
	wg.Wait()
	if got := chkWaitLeaked(tr, 200); len(got) != 200 {
		t.Fatalf("утекло %d ресурсов, ожидали 200 (половина из 400 не закрыта)", len(got))
	}
}
