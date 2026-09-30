package main

func chkOMKeys[K comparable, V any](m *OrderedMap[K, V]) (keys []K) {
	for k := range m.All() {
		keys = append(keys, k)
	}
	return keys
}

func chkOMWithin(t *testing.T, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("паника: %v", r)
			}
		}()
		f()
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("обход не закончился за 3 секунды — зациклился?")
	}
}

func TestOMOrder(t *testing.T) {
	var m OrderedMap[string, int]
	m.Set("b", 1)
	m.Set("a", 2)
	m.Set("c", 3)
	m.Set("b", 10)
	if got := chkOMKeys(&m); !reflect.DeepEqual(got, []string{"b", "a", "c"}) || m.Len() != 3 {
		t.Fatalf("ключи = %q, Len = %d; обновление не должно двигать ключ", got, m.Len())
	}
	if v, _ := m.Get("b"); v != 10 {
		t.Fatalf("Get(b) = %d", v)
	}
	if !m.Delete("b") || m.Delete("b") || m.Delete("zzz") {
		t.Fatalf("Delete: true для существующего, false для остальных")
	}
	m.Set("b", 5)
	if got := chkOMKeys(&m); !reflect.DeepEqual(got, []string{"a", "c", "b"}) {
		t.Fatalf("удалённый и снова добавленный ключ встаёт в конец: %q", got)
	}
	m.Delete("a")
	m.Delete("c")
	m.Delete("b")
	m.Set("x", 1)
	if got := chkOMKeys(&m); !reflect.DeepEqual(got, []string{"x"}) || m.Len() != 1 {
		t.Fatalf("после удаления всех и вставки: %q, Len = %d", got, m.Len())
	}
	for k := range m.All() {
		if k == "x" {
			break
		}
	}
}

func TestOMDeleteDuringIteration(t *testing.T) {
	chkOMWithin(t, func() {
		var m OrderedMap[int, string]
		for i := range 10 {
			m.Set(i, strconv.Itoa(i))
		}
		var seen []int
		for k := range m.All() {
			seen = append(seen, k)
			m.Delete(k)     // текущий
			m.Delete(k + 1) // ещё не выданный
		}
		if !reflect.DeepEqual(seen, []int{0, 2, 4, 6, 8}) || m.Len() != 0 {
			t.Errorf("удаляя текущий и следующий, получили %v, Len = %d; ожидали [0 2 4 6 8] и 0", seen, m.Len())
		}
	})
}

func TestOMAppendDuringIteration(t *testing.T) {
	chkOMWithin(t, func() {
		var m OrderedMap[string, int]
		m.Set("a", 1)
		m.Set("b", 2)
		var seen []string
		for k, v := range m.All() {
			seen = append(seen, k)
			if k == "b" {
				m.Delete("b") // удаляем хвост, на котором стоит обход…
				m.Set("c", v+1)
				m.Set("d", v+2) // …и дописываем новые
			}
		}
		if !reflect.DeepEqual(seen, []string{"a", "b", "c", "d"}) {
			t.Errorf("удалили текущий хвост и дописали c, d — обход выдал %q, ожидали [a b c d]", seen)
		}
	})
}

func TestOMRandomModel(t *testing.T) {
	r := rand.New(rand.NewPCG(31, 32))
	var m OrderedMap[int, int]
	var ref []int
	for i := range 4000 {
		k := r.IntN(40)
		if r.IntN(3) == 0 {
			j := slices.Index(ref, k)
			if m.Delete(k) != (j >= 0) {
				t.Fatalf("Delete(%d) разошёлся с моделью", k)
			}
			if j >= 0 {
				ref = slices.Delete(ref, j, j+1)
			}
		} else {
			m.Set(k, i)
			if !slices.Contains(ref, k) {
				ref = append(ref, k)
			}
		}
	}
	if got := chkOMKeys(&m); !reflect.DeepEqual(got, ref) || m.Len() != len(ref) {
		t.Fatalf("после 4000 операций порядок %v, ожидали %v", got, ref)
	}
}

func TestOMFast(t *testing.T) {
	var m OrderedMap[int, int]
	start := time.Now()
	const n = 200000
	for i := range n {
		m.Set(i, i)
	}
	for i := range n - 1 {
		m.Delete(i)
	}
	if got := chkOMKeys(&m); !reflect.DeepEqual(got, []int{n - 1}) {
		t.Fatalf("осталось %v", got)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("200 000 Set и Delete заняли %v — Delete не O(1)", d)
	}
}
