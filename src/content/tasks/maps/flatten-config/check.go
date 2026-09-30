package main

func TestFlattenBasic(t *testing.T) {
	doc := map[string]any{
		"db":    map[string]any{"host": "x", "ports": []any{5432, 5433}},
		"debug": true,
		"users": []any{map[string]any{"name": "ann"}, map[string]any{"name": "bob"}},
	}
	got, err := Flatten(doc)
	want := map[string]any{
		"db.host": "x", "db.ports.0": 5432, "db.ports.1": 5433, "debug": true,
		"users.0.name": "ann", "users.1.name": "bob",
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Flatten = %v, %v; ожидали %v", got, err, want)
	}
}

func TestFlattenEmptyAndNil(t *testing.T) {
	doc := map[string]any{
		"empty": map[string]any{},
		"list":  []any{},
		"none":  nil,
		"typed": map[string]int{"a": 1},
		"deep":  map[string]any{"x": map[string]any{}},
	}
	got, err := Flatten(doc)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	for _, k := range []string{"empty", "list", "none", "typed", "deep.x"} {
		if _, ok := got[k]; !ok {
			t.Fatalf("нет ключа %q: пустые контейнеры, nil и не-any мапы — листья; получили %v", k, got)
		}
	}
	if len(got) != 5 || got["none"] != nil || !reflect.DeepEqual(got["typed"], map[string]int{"a": 1}) {
		t.Fatalf("Flatten = %v", got)
	}
}

func TestFlattenConflict(t *testing.T) {
	for range 30 { // порядок обхода мапы случаен — ошибка должна быть всегда
		doc := map[string]any{"a.b": 1, "a": map[string]any{"b": 2}, "c": 3}
		if _, err := Flatten(doc); !errors.Is(err, ErrKeyConflict) {
			t.Fatalf("конфликт путей a.b: err = %v, ожидали ErrKeyConflict", err)
		}
	}
	doc := map[string]any{"l.0": nil, "l": []any{"x"}}
	if _, err := Flatten(doc); !errors.Is(err, ErrKeyConflict) {
		t.Fatalf("конфликт l.0 (одно из значений nil): err = %v, ожидали ErrKeyConflict", err)
	}
}

func TestFlattenInputUntouched(t *testing.T) {
	inner := map[string]any{"b": 1}
	doc := map[string]any{"a": inner, "t": "top"}
	got, err := Flatten(doc)
	if err != nil || len(doc) != 2 || len(inner) != 1 || inner["b"] != 1 {
		t.Fatalf("вход изменился: %v (err %v)", doc, err)
	}
	got["a.b"] = 100
	if inner["b"] != 1 {
		t.Fatal("результат делит данные со входом")
	}
	if got, err := Flatten(nil); err != nil || len(got) != 0 {
		t.Fatalf("Flatten(nil) = %v, %v; ожидали пусто", got, err)
	}
}
