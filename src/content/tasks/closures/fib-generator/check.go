package main

func TestFibSequence(t *testing.T) {
	next := FibGen()
	var got []int
	for range 10 {
		got = append(got, next())
	}
	if want := []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}; !reflect.DeepEqual(got, want) {
		t.Fatalf("первые 10: %v, ожидали %v", got, want)
	}
}

func TestFibIndependent(t *testing.T) {
	a, b := FibGen(), FibGen()
	a()
	a()
	a()
	if got := b(); got != 0 {
		t.Fatalf("второй генератор должен начинать с 0, а отдал %d", got)
	}
	if got := a(); got != 2 {
		t.Fatalf("первый генератор сбился: %d, ожидали 2", got)
	}
}

func TestFibFar(t *testing.T) {
	next := FibGen()
	var v int
	for range 51 {
		v = next()
	}
	if v != 12586269025 {
		t.Fatalf("F(50) = %d, ожидали 12586269025", v)
	}
}
