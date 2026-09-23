package main

func TestUniqueKeepsFirstOccurrence(t *testing.T) {
	got := Unique([]int{3, 1, 3, 2, 1, 3})
	if want := []int{3, 1, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Unique([3 1 3 2 1 3]) = %v, ожидали %v", got, want)
	}
}

func TestUniqueStrings(t *testing.T) {
	got := Unique([]string{"go", "rust", "go", "zig", "rust"})
	if want := []string{"go", "rust", "zig"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Unique по строкам = %v, ожидали %v", got, want)
	}
}

func TestUniqueDoesNotTouchInput(t *testing.T) {
	src := []int{1, 1, 2, 2}
	Unique(src)
	if want := []int{1, 1, 2, 2}; !reflect.DeepEqual(src, want) {
		t.Fatalf("исходный слайс изменился: %v", src)
	}
}

func TestUniqueEdgeCases(t *testing.T) {
	if got := Unique([]int(nil)); len(got) != 0 {
		t.Fatalf("Unique(nil) = %v, ожидали пустой результат", got)
	}
	if got := Unique([]int{7, 7, 7}); !reflect.DeepEqual(got, []int{7}) {
		t.Fatalf("Unique([7 7 7]) = %v, ожидали [7]", got)
	}
}
