package main

func TestFormatRanges(t *testing.T) {
	cases := []struct {
		in   []int
		want string
	}{
		{[]int{1, 2, 3, 5, 7, 8}, "1-3,5,7-8"},
		{[]int{8, 3, 1, 2, 7, 5, 3, 3, 1}, "1-3,5,7-8"},
		{[]int{4}, "4"},
		{[]int{4, 5}, "4-5"},
		{[]int{-3, -1, -2, 0, 2}, "-3-0,2"},
		{[]int{-5, -4, 10}, "-5--4,10"},
		{nil, ""},
	}
	for _, c := range cases {
		orig := slices.Clone(c.in)
		if got := FormatRanges(c.in); got != c.want {
			t.Fatalf("FormatRanges(%v) = %q, ожидали %q", orig, got, c.want)
		}
		if !reflect.DeepEqual(c.in, orig) {
			t.Fatalf("FormatRanges переставил вход: %v, был %v", c.in, orig)
		}
	}
}

func TestParseRanges(t *testing.T) {
	cases := []struct {
		in   string
		want []int
	}{
		{"1-3,5,7-8", []int{1, 2, 3, 5, 7, 8}},
		{"-3--1,4", []int{-3, -2, -1, 4}},
		{"10,2-3", []int{10, 2, 3}},
		{"5-5", []int{5}},
		{"-2", []int{-2}},
	}
	for _, c := range cases {
		got, err := ParseRanges(c.in)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Fatalf("ParseRanges(%q) = %v, %v; ожидали %v", c.in, got, err, c.want)
		}
	}
	if got, err := ParseRanges(""); err != nil || len(got) != 0 {
		t.Fatalf("ParseRanges(\"\") = %v, %v; ожидали пустой результат без ошибки", got, err)
	}
}

func TestParseRangesErrors(t *testing.T) {
	for _, s := range []string{"1,,2", "3-1", "x", "1-", "1-2-3", ",", "2,"} {
		if got, err := ParseRanges(s); err == nil {
			t.Fatalf("ParseRanges(%q) = %v без ошибки, ожидали ошибку", s, got)
		}
	}
}

func TestRangesRoundTrip(t *testing.T) {
	in := []int{}
	for i := -20; i < 60; i++ {
		if (i*i+3*i)%7 < 4 {
			in = append(in, i, i)
		}
	}
	back, err := ParseRanges(FormatRanges(in))
	want := slices.Compact(slices.Sorted(slices.Values(in)))
	if err != nil || !reflect.DeepEqual(back, want) {
		t.Fatalf("Parse(Format(x)) = %v, %v; ожидали %v", back, err, want)
	}
}
