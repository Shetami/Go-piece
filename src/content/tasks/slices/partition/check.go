package main

func TestPartitionInts(t *testing.T) {
	even, odd := Partition([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
	if !reflect.DeepEqual(even, []int{2, 4, 6}) || !reflect.DeepEqual(odd, []int{1, 3, 5}) {
		t.Fatalf("Partition чётных: yes=%v no=%v", even, odd)
	}
}

func TestPartitionKeepsOrder(t *testing.T) {
	long, short := Partition([]string{"go", "rust", "c", "zig", "haskell"}, func(s string) bool { return len(s) > 2 })
	if !reflect.DeepEqual(long, []string{"rust", "zig", "haskell"}) {
		t.Fatalf("yes = %v, порядок должен сохраниться", long)
	}
	if !reflect.DeepEqual(short, []string{"go", "c"}) {
		t.Fatalf("no = %v, порядок должен сохраниться", short)
	}
}

func TestPartitionDoesNotTouchInput(t *testing.T) {
	src := []int{5, 1, 4, 2}
	Partition(src, func(n int) bool { return n > 2 })
	if !reflect.DeepEqual(src, []int{5, 1, 4, 2}) {
		t.Fatalf("исходный слайс изменился: %v", src)
	}
}

func TestPartitionAllOrNothing(t *testing.T) {
	yes, no := Partition([]int{1, 2}, func(int) bool { return true })
	if len(yes) != 2 || len(no) != 0 {
		t.Fatalf("всё подходит: yes=%v no=%v", yes, no)
	}
	yes, no = Partition([]int(nil), func(int) bool { return true })
	if len(yes) != 0 || len(no) != 0 {
		t.Fatalf("на nil ожидали две пустые части: %v %v", yes, no)
	}
}
