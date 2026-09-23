package main

func TestLRUGetPut(t *testing.T) {
	c := NewLRU(2)
	c.Put("a", 1)
	c.Put("b", 2)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("Get(a) = %d, %v; ожидали 1, true", v, ok)
	}
	if _, ok := c.Get("нет"); ok {
		t.Fatal("Get несуществующего ключа вернул ok")
	}
}

func TestLRUEvictsLeastRecent(t *testing.T) {
	c := NewLRU(2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a") // теперь самый старый — b
	c.Put("c", 3)
	if _, ok := c.Get("b"); ok {
		t.Fatal("должен был вытесниться b: к нему дольше всего не обращались")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("a вытеснился, хотя к нему обращались недавно")
	}
	if c.Len() != 2 {
		t.Fatalf("Len = %d, ожидали 2", c.Len())
	}
}

func TestLRUPutUpdates(t *testing.T) {
	c := NewLRU(2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("a", 10) // обновление — тоже обращение
	c.Put("c", 3)
	if v, ok := c.Get("a"); !ok || v != 10 {
		t.Fatalf("Get(a) = %d, %v; ожидали 10, true", v, ok)
	}
	if _, ok := c.Get("b"); ok {
		t.Fatal("должен был вытесниться b")
	}
	if c.Len() != 2 {
		t.Fatalf("Len = %d после обновления, ожидали 2", c.Len())
	}
}

func TestLRUManyOps(t *testing.T) {
	c := NewLRU(100)
	for i := range 1000 {
		c.Put(strconv.Itoa(i), i)
	}
	if c.Len() != 100 {
		t.Fatalf("Len = %d, ожидали 100", c.Len())
	}
	if v, ok := c.Get("999"); !ok || v != 999 {
		t.Fatal("последняя запись должна быть в кэше")
	}
	if _, ok := c.Get("899"); ok {
		t.Fatal("запись 899 должна была вытесниться")
	}
}
