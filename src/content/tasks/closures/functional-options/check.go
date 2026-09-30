package main

func TestNewDefaultsAndOrder(t *testing.T) {
	c, err := New("db:5432")
	if err != nil || c == nil {
		t.Fatalf("New без опций: %v, %v", c, err)
	}
	if c.Timeout != 5*time.Second || c.Retries != 3 || c.Tags != nil || c.Addr != "db:5432" {
		t.Fatalf("значения по умолчанию: %+v", *c)
	}
	c, err = New("db", WithTimeout(time.Second), nil, WithRetries(0), WithTimeout(2*time.Second))
	if err != nil {
		t.Fatalf("корректные опции и nil среди них: ошибка %v", err)
	}
	if c.Timeout != 2*time.Second || c.Retries != 0 {
		t.Fatalf("опции применяются по порядку, последняя побеждает; Retries=0 допустим: %+v", *c)
	}
}

func TestNewCollectsAllErrors(t *testing.T) {
	c, err := New("", WithTimeout(-time.Second), WithRetries(99), WithTags("ok", ""))
	if c != nil {
		t.Fatalf("при ошибке конфиг должен быть nil, получили %+v", *c)
	}
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ошибка должна оборачивать ErrInvalid: %v", err)
	}
	for _, word := range []string{"addr", "timeout", "retries", "tag"} {
		if err == nil || !strings.Contains(err.Error(), word) {
			t.Fatalf("в ошибке нет %q — собирайте все ошибки, а не первую: %v", word, err)
		}
	}
}

func TestWithTagsCopiesArgs(t *testing.T) {
	ts := []string{"a", "b"}
	opt := WithTags(ts...)
	ts[0] = "zzz"
	c, _ := New("x", opt)
	if c == nil || !reflect.DeepEqual(c.Tags, []string{"a", "b"}) {
		t.Fatalf("опция должна запомнить теги на момент WithTags: %v", c)
	}
	ts[1] = "yyy"
	if c.Tags[1] != "b" {
		t.Fatalf("конфиг делит массив со слайсом вызывающего: Tags = %v", c.Tags)
	}
}

func TestReusedOptionDoesNotLeak(t *testing.T) {
	base := make([]string, 0, 8)
	base = append(base, "svc")
	common := WithTags(base...)
	a, _ := New("a", common, WithTags("a-only"))
	b, _ := New("b", common, WithTags("b-only"))
	if a == nil || b == nil {
		t.Fatal("New вернул nil на корректных опциях")
	}
	if !reflect.DeepEqual(a.Tags, []string{"svc", "a-only"}) {
		t.Fatalf("конфиг a испорчен после создания b: Tags = %v, ожидали [svc a-only]", a.Tags)
	}
	if !reflect.DeepEqual(b.Tags, []string{"svc", "b-only"}) {
		t.Fatalf("конфиг b: Tags = %v, ожидали [svc b-only]", b.Tags)
	}
}
