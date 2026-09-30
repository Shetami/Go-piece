package main

func TestMoneyString(t *testing.T) {
	cases := []struct {
		m    Money
		want string
	}{
		{Money{1250, "RUB"}, "12.50 RUB"},
		{Money{-5, "USD"}, "-0.05 USD"},
		{Money{-105, "USD"}, "-1.05 USD"},
		{Money{0, "EUR"}, "0.00 EUR"},
		{Money{7, "EUR"}, "0.07 EUR"},
		{Money{math.MinInt64, "RUB"}, "-92233720368547758.08 RUB"},
	}
	for _, c := range cases {
		if got := c.m.String(); got != c.want {
			t.Fatalf("Money{%d, %s}.String() = %q, ожидали %q", c.m.Minor, c.m.Cur, got, c.want)
		}
	}
}

func TestMoneyFormat(t *testing.T) {
	m := Money{1250, "RUB"}
	cases := []struct{ got, want string }{
		{fmt.Sprint(m), "12.50 RUB"},
		{fmt.Sprintf("%v|%s", m, m), "12.50 RUB|12.50 RUB"},
		{fmt.Sprintf("[%12v]", m), "[   12.50 RUB]"},
		{fmt.Sprintf("[%-12s]", m), "[12.50 RUB   ]"},
		{fmt.Sprintf("%+v", m), "+12.50 RUB"},
		{fmt.Sprintf("%+v", Money{0, "RUB"}), "+0.00 RUB"},
		{fmt.Sprintf("%+v", Money{-100, "RUB"}), "-1.00 RUB"},
		{fmt.Sprintf("%d", m), "1250"},
		{fmt.Sprintf("%x", m), "%!x(Money=12.50 RUB)"},
		{fmt.Sprintf("%v", []Money{{1, "USD"}, {-1, "USD"}}), "[0.01 USD -0.01 USD]"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Fatalf("получили %q, ожидали %q", c.got, c.want)
		}
	}
	var _ fmt.Formatter = Money{}
}

func TestMoneySum(t *testing.T) {
	s, err := Sum(Money{150, "RUB"}, Money{-200, "RUB"}, Money{75, "RUB"})
	if err != nil || s != (Money{25, "RUB"}) {
		t.Fatalf("Sum = %v, %v; ожидали 0.25 RUB", s, err)
	}
	if s, err := Sum(); err != nil || s != (Money{}) {
		t.Fatalf("Sum() = %v, %v; ожидали Money{}, nil", s, err)
	}
	_, err = Sum(Money{1, "RUB"}, Money{1, "RUB"}, Money{1, "USD"})
	wrapped := fmt.Errorf("invoice 42: %w", err)
	var ce *CurrencyError
	if !errors.Is(wrapped, ErrCurrencyMismatch) || !errors.As(wrapped, &ce) || ce.Want != "RUB" || ce.Got != "USD" {
		t.Fatalf("разные валюты: %v; ожидали *CurrencyError{RUB, USD}, видимый через errors.Is и errors.As", err)
	}
	if !strings.Contains(err.Error(), "RUB") || !strings.Contains(err.Error(), "USD") {
		t.Fatalf("текст ошибки %q должен называть обе валюты", err)
	}
}
