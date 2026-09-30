package main

// chkRows — итератор по слайсу с ошибкой в конце и журналом.
type chkRows struct {
	xs       []int
	i        int
	err      error
	closeErr error
	closes   int
	nexts    int
}

func (r *chkRows) Next() bool {
	r.nexts++
	if r.closes > 0 {
		panic("Next после Close")
	}
	if r.i >= len(r.xs) {
		return false
	}
	r.i++
	return true
}
func (r *chkRows) Value() int   { return r.xs[r.i-1] }
func (r *chkRows) Err() error   { return r.err }
func (r *chkRows) Close() error { r.closes++; return r.closeErr }

// chkNoClose — без Close.
type chkNoClose struct{ chkRows }

func (chkNoClose) Close(int) {} // не io.Closer: другая сигнатура

func TestAllBasic(t *testing.T) {
	r := &chkRows{xs: []int{1, 2, 3}}
	var got []int
	for v, err := range All[int](r) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if fmt.Sprint(got) != "[1 2 3]" || r.closes != 1 {
		t.Fatalf("получили %v, Close вызван %d раз; ожидали [1 2 3] и 1", got, r.closes)
	}
	nc := &chkNoClose{chkRows{xs: []int{5}}}
	if xs, err := Collect[int](nc); err != nil || fmt.Sprint(xs) != "[5]" {
		t.Fatalf("итератор без Close: %v, %v", xs, err)
	}
}

func TestAllBreak(t *testing.T) {
	r := &chkRows{xs: []int{1, 2, 3, 4}}
	for v := range All[int](r) {
		if v == 2 {
			break
		}
	}
	if r.closes != 1 || r.nexts != 2 {
		t.Fatalf("break на втором: Close %d раз, Next %d раз; ожидали 1 и 2", r.closes, r.nexts)
	}
}

func TestAllErrors(t *testing.T) {
	boom := errors.New("conn lost")
	r := &chkRows{xs: []int{1, 2}, err: boom, closeErr: errors.New("close failed")}
	var pairs []string
	for v, err := range All[int](r) {
		pairs = append(pairs, fmt.Sprint(v, err))
	}
	if strings.Join(pairs, "; ") != "1 <nil>; 2 <nil>; 0 conn lost" || r.closes != 1 {
		t.Fatalf("пары %q, Close %d; ожидали ошибку чтения последней парой, ошибку Close — нет (ошибка чтения важнее)", pairs, r.closes)
	}
	r = &chkRows{xs: []int{7}, closeErr: errors.New("close failed")}
	xs, err := Collect[int](r)
	if xs != nil || err == nil || err.Error() != "close failed" {
		t.Fatalf("ошибка только в Close: %v, %v; ожидали nil и close failed", xs, err)
	}
	r = &chkRows{xs: []int{1, 2, 3}, err: boom}
	if xs, err := Collect[int](r); xs != nil || err != boom || r.closes != 1 {
		t.Fatalf("Collect с ошибкой: %v, %v, Close %d", xs, err, r.closes)
	}
}

func TestFilter(t *testing.T) {
	r := &chkRows{xs: []int{1, 2, 3, 4, 5, 6}}
	even := Filter[int](r, func(v int) bool { return v%2 == 0 })
	xs, err := Collect(even)
	if err != nil || fmt.Sprint(xs) != "[2 4 6]" || r.closes != 1 {
		t.Fatalf("Filter: %v, %v, Close исходного %d; ожидали [2 4 6] и закрытие исходного", xs, err, r.closes)
	}
	boom := errors.New("bad row")
	r = &chkRows{xs: []int{1, 3}, err: boom}
	if _, err := Collect(Filter[int](r, func(v int) bool { return v > 100 })); err != boom {
		t.Fatalf("ошибка исходного через Filter: %v, ожидали bad row", err)
	}
	r = &chkRows{xs: []int{2, 4, 6}}
	for range All(Filter[int](r, func(int) bool { return true })) {
		break
	}
	if r.closes != 1 {
		t.Fatalf("break по Filter: Close исходного %d раз, ожидали 1", r.closes)
	}
}
