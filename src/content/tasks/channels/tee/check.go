package main

func checkCollect(ch <-chan int, dst *[]int, wg *sync.WaitGroup) {
	defer wg.Done()
	for v := range ch {
		*dst = append(*dst, v)
	}
}

func TestTeeBothGetAll(t *testing.T) {
	in := make(chan int)
	a, b := Tee(in)
	if a == nil || b == nil {
		t.Fatal("Tee вернул nil-канал")
	}
	var got1, got2 []int
	var wg sync.WaitGroup
	wg.Add(2)
	go checkCollect(a, &got1, &wg)
	go checkCollect(b, &got2, &wg)

	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("выходы не закрылись после закрытия входа")
	}
	want := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(got1, want) || !reflect.DeepEqual(got2, want) {
		t.Fatalf("первый получил %v, второй %v, ожидали оба %v", got1, got2, want)
	}
}

func TestTeeSlowReaderOnSecond(t *testing.T) {
	// Второй выход сначала читают, первый — нет. Tee не должен
	// требовать, чтобы первым всегда читали первый выход.
	in := make(chan int, 1)
	in <- 42
	close(in)
	a, b := Tee(in)
	select {
	case v := <-b:
		if v != 42 {
			t.Fatalf("второй выход получил %d", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("второй выход не получил значение, пока первый никто не читал")
	}
	select {
	case v := <-a:
		if v != 42 {
			t.Fatalf("первый выход получил %d", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("первый выход не получил значение")
	}
}
