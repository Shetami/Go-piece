package main

func TestDurationValid(t *testing.T) {
	day := 24 * time.Hour
	cases := map[string]time.Duration{
		"36h":         36 * time.Hour,
		"1w2d":        9 * day,
		"1d 12h30m":   day + 12*time.Hour + 30*time.Minute,
		"  2h  15m  ": 2*time.Hour + 15*time.Minute,
		"-1h30m":      -(time.Hour + 30*time.Minute),
		"250ms":       250 * time.Millisecond,
		"1m30s500ms":  time.Minute + 30*time.Second + 500*time.Millisecond,
		"0":           0,
		"15250w1d":    15250*7*day + day,
		"0h0m1s":      time.Second,
	}
	for in, want := range cases {
		got, err := ParseDuration(in)
		if err != nil || got != want {
			t.Fatalf("ParseDuration(%q) = %v, %v; ожидали %v", in, got, err, want)
		}
	}
}

func TestDurationUnitOrder(t *testing.T) {
	for _, in := range []string{"1h2h", "5m1h", "1s1m", "10ms1s", "1d1w"} {
		if _, err := ParseDuration(in); err == nil {
			t.Fatalf("ParseDuration(%q): ожидали ошибку — единицы не по убыванию или повторяются", in)
		}
	}
}

func TestDurationSyntaxErrors(t *testing.T) {
	for _, in := range []string{"", "   ", "-", "10", "1 h", "5y", "h", "1h-30m", "--1h", "1.5h", "1hx"} {
		d, err := ParseDuration(in)
		if err == nil {
			t.Fatalf("ParseDuration(%q) = %v; ожидали ошибку", in, d)
		}
		if !errors.Is(err, ErrInvalidDuration) {
			t.Fatalf("ParseDuration(%q): ошибка %v должна оборачивать ErrInvalidDuration", in, err)
		}
		if in != "" && !strings.Contains(err.Error(), in) {
			t.Fatalf("ParseDuration(%q): текст ошибки %q должен упоминать исходную строку", in, err)
		}
	}
}

func TestDurationOverflow(t *testing.T) {
	for _, in := range []string{"15251w", "15250w2d", "99999999999999999999s", "106752d", "9223372036855ms1"} {
		if d, err := ParseDuration(in); err == nil {
			t.Fatalf("ParseDuration(%q) = %v; ожидали ошибку переполнения, а не тихий мусор", in, d)
		}
	}
}
