package main

func TestKVValid(t *testing.T) {
	in := `  host=db.local  port=5432 name="my app" note="say \"hi\" \\o/" empty= город=Москва q="a=b" `
	got, err := ParseKV(in)
	if err != nil {
		t.Fatalf("ParseKV(%q): неожиданная ошибка %v", in, err)
	}
	want := map[string]string{
		"host": "db.local", "port": "5432", "name": "my app",
		"note": `say "hi" \o/`, "empty": "", "город": "Москва", "q": "a=b",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseKV(%q) = %q, ожидали %q", in, got, want)
	}
	if got, err := ParseKV(""); err != nil || len(got) != 0 {
		t.Fatalf("ParseKV(\"\") = %v, %v; ожидали пустую мапу без ошибки", got, err)
	}
}

func chkKVPos(t *testing.T, in string, wantPos int) {
	t.Helper()
	m, err := ParseKV(in)
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("ParseKV(%q) = %v, %v; ожидали *ParseError", in, m, err)
	}
	if pe.Pos != wantPos {
		t.Fatalf("ParseKV(%q): Pos = %d, ожидали %d (%v)", in, pe.Pos, wantPos, err)
	}
	if m != nil {
		t.Fatalf("ParseKV(%q): при ошибке мапа должна быть nil, получили %v", in, m)
	}
}

func TestKVErrorPositions(t *testing.T) {
	chkKVPos(t, "a=1 bc", 6)     // строка кончилась, '=' нет
	chkKVPos(t, "a=1 b c=2", 5)  // пробел в ключе
	chkKVPos(t, "a=1 =2", 4)     // пустой ключ
	chkKVPos(t, `a=x"y`, 3)      // кавычка в значении без кавычек
	chkKVPos(t, `a=1 b="abc`, 6) // незакрытая — позиция открывающей
	chkKVPos(t, `a="x\ny"`, 4)   // неизвестное экранирование
	chkKVPos(t, `a="x"y b=1`, 5) // мусор за кавычкой
	chkKVPos(t, `a="x \"" b=2 c`, 14)
}

func TestKVPositionsInRunes(t *testing.T) {
	// Позиции — в символах: байтовое смещение здесь было бы почти вдвое больше.
	chkKVPos(t, "имя=Иван фамилия", 16)
	chkKVPos(t, `город="Москва`, 6)
}

func TestKVDuplicate(t *testing.T) {
	in := "a=1 ключ=2 b=3 ключ=4"
	_, err := ParseKV(in)
	if !errors.Is(err, ErrDuplicateKey) {
		t.Fatalf("ParseKV(%q) = %v; ожидали ошибку, для которой errors.Is(err, ErrDuplicateKey)", in, err)
	}
	var pe *ParseError
	if !errors.As(err, &pe) || pe.Pos != 15 {
		t.Fatalf("ParseKV(%q) = %v; ожидали *ParseError с Pos 15 — начало повторного ключа", in, err)
	}
	if _, err := ParseKV(`a="x y" b="z"`); err != nil {
		t.Fatalf("разные ключи — не дубликат: %v", err)
	}
	if _, err := ParseKV(`a=1 a="x`); errors.Is(err, ErrDuplicateKey) {
		t.Fatalf("незакрытая кавычка в повторной паре — синтаксическая ошибка, а не ErrDuplicateKey: %v", err)
	}
}
