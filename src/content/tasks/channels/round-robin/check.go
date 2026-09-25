package main

func source(n int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := range n {
			ch <- i
		}
	}()
	return ch
}

func TestSplitRoundRobin(t *testing.T) {
	outs := Split(source(10), 3)
	if len(outs) != 3 {
		t.Fatalf("выходов %d, ожидали 3", len(outs))
	}
	got := make([][]int, 3)
	var wg sync.WaitGroup
	for i, ch := range outs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch {
				got[i] = append(got[i], v)
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("выходы не закрылись после закрытия входа")
	}
	want := [][]int{{0, 3, 6, 9}, {1, 4, 7}, {2, 5, 8}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %v, ожидали %v", got, want)
	}
}

func TestSplitOne(t *testing.T) {
	outs := Split(source(3), 1)
	var got []int
	for v := range outs[0] {
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("n=1: %v", got)
	}
}

func TestSplitEmpty(t *testing.T) {
	for _, ch := range Split(source(0), 4) {
		select {
		case _, ok := <-ch:
			if ok {
				t.Fatalf("из пустого входа пришло значение")
			}
		case <-time.After(time.Second):
			t.Fatalf("выход не закрылся")
		}
	}
}

func TestSplitPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("Split(in, 0) должен паниковать")
		}
	}()
	Split(source(0), 0)
}
