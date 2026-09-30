package main

func TestMailboxKeepsLatest(t *testing.T) {
	m := NewMailbox[int](3)
	if m.C() == nil || m.C() != m.C() {
		t.Fatal("C() должен возвращать один и тот же не-nil канал")
	}
	var drops []bool
	for i := 1; i <= 5; i++ {
		drops = append(drops, m.Push(i))
	}
	if !reflect.DeepEqual(drops, []bool{false, false, false, true, true}) {
		t.Fatalf("Push вернул %v, ожидали [false false false true true]", drops)
	}
	var got []int
	for range 3 {
		select {
		case v := <-m.C():
			got = append(got, v)
		case <-time.After(time.Second):
			t.Fatalf("в буфере меньше трёх значений: %v", got)
		}
	}
	if !reflect.DeepEqual(got, []int{3, 4, 5}) {
		t.Fatalf("прочитали %v, ожидали три последних [3 4 5]", got)
	}
	if d := m.Dropped(); d != 2 {
		t.Fatalf("Dropped = %d, ожидали 2", d)
	}
}

func TestMailboxNeverBlocks(t *testing.T) {
	m := NewMailbox[int](1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		for g := range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range 500 {
					m.Push(g*1000 + i)
				}
			}()
		}
		wg.Wait()
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("восемь писателей без читателя заблокировались в Push")
	}
	if d := m.Dropped(); d != 8*500-1 {
		t.Fatalf("Dropped = %d, ожидали %d: в буфере на 1 остаётся одно значение", d, 8*500-1)
	}
}

func TestMailboxWithReader(t *testing.T) {
	m := NewMailbox[int](2)
	stop := make(chan struct{})
	var read atomic.Int64
	go func() {
		for {
			select {
			case <-m.C():
				read.Add(1)
			case <-stop:
				return
			}
		}
	}()
	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		for range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range 2000 {
					m.Push(i)
				}
			}()
		}
		wg.Wait()
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		close(stop)
		t.Fatal("Push заблокировался, пока параллельно работал читатель")
	}
	close(stop)
	time.Sleep(10 * time.Millisecond)
	total := read.Load() + int64(m.Dropped()) + int64(len(m.C()))
	if total != 8000 {
		t.Fatalf("прочитано + выброшено + в буфере = %d, ожидали 8000: значения теряются или дублируются", total)
	}
}
