package main

func TestAfterThird(t *testing.T) {
	hits := 0
	g := After(3, func() { hits++ })
	var got []int
	for range 5 {
		g()
		got = append(got, hits)
	}
	if !reflect.DeepEqual(got, []int{0, 0, 1, 2, 3}) {
		t.Fatalf("после каждого вызова hits = %v, ожидали [0 0 1 2 3]", got)
	}
}

func TestAfterOneAndZero(t *testing.T) {
	for _, n := range []int{1, 0, -5} {
		hits := 0
		After(n, func() { hits++ })()
		if hits != 1 {
			t.Fatalf("After(%d): первый вызов должен вызвать f", n)
		}
	}
}

func TestAfterIndependent(t *testing.T) {
	a, b := 0, 0
	ga := After(2, func() { a++ })
	gb := After(2, func() { b++ })
	ga()
	ga()
	gb()
	if a != 1 || b != 0 {
		t.Fatalf("счётчики должны быть независимы: a=%d b=%d, ожидали 1 и 0", a, b)
	}
}
