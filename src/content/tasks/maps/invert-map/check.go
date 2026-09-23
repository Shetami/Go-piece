package main

func TestInvertBasic(t *testing.T) {
	got, err := Invert(map[string]int{"один": 1, "два": 2})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if want := map[int]string{1: "один", 2: "два"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Invert = %v, ожидали %v", got, want)
	}
}

func TestInvertDuplicate(t *testing.T) {
	got, err := Invert(map[string]string{"en": "go", "ru": "го", "de": "go"})
	if err == nil {
		t.Fatalf("два ключа со значением go — ожидали ошибку, получили %v", got)
	}
	if got != nil {
		t.Fatalf("при ошибке ожидали nil-мапу, получили %v", got)
	}
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("ошибка %q должна оборачивать ErrDuplicate", err)
	}
	if !strings.Contains(err.Error(), "go") {
		t.Fatalf("в тексте ошибки %q должно быть повторяющееся значение", err)
	}
}

func TestInvertEmpty(t *testing.T) {
	got, err := Invert(map[int]int{})
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("пустая мапа: got=%v err=%v, ожидали пустую мапу без ошибки", got, err)
	}
}
