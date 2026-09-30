package main

func chkOrder(m *OrderedMap[string, int]) string {
	var parts []string
	for k, v := range m.All() {
		parts = append(parts, fmt.Sprintf("%s=%d", k, v))
	}
	return strings.Join(parts, " ")
}

func TestOrderedMapOrder(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("b", 1)
	m.Set("a", 2)
	m.Set("c", 3)
	m.Set("a", 20) // место не меняется
	if got, want := chkOrder(m), "b=1 a=20 c=3"; got != want {
		t.Fatalf("порядок после обновления: %q, ожидали %q", got, want)
	}
	if !m.Delete("b") || m.Delete("b") {
		t.Fatal("Delete: ожидали true, затем false")
	}
	m.Set("b", 100) // удалённый и добавленный заново — в конец
	if got, want := chkOrder(m), "a=20 c=3 b=100"; got != want {
		t.Fatalf("порядок после Delete+Set: %q, ожидали %q", got, want)
	}
	if v, ok := m.Get("c"); !ok || v != 3 || m.Len() != 3 {
		t.Fatalf("Get(c) = %d, %v, Len = %d; ожидали 3, true, 3", v, ok, m.Len())
	}
}

func TestOrderedMapBreak(t *testing.T) {
	m := NewOrderedMap[string, int]()
	for i := range 5 {
		m.Set(strconv.Itoa(i), i)
	}
	n := 0
	for range m.All() {
		n++
		if n == 2 {
			break // итератор, который продолжит звать yield, вызовет панику
		}
	}
	if n != 2 {
		t.Fatalf("после break сделано %d итераций, ожидали 2", n)
	}
}

func TestOrderedMapDeleteCurrent(t *testing.T) {
	m := NewOrderedMap[string, int]()
	for i := range 5 {
		m.Set(strconv.Itoa(i), i)
	}
	var seen []string
	for k := range m.All() {
		seen = append(seen, k)
		m.Delete(k)
	}
	if strings.Join(seen, "") != "01234" || m.Len() != 0 {
		t.Fatalf("удаление текущего ключа во время обхода: пройдены %v, Len=%d; ожидали все пять и 0", seen, m.Len())
	}
}

func TestOrderedMapMutateAhead(t *testing.T) {
	m := NewOrderedMap[string, int]()
	for _, k := range []string{"a", "b", "c", "d", "e"} {
		m.Set(k, 0)
	}
	var seen []string
	for k, v := range m.All() {
		seen = append(seen, fmt.Sprintf("%s=%d", k, v))
		if k == "a" {
			m.Delete("a")
			m.Delete("b")
			m.Delete("c")
			m.Set("d", 4)
			m.Set("x", 9) // новый ключ — не в этот обход
			m.Delete("e")
			m.Set("e", 5) // тоже новый: удалён и вставлен заново
		}
	}
	if got, want := strings.Join(seen, " "), "a=0 d=4"; got != want {
		t.Fatalf("обход с изменениями: %q, ожидали %q", got, want)
	}
	if got, want := chkOrder(m), "d=4 x=9 e=5"; got != want {
		t.Fatalf("после обхода: %q, ожидали %q", got, want)
	}
}
