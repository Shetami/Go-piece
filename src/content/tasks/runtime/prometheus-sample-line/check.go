package main

func TestSampleFormat(t *testing.T) {
	cases := []struct {
		labels []Label
		value  float64
		want   string
	}{
		{nil, 42, "m 42\n"},
		{[]Label{{"code", "200"}, {"method", "GET"}}, 0.5, "m{code=\"200\",method=\"GET\"} 0.5\n"},
		{[]Label{{"z", "1"}, {"a", "2"}, {"m", "3"}}, 1e6, "m{a=\"2\",m=\"3\",z=\"1\"} 1e+06\n"},
		{[]Label{{"path", `C:\tmp "x"` + "\nend"}}, -1, "m{path=\"C:\\\\tmp \\\"x\\\"\\nend\"} -1\n"},
		{[]Label{{"a", ""}, {"b", "ok"}}, 1, "m{b=\"ok\"} 1\n"},
		{[]Label{{"a", ""}}, 1, "m 1\n"},
		{nil, math.NaN(), "m NaN\n"},
		{nil, math.Inf(1), "m +Inf\n"},
		{nil, math.Inf(-1), "m -Inf\n"},
		{[]Label{{"город", "Москва"}}, 3, "m{город=\"Москва\"} 3\n"},
	}
	for _, c := range cases {
		if got := string(AppendSample(nil, "m", c.labels, c.value)); got != c.want {
			t.Fatalf("AppendSample(%v, %v):\n получили %q\n ожидали  %q", c.labels, c.value, got, c.want)
		}
	}
}

func TestSampleKeepsInput(t *testing.T) {
	labels := []Label{{"z", "1"}, {"a", ""}, {"b", "2"}}
	orig := slices.Clone(labels)
	out := AppendSample([]byte("# TYPE m gauge\n"), "m", labels, 1)
	if !slices.Equal(labels, orig) {
		t.Fatalf("AppendSample изменил слайс меток: %v", labels)
	}
	if string(out) != "# TYPE m gauge\nm{b=\"2\",z=\"1\"} 1\n" {
		t.Fatalf("строка должна дописываться к dst: %q", out)
	}
}

var chkSinkB []byte

func TestSampleNoAllocs(t *testing.T) {
	buf := make([]byte, 0, 512)
	labels := []Label{{"service", "api"}, {"code", "500"}, {"method", "POST"}, {"path", `/v1/"q"`}, {"empty", ""}}
	if a := testing.AllocsPerRun(100, func() {
		chkSinkB = AppendSample(buf[:0], "http_requests_total", labels, 12345)
	}); a != 0 {
		t.Fatalf("AppendSample с 5 метками выделяет память %.0f раз на вызов, ожидали 0", a)
	}
}
