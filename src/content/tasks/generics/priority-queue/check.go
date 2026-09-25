package main

func TestPQInts(t *testing.T) {
	q := NewPQ(func(a, b int) bool { return a < b })
	for _, v := range []int{5, 1, 8, 3, 9, 2, 7} {
		q.Push(v)
	}
	var got []int
	for q.Len() > 0 {
		v, _ := q.Pop()
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []int{1, 2, 3, 5, 7, 8, 9}) {
		t.Fatalf("получили %v", got)
	}
}

func TestPQEmpty(t *testing.T) {
	q := NewPQ(func(a, b string) bool { return a < b })
	if v, ok := q.Pop(); ok || v != "" {
		t.Fatalf("пустая очередь: %q, %v", v, ok)
	}
}

func TestPQMaxHeapStructs(t *testing.T) {
	type task struct {
		name string
		prio int
	}
	q := NewPQ(func(a, b task) bool { return a.prio > b.prio })
	q.Push(task{"почта", 1})
	q.Push(task{"авария", 10})
	q.Push(task{"отчёт", 5})
	if v, _ := q.Pop(); v.name != "авария" {
		t.Fatalf("первой должна идти авария, получили %v", v)
	}
	q.Push(task{"звонок", 7})
	if v, _ := q.Pop(); v.name != "звонок" {
		t.Fatalf("после добавления звонка ожидали его, получили %v", v)
	}
}

func TestPQRandomAgainstSort(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	q := NewPQ(func(a, b int) bool { return a < b })
	var all []int
	for range 2000 {
		v := r.IntN(500)
		q.Push(v)
		all = append(all, v)
		if r.IntN(3) == 0 {
			got, _ := q.Pop()
			slices.Sort(all)
			if got != all[0] {
				t.Fatalf("Pop = %d, а наименьший %d", got, all[0])
			}
			all = all[1:]
		}
	}
}

func TestPQFast(t *testing.T) {
	q := NewPQ(func(a, b int) bool { return a < b })
	start := time.Now()
	for i := range 200000 {
		q.Push((i * 7919) % 200000)
	}
	for range 200000 {
		q.Pop()
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("200 000 Push и Pop заняли %v — не похоже на O(log n)", d)
	}
}
