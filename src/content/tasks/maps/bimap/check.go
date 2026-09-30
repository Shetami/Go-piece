package main

func chkPair(t *testing.T, b *BiMap[string, int], k string, v int) {
	t.Helper()
	if got, ok := b.GetByKey(k); !ok || got != v {
		t.Fatalf("GetByKey(%q) = %d, %v; ожидали %d, true", k, got, ok, v)
	}
	if got, ok := b.GetByValue(v); !ok || got != k {
		t.Fatalf("GetByValue(%d) = %q, %v; ожидали %q, true", v, got, ok, k)
	}
}

func TestBiMapBasic(t *testing.T) {
	b := NewBiMap[string, int]()
	b.Put("a", 1)
	b.Put("b", 2)
	chkPair(t, b, "a", 1)
	chkPair(t, b, "b", 2)
	if _, ok := b.GetByKey("нет"); ok {
		t.Fatal("GetByKey отсутствующего ключа вернул ok")
	}
	b.Put("a", 1)
	if b.Len() != 2 {
		t.Fatalf("повторный Put той же пары: Len = %d, ожидали 2", b.Len())
	}
}

func TestBiMapRebindKey(t *testing.T) {
	b := NewBiMap[string, int]()
	b.Put("a", 1)
	b.Put("a", 2)
	chkPair(t, b, "a", 2)
	if k, ok := b.GetByValue(1); ok {
		t.Fatalf("после Put(a, 2) значение 1 всё ещё указывает на %q — старая обратная связь не удалена", k)
	}
	if b.Len() != 1 {
		t.Fatalf("Len = %d, ожидали 1", b.Len())
	}
}

func TestBiMapStealValue(t *testing.T) {
	b := NewBiMap[string, int]()
	b.Put("a", 1)
	b.Put("b", 2)
	b.Put("a", 2) // рвёт и a→1, и b→2
	chkPair(t, b, "a", 2)
	if v, ok := b.GetByKey("b"); ok {
		t.Fatalf("b всё ещё связан с %d, хотя его значение забрал a", v)
	}
	if _, ok := b.GetByValue(1); ok {
		t.Fatal("значение 1 всё ещё в обратной мапе")
	}
	if b.Len() != 1 {
		t.Fatalf("Len = %d, ожидали 1", b.Len())
	}
}

func TestBiMapDeleteAndZero(t *testing.T) {
	b := NewBiMap[string, int]()
	b.Put("", 0)
	chkPair(t, b, "", 0)
	b.Put("x", 7)
	if !b.DeleteValue(0) {
		t.Fatal("DeleteValue(0) вернул false, хотя пара (\"\", 0) была")
	}
	if _, ok := b.GetByKey(""); ok {
		t.Fatal("после DeleteValue(0) ключ \"\" остался")
	}
	if !b.DeleteKey("x") || b.DeleteKey("x") {
		t.Fatal("DeleteKey(x): первый раз ожидали true, второй — false")
	}
	if _, ok := b.GetByValue(7); ok || b.Len() != 0 {
		t.Fatalf("после удаления всего: GetByValue(7) ok=%v, Len=%d; ожидали false, 0", ok, b.Len())
	}
}
