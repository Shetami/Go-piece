package main

func TestUpperReaderBasic(t *testing.T) {
	got, err := io.ReadAll(NewUpperReader(strings.NewReader("hello, go 1.27!")))
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != "HELLO, GO 1.27!" {
		t.Fatalf("получили %q, ожидали %q", got, "HELLO, GO 1.27!")
	}
}

func TestUpperReaderKeepsNonASCII(t *testing.T) {
	got, _ := io.ReadAll(NewUpperReader(strings.NewReader("go и мир")))
	if string(got) != "GO и мир" {
		t.Fatalf("получили %q, кириллица должна остаться как есть", got)
	}
}

func TestUpperReaderSmallReads(t *testing.T) {
	// Читатель, который отдаёт по одному байту: граница чтения проходит
	// посреди многобайтовых символов.
	r := NewUpperReader(iotest.OneByteReader(strings.NewReader("abc ёж xyz")))
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != "ABC ёж XYZ" {
		t.Fatalf("побайтовое чтение: %q, ожидали %q", got, "ABC ёж XYZ")
	}
}

func TestUpperReaderPropagatesErrors(t *testing.T) {
	boom := errors.New("сеть упала")
	_, err := io.ReadAll(NewUpperReader(iotest.ErrReader(boom)))
	if !errors.Is(err, boom) {
		t.Fatalf("ошибка нижнего reader'а потерялась: %v", err)
	}
}
