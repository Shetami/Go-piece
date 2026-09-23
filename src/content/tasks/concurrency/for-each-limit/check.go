package main

func TestForEachLimitProcessesAll(t *testing.T) {
	var sum atomic.Int64
	items := make([]int, 100)
	for i := range items {
		items[i] = i + 1
	}
	ForEachLimit(items, 7, func(n int) { sum.Add(int64(n)) })
	if sum.Load() != 5050 {
		t.Fatalf("сумма %d, ожидали 5050: обработаны не все элементы или ForEachLimit вернулся раньше", sum.Load())
	}
}

func TestForEachLimitRespectsLimit(t *testing.T) {
	var cur, peak atomic.Int64
	ForEachLimit(make([]int, 12), 3, func(int) {
		n := cur.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		cur.Add(-1)
	})
	if peak.Load() > 3 {
		t.Fatalf("одновременно работало %d вызовов, лимит 3", peak.Load())
	}
	if peak.Load() < 3 {
		t.Fatalf("одновременно работало не больше %d вызовов — параллельности нет, а лимит 3", peak.Load())
	}
}

func TestForEachLimitEmpty(t *testing.T) {
	ForEachLimit(nil, 2, func(int) { t.Fatal("f не должна вызываться") })
}
