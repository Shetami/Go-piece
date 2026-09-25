package main

var errDisk = errors.New("диск полон")

func TestOpErrorText(t *testing.T) {
	err := Wrap("save", errDisk)
	if err == nil || err.Error() != "save: диск полон" {
		t.Fatalf("Wrap(\"save\", errDisk) = %v, ожидали \"save: диск полон\"", err)
	}
}

func TestOpErrorNil(t *testing.T) {
	if err := Wrap("save", nil); err != nil {
		t.Fatalf("Wrap(op, nil) = %#v, ожидали настоящий nil", err)
	}
}

func TestOpErrorIs(t *testing.T) {
	err := fmt.Errorf("запрос: %w", Wrap("load", Wrap("read", errDisk)))
	if !errors.Is(err, errDisk) {
		t.Fatalf("errors.Is не нашёл errDisk в %v — нужен Unwrap", err)
	}
	if err.Error() != "запрос: load: read: диск полон" {
		t.Fatalf("текст цепочки: %q", err)
	}
}

func TestOpErrorAs(t *testing.T) {
	err := fmt.Errorf("запрос: %w", Wrap("load", errDisk))
	var op *OpError
	if !errors.As(err, &op) || op.Op != "load" {
		t.Fatalf("errors.As должен найти *OpError с Op=load")
	}
}
