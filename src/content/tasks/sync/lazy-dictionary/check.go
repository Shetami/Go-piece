package main

func TestDirectoryLazyAndOnce(t *testing.T) {
	var calls atomic.Int32
	d := NewDirectory(func() (map[string]string, error) {
		calls.Add(1)
		time.Sleep(10 * time.Millisecond)
		return map[string]string{"RU": "Россия", "KZ": "Казахстан"}, nil
	})
	if c := calls.Load(); c != 0 {
		t.Fatalf("load вызван %d раз ещё до первого Lookup — справочник должен грузиться лениво", c)
	}
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if name, err := d.Lookup("KZ"); err != nil || name != "Казахстан" {
				t.Errorf("Lookup(KZ) = %q, %v, ожидали Казахстан", name, err)
			}
		}()
	}
	wg.Wait()
	if c := calls.Load(); c != 1 {
		t.Fatalf("50 одновременных Lookup вызвали load %d раз, ожидали 1", c)
	}
	if _, err := d.Lookup("XX"); !errors.Is(err, ErrUnknownCode) {
		t.Fatalf("Lookup(XX) вернул %v, ожидали ErrUnknownCode", err)
	}
}

func TestDirectoryErrorRemembered(t *testing.T) {
	boom := errors.New("база недоступна")
	calls := 0
	d := NewDirectory(func() (map[string]string, error) { calls++; return nil, boom })
	for range 3 {
		if _, err := d.Lookup("RU"); !errors.Is(err, boom) {
			t.Fatalf("Lookup при упавшей загрузке вернул %v, ожидали ошибку загрузки", err)
		}
	}
	if calls != 1 {
		t.Fatalf("load вызван %d раз, ожидали 1: ошибка запоминается, как в sync.Once", calls)
	}
}

func TestDirectoryOwnCopy(t *testing.T) {
	src := map[string]string{"RU": "Россия"}
	d := NewDirectory(func() (map[string]string, error) { return src, nil })
	d.Lookup("RU")
	src["RU"] = "испорчено"
	src["US"] = "США"
	if name, _ := d.Lookup("RU"); name != "Россия" {
		t.Fatalf("загрузчик поменял свою мапу — и справочник вернул %q: нужна своя копия", name)
	}
	if _, err := d.Lookup("US"); !errors.Is(err, ErrUnknownCode) {
		t.Fatalf("ключ, добавленный в мапу загрузчика после загрузки, виден в справочнике")
	}
}
