package main

func TestKeyRoundTrip(t *testing.T) {
	user := NewKey[string]("user")
	ctx := user.With(context.Background(), "alice")
	if v, ok := user.Get(ctx); !ok || v != "alice" {
		t.Fatalf("Get = %q, %v; ожидали alice, true", v, ok)
	}
	if _, ok := user.Get(context.Background()); ok {
		t.Fatalf("Get на пустом контексте вернул ok = true")
	}
	ctx = user.With(ctx, "bob")
	if v := user.MustGet(ctx); v != "bob" {
		t.Fatalf("после второго With значение %q, ожидали ближайшее — bob", v)
	}
}

func TestKeyNoCollisions(t *testing.T) {
	a := NewKey[string]("id")
	b := NewKey[string]("id") // то же имя и тот же тип — но другой ключ
	n := NewKey[int]("id")
	ctx := a.With(context.Background(), "из a")
	ctx = context.WithValue(ctx, "id", "строковый ключ")
	if _, ok := b.Get(ctx); ok {
		t.Fatalf("ключи из двух NewKey с одним именем видят значения друг друга")
	}
	if _, ok := n.Get(ctx); ok {
		t.Fatalf("Key[int] увидел значение Key[string] с тем же именем")
	}
	if v, _ := a.Get(ctx); v != "из a" {
		t.Fatalf("строковый ключ \"id\" перекрыл значение ключа: %q", v)
	}
}

func TestKeyNilValue(t *testing.T) {
	lastErr := NewKey[error]("last-error")
	ctx := lastErr.With(context.Background(), nil)
	v, ok := lastErr.Get(ctx)
	if !ok || v != nil {
		t.Fatalf("положили nil-ошибку: Get = %v, %v; ожидали nil, true — «положили nil» ≠ «не клали»", v, ok)
	}
	ptr := NewKey[*int]("ptr")
	if _, ok := ptr.Get(ptr.With(context.Background(), nil)); !ok {
		t.Fatalf("положили nil-указатель, а Get говорит, что значения нет")
	}
}

func TestKeyMustGetPanics(t *testing.T) {
	k := NewKey[int]("tenant-id")
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(fmt.Sprint(r), "tenant-id") {
			t.Fatalf("MustGet без значения: recover() = %v; ожидали панику с именем ключа", r)
		}
	}()
	k.MustGet(context.Background())
}
