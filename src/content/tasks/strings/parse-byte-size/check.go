package main

func TestSizeValid(t *testing.T) {
	cases := map[string]int64{
		"512":                 512,
		"0":                   0,
		"10MB":                10_000_000,
		"10mb":                10_000_000,
		"2k":                  2000,
		"  3 KiB ":            3072,
		"1.5 GiB":             1_610_612_736,
		"1.5KiB":              1536,
		"1.1GB":               1_100_000_000,
		"0.3 KB":              300,
		"4.35 MB":             4_350_000,
		"2 Ti":                2 << 40,
		"7B":                  7,
		"8388607TiB":          8388607 << 40,
		"9223372036854775807": math.MaxInt64,
	}
	for in, want := range cases {
		got, err := ParseSize(in)
		if err != nil || got != want {
			t.Fatalf("ParseSize(%q) = %d, %v; ожидали %d", in, got, err, want)
		}
	}
}

func TestSizeSyntax(t *testing.T) {
	for _, in := range []string{"", "MB", "1,5MB", "1.", ".5K", "1..5K", "1.2.3", "-1", "+1", "5 XB", "5 M B", "10 байт", "1e3", "0x10"} {
		got, err := ParseSize(in)
		if !errors.Is(err, ErrSize) {
			t.Fatalf("ParseSize(%q) = %d, %v; ожидали ErrSize", in, got, err)
		}
		if in != "" && !strings.Contains(err.Error(), in) {
			t.Fatalf("ParseSize(%q): ошибка %q должна содержать исходную строку", in, err)
		}
	}
}

func TestSizeFractionalBytes(t *testing.T) {
	for _, in := range []string{"0.5B", "1.0001KiB", "0.1234 KB", "1.5"} {
		if got, err := ParseSize(in); !errors.Is(err, ErrSize) {
			t.Fatalf("ParseSize(%q) = %d, %v; нецелое число байт — ожидали ErrSize", in, got, err)
		}
	}
	if got, err := ParseSize("1.000KiB"); err != nil || got != 1024 {
		t.Fatalf("ParseSize(\"1.000KiB\") = %d, %v; ожидали 1024", got, err)
	}
}

func TestSizeOverflow(t *testing.T) {
	for _, in := range []string{"8388608TiB", "9223372036854775808", "99999999999999999999999 B", "9300000 TB"} {
		if got, err := ParseSize(in); !errors.Is(err, ErrOverflow) {
			t.Fatalf("ParseSize(%q) = %d, %v; ожидали ErrOverflow", in, got, err)
		}
	}
}
