package main

var chkTemp = errors.New("timeout")

// chkStore — фейковое хранилище: считает обращения, может вернуть ошибку.
type chkStore struct {
	mu      sync.Mutex
	m       map[string][]byte
	gets    int
	failGet int   // сколько ближайших Get вернут chkTemp
	failPut error // ошибка для Put
}

func chkNew() *chkStore { return &chkStore{m: map[string][]byte{}} }

func (s *chkStore) Get(_ context.Context, k string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gets++
	if s.failGet > 0 {
		s.failGet--
		return nil, chkTemp
	}
	v, ok := s.m[k]
	if !ok {
		return nil, fmt.Errorf("key %q: %w", k, ErrNotFound)
	}
	return v, nil // отдаём внутренний срез — кэш обязан копировать
}

func (s *chkStore) Put(_ context.Context, k string, v []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failPut != nil {
		return s.failPut
	}
	s.m[k] = append([]byte{}, v...)
	return nil
}

func (s *chkStore) Delete(_ context.Context, k string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, k)
	return nil
}

func TestCacheHitMiss(t *testing.T) {
	var _ Storage = (*Cache)(nil)
	ctx := context.Background()
	s := chkNew()
	s.m["a"] = []byte("1")
	c := Cached(s)
	for range 3 {
		if v, err := c.Get(ctx, "a"); err != nil || string(v) != "1" {
			t.Fatalf("Get(a) = %q, %v", v, err)
		}
	}
	if h, m := c.Stats(); h != 2 || m != 1 || s.gets != 1 {
		t.Fatalf("три Get: hits=%d misses=%d обращений к хранилищу=%d; ожидали 2, 1, 1", h, m, s.gets)
	}
}

func TestCacheNegative(t *testing.T) {
	ctx := context.Background()
	s := chkNew()
	c := Cached(s)
	_, err1 := c.Get(ctx, "nope")
	_, err2 := c.Get(ctx, "nope")
	if !errors.Is(err1, ErrNotFound) || !errors.Is(err2, ErrNotFound) || s.gets != 1 {
		t.Fatalf("отсутствующий ключ: %v, %v, обращений %d; ожидали ErrNotFound дважды и одно обращение", err1, err2, s.gets)
	}
	c.Put(ctx, "nope", []byte{})
	v, err := c.Get(ctx, "nope")
	if err != nil || v == nil || len(v) != 0 || s.gets != 1 {
		t.Fatalf("после Put пустого значения Get = %v, %v (обращений %d); ожидали пустой не-nil срез из кэша", v, err, s.gets)
	}
	c.Delete(ctx, "nope")
	if _, err := c.Get(ctx, "nope"); !errors.Is(err, ErrNotFound) || s.gets != 1 {
		t.Fatalf("после Delete Get = %v (обращений %d); ожидали ErrNotFound из кэша", err, s.gets)
	}
}

func TestCacheTransientNotCached(t *testing.T) {
	ctx := context.Background()
	s := chkNew()
	s.m["k"] = []byte("v")
	s.failGet = 1
	c := Cached(s)
	if _, err := c.Get(ctx, "k"); err != chkTemp {
		t.Fatalf("первый Get = %v, ожидали временную ошибку как есть", err)
	}
	if v, err := c.Get(ctx, "k"); err != nil || string(v) != "v" {
		t.Fatalf("второй Get = %q, %v; временная ошибка не должна кэшироваться", v, err)
	}
}

func TestCacheCopies(t *testing.T) {
	ctx := context.Background()
	s := chkNew()
	s.m["x"] = []byte("abc")
	c := Cached(s)
	v, _ := c.Get(ctx, "x")
	v[0] = 'Z'
	if w, _ := c.Get(ctx, "x"); string(w) != "abc" {
		t.Fatalf("изменили результат Get — кэш вернул %q, ожидали \"abc\"", w)
	}
	if string(s.m["x"]) != "abc" {
		t.Fatalf("изменили результат Get — испортили само хранилище: %q", s.m["x"])
	}
	buf := []byte("new")
	c.Put(ctx, "y", buf)
	buf[0] = 'X'
	if w, _ := c.Get(ctx, "y"); string(w) != "new" {
		t.Fatalf("изменили срез после Put — кэш вернул %q, ожидали \"new\"", w)
	}
}

func TestCachePutFailure(t *testing.T) {
	ctx := context.Background()
	s := chkNew()
	s.m["k"] = []byte("old")
	c := Cached(s)
	c.Get(ctx, "k")
	s.failPut = errors.New("write failed")
	if err := c.Put(ctx, "k", []byte("new")); err == nil {
		t.Fatal("ошибка Put должна вернуться вызывающему")
	}
	s.failPut = nil
	s.m["k"] = []byte("fresh")
	if v, _ := c.Get(ctx, "k"); string(v) != "fresh" {
		t.Fatalf("после неудачного Put Get = %q; запись кэша должна быть сброшена, ожидали \"fresh\" из хранилища", v)
	}
}

func TestCacheConcurrent(t *testing.T) {
	ctx := context.Background()
	s := chkNew()
	c := Cached(s)
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			k := strconv.Itoa(i % 3)
			for range 50 {
				c.Put(ctx, k, []byte(k))
				c.Get(ctx, k)
				c.Stats()
			}
		}()
	}
	wg.Wait()
	if h, m := c.Stats(); h+m != 400 {
		t.Fatalf("hits+misses = %d, ожидали 400 — счётчики теряются без синхронизации", h+m)
	}
}
