package main

// chkTags — несравнимый тип (слайс) с Keyer: порядок тегов не важен.
type chkTags []string

func (t chkTags) DedupKey() string {
	s := slices.Clone([]string(t))
	slices.Sort(s)
	return strings.Join(s, ",")
}

type chkUserID int

func (u chkUserID) DedupKey() string { return "user" }

type chkBox struct{ V any }

func chkDedup(t *testing.T, in []any) (out []any) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Dedupe(%#v) паникует: %v", in, r)
		}
	}()
	return Dedupe(in)
}

func TestDedupeComparable(t *testing.T) {
	in := []any{1, "a", 1, int64(1), "a", true, 1.5, true}
	got := chkDedup(t, in)
	if want := []any{1, "a", int64(1), true, 1.5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Dedupe = %#v, ожидали %#v (int и int64 — разные)", got, want)
	}
	if !reflect.DeepEqual(in, []any{1, "a", 1, int64(1), "a", true, 1.5, true}) {
		t.Fatal("Dedupe не должен менять входной слайс")
	}
}

func TestDedupeUncomparable(t *testing.T) {
	in := []any{[]int{1, 2}, map[string]int{"a": 1}, []int{1, 2}, []int{2, 1}, map[string]int{"a": 1}, []int(nil)}
	got := chkDedup(t, in)
	if len(got) != 4 || !reflect.DeepEqual(got[2], []int{2, 1}) {
		t.Fatalf("слайсы и мапы: %#v; ожидали 4 значения — равные по содержимому склеиваются", got)
	}
	in = []any{chkBox{1}, chkBox{[]int{1}}, chkBox{1}, chkBox{[]int{1}}, [1]any{[]int{1}}, [1]any{[]int{1}}}
	got = chkDedup(t, in)
	if len(got) != 3 {
		t.Fatalf("структуры и массивы со слайсом в поле any: %#v; ожидали 3 значения без паники", got)
	}
}

func TestDedupeNaNAndNil(t *testing.T) {
	var p *int
	var q *string
	in := []any{math.NaN(), nil, math.NaN(), p, nil, float32(math.NaN()), p, q, float32(math.NaN()), 0.0}
	got := chkDedup(t, in)
	if len(got) != 6 {
		t.Fatalf("NaN и nil: %#v; ожидали 6 значений: NaN64, nil, (*int)(nil), NaN32, (*string)(nil), 0.0", got)
	}
	if got[1] != nil || got[2] == nil {
		t.Fatalf("nil и типизированный nil-указатель должны остаться разными: %#v", got)
	}
}

func TestDedupeKeyer(t *testing.T) {
	in := []any{chkTags{"go", "db"}, chkTags{"db", "go"}, chkTags{"go"}, chkUserID(1), chkUserID(2), "user"}
	got := chkDedup(t, in)
	if len(got) != 4 || !reflect.DeepEqual(got[0], chkTags{"go", "db"}) || got[2] != any(chkUserID(1)) || got[3] != "user" {
		t.Fatalf("Keyer: %#v; ожидали [go db] [go] chkUserID(1) \"user\" — ключ сравнивается в пределах типа", got)
	}
}
