package main

func TestParallelMapKeepsOrder(t *testing.T) {
	items := []int{5, 1, 4, 2, 8, 3, 7, 6}
	got := ParallelMap(items, 3, func(n int) int { return n * 10 })

	want := []int{50, 10, 40, 20, 80, 30, 70, 60}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParallelMap = %v, ожидали %v (порядок должен совпадать со входом)", got, want)
	}
}

func TestParallelMapEmpty(t *testing.T) {
	if got := ParallelMap(nil, 4, func(n int) int { return n }); len(got) != 0 {
		t.Fatalf("на пустом входе ожидали пустой результат, получили %v", got)
	}
}

func TestParallelMapRunsInParallel(t *testing.T) {
	// Каждый вызов f занимает 100 мс. Восемь элементов на четырёх воркерах —
	// это примерно 200 мс; последовательный перебор занял бы 800 мс.
	start := time.Now()
	got := ParallelMap([]int{1, 2, 3, 4, 5, 6, 7, 8}, 4, func(n int) int {
		time.Sleep(100 * time.Millisecond)
		return n
	})
	elapsed := time.Since(start)

	if len(got) != 8 {
		t.Fatalf("ожидали 8 результатов, получили %d", len(got))
	}
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("ушло %v — похоже, элементы обрабатываются по очереди, а не параллельно", elapsed)
	}
}

func TestParallelMapRespectsWorkerLimit(t *testing.T) {
	var mu sync.Mutex
	running, peak := 0, 0

	ParallelMap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3, func(n int) int {
		mu.Lock()
		running++
		if running > peak {
			peak = running
		}
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		running--
		mu.Unlock()
		return n
	})

	if peak > 3 {
		t.Fatalf("одновременно работало %d горутин при workers=3", peak)
	}
	if peak < 2 {
		t.Fatalf("одновременно работала всего %d горутина — параллелизма нет", peak)
	}
}
