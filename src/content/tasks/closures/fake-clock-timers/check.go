package main

var chkStart = time.Unix(1_700_000_000, 0)

func chkSince(c *FakeClock) time.Duration { return c.Now().Sub(chkStart) }

// chkRun выполняет f с таймаутом: дедлок превращается в понятное падение.
func chkRun(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { defer close(done); f() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: зависло — f, похоже, вызывается под блокировкой", what)
	}
}

func TestFakeClockOrder(t *testing.T) {
	c := NewFakeClock(chkStart)
	var got []string
	add := func(d time.Duration, name string) {
		c.AfterFunc(d, func() { got = append(got, fmt.Sprintf("%s@%v", name, chkSince(c))) })
	}
	add(3*time.Second, "c")
	add(time.Second, "a")
	add(3*time.Second, "d")
	add(2*time.Second, "b")
	add(10*time.Second, "late")
	chkRun(t, "Advance", func() { c.Advance(5 * time.Second) })
	want := []string{"a@1s", "b@2s", "c@3s", "d@3s"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("сработали %v, ожидали %v — по времени, при равенстве по порядку планирования, Now() — момент таймера", got, want)
	}
	if s := chkSince(c); s != 5*time.Second {
		t.Fatalf("после Advance(5s) Now() = старт+%v, ожидали +5s", s)
	}
}

func TestFakeClockStop(t *testing.T) {
	c := NewFakeClock(chkStart)
	fired := 0
	stop := c.AfterFunc(time.Second, func() { fired++ })
	if !stop() {
		t.Fatal("первый stop несработавшего таймера должен вернуть true")
	}
	if stop() {
		t.Fatal("повторный stop должен вернуть false")
	}
	stop2 := c.AfterFunc(time.Second, func() { fired++ })
	c.Advance(time.Second)
	if fired != 1 {
		t.Fatalf("сработало %d таймеров, ожидали 1 — отменённый не должен срабатывать", fired)
	}
	if stop2() {
		t.Fatal("stop после срабатывания должен вернуть false")
	}
}

func TestFakeClockReentrant(t *testing.T) {
	c := NewFakeClock(chkStart)
	var ticks []time.Duration
	var tick func()
	tick = func() {
		ticks = append(ticks, chkSince(c))
		if len(ticks) < 10 {
			c.AfterFunc(time.Second, tick) // перепланирует себя, как Ticker
		}
	}
	c.AfterFunc(time.Second, tick)
	var victimFired bool
	var stopVictim func() bool
	c.AfterFunc(2*time.Second, func() { stopVictim() })
	stopVictim = c.AfterFunc(2*time.Second, func() { victimFired = true })
	zero := false
	c.AfterFunc(0, func() { zero = true })
	chkRun(t, "Advance с f, которая зовёт AfterFunc, Now и stop", func() { c.Advance(3500 * time.Millisecond) })
	if !reflect.DeepEqual(ticks, []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}) {
		t.Fatalf("самоперепланирующийся таймер сработал в моменты %v, ожидали [1s 2s 3s]", ticks)
	}
	if victimFired {
		t.Fatal("таймер, отменённый другим таймером того же момента, всё равно сработал")
	}
	if !zero {
		t.Fatal("AfterFunc(0) должен сработать при ближайшем Advance")
	}
}

func TestFakeClockAdvanceZero(t *testing.T) {
	c := NewFakeClock(chkStart)
	fired := false
	c.AfterFunc(-time.Second, func() { fired = true })
	c.Advance(0)
	if !fired || chkSince(c) != 0 {
		t.Fatalf("AfterFunc(-1s) + Advance(0): сработал=%v, Now()=старт+%v; ожидали true и +0s", fired, chkSince(c))
	}
}
