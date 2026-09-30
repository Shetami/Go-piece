package main

func TestRegistryBasic(t *testing.T) {
	r := NewRegistry()
	r.Set("db", "host", "x")
	r.Set("db", "port", "5432")
	r.Set("log", "level", "")
	if v, ok := r.Get("db", "host"); !ok || v != "x" {
		t.Fatalf("Get(db, host) = %q, %v; ожидали x, true", v, ok)
	}
	if _, ok := r.Get("log", "level"); !ok {
		t.Fatal("пустое значение — тоже значение: Get(log, level) должен вернуть ok")
	}
	if _, ok := r.Get("нет", "host"); ok {
		t.Fatal("Get из отсутствующего раздела вернул ok")
	}
}

func TestRegistryDeleteSection(t *testing.T) {
	r := NewRegistry()
	r.Set("db", "host", "x")
	r.Delete("db", "host")
	r.Delete("нет", "host")
	if s := r.Snapshot(); len(s) != 0 {
		t.Fatalf("после удаления последнего ключа раздел должен исчезнуть, снимок: %v", s)
	}
}

func TestRegistrySnapshotIsDeep(t *testing.T) {
	r := NewRegistry()
	if s := r.Snapshot(); s == nil {
		t.Fatal("снимок пустого реестра — nil, ожидали пустую мапу")
	}
	r.Set("db", "host", "x")
	s := r.Snapshot()
	if s["db"]["host"] != "x" {
		t.Fatalf("снимок = %v, ожидали db.host=x", s)
	}
	s["db"]["host"] = "ИСПОРЧЕНО"
	s["db"]["new"] = "1"
	s["other"] = map[string]string{}
	if v, _ := r.Get("db", "host"); v != "x" {
		t.Fatalf("правка раздела в снимке изменила реестр: db.host = %q", v)
	}
	if _, ok := r.Get("db", "new"); ok {
		t.Fatal("ключ, добавленный в снимок, появился в реестре")
	}
	r.Set("db", "port", "1")
	if _, ok := s["db"]["port"]; ok {
		t.Fatal("правка реестра видна в старом снимке")
	}
}

func TestRegistryConcurrent(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				sec := strconv.Itoa(i % 4)
				r.Set(sec, strconv.Itoa(g), "v")
				for _, m := range r.Snapshot() {
					_ = len(m)
				}
				r.Delete(sec, strconv.Itoa(g))
			}
		}()
	}
	wg.Wait()
	if s := r.Snapshot(); len(s) != 0 {
		t.Fatalf("после всех Set/Delete снимок = %v, ожидали пусто", s)
	}
}
