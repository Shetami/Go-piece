package main

func TestMergeBasic(t *testing.T) {
	got := MergeSorted([]int{1, 4, 9}, []int{2, 3, 10, 11})
	if want := []int{1, 2, 3, 4, 9, 10, 11}; !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeSorted = %v, ожидали %v", got, want)
	}
}

func TestMergeDuplicates(t *testing.T) {
	got := MergeSorted([]int{1, 2, 2}, []int{2, 3})
	if want := []int{1, 2, 2, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("дубликаты должны сохраниться: %v, ожидали %v", got, want)
	}
}

func TestMergeEmpty(t *testing.T) {
	if got := MergeSorted(nil, []int{1, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("MergeSorted(nil, [1 2]) = %v", got)
	}
	if got := MergeSorted([]int{5}, nil); !reflect.DeepEqual(got, []int{5}) {
		t.Fatalf("MergeSorted([5], nil) = %v", got)
	}
	if got := MergeSorted(nil, nil); len(got) != 0 {
		t.Fatalf("MergeSorted(nil, nil) = %v, ожидали пустой", got)
	}
}

func TestMergeDoesNotTouchInputs(t *testing.T) {
	a := make([]int, 2, 10)
	a[0], a[1] = 1, 5
	b := []int{2, 3}
	MergeSorted(a, b)
	if a[:3][2] != 0 {
		t.Fatalf("MergeSorted записал в свободную вместимость первого слайса: %v", a[:3])
	}
	if !reflect.DeepEqual(b, []int{2, 3}) {
		t.Fatalf("второй слайс изменился: %v", b)
	}
}

func TestMergeMatchesSort(t *testing.T) {
	a := []int{-5, 0, 0, 3, 8, 13, 21}
	b := []int{-7, -5, 1, 2, 3, 34}
	want := append(append([]int(nil), a...), b...)
	sort.Ints(want)
	if got := MergeSorted(a, b); !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeSorted = %v, ожидали %v", got, want)
	}
}
