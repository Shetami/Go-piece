package main

func TestMergeCollectsEverything(t *testing.T) {
	a := make(chan int, 3)
	b := make(chan int, 2)
	for _, v := range []int{1, 2, 3} {
		a <- v
	}
	for _, v := range []int{10, 20} {
		b <- v
	}
	close(a)
	close(b)

	var got []int
	for v := range Merge(a, b) {
		got = append(got, v)
	}
	sort.Ints(got)

	want := []int{1, 2, 3, 10, 20}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Merge собрал %v, а должен был %v", got, want)
	}
}

func TestMergeClosesOutput(t *testing.T) {
	a := make(chan int)
	close(a)

	out := Merge(a)
	select {
	case v, ok := <-out:
		if ok {
			t.Fatalf("из пустого входа пришло значение %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("выходной канал не закрылся после закрытия входного")
	}
}

func TestMergeNoInputs(t *testing.T) {
	out := Merge()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("Merge без входов выдал значение")
		}
	case <-time.After(time.Second):
		t.Fatal("Merge без входов не закрыл выходной канал")
	}
}

func TestMergeStreamsWithoutBuffering(t *testing.T) {
	// Небуферизованный вход: если Merge сначала собирает всё в слайс,
	// он не отдаст первое значение, пока вход не закроется.
	a := make(chan int)
	go func() {
		a <- 1
		a <- 2
		close(a)
	}()

	out := Merge(a)
	select {
	case v := <-out:
		if v != 1 {
			t.Fatalf("первым пришло %d, ожидали 1", v)
		}
	case <-time.After(time.Second):
		t.Fatal("Merge не отдал первое значение — похоже, он копит их у себя")
	}
}

func TestMergeReadsInputsConcurrently(t *testing.T) {
	// Главный тест. Вход a не закроется, пока из b не придёт значение, —
	// значит, читать входы по очереди нельзя: на первом же и встанешь.
	a := make(chan int)
	b := make(chan int)
	release := make(chan struct{})

	go func() {
		a <- 1
		<-release
		close(a)
	}()
	go func() {
		b <- 2
		close(b)
	}()

	done := make(chan []int, 1)
	go func() {
		var got []int
		for v := range Merge(a, b) {
			got = append(got, v)
			if len(got) == 2 {
				close(release)
			}
		}
		done <- got
	}()

	select {
	case got := <-done:
		sort.Ints(got)
		if want := []int{1, 2}; !reflect.DeepEqual(got, want) {
			t.Fatalf("Merge собрал %v, а должен был %v", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Merge читает входы по очереди, а не одновременно: на неисчерпанном канале он встаёт намертво")
	}
}
