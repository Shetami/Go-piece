package main

func TestPipeOrder(t *testing.T) {
	inc := func(x int) int { return x + 1 }
	dbl := func(x int) int { return x * 2 }
	if got := Pipe(inc, dbl)(3); got != 8 {
		t.Fatalf("Pipe(inc, dbl)(3) = %d, ожидали 8: сначала +1, потом ×2", got)
	}
	if got := Pipe(dbl, inc)(3); got != 7 {
		t.Fatalf("Pipe(dbl, inc)(3) = %d, ожидали 7", got)
	}
}

func TestPipeEmpty(t *testing.T) {
	id := Pipe()
	if id == nil {
		t.Fatal("Pipe() должен вернуть функцию, а не nil")
	}
	if got := id(42); got != 42 {
		t.Fatalf("Pipe()(42) = %d, ожидали 42", got)
	}
}

func TestPipeReusable(t *testing.T) {
	p := Pipe(func(x int) int { return x * 10 })
	if p(1) != 10 || p(2) != 20 {
		t.Fatal("собранную функцию можно вызывать много раз")
	}
}

func TestPipeDoesNotAliasArgs(t *testing.T) {
	fs := []func(int) int{func(x int) int { return x + 1 }}
	p := Pipe(fs...)
	fs[0] = func(x int) int { return -1 }
	if got := p(1); got != 2 {
		t.Fatalf("Pipe должен зафиксировать функции в момент сборки: p(1) = %d, ожидали 2", got)
	}
}
