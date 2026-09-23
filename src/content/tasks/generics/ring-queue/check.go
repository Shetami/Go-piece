package main

func TestQueueFIFO(t *testing.T) {
	var q Queue[string]
	q.Push("a")
	q.Push("b")
	q.Push("c")
	for _, want := range []string{"a", "b", "c"} {
		if got, ok := q.Pop(); !ok || got != want {
			t.Fatalf("Pop = %q, %v; ожидали %q", got, ok, want)
		}
	}
	if _, ok := q.Pop(); ok {
		t.Fatal("Pop на пустой очереди вернул ok")
	}
}

func TestQueueWrapAround(t *testing.T) {
	var q Queue[int]
	next := 0
	// Чередуем вставки и изъятия, чтобы голова ушла по кругу много раз.
	for i := range 1000 {
		q.Push(i)
		if i%3 == 2 {
			for range 2 {
				v, ok := q.Pop()
				if !ok || v != next {
					t.Fatalf("Pop = %d, %v; ожидали %d", v, ok, next)
				}
				next++
			}
		}
	}
	if q.Len() != 1000-next {
		t.Fatalf("Len = %d, ожидали %d", q.Len(), 1000-next)
	}
	for q.Len() > 0 {
		v, _ := q.Pop()
		if v != next {
			t.Fatalf("после роста порядок сбился: %d, ожидали %d", v, next)
		}
		next++
	}
}

func TestQueueZeroValueUsable(t *testing.T) {
	var q Queue[float64]
	if q.Len() != 0 {
		t.Fatal("нулевая очередь должна быть пустой")
	}
	q.Push(1.5)
	if v, ok := q.Pop(); !ok || v != 1.5 {
		t.Fatalf("Pop = %v, %v", v, ok)
	}
}
