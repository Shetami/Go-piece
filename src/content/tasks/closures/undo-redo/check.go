package main

func chkState(s *Store, keys ...string) string {
	var parts []string
	for _, k := range keys {
		if v, ok := s.Get(k); ok {
			parts = append(parts, k+"="+v)
		} else {
			parts = append(parts, k+"∅")
		}
	}
	return strings.Join(parts, " ")
}

func TestUndoRestoresAbsence(t *testing.T) {
	s := NewStore()
	s.Set("a", "1")
	s.Set("a", "2")
	s.Delete("a")
	s.Delete("nope") // ничего не меняет — не в истории
	steps := []string{"a=2", "a=1", "a∅"}
	for i, want := range steps {
		if !s.Undo() {
			t.Fatalf("Undo №%d вернул false", i+1)
		}
		if got := chkState(s, "a", "nope"); got != want+" nope∅" {
			t.Fatalf("после Undo №%d: %s, ожидали %s nope∅", i+1, got, want)
		}
	}
	if s.Undo() {
		t.Fatal("история пуста — Undo должен вернуть false")
	}
}

func TestRedoAndClear(t *testing.T) {
	s := NewStore()
	s.Set("x", "1")
	s.Set("x", "2")
	s.Undo()
	s.Undo()
	if !s.Redo() || chkState(s, "x") != "x=1" {
		t.Fatalf("Redo: %s, ожидали x=1", chkState(s, "x"))
	}
	if !s.Redo() || chkState(s, "x") != "x=2" {
		t.Fatalf("второй Redo: %s, ожидали x=2", chkState(s, "x"))
	}
	if s.Redo() {
		t.Fatal("повторять больше нечего — Redo должен вернуть false")
	}
	s.Undo()
	s.Set("y", "new")
	if s.Redo() {
		t.Fatal("новое изменение после Undo должно очищать redo")
	}
	s.Undo()
	if got := chkState(s, "x", "y"); got != "x=1 y∅" {
		t.Fatalf("после Undo: %s, ожидали x=1 y∅", got)
	}
}

func TestBatch(t *testing.T) {
	s := NewStore()
	s.Set("keep", "k")
	s.Batch(func() {
		s.Set("a", "1")
		s.Set("a", "2")
		s.Batch(func() { s.Set("b", "1") })
		s.Delete("keep")
	})
	s.Batch(func() {}) // пустой — не в истории
	if !s.Undo() {
		t.Fatal("Undo после Batch вернул false")
	}
	if got := chkState(s, "a", "b", "keep"); got != "a∅ b∅ keep=k" {
		t.Fatalf("один Undo отменяет весь Batch (и вложенный): %s, ожидали a∅ b∅ keep=k", got)
	}
	if !s.Redo() {
		t.Fatal("Redo после отмены Batch вернул false")
	}
	if got := chkState(s, "a", "b", "keep"); got != "a=2 b=1 keep∅" {
		t.Fatalf("Redo повторяет Batch в исходном порядке: %s, ожидали a=2 b=1 keep∅", got)
	}
	s.Undo()
	s.Undo()
	if got := chkState(s, "keep"); got != "keep∅" {
		t.Fatalf("второй Undo отменяет Set(keep): %s", got)
	}
}

func TestBatchPanic(t *testing.T) {
	s := NewStore()
	func() {
		defer func() { recover() }()
		s.Batch(func() {
			s.Set("a", "1")
			panic("boom")
		})
	}()
	s.Set("b", "2")
	s.Undo()
	if got := chkState(s, "a", "b"); got != "a=1 b∅" {
		t.Fatalf("после паники в Batch Store должен выйти из режима пачки: после Undo %s, ожидали a=1 b∅", got)
	}
	s.Undo()
	if got := chkState(s, "a"); got != "a∅" {
		t.Fatalf("изменения до паники — отдельная запись в истории: %s, ожидали a∅", got)
	}
}
