package main

type chkEmp struct {
	Dept   string
	Salary int
	Name   string
}

func chkNames(es []chkEmp) []string {
	var out []string
	for _, e := range es {
		out = append(out, e.Name)
	}
	return out
}

func TestByChain(t *testing.T) {
	es := []chkEmp{
		{"ops", 100, "Вера"}, {"dev", 200, "Олег"}, {"dev", 300, "Анна"},
		{"ops", 100, "Борис"}, {"dev", 200, "Ира"},
	}
	slices.SortFunc(es, By(
		Key(func(e chkEmp) string { return e.Dept }),
		Desc(Key(func(e chkEmp) int { return e.Salary })),
		Key(func(e chkEmp) string { return e.Name }),
	))
	want := []string{"Анна", "Ира", "Олег", "Борис", "Вера"}
	if got := chkNames(es); !reflect.DeepEqual(got, want) {
		t.Fatalf("отдел ↑, зарплата ↓, имя ↑: получили %v, ожидали %v", got, want)
	}
}

func TestByEmptyIsStable(t *testing.T) {
	es := []chkEmp{{Name: "в"}, {Name: "а"}, {Name: "б"}}
	slices.SortStableFunc(es, By[chkEmp]())
	if got := chkNames(es); !reflect.DeepEqual(got, []string{"в", "а", "б"}) {
		t.Fatalf("By() должен считать всё равным, порядок не меняется: %v", got)
	}
}

func TestDescExtremeComparator(t *testing.T) {
	// Законный компаратор: важен только знак.
	raw := func(a, b int) int {
		switch {
		case a < b:
			return math.MinInt
		case a > b:
			return math.MaxInt
		}
		return 0
	}
	if r := Desc(raw)(1, 2); r <= 0 {
		t.Fatalf("Desc(c)(1, 2) = %d, ожидали > 0: при развороте 1 идёт после 2", r)
	}
	xs := []int{3, 1, 2}
	slices.SortFunc(xs, Desc(raw))
	if !reflect.DeepEqual(xs, []int{3, 2, 1}) {
		t.Fatalf("сортировка по Desc: %v, ожидали [3 2 1]", xs)
	}
}

func TestByCopiesArgs(t *testing.T) {
	cs := []func(a, b int) int{func(a, b int) int { return cmp.Compare(a, b) }}
	c := By(cs...)
	cs[0] = func(a, b int) int { return 0 }
	if c(1, 2) >= 0 {
		t.Fatal("By должен зафиксировать компараторы при сборке: поменяли слайс снаружи — поменялось поведение")
	}
}
