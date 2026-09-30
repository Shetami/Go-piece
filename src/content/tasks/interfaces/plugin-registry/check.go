package main

// chkPlug — плагин с журналом; умеет Start и Close.
type chkPlug struct {
	name     string
	log      *[]string
	startErr error
	closeErr error
}

func (p *chkPlug) Name() string { return p.name }
func (p *chkPlug) Start(ctx context.Context) error {
	*p.log = append(*p.log, "start "+p.name)
	return p.startErr
}
func (p *chkPlug) Close() error { *p.log = append(*p.log, "close "+p.name); return p.closeErr }

// chkBare — плагин без опциональных методов.
type chkBare struct{ name string }

func (b chkBare) Name() string { return b.name }

func chkReg(log *[]string, calls *int) *Registry {
	r := NewRegistry()
	mk := func(name string) Factory {
		return func(cfg map[string]string) (Plugin, error) {
			*calls++
			p := &chkPlug{name: name, log: log}
			if cfg["start"] == "fail" {
				p.startErr = errors.New("port busy")
			}
			if cfg["close"] == "fail" {
				p.closeErr = errors.New("flush failed")
			}
			if cfg["create"] == "fail" {
				return nil, errors.New("bad config")
			}
			return p, nil
		}
	}
	for _, n := range []string{"db", "cache", "http"} {
		r.Register(n, mk(n))
	}
	r.Register("bare", func(map[string]string) (Plugin, error) { *calls++; return chkBare{"bare"}, nil })
	return r
}

func TestRegistryRegister(t *testing.T) {
	var log []string
	var calls int
	r := chkReg(&log, &calls)
	if got := strings.Join(r.Names(), ","); got != "bare,cache,db,http" {
		t.Fatalf("Names = %s, ожидали bare,cache,db,http", got)
	}
	err := r.Register("db", func(map[string]string) (Plugin, error) { return nil, nil })
	if !errors.Is(err, ErrDuplicate) || !strings.Contains(err.Error(), "db") {
		t.Fatalf("повторная регистрация: %v; ожидали ErrDuplicate с именем", err)
	}
	if r.Register("", func(map[string]string) (Plugin, error) { return nil, nil }) == nil || r.Register("x", nil) == nil {
		t.Fatal("пустое имя и nil-фабрика должны давать ошибку")
	}
}

func TestRegistryLoadOK(t *testing.T) {
	var log []string
	var calls int
	r := chkReg(&log, &calls)
	set, err := r.Load(context.Background(), []string{"db", "bare", "http"}, nil)
	if err != nil || len(set.Plugins()) != 3 || set.Plugins()[1].Name() != "bare" {
		t.Fatalf("Load = %v, %v", set, err)
	}
	if err := set.Close(); err != nil {
		t.Fatalf("Close = %v", err)
	}
	set.Close()
	if got := strings.Join(log, "; "); got != "start db; start http; close http; close db" {
		t.Fatalf("журнал: %q; ожидали запуск по порядку, закрытие в обратном и один раз", got)
	}
}

func TestRegistryUnknown(t *testing.T) {
	var log []string
	var calls int
	r := chkReg(&log, &calls)
	_, err := r.Load(context.Background(), []string{"db", "nope"}, nil)
	if !errors.Is(err, ErrUnknown) || !strings.Contains(err.Error(), "nope") || calls != 0 {
		t.Fatalf("неизвестный плагин: %v, вызвано фабрик %d; ожидали ErrUnknown с именем и ни одной фабрики", err, calls)
	}
}

func TestRegistryRollback(t *testing.T) {
	var log []string
	var calls int
	r := chkReg(&log, &calls)
	cfg := map[string]map[string]string{"db": {"close": "fail"}, "http": {"start": "fail"}}
	set, err := r.Load(context.Background(), []string{"db", "cache", "http", "bare"}, cfg)
	if set != nil || err == nil || !strings.Contains(err.Error(), "port busy") || !strings.Contains(err.Error(), "flush failed") {
		t.Fatalf("Load с падением на Start: %v, %v; ожидали nil и ошибку с обеими причинами", set, err)
	}
	if got := strings.Join(log, "; "); got != "start db; start cache; start http; close http; close cache; close db" {
		t.Fatalf("журнал: %q; упавший тоже закрывается, всё в обратном порядке, дальше не идём", got)
	}
	log = nil
	_, err = r.Load(context.Background(), []string{"db", "cache"}, map[string]map[string]string{"cache": {"create": "fail"}})
	if err == nil || strings.Join(log, "; ") != "start db; close db" {
		t.Fatalf("ошибка фабрики: %v, журнал %q; ожидали откат уже созданных", err, log)
	}
}

func TestRegistryCanceled(t *testing.T) {
	var log []string
	var calls int
	r := chkReg(&log, &calls)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := r.Load(ctx, []string{"bare", "db"}, nil)
	if !errors.Is(err, context.Canceled) || strings.Join(log, "; ") != "close db" {
		t.Fatalf("отменённый контекст: %v, журнал %q; ожидали context.Canceled, Start не вызывается, созданный закрыт", err, log)
	}
}

func TestRegistryConcurrent(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Register(fmt.Sprint("p", i), func(map[string]string) (Plugin, error) { return chkBare{"x"}, nil })
			r.Names()
		}()
	}
	wg.Wait()
	if len(r.Names()) != 20 {
		t.Fatalf("после 20 параллельных Register имён %d", len(r.Names()))
	}
}
