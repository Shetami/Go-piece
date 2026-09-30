package main

var chkErrDB = errors.New("нет соединения")

func TestFieldsBasic(t *testing.T) {
	if WithFields(nil, "a", 1) != nil {
		t.Fatal("WithFields(nil) должен вернуть nil")
	}
	err := fmt.Errorf("сохранить заказ: %w", WithFields(chkErrDB, "host", "db-1", "attempt", 3))
	if err.Error() != "сохранить заказ: нет соединения" {
		t.Fatalf("текст изменился: %q", err.Error())
	}
	if !errors.Is(err, chkErrDB) {
		t.Fatal("errors.Is не видит исходную ошибку сквозь WithFields")
	}
	got := Fields(err)
	if !maps.Equal(got, map[string]any{"host": "db-1", "attempt": 3}) {
		t.Fatalf("Fields = %v, ожидали map[attempt:3 host:db-1]", got)
	}
	if f := Fields(io.EOF); f == nil || len(f) != 0 {
		t.Fatalf("ошибка без полей: ожидали пустую не-nil мапу, получили %#v", f)
	}
}

func TestFieldsPrecedence(t *testing.T) {
	inner := WithFields(chkErrDB, "user", "u-inner", "table", "orders")
	err := WithFields(fmt.Errorf("repo: %w", inner), "user", "u-outer", "req", "r1")
	want := map[string]any{"user": "u-inner", "table": "orders", "req": "r1"}
	if got := Fields(err); !maps.Equal(got, want) {
		t.Fatalf("Fields = %v, ожидали %v (глубокое поле побеждает)", got, want)
	}
}

func TestFieldsJoin(t *testing.T) {
	err := WithFields(errors.Join(
		WithFields(io.EOF, "part", 1, "shard", "a"),
		fmt.Errorf("x: %w", WithFields(chkErrDB, "part", 2, "host", "h")),
	), "job", "import")
	want := map[string]any{"part": 1, "shard": "a", "host": "h", "job": "import"}
	if got := Fields(err); !maps.Equal(got, want) {
		t.Fatalf("Fields = %v, ожидали %v (поля из всех веток Join, ранняя ветка побеждает)", got, want)
	}
}

func TestFieldsNoAlias(t *testing.T) {
	kv := []any{"k", "v1"}
	err := WithFields(io.EOF, kv...)
	kv[1] = "v2"
	if got := Fields(err)["k"]; got != "v1" {
		t.Fatalf("после изменения исходного слайса поле стало %v — WithFields должен скопировать kv", got)
	}
}

func TestAttrs(t *testing.T) {
	if Attrs(nil) != nil {
		t.Fatal("Attrs(nil) должен вернуть nil")
	}
	err := fmt.Errorf("импорт: %w", WithFields(chkErrDB, "b", 2, "a", 1))
	var sb strings.Builder
	for _, a := range Attrs(err) {
		sb.WriteString(a.String() + ";")
	}
	if want := "a=1;b=2;err=импорт: нет соединения;"; sb.String() != want {
		t.Fatalf("Attrs = %q, ожидали %q", sb.String(), want)
	}
}
