package main

type chkRec struct {
	mu      sync.Mutex
	batches [][]int
}

func (r *chkRec) flush(b []int) {
	r.mu.Lock()
	r.batches = append(r.batches, b)
	r.mu.Unlock()
}

func (r *chkRec) get() [][]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([][]int, len(r.batches))
	for i, b := range r.batches {
		out[i] = slices.Clone(b)
	}
	return out
}

func TestBatcherBySizeAndClose(t *testing.T) {
	var r chkRec
	b := NewBatcher(3, time.Hour, r.flush)
	for i := 1; i <= 7; i++ {
		b.Add(i)
	}
	b.Close()
	want := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	if got := r.get(); !reflect.DeepEqual(got, want) {
		t.Fatalf("пачки %v, ожидали %v (Close отправляет остаток и ждёт flush)", got, want)
	}
}

func TestBatcherByTimer(t *testing.T) {
	var r chkRec
	b := NewBatcher(100, 50*time.Millisecond, r.flush)
	defer b.Close()
	b.Add(1)
	b.Add(2)
	if got := r.get(); len(got) != 0 {
		t.Fatalf("пачка %v ушла сразу, хотя не набралась и время не вышло", got)
	}
	for i := 0; i < 100 && len(r.get()) == 0; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if got := r.get(); !reflect.DeepEqual(got, [][]int{{1, 2}}) {
		t.Fatalf("через maxWait ожидали пачку [[1 2]], получили %v", got)
	}
}

func TestBatcherTimerStartsAtFirstItem(t *testing.T) {
	var r chkRec
	b := NewBatcher(2, 200*time.Millisecond, r.flush)
	defer b.Close()
	b.Add(1)
	b.Add(2) // пачка ушла по размеру
	time.Sleep(150 * time.Millisecond)
	b.Add(3) // новая пачка: её отсчёт начинается сейчас
	time.Sleep(100 * time.Millisecond)
	if got := r.get(); len(got) != 1 {
		t.Fatalf("элемент 3 ушёл через 100 мс после добавления при maxWait = 200 мс: %v — таймер прошлой пачки не остановлен?", got)
	}
	time.Sleep(400 * time.Millisecond)
	if got := r.get(); !reflect.DeepEqual(got, [][]int{{1, 2}, {3}}) {
		t.Fatalf("пачки %v, ожидали [[1 2] [3]]", got)
	}
}

func TestBatcherDoesNotReuseSlice(t *testing.T) {
	var mu sync.Mutex
	var kept [][]int // храним слайсы как есть, без копирования
	b := NewBatcher(2, time.Hour, func(s []int) { mu.Lock(); kept = append(kept, s); mu.Unlock() })
	for i := 1; i <= 6; i++ {
		b.Add(i)
	}
	b.Close()
	mu.Lock()
	defer mu.Unlock()
	if want := [][]int{{1, 2}, {3, 4}, {5, 6}}; !reflect.DeepEqual(kept, want) {
		t.Fatalf("сохранённые пачки %v, ожидали %v — Batcher переписал слайс, отданный в flush", kept, want)
	}
}

func TestBatcherConcurrentAddClose(t *testing.T) {
	base := runtime.NumGoroutine()
	var r chkRec
	b := NewBatcher(7, 5*time.Millisecond, r.flush)
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for g := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 50 {
				if b.Add(g*100+i) == nil {
					accepted.Add(1)
				}
			}
		}()
	}
	time.Sleep(time.Millisecond)
	b.Close()
	wg.Wait()
	b.Close() // повторный Close не должен зависать или паниковать
	if err := b.Add(1); !errors.Is(err, ErrClosed) {
		t.Fatalf("Add после Close = %v, ожидали ErrClosed", err)
	}
	total := 0
	for _, batch := range r.get() {
		if len(batch) == 0 || len(batch) > 7 {
			t.Fatalf("пачка размера %d, ожидали от 1 до 7", len(batch))
		}
		total += len(batch)
	}
	if total != int(accepted.Load()) {
		t.Fatalf("принято %d элементов, а во flush попало %d", accepted.Load(), total)
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > base {
		t.Fatalf("после Close осталось %d лишних горутин", n-base)
	}
}
