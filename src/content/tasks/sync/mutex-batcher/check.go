package main

func TestBatcherRetainedBatches(t *testing.T) {
	var got [][]int
	b := NewBatcher(3, func(xs []int) { got = append(got, xs) })
	for i := 1; i <= 7; i++ {
		b.Add(i)
	}
	b.Flush()
	b.Flush()
	want := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("сохранённые пачки = %v, ожидали %v — flush сохраняет слайс, а его массив переиспользовали?", got, want)
	}
}

func TestBatcherConcurrentExactlyOnce(t *testing.T) {
	var mu sync.Mutex
	seen := map[int]int{}
	b := NewBatcher(10, func(xs []int) {
		if len(xs) == 0 || len(xs) > 10 {
			t.Errorf("flush получил пачку из %d элементов", len(xs))
		}
		mu.Lock()
		for _, x := range xs {
			seen[x]++
		}
		mu.Unlock()
	})
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 500 {
				b.Add(g*1000 + i)
			}
		}()
	}
	wg.Wait()
	b.Flush()
	if len(seen) != 4000 {
		t.Fatalf("во flush дошло %d разных элементов из 4000", len(seen))
	}
	for x, n := range seen {
		if n != 1 {
			t.Fatalf("элемент %d отправлен %d раз", x, n)
		}
	}
}

func TestBatcherFlushOutsideLock(t *testing.T) {
	release := make(chan struct{})
	first := true
	b := NewBatcher(2, func(xs []int) {
		if first {
			first = false
			<-release // медленная запись первой пачки
		}
	})
	go func() { b.Add(1); b.Add(2) }()
	time.Sleep(20 * time.Millisecond)
	done := make(chan struct{})
	go func() { b.Add(3); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("Add ждёт, пока выполняется flush другой пачки — flush вызывается под мьютексом?")
	}
	close(release)
}
