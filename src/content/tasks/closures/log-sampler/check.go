package main

func chkT(ms int) time.Time { return time.Unix(1_700_000_000, 0).Add(time.Duration(ms) * time.Millisecond) }

func chkPassed(allow func(string, time.Time) bool, key string, n int, at time.Time) []int {
	var out []int
	for i := 1; i <= n; i++ {
		if allow(key, at) {
			out = append(out, i)
		}
	}
	return out
}

func TestSamplerFirstThereafter(t *testing.T) {
	allow := NewSampler(2, 3, time.Second)
	got := chkPassed(allow, "db timeout", 11, chkT(0))
	if want := []int{1, 2, 5, 8, 11}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first=2, thereafter=3: пропущены сообщения %v, ожидали %v", got, want)
	}
}

func TestSamplerThereafterZero(t *testing.T) {
	allow := NewSampler(3, 0, time.Second)
	var got []int
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("thereafter=0: паника %v", p)
			}
		}()
		got = chkPassed(allow, "k", 10, chkT(0))
	}()
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("thereafter=0: пропущены %v, ожидали [1 2 3]", got)
	}
}

func TestSamplerKeysAndWindows(t *testing.T) {
	allow := NewSampler(1, 0, time.Second)
	if !allow("a", chkT(0)) || allow("a", chkT(10)) {
		t.Fatal("ключ a: первое сообщение проходит, второе в том же окне — нет")
	}
	if !allow("b", chkT(500)) {
		t.Fatal("у ключа b свой счётчик: его первое сообщение должно пройти")
	}
	if allow("a", chkT(999)) {
		t.Fatal("ключ a, 999 мс — ещё то же окно")
	}
	if !allow("a", chkT(1000)) {
		t.Fatal("ключ a, ровно через tick — новое окно, сообщение проходит")
	}
	if allow("b", chkT(1400)) {
		t.Fatal("окно ключа b началось в 500 мс и длится до 1500 — сообщение в 1400 не проходит")
	}
	if !allow("b", chkT(1500)) {
		t.Fatal("ключ b, 1500 мс — его новое окно")
	}
}

func TestSamplerConcurrent(t *testing.T) {
	allow := NewSampler(10, 0, time.Minute)
	var ok atomic.Int32
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if allow("hot", chkT(0)) {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	if n := ok.Load(); n != 10 {
		t.Fatalf("100 горутин, first=10: пропущено %d, ожидали ровно 10", n)
	}
}
