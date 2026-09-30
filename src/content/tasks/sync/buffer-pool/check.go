package main

func TestPoolGetEmpty(t *testing.T) {
	p := NewBufferPool(1024)
	for range 10 {
		b := p.Get()
		if b == nil || b.Len() != 0 {
			t.Fatalf("Get вернул непустой буфер: %q — забыли Reset?", b.String())
		}
		b.WriteString("мусор от прошлого запроса")
		p.Put(b)
	}
	p.Put(nil)
}

func TestPoolDropsHuge(t *testing.T) {
	p := NewBufferPool(1024)
	for range 10 {
		b := p.Get()
		b.Grow(1 << 20)
		b.WriteString("x")
		b.Reset() // длина 0, но под буфером мегабайт
		p.Put(b)
		if c := p.Get().Cap(); c > 1024 {
			t.Fatalf("Get вернул буфер ёмкостью %d > maxCap 1024 — выросший буфер вернулся в пул", c)
		}
	}
}

func TestFormatKV(t *testing.T) {
	p := NewBufferPool(1024)
	cases := []struct {
		m    map[string]int
		want string
	}{
		{map[string]int{"b": 22, "a": 1, "c": -3}, "a=1 b=22 c=-3\n"},
		{map[string]int{}, "\n"},
		{map[string]int{"z": 0}, "z=0\n"},
	}
	for _, c := range cases {
		if got := string(p.FormatKV(c.m)); got != c.want {
			t.Fatalf("FormatKV(%v) = %q, ожидали %q", c.m, got, c.want)
		}
	}
}

func TestFormatKVOwnsResult(t *testing.T) {
	p := NewBufferPool(1024)
	first := p.FormatKV(map[string]int{"user": 1})
	for i := range 20 {
		p.FormatKV(map[string]int{"zzzz": 9999, "yyyy": i})
	}
	if string(first) != "user=1\n" {
		t.Fatalf("первый результат после следующих вызовов стал %q, ожидали \"user=1\\n\" — он смотрит в память буфера из пула", first)
	}
}
