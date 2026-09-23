package main

func TestParseConfigOK(t *testing.T) {
	cfg, err := ParseConfig("# комментарий\nhost = localhost\n\nport=8080\nurl=a=b")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	want := map[string]string{"host": "localhost", "port": "8080", "url": "a=b"}
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("разобрали %v, ожидали %v", cfg, want)
	}
}

func TestParseConfigLineNumber(t *testing.T) {
	_, err := ParseConfig("a=1\n# ок\nсломано\nb=2")
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("ожидали *ParseError, получили %v", err)
	}
	if pe.Line != 3 {
		t.Fatalf("ошибка в строке %d, ожидали 3", pe.Line)
	}
	if !errors.Is(err, ErrNoEquals) {
		t.Fatalf("причина должна находиться через errors.Is(err, ErrNoEquals): %v", err)
	}
	if !strings.Contains(err.Error(), "3") {
		t.Fatalf("в тексте %q должен быть номер строки", err)
	}
}

func TestParseConfigEmptyKey(t *testing.T) {
	_, err := ParseConfig("  = значение")
	if !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("ожидали ErrEmptyKey, получили %v", err)
	}
}
