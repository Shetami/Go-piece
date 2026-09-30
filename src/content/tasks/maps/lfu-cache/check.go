package main

func chkHas(t *testing.T, c *LFU[string, int], k string, want bool) {
	t.Helper()
	if _, ok := c.Get(k); ok != want {
		if want {
			t.Fatalf("ключ %q вытеснен, а не должен был", k)
		}
		t.Fatalf("ключ %q должен был вытесниться, но он в кэше", k)
	}
}

func TestLFUClassic(t *testing.T) {
	c := NewLFU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")    // a:2 b:1
	c.Put("c", 3) // вытесняет b
	chkHas(t, c, "b", false)
	chkHas(t, c, "c", true) // c:2 a:2 — ничья, давнее обращение у a
	c.Put("d", 4)           // вытесняет a
	chkHas(t, c, "a", false)
	chkHas(t, c, "c", true)
	chkHas(t, c, "d", true)
	if c.Len() != 2 {
		t.Fatalf("Len = %d, ожидали 2", c.Len())
	}
}

func TestLFUTieIsLRU(t *testing.T) {
	c := NewLFU[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	c.Get("a")
	c.Get("b")
	c.Get("c")
	c.Get("a") // частоты a:3 b:2 c:2; из b и c давнее — b
	c.Put("d", 4)
	chkHas(t, c, "b", false)
	chkHas(t, c, "c", true)
}

func TestLFUPutCountsAsUse(t *testing.T) {
	c := NewLFU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("a", 10) // a:2
	c.Put("c", 3)
	chkHas(t, c, "b", false)
	if v, ok := c.Get("a"); !ok || v != 10 {
		t.Fatalf("Get(a) = %d, %v; ожидали 10, true", v, ok)
	}
}

func TestLFUNewKeyAfterEviction(t *testing.T) {
	c := NewLFU[string, int](2)
	c.Put("a", 1)
	for range 5 {
		c.Get("a")
	}
	c.Put("b", 2)
	for range 5 {
		c.Get("b")
	}
	c.Put("c", 3) // вытесняет a (частоты равны, a давнее); c:1
	c.Put("a", 1) // a вернулся с частотой 1, вытесняет c — самый редкий
	chkHas(t, c, "c", false)
	chkHas(t, c, "b", true)
	c.Put("e", 5) // у a по-прежнему частота 1 — вытесняется он
	chkHas(t, c, "a", false)
	chkHas(t, c, "e", true)
}

func TestLFUZeroCapacity(t *testing.T) {
	c := NewLFU[string, int](0)
	c.Put("a", 1)
	c.Get("a")
	if c.Len() != 0 {
		t.Fatalf("capacity 0: Len = %d, ожидали 0", c.Len())
	}
	chkHas(t, c, "a", false)
}
