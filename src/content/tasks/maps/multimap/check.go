package main

func chkVals(t *testing.T, what string, got, want []string) {
	t.Helper()
	if len(got) != len(want) || (len(got) > 0 && !reflect.DeepEqual(got, want)) {
		t.Fatalf("%s = %q, ожидали %q", what, got, want)
	}
}

func TestMultiMapAddOrder(t *testing.T) {
	m := NewMultiMap[string, string]()
	for _, v := range []string{"go", "sql", "go", "k8s"} {
		m.Add("anna", v)
	}
	m.Add("boris", "go")
	chkVals(t, "Get(anna)", m.Get("anna"), []string{"go", "sql", "k8s"})
	if m.Add("anna", "sql") {
		t.Fatal("Add существующей пары вернул true")
	}
	if m.Len() != 2 || m.Size() != 4 {
		t.Fatalf("Len=%d Size=%d, ожидали 2 и 4", m.Len(), m.Size())
	}
	chkVals(t, "Get(нет)", m.Get("нет"), nil)
}

func TestMultiMapRemoveEmptiesKey(t *testing.T) {
	m := NewMultiMap[string, string]()
	m.Add("a", "1")
	m.Add("a", "2")
	if !m.Remove("a", "1") || m.Remove("a", "1") || m.Remove("нет", "1") {
		t.Fatal("Remove: ожидали true для существующей пары и false для отсутствующей")
	}
	chkVals(t, "Get(a) после Remove", m.Get("a"), []string{"2"})
	m.Remove("a", "2")
	if m.Has("a") || m.Len() != 0 || m.Size() != 0 {
		t.Fatalf("после удаления последнего значения: Has=%v Len=%d Size=%d; ключ должен исчезнуть", m.Has("a"), m.Len(), m.Size())
	}
	if !m.Add("a", "2") {
		t.Fatal("после удаления пару можно добавить снова")
	}
	chkVals(t, "Get(a)", m.Get("a"), []string{"2"})
}

func TestMultiMapGetIsCopy(t *testing.T) {
	m := NewMultiMap[string, string]()
	m.Add("a", "x")
	m.Add("a", "y")
	m.Add("a", "z")
	m.Remove("a", "z") // во внутреннем слайсе остаётся запас ёмкости
	got := m.Get("a")
	got[0] = "ИСПОРЧЕНО"
	_ = append(got, "ЛИШНЕЕ")
	chkVals(t, "Get(a) после правки результата", m.Get("a"), []string{"x", "y"})
	m.Add("a", "w")
	chkVals(t, "Get(a)", m.Get("a"), []string{"x", "y", "w"})
}

func TestMultiMapRemoveAll(t *testing.T) {
	m := NewMultiMap[int, int]()
	for i := range 5 {
		m.Add(i%2, i)
	}
	chkInts := m.RemoveAll(0)
	if !reflect.DeepEqual(chkInts, []int{0, 2, 4}) {
		t.Fatalf("RemoveAll(0) = %v, ожидали [0 2 4]", chkInts)
	}
	if m.Has(0) || m.Len() != 1 || m.Size() != 2 {
		t.Fatalf("после RemoveAll: Has(0)=%v Len=%d Size=%d; ожидали false, 1, 2", m.Has(0), m.Len(), m.Size())
	}
	if got := m.RemoveAll(42); len(got) != 0 {
		t.Fatalf("RemoveAll отсутствующего = %v, ожидали пусто", got)
	}
}
