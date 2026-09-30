package main

func chkBase() map[string]any {
	return map[string]any{
		"db":    map[string]any{"host": "localhost", "port": 5432, "tls": map[string]any{"on": false, "ca": "x"}},
		"tags":  []any{"a", "b"},
		"level": "info",
		"keep":  map[string]any{"deep": []any{map[string]any{"n": 1}}},
	}
}

func TestMergeRecursive(t *testing.T) {
	over := map[string]any{
		"db":    map[string]any{"host": "prod", "tls": map[string]any{"on": true}},
		"tags":  []any{"c"},
		"level": map[string]any{"root": "warn"}, // скаляр заменяется мапой
		"new":   1,
	}
	got := Merge(chkBase(), over)
	want := map[string]any{
		"db":    map[string]any{"host": "prod", "port": 5432, "tls": map[string]any{"on": true, "ca": "x"}},
		"tags":  []any{"c"},
		"level": map[string]any{"root": "warn"},
		"keep":  map[string]any{"deep": []any{map[string]any{"n": 1}}},
		"new":   1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Merge =\n  %v\nожидали\n  %v", got, want)
	}
}

func TestMergeNilDeletes(t *testing.T) {
	got := Merge(chkBase(), map[string]any{"level": nil, "db": map[string]any{"tls": map[string]any{"ca": nil}}, "ghost": nil})
	if _, ok := got["level"]; ok {
		t.Fatal("nil в override должен удалить ключ level")
	}
	if _, ok := got["ghost"]; ok {
		t.Fatal("nil для отсутствующего ключа не должен создавать ключ")
	}
	tls := got["db"].(map[string]any)["tls"].(map[string]any)
	if _, ok := tls["ca"]; ok || tls["on"] != false {
		t.Fatalf("db.tls = %v, ожидали только on=false", tls)
	}
	if got := Merge(nil, nil); got == nil || len(got) != 0 {
		t.Fatalf("Merge(nil, nil) = %#v, ожидали пустую непустую мапу", got)
	}
}

func TestMergeNoAliasing(t *testing.T) {
	base := chkBase()
	over := map[string]any{"db": map[string]any{"port": 6432}, "list": []any{map[string]any{"x": 1}}}
	got := Merge(base, over)

	got["db"].(map[string]any)["host"] = "ИСПОРЧЕНО"
	got["db"].(map[string]any)["tls"].(map[string]any)["ca"] = "ИСПОРЧЕНО"
	got["tags"].([]any)[0] = "ИСПОРЧЕНО"
	got["keep"].(map[string]any)["deep"].([]any)[0].(map[string]any)["n"] = 100
	got["list"].([]any)[0].(map[string]any)["x"] = 100

	if !reflect.DeepEqual(base, chkBase()) {
		t.Fatalf("правка результата изменила base: %v", base)
	}
	if over["list"].([]any)[0].(map[string]any)["x"] != 1 {
		t.Fatal("правка результата изменила override")
	}
	if over["db"].(map[string]any)["host"] != nil {
		t.Fatal("в мапу override дописали ключи из base")
	}
}

func TestMergeInputsUntouched(t *testing.T) {
	base := chkBase()
	over := map[string]any{"db": map[string]any{"host": nil}, "tags": nil}
	Merge(base, over)
	if !reflect.DeepEqual(base, chkBase()) {
		t.Fatalf("Merge изменил base: %v", base)
	}
	if len(over) != 2 || len(over["db"].(map[string]any)) != 1 {
		t.Fatalf("Merge изменил override: %v", over)
	}
}
