package main

func chkGet(t *testing.T, m *TimeMap, key string, ts int, want string, wantOK bool) {
	t.Helper()
	if got, ok := m.Get(key, ts); got != want || ok != wantOK {
		t.Fatalf("Get(%q, %d) = %q, %v; ожидали %q, %v", key, ts, got, ok, want, wantOK)
	}
}

func TestTimeMapBasic(t *testing.T) {
	m := NewTimeMap()
	m.Set("price", "100", 10)
	m.Set("price", "120", 20)
	chkGet(t, m, "price", 10, "100", true)
	chkGet(t, m, "price", 15, "100", true)
	chkGet(t, m, "price", 20, "120", true)
	chkGet(t, m, "price", 1000, "120", true)
	chkGet(t, m, "price", 9, "", false)
	chkGet(t, m, "нет", 10, "", false)
}

func TestTimeMapOutOfOrder(t *testing.T) {
	m := NewTimeMap()
	m.Set("k", "c", 30)
	m.Set("k", "a", 10)
	m.Set("k", "d", 40)
	m.Set("k", "b", 20)
	for ts, want := range map[int]string{10: "a", 19: "a", 20: "b", 35: "c", 40: "d", 99: "d"} {
		chkGet(t, m, "k", ts, want, true)
	}
	chkGet(t, m, "k", 5, "", false)
}

func TestTimeMapOverwrite(t *testing.T) {
	m := NewTimeMap()
	m.Set("k", "old", 10)
	m.Set("k", "x", 20)
	m.Set("k", "new", 10)
	chkGet(t, m, "k", 15, "new", true)
	m.Set("k", "", 25) // пустая строка — тоже значение
	chkGet(t, m, "k", 30, "", true)
	chkGet(t, m, "k", 20, "x", true)
}

func TestTimeMapNegativeAndZero(t *testing.T) {
	m := NewTimeMap()
	m.Set("k", "zero", 0)
	m.Set("k", "neg", -5)
	chkGet(t, m, "k", -1, "neg", true)
	chkGet(t, m, "k", 0, "zero", true)
	chkGet(t, m, "k", -6, "", false)
}

func TestTimeMapMany(t *testing.T) {
	m := NewTimeMap()
	for i := 0; i < 2000; i += 2 {
		m.Set("k", strconv.Itoa(i), i)
	}
	for i := 1999; i >= 1; i -= 2 { // догружаем нечётные с конца
		m.Set("k", strconv.Itoa(i), i)
	}
	for _, ts := range []int{0, 1, 2, 777, 1998, 1999, 5000} {
		chkGet(t, m, "k", ts, strconv.Itoa(min(ts, 1999)), true)
	}
}
