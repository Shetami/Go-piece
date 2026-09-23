package main

func TestRequestIDRoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "r-42")
	id, ok := RequestID(ctx)
	if !ok || id != "r-42" {
		t.Fatalf("RequestID = %q, %v; ожидали r-42, true", id, ok)
	}
}

func TestRequestIDMissing(t *testing.T) {
	if id, ok := RequestID(context.Background()); ok || id != "" {
		t.Fatalf("в пустом контексте: %q, %v", id, ok)
	}
}

func TestRequestIDOverride(t *testing.T) {
	parent := WithRequestID(context.Background(), "внешний")
	child := WithRequestID(parent, "внутренний")
	if id, _ := RequestID(child); id != "внутренний" {
		t.Fatalf("дочерний контекст: %q", id)
	}
	if id, _ := RequestID(parent); id != "внешний" {
		t.Fatalf("родитель не должен меняться: %q", id)
	}
}

func TestRequestIDNotReachableByString(t *testing.T) {
	ctx := WithRequestID(context.Background(), "секрет")
	for _, key := range []any{"request_id", "requestID", "request-id", "id"} {
		if v := ctx.Value(key); v != nil {
			t.Fatalf("значение достаётся по строковому ключу %q — ключ должен быть своим типом", key)
		}
	}
	// И чужой код не должен суметь подменить значение строкой.
	ctx = context.WithValue(ctx, "request_id", "подделка")
	if id, _ := RequestID(ctx); id != "секрет" {
		t.Fatalf("подделка по строковому ключу сработала: %q", id)
	}
}
