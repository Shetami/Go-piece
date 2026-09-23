package main

func TestRetrySucceeds(t *testing.T) {
	calls := 0
	err := Retry(5, func() error {
		calls++
		if calls < 3 {
			return errors.New("временно недоступно")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("успех на третьей попытке: err=%v calls=%d", err, calls)
	}
}

func TestRetryGivesUp(t *testing.T) {
	calls := 0
	last := errors.New("таймаут")
	err := Retry(4, func() error { calls++; return last })
	if calls != 4 {
		t.Fatalf("ожидали 4 попытки, было %d", calls)
	}
	if !errors.Is(err, last) {
		t.Fatalf("итоговая ошибка должна оборачивать последнюю: %v", err)
	}
	if !strings.Contains(err.Error(), "4") {
		t.Fatalf("в ошибке %q должно быть число попыток", err)
	}
}

func TestRetryStopsOnPermanent(t *testing.T) {
	calls := 0
	perm := fmt.Errorf("неверный пароль: %w", ErrPermanent)
	err := Retry(5, func() error { calls++; return perm })
	if calls != 1 {
		t.Fatalf("постоянную ошибку повторять не надо, а было %d попыток", calls)
	}
	if !errors.Is(err, ErrPermanent) {
		t.Fatalf("ожидали постоянную ошибку, получили %v", err)
	}
}
