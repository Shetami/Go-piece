package main

type chkItem struct {
	id  int
	pad *[64]byte
}

func TestQueueFIFO(t *testing.T) {
	var q Queue[int]
	if _, ok := q.Pop(); ok {
		t.Fatalf("Pop из пустой очереди вернул ok=true")
	}
	next, want := 0, 0
	for round := 0; round < 200; round++ {
		for i := 0; i < round%7+1; i++ {
			q.Push(next)
			next++
		}
		for i := 0; i < round%5+1 && q.Len() > 0; i++ {
			v, ok := q.Pop()
			if !ok || v != want {
				t.Fatalf("Pop = %d, %v; ожидали %d, true", v, ok, want)
			}
			want++
		}
	}
	if q.Len() != next-want {
		t.Fatalf("Len = %d, ожидали %d", q.Len(), next-want)
	}
}

func TestQueueShrinks(t *testing.T) {
	var q Queue[int]
	for i := 0; i < 100000; i++ {
		q.Push(i)
	}
	for q.Len() > 0 {
		q.Pop()
		if c, n := q.Cap(), q.Len(); c > max(64, 4*n) {
			t.Fatalf("после Pop: Len=%d, Cap=%d — очередь держит лишнюю память (ожидали Cap <= max(64, 4*Len))", n, c)
		}
	}
	for i := 0; i < 100000; i++ {
		q.Push(i)
		q.Push(i)
		q.Pop()
		if i%1000 == 0 && q.Cap() > 4*q.Len()+64 {
			t.Fatalf("очередь с Len=%d держит Cap=%d", q.Len(), q.Cap())
		}
	}
}

func TestQueueSteadyState(t *testing.T) {
	var q Queue[int]
	q.Push(-1)
	for i := 0; i < 100000; i++ {
		q.Push(i)
		q.Pop()
	}
	if q.Len() != 1 || q.Cap() > 64 {
		t.Fatalf("после 100000 пар Push/Pop с одним элементом: Len=%d, Cap=%d, ожидали Len=1, Cap<=64", q.Len(), q.Cap())
	}
}

func TestQueueReleasesPopped(t *testing.T) {
	var q Queue[*chkItem]
	var freed atomic.Int32
	for i := 0; i < 10; i++ {
		it := &chkItem{id: i, pad: new([64]byte)}
		runtime.SetFinalizer(it, func(*chkItem) { freed.Add(1) })
		q.Push(it)
	}
	for i := 0; i < 9; i++ {
		q.Pop()
	}
	for i := 0; i < 50 && freed.Load() < 9; i++ {
		runtime.GC()
		time.Sleep(time.Millisecond)
	}
	if got := freed.Load(); got < 9 {
		t.Fatalf("после Pop девяти элементов и GC собрано %d из 9 — очередь удерживает извлечённые объекты", got)
	}
	if v, ok := q.Pop(); !ok || v.id != 9 {
		t.Fatalf("последний элемент потерян")
	}
}
