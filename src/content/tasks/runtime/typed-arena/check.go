package main

type chkNode struct {
	id   int
	next *chkNode
	blob *chkBlob
}

type chkBlob struct{ data [128]byte }

func TestArenaDistinct(t *testing.T) {
	a := NewArena[chkNode](64)
	seen := map[*chkNode]bool{}
	var ptrs []*chkNode
	for i := 0; i < 1000; i++ {
		p := a.New()
		if p == nil || seen[p] {
			t.Fatalf("New #%d вернул nil или уже выданный указатель", i)
		}
		if p.id != 0 || p.next != nil {
			t.Fatalf("New #%d вернул необнулённый объект", i)
		}
		seen[p] = true
		p.id = i
		ptrs = append(ptrs, p)
	}
	for i, p := range ptrs {
		if p.id != i {
			t.Fatalf("объект #%d перезаписан: id=%d", i, p.id)
		}
	}
	if a.Len() != 1000 {
		t.Fatalf("Len = %d, ожидали 1000", a.Len())
	}
}

func TestArenaResetZeroes(t *testing.T) {
	a := NewArena[chkNode](7)
	for i := 0; i < 100; i++ {
		p := a.New()
		p.id = i + 1
		p.next = p
	}
	a.Reset()
	if a.Len() != 0 {
		t.Fatalf("Len после Reset = %d", a.Len())
	}
	for i := 0; i < 150; i++ {
		if p := a.New(); p.id != 0 || p.next != nil {
			t.Fatalf("New #%d после Reset вернул объект со старыми данными: id=%d", i, p.id)
		}
	}
}

func TestArenaNoAllocsAfterReset(t *testing.T) {
	a := NewArena[chkNode](64)
	allocs := testing.AllocsPerRun(20, func() {
		a.Reset()
		for i := 0; i < 1000; i++ {
			a.New().id = i
		}
	})
	if allocs != 0 {
		t.Fatalf("Reset + 1000×New выделяют память %.0f раз на круг, ожидали 0 — после Reset блоки надо переиспользовать", allocs)
	}
}

func TestArenaResetReleases(t *testing.T) {
	a := NewArena[chkNode](16)
	var freed atomic.Int32
	for i := 0; i < 40; i++ {
		b := &chkBlob{}
		runtime.SetFinalizer(b, func(*chkBlob) { freed.Add(1) })
		a.New().blob = b
	}
	a.Reset()
	for i := 0; i < 50 && freed.Load() < 40; i++ {
		runtime.GC()
		time.Sleep(time.Millisecond)
	}
	if got := freed.Load(); got < 40 {
		t.Fatalf("после Reset и GC собрано %d из 40 объектов, на которые ссылались элементы арены — Reset должен их отпустить", got)
	}
	runtime.KeepAlive(a)
}
