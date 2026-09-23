package main

func checkFind(id string) error {
	return fmt.Errorf("загрузка профиля: %w", &NotFoundError{Resource: "user", ID: id})
}

func TestNotFoundMessage(t *testing.T) {
	err := &NotFoundError{Resource: "order", ID: "42"}
	if err.Error() != "order 42: не найдено" {
		t.Fatalf("Error() = %q, ожидали %q", err.Error(), "order 42: не найдено")
	}
}

func TestNotFoundIsSentinel(t *testing.T) {
	err := checkFind("7")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("errors.Is(%q, ErrNotFound) должно быть истинно", err)
	}
}

func TestNotFoundAs(t *testing.T) {
	var nf *NotFoundError
	if !errors.As(checkFind("7"), &nf) || nf.ID != "7" {
		t.Fatal("errors.As должен достать *NotFoundError с ID 7")
	}
}

func TestNotFoundNotOtherErrors(t *testing.T) {
	if errors.Is(checkFind("7"), io.EOF) {
		t.Fatal("NotFoundError не должен совпадать с любой ошибкой")
	}
}
