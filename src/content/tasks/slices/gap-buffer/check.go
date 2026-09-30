package main

func TestEditorBasic(t *testing.T) {
	var e Editor
	e.Insert("привет мир")
	e.Move(-4)
	e.Insert(",")
	if got := e.String(); got != "привет, мир" || e.Cursor() != 7 {
		t.Fatalf("текст %q, курсор %d; ожидали \"привет, мир\" и 7 (курсор — в символах)", got, e.Cursor())
	}
	if n := e.Backspace(100); n != 7 || e.String() != " мир" || e.Cursor() != 0 {
		t.Fatalf("Backspace(100) = %d, текст %q; ожидали 7 и \" мир\"", n, e.String())
	}
	e.Move(-5)
	e.Move(100)
	if e.Cursor() != 4 {
		t.Fatalf("курсор %d, ожидали 4 — Move не выходит за границы", e.Cursor())
	}
	e.Move(-4)
	if n := e.Delete(1); n != 1 || e.String() != "мир" {
		t.Fatalf("Delete(1) = %d, текст %q", n, e.String())
	}
	if n := e.Delete(-3); n != 0 || e.Backspace(-1) != 0 {
		t.Fatalf("отрицательное n ничего не удаляет")
	}
}

func TestEditorAgainstModel(t *testing.T) {
	var e Editor
	var model []rune
	cur := 0
	alphabet := []rune("abcЖЯ😀")
	seed := uint32(3)
	rnd := func(n int) int {
		seed = seed*1664525 + 1013904223
		return int(seed>>16) % n
	}
	for step := range 5000 {
		switch op := rnd(4); op {
		case 0:
			s := string(alphabet[rnd(len(alphabet))]) + string(alphabet[rnd(len(alphabet))])
			e.Insert(s)
			model = slices.Insert(model, cur, []rune(s)...)
			cur += 2
		case 1:
			d := rnd(40) - 20
			e.Move(d)
			cur = min(max(cur+d, 0), len(model))
		case 2:
			n := rnd(4)
			k := min(n, cur)
			if got := e.Backspace(n); got != k {
				t.Fatalf("шаг %d: Backspace(%d) = %d, ожидали %d", step, n, got, k)
			}
			model = slices.Delete(model, cur-k, cur)
			cur -= k
		case 3:
			n := rnd(3)
			k := min(n, len(model)-cur)
			if got := e.Delete(n); got != k {
				t.Fatalf("шаг %d: Delete(%d) = %d, ожидали %d", step, n, got, k)
			}
			model = slices.Delete(model, cur, cur+k)
		}
		if e.Cursor() != cur || e.String() != string(model) {
			t.Fatalf("шаг %d: текст %q курсор %d, ожидали %q и %d", step, e.String(), e.Cursor(), string(model), cur)
		}
	}
}

func TestEditorTypingInMiddle(t *testing.T) {
	var e Editor
	start := time.Now()
	for range 100000 {
		e.Insert("a")
	}
	e.Move(-50000)
	for range 100000 {
		e.Insert("b")
	}
	e.Move(-1000)
	e.Move(1000)
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("200000 нажатий в середине текста заняли %v — каждая вставка сдвигает хвост?", d)
	}
	s := e.String()
	if len(s) != 200000 || s[49999] != 'a' || s[50000] != 'b' || s[149999] != 'b' || s[150000] != 'a' {
		t.Fatalf("после набора в середине текст неверный")
	}
}
