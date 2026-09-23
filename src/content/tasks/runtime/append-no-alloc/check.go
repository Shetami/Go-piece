package main

func TestAppendPointFormat(t *testing.T) {
	got := AppendPoint([]byte("точка "), 3, -14)
	if string(got) != "точка (3, -14)" {
		t.Fatalf("AppendPoint = %q", got)
	}
	if got := AppendPoint(nil, 0, 0); string(got) != "(0, 0)" {
		t.Fatalf("AppendPoint(nil) = %q", got)
	}
}

func TestAppendPointNoAlloc(t *testing.T) {
	buf := make([]byte, 0, 64)
	allocs := testing.AllocsPerRun(1000, func() {
		buf = AppendPoint(buf[:0], 123456, -789)
	})
	if allocs != 0 {
		t.Fatalf("при достаточной вместимости ожидали 0 аллокаций, а их %.0f на вызов", allocs)
	}
}
