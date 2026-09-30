package main

var chkT0 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func chkAt(sec float64) time.Time { return chkT0.Add(time.Duration(sec * float64(time.Second))) }

func TestDedupBoundary(t *testing.T) {
	d := NewDedup(10 * time.Second)
	steps := []struct {
		id   string
		at   float64
		want bool
	}{
		{"a", 0, false}, {"a", 0, true}, {"b", 1, false}, {"a", 9.9, true},
		{"a", 10, false}, // ровно через окно — новое
		{"a", 19, true}, {"b", 19, false},
	}
	for _, s := range steps {
		if got := d.Seen(s.id, chkAt(s.at)); got != s.want {
			t.Fatalf("Seen(%s, %vс) = %v, ожидали %v", s.id, s.at, got, s.want)
		}
	}
}

func TestDedupRepeatsDoNotExtend(t *testing.T) {
	d := NewDedup(10 * time.Second)
	d.Seen("x", chkAt(0))
	for s := 3.0; s < 10; s += 3 {
		if !d.Seen("x", chkAt(s)) {
			t.Fatalf("на %vс ожидали повтор", s)
		}
	}
	if d.Seen("x", chkAt(10)) {
		t.Fatal("на 10с событие должно приниматься: повторы на 3, 6, 9 с не продлевают окно")
	}
}

func TestDedupStaleEvictionKeepsFresh(t *testing.T) {
	d := NewDedup(10 * time.Second)
	d.Seen("a", chkAt(0))
	d.Seen("a", chkAt(10)) // принят заново
	d.Seen("b", chkAt(15)) // чистка: старая запись a@0 истекла, но a@10 жив
	if !d.Seen("a", chkAt(16)) {
		t.Fatal("a принят на 10с, на 16с это повтор — чистка удалила свежую запись по старой отметке")
	}
	if d.Len() != 2 {
		t.Fatalf("Len = %d, ожидали 2 (a и b)", d.Len())
	}
}

func TestDedupMemoryBounded(t *testing.T) {
	d := NewDedup(time.Second)
	for i := range 20000 {
		d.Seen(strconv.Itoa(i), chkAt(float64(i)/100)) // 100 событий в секунду
	}
	if n := d.Len(); n > 101 {
		t.Fatalf("Len = %d после 20000 уникальных id при окне в 1с (≈100 событий) — старые записи не удаляются", n)
	}
	d.Seen("late", chkAt(1000))
	if n := d.Len(); n != 1 {
		t.Fatalf("через много времени Len = %d, ожидали 1", n)
	}
}
