package main

func TestShardedBasic(t *testing.T) {
	for _, n := range []int{0, 1, 16} {
		s := NewShardedMap[int](n)
		for i := range 100 {
			s.Set(strconv.Itoa(i), i)
		}
		s.Delete("5")
		s.Delete("нет")
		if v, ok := s.Get("42"); !ok || v != 42 {
			t.Fatalf("n=%d: Get(42) = %d, %v; ожидали 42, true", n, v, ok)
		}
		if _, ok := s.Get("5"); ok || s.Len() != 99 {
			t.Fatalf("n=%d: после Delete(5): ok=%v Len=%d; ожидали false, 99", n, ok, s.Len())
		}
	}
}

func TestShardedUpdateAtomic(t *testing.T) {
	s := NewShardedMap[int](4)
	var wg sync.WaitGroup
	for g := range 40 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				s.Update(strconv.Itoa(g%3), func(v int, ok bool) (int, bool) {
					runtime.Gosched() // между чтением и записью — повод переключиться
					return v + 1, true
				})
			}
		}()
	}
	wg.Wait()
	total := 0
	for _, k := range []string{"0", "1", "2"} {
		v, _ := s.Get(k)
		total += v
	}
	if total != 2000 {
		t.Fatalf("сумма после 2000 Update(+1) = %d — обновления теряются, Update не атомарен", total)
	}
	s.Update("0", func(int, bool) (int, bool) { return 0, false })
	if _, ok := s.Get("0"); ok {
		t.Fatal("Update с keep=false должен удалить ключ")
	}
	s.Update("new", func(v int, ok bool) (int, bool) {
		if ok {
			t.Error("для отсутствующего ключа fn получила ok=true")
		}
		return 7, true
	})
	if v, _ := s.Get("new"); v != 7 {
		t.Fatalf("Update отсутствующего ключа: Get = %d, ожидали 7", v)
	}
}

func TestShardedRangeCanWrite(t *testing.T) {
	s := NewShardedMap[int](2)
	for i := range 20 {
		s.Set(strconv.Itoa(i), i)
	}
	done := make(chan int)
	go func() {
		n := 0
		s.Range(func(k string, v int) bool {
			n++
			s.Set(k, v*10)
			s.Delete("x" + k)
			return true
		})
		done <- n
	}()
	select {
	case n := <-done:
		if n != 20 {
			t.Fatalf("Range прошёл %d пар, ожидали 20", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Range с вызовом Set внутри fn завис — блокировка шарда держится во время fn")
	}
	if v, _ := s.Get("7"); v != 70 {
		t.Fatalf("Get(7) = %d после Range, ожидали 70", v)
	}
}

func TestShardedRangeStop(t *testing.T) {
	s := NewShardedMap[int](8)
	for i := range 50 {
		s.Set(strconv.Itoa(i), i)
	}
	n := 0
	s.Range(func(string, int) bool { n++; return n < 3 })
	if n != 3 {
		t.Fatalf("fn вернула false на третьем вызове, а вызвана %d раз", n)
	}
}
