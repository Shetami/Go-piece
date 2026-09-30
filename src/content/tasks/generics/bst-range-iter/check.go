package main

func chkTreeKeys[K cmp.Ordered, V any](t *testing.T, seq iter.Seq2[K, V], limit int) (keys []K) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("паника при обходе: %v", r)
		}
	}()
	for k := range seq {
		keys = append(keys, k)
		if len(keys) == limit {
			break
		}
	}
	return keys
}

func TestTreeBasic(t *testing.T) {
	var tr Tree[string, int]
	for i, k := range []string{"m", "c", "x", "a", "e", "c"} {
		tr.Put(k, i)
	}
	if v, ok := tr.Get("c"); !ok || v != 5 || tr.Len() != 5 {
		t.Fatalf("Get(c) = (%v, %v), Len = %d — ожидали перезапись (5, true) и 5 ключей", v, ok, tr.Len())
	}
	if _, ok := tr.Get("b"); ok {
		t.Fatalf("Get(b) нашёл несуществующий ключ")
	}
	if got := chkTreeKeys(t, tr.All(), 100); !reflect.DeepEqual(got, []string{"a", "c", "e", "m", "x"}) {
		t.Fatalf("All = %q, ожидали по возрастанию", got)
	}
	if got := chkTreeKeys(t, tr.All(), 2); !reflect.DeepEqual(got, []string{"a", "c"}) {
		t.Fatalf("break после двух: %q", got)
	}
}

func TestTreeDelete(t *testing.T) {
	var tr Tree[int, string]
	for _, k := range []int{50, 30, 70, 20, 40, 60, 80, 35, 45, 65} {
		tr.Put(k, strconv.Itoa(k))
	}
	for _, k := range []int{30, 50, 20, 99, 70} {
		ok := tr.Delete(k)
		if ok != (k != 99) {
			t.Fatalf("Delete(%d) = %v", k, ok)
		}
	}
	want := []int{35, 40, 45, 60, 65, 80}
	if got := chkTreeKeys(t, tr.All(), 100); !reflect.DeepEqual(got, want) || tr.Len() != 6 {
		t.Fatalf("после удаления 30, 50 (корень с двумя детьми), 20, 70: %v, Len=%d, ожидали %v", got, tr.Len(), want)
	}
	for _, k := range want {
		if v, ok := tr.Get(k); !ok || v != strconv.Itoa(k) {
			t.Fatalf("после удалений Get(%d) = (%q, %v)", k, v, ok)
		}
	}
}

func TestTreeRandomAgainstMap(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	var tr Tree[int, int]
	ref := map[int]int{}
	for i := range 3000 {
		k := r.IntN(300)
		if r.IntN(3) == 0 {
			if tr.Delete(k) != (func() bool { _, ok := ref[k]; return ok })() {
				t.Fatalf("Delete(%d) разошёлся с мапой", k)
			}
			delete(ref, k)
		} else {
			tr.Put(k, i)
			ref[k] = i
		}
	}
	keys := slices.Sorted(maps.Keys(ref))
	if got := chkTreeKeys(t, tr.All(), 1000); !reflect.DeepEqual(got, keys) || tr.Len() != len(ref) {
		t.Fatalf("после 3000 случайных операций ключи разошлись с мапой (Len=%d, ожидали %d)", tr.Len(), len(ref))
	}
	lo, hi := 100, 200
	var want []int
	for _, k := range keys {
		if k >= lo && k < hi {
			want = append(want, k)
		}
	}
	if got := chkTreeKeys(t, tr.Range(lo, hi), 1000); !reflect.DeepEqual(got, want) {
		t.Fatalf("Range(100, 200) = %v, ожидали %v", got, want)
	}
}

func TestTreeRangeBounds(t *testing.T) {
	var tr Tree[float64, bool]
	for _, k := range []float64{1, 2, 2.5, 3, 4} {
		tr.Put(k, true)
	}
	if got := chkTreeKeys(t, tr.Range(2, 4), 10); !reflect.DeepEqual(got, []float64{2, 2.5, 3}) {
		t.Fatalf("Range(2, 4) = %v — lo включается, hi нет", got)
	}
	if got := chkTreeKeys(t, tr.Range(5, 9), 10); len(got) != 0 {
		t.Fatalf("Range(5, 9) = %v, ожидали пусто", got)
	}
}

func TestTreeRangePrunes(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 6))
	var tr Tree[int, int]
	for _, k := range r.Perm(100000) {
		tr.Put(k, k)
	}
	start := time.Now()
	total := 0
	for i := range 20000 {
		for range tr.Range(i*5, i*5+3) {
			total++
		}
	}
	if total != 60000 {
		t.Fatalf("20000 узких Range выдали %d ключей, ожидали 60000", total)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("20000 узких Range на 100 000 ключей заняли %v — Range обходит всё дерево?", d)
	}
}
