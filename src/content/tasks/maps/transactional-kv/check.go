package main

func chkKV(t *testing.T, s *Store, k, want string, wantOK bool) {
	t.Helper()
	if v, ok := s.Get(k); v != want || ok != wantOK {
		t.Fatalf("Get(%q) = %q, %v; ожидали %q, %v", k, v, ok, want, wantOK)
	}
}

func TestStoreRollbackRestoresAll(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Set("b", "1")
	s.Begin()
	s.Set("a", "2")
	s.Set("a", "3") // два изменения одного ключа
	s.Delete("b")
	s.Set("c", "1")
	chkKV(t, s, "b", "", false)
	if s.Count("1") != 1 || s.Count("3") != 1 {
		t.Fatalf("внутри транзакции Count(1)=%d Count(3)=%d, ожидали 1 и 1", s.Count("1"), s.Count("3"))
	}
	if err := s.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	chkKV(t, s, "a", "1", true)
	chkKV(t, s, "b", "1", true)
	chkKV(t, s, "c", "", false)
	if s.Count("1") != 2 || s.Count("3") != 0 {
		t.Fatalf("после Rollback Count(1)=%d Count(3)=%d, ожидали 2 и 0", s.Count("1"), s.Count("3"))
	}
}

func TestStoreNestedCommitThenRollback(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Begin()
	s.Set("a", "2")
	s.Begin()
	s.Set("a", "3")
	s.Set("x", "new")
	if err := s.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	chkKV(t, s, "a", "3", true)
	s.Rollback() // откатывает и внешнюю, и влитую в неё внутреннюю
	chkKV(t, s, "a", "1", true)
	chkKV(t, s, "x", "", false)
	if s.Count("1") != 1 || s.Count("2") != 0 || s.Count("new") != 0 {
		t.Fatalf("Count после отката: 1→%d 2→%d new→%d, ожидали 1, 0, 0", s.Count("1"), s.Count("2"), s.Count("new"))
	}
}

func TestStoreInnerRollbackOnly(t *testing.T) {
	s := NewStore()
	s.Begin()
	s.Set("a", "outer")
	s.Begin()
	s.Delete("a")
	s.Set("b", "inner")
	s.Rollback()
	chkKV(t, s, "a", "outer", true)
	chkKV(t, s, "b", "", false)
	if err := s.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if err := s.Rollback(); !errors.Is(err, ErrNoTx) {
		t.Fatalf("Rollback без транзакции: %v, ожидали ErrNoTx", err)
	}
	if err := s.Commit(); !errors.Is(err, ErrNoTx) {
		t.Fatalf("Commit без транзакции: %v, ожидали ErrNoTx", err)
	}
	chkKV(t, s, "a", "outer", true)
}

func TestStoreDeleteAbsentAndEmptyValue(t *testing.T) {
	s := NewStore()
	s.Set("e", "")
	chkKV(t, s, "e", "", true)
	s.Begin()
	s.Delete("нет")
	s.Delete("e")
	s.Set("e", "x")
	s.Delete("e")
	if s.Count("") != 0 {
		t.Fatalf("Count(\"\") = %d, ожидали 0", s.Count(""))
	}
	s.Rollback()
	chkKV(t, s, "e", "", true)
	chkKV(t, s, "нет", "", false)
	if s.Count("") != 1 || s.Count("x") != 0 {
		t.Fatalf("после отката Count(\"\")=%d Count(x)=%d, ожидали 1 и 0", s.Count(""), s.Count("x"))
	}
}

func TestStoreRollbackIsCheap(t *testing.T) {
	s := NewStore()
	for i := range 100000 {
		s.Set(strconv.Itoa(i), "v")
	}
	start := time.Now()
	for range 2000 {
		s.Begin()
		s.Set("0", "w")
		s.Rollback()
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("2000 коротких транзакций на хранилище в 100 000 ключей заняли %v — Rollback копирует всё хранилище?", d)
	}
	if s.Count("v") != 100000 {
		t.Fatalf("Count(v) = %d, ожидали 100000", s.Count("v"))
	}
}
