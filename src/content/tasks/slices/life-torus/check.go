package main

func chkLive(l *Life, h, w int) [][2]int {
	out := [][2]int{}
	for r := range h {
		for c := range w {
			if l.Alive(r, c) {
				out = append(out, [2]int{r, c})
			}
		}
	}
	return out
}

func TestLifeBlinker(t *testing.T) {
	l := NewLife(5, 5)
	for c := 1; c <= 3; c++ {
		l.Set(2, c, true)
	}
	l.Step()
	if got, want := chkLive(l, 5, 5), [][2]int{{1, 2}, {2, 2}, {3, 2}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("мигалка после шага: %v, ожидали %v (клетки меняются одновременно, по состоянию до шага)", got, want)
	}
	l.Step()
	if got, want := chkLive(l, 5, 5), [][2]int{{2, 1}, {2, 2}, {2, 3}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("мигалка через два шага: %v, ожидали %v", got, want)
	}
}

func TestLifeWrapEdge(t *testing.T) {
	l := NewLife(6, 7)
	// Вертикальная мигалка через верхний край: строки 5, 0, 1 в столбце 0.
	l.Set(5, 0, true)
	l.Set(0, 0, true)
	l.Set(1, 0, true)
	l.Step()
	if got, want := chkLive(l, 6, 7), [][2]int{{0, 0}, {0, 1}, {0, 6}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("мигалка на краю тора: %v, ожидали %v — края склеены", got, want)
	}
}

func TestLifeGliderLoops(t *testing.T) {
	l := NewLife(8, 8)
	glider := [][2]int{{0, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}}
	for _, p := range glider {
		l.Set(p[0], p[1], true)
	}
	start := chkLive(l, 8, 8)
	for range 32 { // планер сдвигается на клетку по диагонали за 4 шага
		l.Step()
	}
	if got := chkLive(l, 8, 8); !reflect.DeepEqual(got, start) {
		t.Fatalf("планер на торе 8×8 за 32 шага должен вернуться на место: %v, ожидали %v", got, start)
	}
}

func TestLifeStepNoAlloc(t *testing.T) {
	l := NewLife(30, 40)
	for i := range 30 {
		l.Set(i, (i*7)%40, true)
		l.Set(i, (i*3+1)%40, true)
	}
	if allocs := testing.AllocsPerRun(20, l.Step); allocs > 0 {
		t.Fatalf("Step выделяет %.0f раз за шаг — второе поле надо переиспользовать", allocs)
	}
}

func TestLifeSmallPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("NewLife(2, 5): ожидали панику")
		}
	}()
	NewLife(2, 5)
}
