package main

type chkMigration int

func TestTopoDeterministic(t *testing.T) {
	deps := map[string][]string{
		"api":    {"db", "cache", "config"},
		"worker": {"db", "queue"},
		"db":     {"config"},
		"cache":  {"config"},
		"queue":  {},
	}
	want := []string{"config", "cache", "db", "api", "queue", "worker"}
	for range 20 {
		got, err := TopoSort(deps)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("TopoSort = %q, %v; ожидали %q (из готовых — наименьший)", got, err, want)
		}
	}
}

func TestTopoImplicitAndDuplicates(t *testing.T) {
	deps := map[chkMigration][]chkMigration{
		30: {10, 20, 10},
		20: {10},
		40: {5},
	}
	got, err := TopoSort(deps)
	if err != nil || !reflect.DeepEqual(got, []chkMigration{5, 10, 20, 30, 40}) {
		t.Fatalf("TopoSort = %v, %v; ожидали [5 10 20 30 40] — 5 и 10 есть только в зависимостях, 10 у 30 указан дважды", got, err)
	}
	if got, err := TopoSort(map[int][]int{}); err != nil || len(got) != 0 {
		t.Fatalf("пустой граф: %v, %v", got, err)
	}
}

func TestTopoCycle(t *testing.T) {
	deps := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
		"d": {"c"},
		"e": {},
		"f": {"e"},
	}
	got, err := TopoSort(deps)
	if got != nil || !errors.Is(err, ErrCycle) {
		t.Fatalf("при цикле ожидали nil и ErrCycle, получили %v, %v", got, err)
	}
	if want := "цикл в зависимостях: [a b c d]"; err.Error() != want {
		t.Fatalf("текст ошибки %q, ожидали %q", err.Error(), want)
	}
	_, err = TopoSort(map[int][]int{1: {1}, 2: {}})
	if !errors.Is(err, ErrCycle) || err.Error() != "цикл в зависимостях: [1]" {
		t.Fatalf("зависимость от самого себя: %v", err)
	}
}
