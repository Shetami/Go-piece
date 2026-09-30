package main

func chkAt(ms int) time.Time {
	return time.Unix(1_700_000_000, 0).Add(time.Duration(ms) * time.Millisecond)
}

func TestDebounceLastValue(t *testing.T) {
	var got []string
	call, tick := NewDebouncer(100*time.Millisecond, 0, func(s string) { got = append(got, s) })
	tick(chkAt(0))
	call("a", chkAt(0))
	call("b", chkAt(50))
	tick(chkAt(120))
	if len(got) != 0 {
		t.Fatalf("через 70 мс после последнего call уже вызвали f(%v) — таймер должен сбрасываться каждым call", got)
	}
	tick(chkAt(149))
	tick(chkAt(150))
	tick(chkAt(300))
	if !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("f вызвана с %q, ожидали ровно один вызов с последним значением [b] ровно через wait", got)
	}
	call("c", chkAt(400))
	tick(chkAt(500))
	if !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Fatalf("новая серия после срабатывания: %q, ожидали [b c]", got)
	}
}

func TestDebounceZeroValue(t *testing.T) {
	calls := 0
	var last = -1
	call, tick := NewDebouncer(10*time.Millisecond, 0, func(v int) { calls++; last = v })
	call(0, chkAt(0))
	tick(chkAt(10))
	if calls != 1 || last != 0 {
		t.Fatalf("отложенное значение 0 — тоже значение: вызовов %d, аргумент %d; ожидали 1 вызов с 0", calls, last)
	}
}

func TestDebounceMaxWait(t *testing.T) {
	var got []int
	call, tick := NewDebouncer(100*time.Millisecond, 250*time.Millisecond, func(v int) { got = append(got, v) })
	for ms := 0; ms <= 400; ms += 50 {
		call(ms, chkAt(ms))
		tick(chkAt(ms))
	}
	tick(chkAt(499))
	tick(chkAt(500))
	if !reflect.DeepEqual(got, []int{250, 400}) {
		t.Fatalf("вызовы каждые 50 мс, wait=100, maxWait=250: f получила %v, ожидали [250 400] — maxWait считается от начала серии", got)
	}
}
