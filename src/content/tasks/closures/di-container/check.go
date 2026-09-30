package main

// chkNode — объект, который помнит, из чего собран.
type chkNode struct {
	Name string
	Deps []*chkNode
}

// chkDeps регистрирует фабрику name, зависящую от deps, и считает её вызовы.
func chkDeps(c *Container, calls map[string]int, name string, deps ...string) {
	c.Provide(name, func(get func(string) (any, error)) (any, error) {
		calls[name]++
		n := &chkNode{Name: name}
		for _, d := range deps {
			v, err := get(d)
			if err != nil {
				return nil, err
			}
			n.Deps = append(n.Deps, v.(*chkNode))
		}
		return n, nil
	})
}

func TestContainerDiamond(t *testing.T) {
	c := NewContainer()
	calls := map[string]int{}
	chkDeps(c, calls, "app", "users", "orders")
	chkDeps(c, calls, "users", "db")
	chkDeps(c, calls, "orders", "db", "users")
	chkDeps(c, calls, "db")
	v, err := c.Get("app")
	if err != nil {
		t.Fatalf("ромб зависимостей — не цикл, а получили ошибку: %v", err)
	}
	app, ok := v.(*chkNode)
	if !ok || len(app.Deps) != 2 || len(app.Deps[0].Deps) != 1 || len(app.Deps[1].Deps) != 2 {
		t.Fatalf("Get(app) вернул %#v, ожидали собранный app с двумя зависимостями", v)
	}
	if app.Deps[0].Deps[0] != app.Deps[1].Deps[0] || app.Deps[1].Deps[1] != app.Deps[0] {
		t.Fatal("db и users должны быть одними и теми же объектами у всех, кто от них зависит")
	}
	c.Get("app")
	c.Get("db")
	for name, n := range calls {
		if n != 1 {
			t.Fatalf("фабрика %s вызвана %d раз, ожидали 1", name, n)
		}
	}
}

func TestContainerCycle(t *testing.T) {
	c := NewContainer()
	calls := map[string]int{}
	chkDeps(c, calls, "api", "auth")
	chkDeps(c, calls, "auth", "session")
	chkDeps(c, calls, "session", "auth")
	chkDeps(c, calls, "self", "self")
	_, err := c.Get("api")
	if !errors.Is(err, ErrCycle) {
		t.Fatalf("auth -> session -> auth: ожидали ErrCycle, получили %v", err)
	}
	if !strings.Contains(err.Error(), "auth -> session -> auth") {
		t.Fatalf("в ошибке нет пути цикла \"auth -> session -> auth\": %v", err)
	}
	if _, err := c.Get("self"); !errors.Is(err, ErrCycle) || !strings.Contains(err.Error(), "self -> self") {
		t.Fatalf("зависимость от самого себя: %v", err)
	}
}

func TestContainerUnknownAndRetry(t *testing.T) {
	c := NewContainer()
	calls := map[string]int{}
	chkDeps(c, calls, "svc", "cache")
	_, err := c.Get("svc")
	if !errors.Is(err, ErrUnknown) || !strings.Contains(err.Error(), "cache") {
		t.Fatalf("нет фабрики cache: ожидали ErrUnknown с именем, получили %v", err)
	}
	chkDeps(c, calls, "cache")
	if _, err := c.Get("svc"); err != nil {
		t.Fatalf("после регистрации cache svc должен собраться — ошибки не кэшируются: %v", err)
	}
}

func TestContainerFlakyFactory(t *testing.T) {
	c := NewContainer()
	attempts := 0
	c.Provide("conn", func(get func(string) (any, error)) (any, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("connection refused")
		}
		return attempts, nil
	})
	if _, err := c.Get("conn"); err == nil {
		t.Fatal("первая попытка должна вернуть ошибку фабрики")
	}
	v, err := c.Get("conn")
	if err != nil {
		t.Fatalf("вторая попытка: %v — после ошибки имя не должно считаться «строящимся»", err)
	}
	if v != 2 {
		t.Fatalf("Get(conn) = %v, ожидали 2", v)
	}
}
