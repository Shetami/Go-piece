package main

func TestUnflattenBasic(t *testing.T) {
	got, err := Unflatten(map[string]any{
		"db.host": "x", "db.port": 5432, "db.opts.ssl": true, "debug": nil, "l.0": "a",
	})
	want := map[string]any{
		"db":    map[string]any{"host": "x", "port": 5432, "opts": map[string]any{"ssl": true}},
		"debug": nil,
		"l":     map[string]any{"0": "a"},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Unflatten = %v, %v; ожидали %v", got, err, want)
	}
	if got, err := Unflatten(nil); err != nil || len(got) != 0 {
		t.Fatalf("Unflatten(nil) = %v, %v; ожидали пусто", got, err)
	}
}

func TestUnflattenConflictAnyOrder(t *testing.T) {
	cases := []map[string]any{
		{"a": 1, "a.b": 2},
		{"a.b": 2, "a": nil, "z": 0},
		{"x.y": 1, "x.y.z": 2},
		{"p.q.r": 1, "p": "leaf"},
	}
	for _, c := range cases {
		for range 30 {
			got, err := func() (m map[string]any, err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("паника: %v", r)
					}
				}()
				return Unflatten(c)
			}()
			if !errors.Is(err, ErrConflict) || got != nil {
				t.Fatalf("Unflatten(%v) = %v, %v; ожидали nil и ErrConflict при любом порядке", c, got, err)
			}
		}
	}
}

func TestUnflattenMapLeafUntouched(t *testing.T) {
	leaf := map[string]any{"x": 1}
	flat := map[string]any{"a": leaf, "a.b": 2}
	for range 30 {
		_, err := func() (m map[string]any, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("паника: %v", r)
				}
			}()
			return Unflatten(flat)
		}()
		if !errors.Is(err, ErrConflict) {
			t.Fatalf("лист-мапа a и путь a.b: err = %v, ожидали ErrConflict", err)
		}
		if len(leaf) != 1 {
			t.Fatalf("в мапу-значение из входа дописали ключи: %v", leaf)
		}
	}
	got, err := Unflatten(map[string]any{"cfg": map[string]any{}, "n": 1})
	if err != nil || !reflect.DeepEqual(got, map[string]any{"cfg": map[string]any{}, "n": 1}) {
		t.Fatalf("пустая мапа как лист: %v, %v", got, err)
	}
}

func TestUnflattenBadPath(t *testing.T) {
	for _, p := range []string{"", "a..b", ".a", "a."} {
		got, err := Unflatten(map[string]any{p: 1, "ok": 2})
		if !errors.Is(err, ErrBadPath) || got != nil {
			t.Fatalf("путь %q: %v, %v; ожидали nil и ErrBadPath", p, got, err)
		}
	}
}
