package main

func TestBuildOrderBasic(t *testing.T) {
	deps := map[string][]string{
		"app": {"db", "log"},
		"db":  {"log"},
		"api": {"app"},
		"cli": {},
	}
	for range 10 {
		got, err := BuildOrder(deps)
		want := []string{"cli", "log", "db", "app", "api"}
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("BuildOrder = %v, %v; ожидали %v, nil", got, err, want)
		}
	}
}

func TestBuildOrderImplicitAndDup(t *testing.T) {
	got, err := BuildOrder(map[string][]string{"z": {"b", "b", "a"}, "a": {"b"}})
	want := []string{"b", "a", "z"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildOrder = %v, %v; ожидали %v — b есть только в списках зависимостей, а повтор b — одна зависимость", got, err, want)
	}
	got, err = BuildOrder(nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("пустой граф: %v, %v; ожидали пусто и nil", got, err)
	}
}

func TestBuildOrderAlphabeticalAmongReady(t *testing.T) {
	// Жадно по алфавиту среди готовых, а не «сначала все листья».
	got, _ := BuildOrder(map[string][]string{"b": {"a"}, "c": {}, "d": {}})
	want := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildOrder = %v, ожидали %v", got, want)
	}
}

func TestBuildOrderCycle(t *testing.T) {
	deps := map[string][]string{
		"a": {"b"}, "b": {"c"}, "c": {"a"},
		"d": {"c"},   // зависит от цикла
		"e": {"log"}, // собирается нормально
	}
	got, err := BuildOrder(deps)
	if got != nil {
		t.Fatalf("при цикле ожидали nil вместо порядка, получили %v", got)
	}
	if !errors.Is(err, ErrCycle) {
		t.Fatalf("errors.Is(err, ErrCycle) = false, err = %v", err)
	}
	var ce *CycleError
	if !errors.As(err, &ce) || !reflect.DeepEqual(ce.Nodes, []string{"a", "b", "c", "d"}) {
		t.Fatalf("ожидали *CycleError с Nodes [a b c d], получили %v", err)
	}
}

func TestBuildOrderSelfLoop(t *testing.T) {
	_, err := BuildOrder(map[string][]string{"a": {"a"}, "b": {}})
	var ce *CycleError
	if !errors.As(err, &ce) || !reflect.DeepEqual(ce.Nodes, []string{"a"}) {
		t.Fatalf("петля a → a: ожидали *CycleError с Nodes [a], получили %v", err)
	}
}
