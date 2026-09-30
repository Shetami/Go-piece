package main

type chkEvent struct {
	User string
	N    int
}

func chkEvents(pulled *int, es ...chkEvent) iter.Seq[chkEvent] {
	return func(yield func(chkEvent) bool) {
		for _, e := range es {
			*pulled++
			if !yield(e) {
				return
			}
		}
	}
}

func chkUser(e chkEvent) string { return e.User }

func TestRunsBasic(t *testing.T) {
	var pulled int
	seq := chkEvents(&pulled, chkEvent{"", 1}, chkEvent{"", 2}, chkEvent{"ann", 3}, chkEvent{"", 4}, chkEvent{"bob", 5}, chkEvent{"bob", 6})
	var keys []string
	var groups [][]chkEvent
	for k, g := range Runs(seq, chkUser) {
		keys = append(keys, k)
		groups = append(groups, g)
	}
	if !reflect.DeepEqual(keys, []string{"", "ann", "", "bob"}) {
		t.Fatalf("ключи групп %q, ожидали [\"\" ann \"\" bob] — пустой ключ тоже ключ, а соседство важно", keys)
	}
	var sizes []int
	for _, g := range groups {
		sizes = append(sizes, len(g))
	}
	if !reflect.DeepEqual(sizes, []int{2, 1, 1, 2}) {
		t.Fatalf("размеры групп %v, ожидали [2 1 1 2]", sizes)
	}
	if groups[0][0].N != 1 || groups[0][1].N != 2 || groups[1][0].N != 3 || groups[3][1].N != 6 {
		t.Fatalf("группы, сохранённые потребителем, испорчены: %v", groups)
	}
}

func TestRunsEmpty(t *testing.T) {
	var pulled int
	for k, g := range Runs(chkEvents(&pulled), chkUser) {
		t.Fatalf("пустой источник, а пришла группа %q %v", k, g)
	}
}

func TestRunsBreakAndLazy(t *testing.T) {
	var pulled int
	seq := chkEvents(&pulled, chkEvent{"a", 1}, chkEvent{"a", 2}, chkEvent{"b", 3}, chkEvent{"b", 4}, chkEvent{"c", 5})
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("паника после break: %v — Runs продолжает вызывать yield после false", p)
		}
	}()
	for k := range Runs(seq, chkUser) {
		if k == "a" {
			break
		}
	}
	if pulled > 3 {
		t.Fatalf("после break на первой группе из источника прочитано %d элементов, ожидали не больше 3", pulled)
	}
}

func TestRunsReusable(t *testing.T) {
	var pulled int
	runs := Runs(chkEvents(&pulled, chkEvent{"x", 1}, chkEvent{"y", 2}), chkUser)
	count := func() int {
		n := 0
		for range runs {
			n++
		}
		return n
	}
	if a, b := count(), count(); a != 2 || b != 2 {
		t.Fatalf("два обхода дали %d и %d групп, ожидали 2 и 2 — состояние должно быть своим у каждого обхода", a, b)
	}
}
