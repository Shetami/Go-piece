package main

type chkUserID int

func TestBiMapBasic(t *testing.T) {
	m := NewBiMap[chkUserID, string]()
	m.Put(1, "neo@matrix.io")
	m.Put(2, "trinity@matrix.io")
	m.Put(1, "neo@matrix.io")
	if v, ok := m.Get(1); !ok || v != "neo@matrix.io" || m.Len() != 2 {
		t.Fatalf("Get(1) = (%q, %v), Len = %d", v, ok, m.Len())
	}
	if k, ok := m.GetKey("trinity@matrix.io"); !ok || k != 2 {
		t.Fatalf("GetKey(trinity) = (%v, %v)", k, ok)
	}
	if _, ok := m.GetKey("smith@matrix.io"); ok {
		t.Fatalf("GetKey нашёл то, чего нет")
	}
}

func TestBiMapReassign(t *testing.T) {
	m := NewBiMap[chkUserID, string]()
	m.Put(1, "old@x.io")
	m.Put(1, "new@x.io") // пользователь сменил почту
	if _, ok := m.GetKey("old@x.io"); ok {
		t.Fatalf("после смены почты старый адрес всё ещё ведёт к пользователю 1")
	}
	m.Put(2, "new@x.io") // почту забрал другой пользователь
	if _, ok := m.Get(1); ok {
		t.Fatalf("адрес перешёл пользователю 2, а у 1 он остался")
	}
	if k, _ := m.GetKey("new@x.io"); k != 2 || m.Len() != 1 {
		t.Fatalf("GetKey(new) = %v, Len = %d; ожидали 2 и одну пару", k, m.Len())
	}
}

func TestBiMapCrossReassign(t *testing.T) {
	m := NewBiMap[string, string]()
	m.Put("a", "x")
	m.Put("b", "y")
	m.Put("a", "y") // разрывает и a→x, и b→y
	if m.Len() != 1 {
		t.Fatalf("после Put(a, y) Len = %d, ожидали 1", m.Len())
	}
	if _, ok := m.Get("b"); ok {
		t.Fatalf("b всё ещё связан, хотя его значение y отдали a")
	}
	if _, ok := m.GetKey("x"); ok {
		t.Fatalf("x всё ещё связан, хотя у a теперь y")
	}
	if v, _ := m.Get("a"); v != "y" {
		t.Fatalf("Get(a) = %q", v)
	}
	m.Put("s", "s")
	if !m.Delete("s") || m.Delete("s") {
		t.Fatalf("Delete: первый раз true, второй false")
	}
	if _, ok := m.GetKey("s"); ok || m.Len() != 1 {
		t.Fatalf("после Delete значение осталось в обратной стороне")
	}
}

func TestBiMapInverseView(t *testing.T) {
	m := NewBiMap[string, int]()
	m.Put("ru", 7)
	inv := m.Inverse()
	if k, ok := inv.Get(7); !ok || k != "ru" {
		t.Fatalf("Inverse().Get(7) = (%q, %v)", k, ok)
	}
	inv.Put(1, "us")
	if v, ok := m.Get("us"); !ok || v != 1 {
		t.Fatalf("Put через Inverse не виден в исходной: (%v, %v)", v, ok)
	}
	m.Put("kz", 7)
	if k, _ := inv.Get(7); k != "kz" || inv.Len() != 2 {
		t.Fatalf("изменение исходной не видно через Inverse: Get(7) = %q, Len = %d", k, inv.Len())
	}
}
