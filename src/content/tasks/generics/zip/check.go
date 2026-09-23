package main

func TestZipEqual(t *testing.T) {
	got := Zip([]string{"a", "b"}, []int{1, 2})
	want := []Pair[string, int]{{"a", 1}, {"b", 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Zip = %v, ожидали %v", got, want)
	}
}

func TestZipShorter(t *testing.T) {
	if got := Zip([]int{1, 2, 3}, []bool{true}); len(got) != 1 || got[0] != (Pair[int, bool]{1, true}) {
		t.Fatalf("по короткому: %v", got)
	}
	if got := Zip([]int{1}, []string{"x", "y", "z"}); len(got) != 1 {
		t.Fatalf("по короткому (второй длиннее): %v", got)
	}
}

func TestZipEmpty(t *testing.T) {
	if got := Zip([]int(nil), []int{1}); len(got) != 0 {
		t.Fatalf("Zip с пустым: %v", got)
	}
}
