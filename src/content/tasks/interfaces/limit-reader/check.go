package main

func TestLimitCuts(t *testing.T) {
	got, err := io.ReadAll(Limit(strings.NewReader("hello, world"), 5))
	if err != nil || string(got) != "hello" {
		t.Fatalf("Limit(…, 5) прочитал %q, %v; ожидали \"hello\"", got, err)
	}
}

func TestLimitShortSource(t *testing.T) {
	got, err := io.ReadAll(Limit(strings.NewReader("hi"), 100))
	if err != nil || string(got) != "hi" {
		t.Fatalf("источник короче лимита: %q, %v", got, err)
	}
}

func TestLimitZero(t *testing.T) {
	n, err := Limit(strings.NewReader("abc"), 0).Read(make([]byte, 10))
	if n != 0 || err != io.EOF {
		t.Fatalf("лимит 0: n=%d err=%v, ожидали 0 и io.EOF", n, err)
	}
}

func TestLimitDoesNotOverread(t *testing.T) {
	src := strings.NewReader("0123456789")
	got, _ := io.ReadAll(Limit(src, 4))
	rest, _ := io.ReadAll(src)
	if string(got) != "0123" || string(rest) != "456789" {
		t.Fatalf("прочитали %q, в источнике осталось %q; лишнее из источника брать нельзя", got, rest)
	}
}

func TestLimitSmallBuffers(t *testing.T) {
	r := Limit(strings.NewReader("abcdefgh"), 5)
	var out []byte
	buf := make([]byte, 2)
	for {
		n, err := r.Read(buf)
		out = append(out, buf[:n]...)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if string(out) != "abcde" {
		t.Fatalf("чтение по 2 байта дало %q, ожидали \"abcde\"", out)
	}
}
