package main

func chkFilled() *MapStore {
	s := NewMapStore()
	s.SetMany(map[string]string{"a": "1", "b": "2"})
	return s
}

func TestReadOnly(t *testing.T) {
	s := chkFilled()
	ro := ReadOnly(s)
	if v, ok := ro.Get("a"); !ok || v != "1" {
		t.Fatalf("ro.Get(a) = %q, %v", v, ok)
	}
	if err := ro.Set("a", "X"); err != ErrReadOnly {
		t.Fatalf("ro.Set = %v, ожидали ErrReadOnly", err)
	}
	if err := ro.Delete("b"); err != ErrReadOnly {
		t.Fatalf("ro.Delete = %v, ожидали ErrReadOnly", err)
	}
	if err := ro.SetMany(map[string]string{"a": "X", "c": "3"}); err != ErrReadOnly {
		t.Fatalf("ro.SetMany = %v, ожидали ErrReadOnly — встроенный метод, который забыли переопределить, пишет в хранилище", err)
	}
	if got := strings.Join(s.Keys(), ","); got != "a,b" || s.m["a"] != "1" {
		t.Fatalf("после попыток записи через ReadOnly хранилище изменилось: ключи %s, a=%q", got, s.m["a"])
	}
	s.Set("z", "9")
	if got := strings.Join(ro.Keys(), ","); got != "a,b,z" {
		t.Fatalf("ro.Keys() = %s — ReadOnly должен быть видом, а не копией", got)
	}
}

func TestPrefixed(t *testing.T) {
	s := chkFilled()
	p := Prefixed(s, "user/")
	p.Set("ann", "admin")
	kv := map[string]string{"bob": "dev", "cat": "qa"}
	p.SetMany(kv)
	if _, ok := kv["user/bob"]; ok || len(kv) != 2 {
		t.Fatalf("SetMany изменил переданную мапу: %v", kv)
	}
	if got := strings.Join(s.Keys(), ","); got != "a,b,user/ann,user/bob,user/cat" {
		t.Fatalf("ключи в хранилище: %s", got)
	}
	if got := strings.Join(p.Keys(), ","); got != "ann,bob,cat" {
		t.Fatalf("p.Keys() = %s, ожидали ann,bob,cat — только свои и без префикса", got)
	}
	if v, ok := p.Get("bob"); !ok || v != "dev" {
		t.Fatalf("p.Get(bob) = %q, %v", v, ok)
	}
	if _, ok := p.Get("a"); ok {
		t.Fatal("p.Get(a) нашёл ключ без префикса")
	}
	p.Delete("ann")
	if _, ok := s.Get("user/ann"); ok {
		t.Fatal("p.Delete(ann) не удалил user/ann")
	}
}

func TestComposition(t *testing.T) {
	s := NewMapStore()
	inner := Prefixed(s, "a/")
	outer := Prefixed(inner, "b/")
	outer.Set("x", "1")
	outer.SetMany(map[string]string{"y": "2"})
	inner.Set("other", "3")
	if got := strings.Join(s.Keys(), ","); got != "a/b/x,a/b/y,a/other" {
		t.Fatalf("вложенные префиксы: ключи %s, ожидали a/b/x,a/b/y,a/other", got)
	}
	if got := strings.Join(outer.Keys(), ","); got != "x,y" {
		t.Fatalf("outer.Keys() = %s, ожидали x,y", got)
	}
	ro := ReadOnly(outer)
	if err := ro.SetMany(map[string]string{"z": "1"}); err != ErrReadOnly {
		t.Fatalf("ReadOnly(Prefixed).SetMany = %v", err)
	}
	if v, _ := ro.Get("x"); v != "1" || len(s.Keys()) != 3 {
		t.Fatalf("ReadOnly(Prefixed): Get(x) = %q, ключей %d", v, len(s.Keys()))
	}
}
