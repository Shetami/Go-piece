package main

func eqKeys(a, b []string) bool {
	return len(a) == len(b) && (len(a) == 0 || reflect.DeepEqual(a, b))
}

func TestDiffBasic(t *testing.T) {
	old := map[string]string{"port": "80", "host": "a", "debug": "true", "tls": "off"}
	cur := map[string]string{"port": "8080", "host": "a", "tls": "on", "user": "x", "log": "json"}
	a, r, c := Diff(old, cur)
	if !eqKeys(a, []string{"log", "user"}) {
		t.Fatalf("added = %q, ожидали [log user]", a)
	}
	if !eqKeys(r, []string{"debug"}) {
		t.Fatalf("removed = %q, ожидали [debug]", r)
	}
	if !eqKeys(c, []string{"port", "tls"}) {
		t.Fatalf("changed = %q, ожидали [port tls]", c)
	}
}

func TestDiffEmptyValue(t *testing.T) {
	a, r, c := Diff(map[string]string{}, map[string]string{"x": ""})
	if !eqKeys(a, []string{"x"}) || len(r) != 0 || len(c) != 0 {
		t.Fatalf("ключ с пустым значением появился: added=%q removed=%q changed=%q, ожидали added=[x]", a, r, c)
	}
	a, r, c = Diff(map[string]string{"x": ""}, map[string]string{})
	if len(a) != 0 || !eqKeys(r, []string{"x"}) || len(c) != 0 {
		t.Fatalf("ключ с пустым значением исчез: added=%q removed=%q changed=%q, ожидали removed=[x]", a, r, c)
	}
}

func TestDiffSame(t *testing.T) {
	m := map[string]string{"a": "1", "b": "2"}
	a, r, c := Diff(m, m)
	if len(a)+len(r)+len(c) != 0 {
		t.Fatalf("одинаковые мапы: added=%q removed=%q changed=%q, ожидали пусто", a, r, c)
	}
}

func TestDiffSorted(t *testing.T) {
	cur := map[string]string{}
	var want []string
	for i := range 50 {
		k := fmt.Sprintf("k%02d", i)
		cur[k] = "v"
		want = append(want, k)
	}
	for range 5 {
		if a, _, _ := Diff(nil, cur); !reflect.DeepEqual(a, want) {
			t.Fatalf("added не отсортирован: %q", a)
		}
	}
}
