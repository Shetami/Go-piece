package main

func TestSortedKeysInts(t *testing.T) {
	got := SortedKeys(map[int]string{3: "c", 1: "a", 2: "b"})
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("SortedKeys = %v, ожидали [1 2 3]", got)
	}
}

func TestSortedKeysStrings(t *testing.T) {
	got := SortedKeys(map[string]bool{"zig": true, "go": true, "c": false})
	if !reflect.DeepEqual(got, []string{"c", "go", "zig"}) {
		t.Fatalf("SortedKeys = %v", got)
	}
}

func TestSortedKeysNamedType(t *testing.T) {
	type userID int64
	got := SortedKeys(map[userID]int{20: 1, 10: 2})
	if !reflect.DeepEqual(got, []userID{10, 20}) {
		t.Fatalf("для своего типа ключа: %v", got)
	}
}

func TestSortedKeysEmpty(t *testing.T) {
	if got := SortedKeys(map[string]int{}); len(got) != 0 {
		t.Fatalf("пустая мапа: %v", got)
	}
}
