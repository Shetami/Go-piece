package main

func chkTop[K cmp.Ordered](t *testing.T, got, want []Entry[K]) {
	t.Helper()
	if len(got) != len(want) || (len(got) > 0 && !reflect.DeepEqual(got, want)) {
		t.Fatalf("TopK = %v, ожидали %v", got, want)
	}
}

func TestTopKBasic(t *testing.T) {
	c := NewCounter[string]()
	for _, w := range strings.Fields("go sql go k8s go sql redis") {
		c.Add(w, 1)
	}
	chkTop(t, c.TopK(2), []Entry[string]{{"go", 3}, {"sql", 2}})
	chkTop(t, c.TopK(2), []Entry[string]{{"go", 3}, {"sql", 2}}) // TopK ничего не портит
	if c.Count("go") != 3 || c.Len() != 4 {
		t.Fatalf("после TopK: Count(go)=%d Len=%d, ожидали 3 и 4", c.Count("go"), c.Len())
	}
}

func TestTopKTies(t *testing.T) {
	c := NewCounter[string]()
	for _, w := range []string{"e", "d", "c", "b", "a"} {
		c.Add(w, 5)
	}
	c.Add("z", 9)
	for range 20 {
		chkTop(t, c.TopK(3), []Entry[string]{{"z", 9}, {"a", 5}, {"b", 5}})
	}
}

func TestTopKNegativeDelta(t *testing.T) {
	c := NewCounter[int]()
	c.Add(1, 5)
	c.Add(2, 3)
	c.Add(2, -3)
	c.Add(3, -4)
	c.Add(1, -2)
	if c.Len() != 1 || c.Count(2) != 0 || c.Count(3) != 0 {
		t.Fatalf("Len=%d Count(2)=%d Count(3)=%d; ожидали 1, 0, 0 — ключи с ≤0 удаляются", c.Len(), c.Count(2), c.Count(3))
	}
	chkTop(t, c.TopK(5), []Entry[int]{{1, 3}})
	c.Add(3, 1) // после ухода в минус ключ начинает с нуля
	chkTop(t, c.TopK(5), []Entry[int]{{1, 3}, {3, 1}})
}

func TestTopKEdges(t *testing.T) {
	c := NewCounter[string]()
	chkTop(t, c.TopK(3), nil)
	c.Add("x", 1)
	chkTop(t, c.TopK(0), nil)
	chkTop(t, c.TopK(-1), nil)
	chkTop(t, c.TopK(100), []Entry[string]{{"x", 1}})
}

func TestTopKMany(t *testing.T) {
	c := NewCounter[int]()
	for i := range 2000 {
		c.Add(i, i%100+1)
	}
	got := c.TopK(3)
	chkTop(t, got, []Entry[int]{{99, 100}, {199, 100}, {299, 100}})
}
