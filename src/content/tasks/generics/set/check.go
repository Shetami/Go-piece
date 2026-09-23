package main

func TestSetBasic(t *testing.T) {
	s := NewSet(1, 2, 2, 3)
	if len(s) != 3 {
		t.Fatalf("NewSet(1, 2, 2, 3): %d элементов, ожидали 3", len(s))
	}
	s.Add(4)
	if !s.Has(4) || s.Has(5) {
		t.Fatal("Add/Has работают неправильно")
	}
}

func TestSetEmpty(t *testing.T) {
	s := NewSet[string]()
	s.Add("go") // не должно паниковать
	if !s.Has("go") {
		t.Fatal("пустое множество из NewSet должно принимать элементы")
	}
}

func TestSetUnionIntersect(t *testing.T) {
	a := NewSet("go", "rust", "zig")
	b := NewSet("rust", "zig", "c")
	u := a.Union(b)
	if len(u) != 4 || !u.Has("c") || !u.Has("go") {
		t.Fatalf("Union = %v", u)
	}
	i := a.Intersect(b)
	if len(i) != 2 || !i.Has("rust") || !i.Has("zig") {
		t.Fatalf("Intersect = %v", i)
	}
	if len(a) != 3 || len(b) != 3 {
		t.Fatal("Union и Intersect не должны менять исходные множества")
	}
	u.Add("haskell")
	if a.Has("haskell") {
		t.Fatal("результат Union делит память с исходным множеством")
	}
}
