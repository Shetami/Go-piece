package main

func TestCodeIs(t *testing.T) {
	c := Code("db.conn.timeout")
	cases := []struct {
		target Code
		want   bool
	}{{"db", true}, {"db.conn", true}, {"db.conn.timeout", true}, {"d", false}, {"db.co", false}, {"db.conn.timeout.x", false}, {"api", false}, {"", false}}
	for _, tc := range cases {
		if got := errors.Is(c, tc.target); got != tc.want {
			t.Fatalf("errors.Is(Code(%q), Code(%q)) = %v, ожидали %v", c, tc.target, got, tc.want)
		}
	}
	wrapped := fmt.Errorf("пул: %w", Code("db.conn"))
	if !errors.Is(wrapped, Code("db")) {
		t.Fatal("обёрнутый Code(\"db.conn\") должен отвечать на errors.Is(err, Code(\"db\"))")
	}
}

func TestEError(t *testing.T) {
	e := &E{Code: "db.conn", Msg: "не подключились", Err: io.EOF}
	if e.Error() != "db.conn: не подключились: EOF" {
		t.Fatalf("Error() = %q", e.Error())
	}
	if (&E{Code: "auth", Msg: "нет токена"}).Error() != "auth: нет токена" {
		t.Fatalf("без причины: %q", (&E{Code: "auth", Msg: "нет токена"}).Error())
	}
}

func TestEChain(t *testing.T) {
	inner := &E{Code: "db.conn.timeout", Msg: "5s", Err: context.DeadlineExceeded}
	outer := fmt.Errorf("handler: %w", &E{Code: "order.save", Msg: "заказ 7", Err: inner})
	for _, c := range []Code{"order", "order.save", "db", "db.conn.timeout"} {
		if !errors.Is(outer, c) {
			t.Fatalf("errors.Is(err, %q) = false, ожидали true: коды есть на разных уровнях цепочки", c)
		}
	}
	if errors.Is(outer, Code("order.save.x")) || errors.Is(outer, Code("ord")) {
		t.Fatal("коды не должны совпадать по кускам сегментов или по потомкам")
	}
	if !errors.Is(outer, context.DeadlineExceeded) {
		t.Fatal("причина внутри E должна быть видна errors.Is")
	}
	if got := CodeOf(outer); got != "order.save" {
		t.Fatalf("CodeOf = %q, ожидали самый внешний order.save", got)
	}
	if CodeOf(io.EOF) != "" {
		t.Fatal("CodeOf ошибки без E должен вернуть пустой код")
	}
}
