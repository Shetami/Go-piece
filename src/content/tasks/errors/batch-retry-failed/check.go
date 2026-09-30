package main

type chkTempErr struct{ msg string }

func (e *chkTempErr) Error() string   { return e.msg }
func (e *chkTempErr) Temporary() bool { return true }

var (
	chkBusy    = &chkTempErr{"перегружен"}
	chkInvalid = errors.New("невалидный payload")
)

func chkItems(ids ...string) []Item {
	out := make([]Item, len(ids))
	for i, id := range ids {
		out[i] = Item{ID: id}
	}
	return out
}

func chkIDs(b []Item) string {
	var ids []string
	for _, it := range b {
		ids = append(ids, it.ID)
	}
	return strings.Join(ids, ",")
}

// chkSender отвечает по сценарию: i-й вызов — replies[i].
func chkSender(batches *[]string, replies ...error) func(context.Context, []Item) error {
	return func(_ context.Context, b []Item) error {
		*batches = append(*batches, chkIDs(b))
		if n := len(*batches) - 1; n < len(replies) {
			return replies[n]
		}
		return nil
	}
}

func TestSendAllPartial(t *testing.T) {
	var batches []string
	send := chkSender(&batches,
		fmt.Errorf("api: %w", &BatchError{Failed: map[int]error{1: chkBusy, 3: chkInvalid}}),
		&BatchError{Failed: map[int]error{0: chkBusy}}, // индекс 0 — это b в пачке [b]
	)
	err := SendAll(context.Background(), chkItems("a", "b", "c", "d"), 3, send)
	if want := []string{"a,b,c,d", "b", "b"}; !slices.Equal(batches, want) {
		t.Fatalf("пачки %v, ожидали %v (повторять только непринятые; индексы Failed — в текущей пачке)", batches, want)
	}
	var ie *ItemError
	if !errors.As(err, &ie) || ie.ID != "d" || ie.Attempts != 1 || !errors.Is(err, chkInvalid) {
		t.Fatalf("ожидали единственную окончательную неудачу d после 1 попытки с постоянной ошибкой, получили %v", err)
	}
	if strings.Contains(err.Error(), "\n") {
		t.Fatalf("b в итоге принят, неудача должна быть одна: %q", err)
	}
}

func TestSendAllWholeBatch(t *testing.T) {
	var batches []string
	err := SendAll(context.Background(), chkItems("a", "b"), 3, chkSender(&batches, fmt.Errorf("503: %w", chkBusy)))
	if err != nil || !slices.Equal(batches, []string{"a,b", "a,b"}) {
		t.Fatalf("временная ошибка всей пачки: повторить всю; err=%v пачки=%v", err, batches)
	}
	batches = nil
	err = SendAll(context.Background(), chkItems("a", "b"), 3, chkSender(&batches, chkInvalid))
	if len(batches) != 1 || !errors.Is(err, chkInvalid) || strings.Count(err.Error(), "\n") != 1 {
		t.Fatalf("постоянная ошибка всей пачки: одна отправка и две неудачи; пачки=%v err=%v", batches, err)
	}
}

func TestSendAllExhausted(t *testing.T) {
	var batches []string
	always := func(_ context.Context, b []Item) error {
		batches = append(batches, chkIDs(b))
		return &BatchError{Failed: map[int]error{len(b) - 1: chkBusy}}
	}
	err := SendAll(context.Background(), chkItems("a", "b", "c"), 4, always)
	if want := []string{"a,b,c", "c", "c", "c"}; !slices.Equal(batches, want) {
		t.Fatalf("пачки %v, ожидали %v", batches, want)
	}
	var ie *ItemError
	if !errors.As(err, &ie) || ie.ID != "c" || ie.Attempts != 4 || !errors.Is(err, chkBusy) {
		t.Fatalf("ожидали неудачу c после 4 попыток, получили %v", err)
	}
	if want := "c (попыток: 4): перегружен"; err.Error() != want {
		t.Fatalf("текст %q, ожидали %q", err.Error(), want)
	}
}

func TestSendAllOrder(t *testing.T) {
	var batches []string
	send := chkSender(&batches,
		&BatchError{Failed: map[int]error{0: chkBusy, 2: chkInvalid}},
		&BatchError{Failed: map[int]error{0: chkInvalid}},
	)
	err := SendAll(context.Background(), chkItems("a", "b", "c"), 5, send)
	if want := "a (попыток: 2): невалидный payload\nc (попыток: 1): невалидный payload"; err == nil || err.Error() != want {
		t.Fatalf("ошибка %q, ожидали %q (в исходном порядке элементов)", err, want)
	}
}

func TestSendAllContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var batches []string
	send := func(_ context.Context, b []Item) error {
		batches = append(batches, chkIDs(b))
		cancel()
		return chkBusy
	}
	err := SendAll(ctx, chkItems("a", "b"), 5, send)
	var ie *ItemError
	if len(batches) != 1 || !errors.Is(err, context.Canceled) || !errors.As(err, &ie) || ie.Attempts != 1 {
		t.Fatalf("после отмены не отправлять: пачки=%v err=%v", batches, err)
	}
	called := false
	if err := SendAll(context.Background(), nil, 3, func(context.Context, []Item) error { called = true; return nil }); err != nil || called {
		t.Fatalf("пустой вход: send не вызывается, результат nil; called=%v err=%v", called, err)
	}
}
