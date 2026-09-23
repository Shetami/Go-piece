package main

func checkNext(t *testing.T, ch <-chan []int, within time.Duration) ([]int, bool) {
	t.Helper()
	if ch == nil {
		t.Fatal("Batch вернул nil-канал")
	}
	select {
	case b, ok := <-ch:
		return b, ok
	case <-time.After(within):
		t.Fatalf("за %v не пришло ни пачки, ни закрытия", within)
	}
	return nil, false
}

func TestBatchBySize(t *testing.T) {
	in := make(chan int)
	out := Batch(in, 3, time.Hour)
	go func() {
		for i := 1; i <= 7; i++ {
			in <- i
		}
		close(in)
	}()
	var got [][]int
	for {
		b, ok := checkNext(t, out, 2*time.Second)
		if !ok {
			break
		}
		got = append(got, b)
	}
	want := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("пачки %v, ожидали %v (хвост — при закрытии входа)", got, want)
	}
}

func TestBatchByTime(t *testing.T) {
	in := make(chan int)
	out := Batch(in, 100, 50*time.Millisecond)
	in <- 1
	in <- 2
	b, ok := checkNext(t, out, time.Second)
	if !ok || !reflect.DeepEqual(b, []int{1, 2}) {
		t.Fatalf("по таймауту ожидали [1 2], получили %v (ok=%v)", b, ok)
	}
	in <- 3
	b, _ = checkNext(t, out, time.Second)
	if !reflect.DeepEqual(b, []int{3}) {
		t.Fatalf("следующая пачка по таймауту: %v, ожидали [3]", b)
	}
	close(in)
	if _, ok := checkNext(t, out, time.Second); ok {
		t.Fatal("после закрытия входа без данных ожидали закрытие выхода, а не пустую пачку")
	}
}

func TestBatchNoEmptyBatches(t *testing.T) {
	in := make(chan int)
	out := Batch(in, 5, 10*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	close(in)
	if b, ok := checkNext(t, out, time.Second); ok {
		t.Fatalf("пришла пачка %v, хотя данных не было", b)
	}
}
