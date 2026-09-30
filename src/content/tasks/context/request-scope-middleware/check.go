package main

var chkTokens = map[string]string{"t1": "alice", "t2": "bob"}

func chkApp() Handler {
	n := 0
	gen := func() string { n++; return fmt.Sprintf("gen-%d", n) }
	h := func(ctx context.Context, req Request) string { return Logf(ctx, "заказ %s", req.Body) }
	return Chain(h, WithRequestID(gen), WithUser(chkTokens))
}

func TestMiddlewareFull(t *testing.T) {
	app := chkApp()
	got := app(context.Background(), Request{Header: map[string]string{"X-Request-Id": "  r-1 ", "Authorization": "Bearer t1"}, Body: "42"})
	if got != "[r-1 alice] заказ 42" {
		t.Fatalf("лог = %q, ожидали %q", got, "[r-1 alice] заказ 42")
	}
}

func TestMiddlewareDefaults(t *testing.T) {
	app := chkApp()
	cases := []struct {
		hdr  map[string]string
		want string
	}{
		{nil, "[gen-1 anon] заказ 1"},
		{map[string]string{"X-Request-Id": " ", "Authorization": "Bearer nope"}, "[gen-2 anon] заказ 1"},
		{map[string]string{"Authorization": "Basic t2"}, "[gen-3 anon] заказ 1"},
		{map[string]string{"Authorization": "Bearer "}, "[gen-4 anon] заказ 1"},
		{map[string]string{"Authorization": "Bearer t2"}, "[gen-5 bob] заказ 1"},
	}
	for _, c := range cases {
		if got := app(context.Background(), Request{Header: c.hdr, Body: "1"}); got != c.want {
			t.Fatalf("заголовки %v: лог = %q, ожидали %q", c.hdr, got, c.want)
		}
	}
}

func TestMiddlewareOrder(t *testing.T) {
	var trace []string
	mw := func(name string) Middleware {
		return func(next Handler) Handler {
			return func(ctx context.Context, r Request) string {
				trace = append(trace, name)
				return next(ctx, r)
			}
		}
	}
	h := Chain(func(context.Context, Request) string { trace = append(trace, "h"); return "" }, mw("a"), mw("b"), mw("c"))
	h(context.Background(), Request{})
	if !slices.Equal(trace, []string{"a", "b", "c", "h"}) {
		t.Fatalf("порядок вызова %v, ожидали [a b c h] — mws[0] внешний", trace)
	}
}

func TestMiddlewareKeysArePrivate(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "поддельный")
	ctx = context.WithValue(ctx, "user", "admin")
	for _, k := range []string{"rid", "requestID", "RequestID", "X-Request-Id", "userKey", "User"} {
		ctx = context.WithValue(ctx, k, "admin")
	}
	if RequestID(ctx) != "" {
		t.Fatalf("RequestID прочитал значение по строковому ключу — ключ должен быть приватным типом")
	}
	if u, ok := User(ctx); ok {
		t.Fatalf("User = %q — чужой код подделал пользователя строковым ключом", u)
	}
	if got := Logf(context.Background(), "x=%d", 1); got != "[- anon] x=1" {
		t.Fatalf("Logf без значений = %q, ожидали %q", got, "[- anon] x=1")
	}
}

func TestMiddlewareUserDoesNotLeak(t *testing.T) {
	// Внешний слой положил пользователя (например, сервисный токен),
	// а в самом запросе токен неизвестен — пользователя быть не должно.
	outer := Chain(func(ctx context.Context, r Request) string { return Logf(ctx, "ok") },
		WithUser(chkTokens), WithUser(map[string]string{}))
	got := outer(context.Background(), Request{Header: map[string]string{"Authorization": "Bearer t1"}})
	if got != "[- anon] ok" {
		t.Fatalf("лог = %q; ожидали [- anon] — неизвестный токен не должен наследовать пользователя", got)
	}
}
