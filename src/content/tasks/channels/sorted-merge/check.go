package main

type chkEv struct {
	At  int
	Src string
}

func chkFrom[T any](vals ...T) <-chan T {
	ch := make(chan T)
	go func() {
		for _, v := range vals {
			ch <- v
			time.Sleep(time.Duration(len(vals)) * 50 * time.Microsecond)
		}
		close(ch)
	}()
	return ch
}

func chkCollect[T any](t *testing.T, out <-chan T) []T {
	t.Helper()
	if out == nil {
		t.Fatal("MergeSorted вернул nil-канал")
	}
	var got []T
	for {
		select {
		case v, ok := <-out:
			if !ok {
				return got
			}
			got = append(got, v)
		case <-time.After(2 * time.Second):
			t.Fatalf("выход не закрылся, получено %v", got)
		}
	}
}

func TestMergeSortedInts(t *testing.T) {
	got := chkCollect(t, MergeSorted(context.Background(), cmp.Compare[int],
		chkFrom(0, 4, 9), chkFrom[int](), chkFrom(0, 1, 2, 3, 10, 11), chkFrom(5)))
	want := []int{0, 0, 1, 2, 3, 4, 5, 9, 10, 11}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %v, ожидали %v (нули — тоже данные, пустой канал не мешает)", got, want)
	}
}

func TestMergeSortedStableTies(t *testing.T) {
	byAt := func(a, b chkEv) int { return cmp.Compare(a.At, b.At) }
	got := chkCollect(t, MergeSorted(context.Background(), byAt,
		chkFrom(chkEv{1, "a"}, chkEv{3, "a"}),
		chkFrom(chkEv{1, "b"}, chkEv{2, "b"}, chkEv{3, "b"}),
		chkFrom(chkEv{1, "c"})))
	want := []chkEv{{1, "a"}, {1, "b"}, {1, "c"}, {2, "b"}, {3, "a"}, {3, "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %v,\nожидали %v (при равенстве — меньший индекс входа первым)", got, want)
	}
}

func TestMergeSortedWaitsSlowInput(t *testing.T) {
	slow := make(chan int)
	go func() {
		time.Sleep(30 * time.Millisecond)
		slow <- 1
		close(slow)
	}()
	got := chkCollect(t, MergeSorted(context.Background(), cmp.Compare[int], chkFrom(2, 3), slow))
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("получили %v, ожидали [1 2 3]: нельзя отдавать значение, пока не известна голова медленного входа", got)
	}
}

func TestMergeSortedEmptyAndCancel(t *testing.T) {
	if got := chkCollect(t, MergeSorted[int](context.Background(), cmp.Compare[int])); len(got) != 0 {
		t.Fatalf("без входов получили %v", got)
	}
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	stuck := make(chan int) // никогда не закроется
	out := MergeSorted(ctx, cmp.Compare[int], chkFrom(1, 2), stuck)
	time.Sleep(10 * time.Millisecond)
	cancel()
	chkCollect(t, out)
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before+1 { // +1: генератор chkFrom, которого никто не дочитал
		t.Fatalf("после отмены остались горутины MergeSorted: %d лишних", n-before)
	}
}
