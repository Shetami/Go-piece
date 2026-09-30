package main

func chkKeys[V any](t *testing.T, seq iter.Seq2[string, V], limit int) (keys []string) {
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

func TestTrieGetPut(t *testing.T) {
	var tr Trie[int]
	tr.Put("cat", 1)
	tr.Put("car", 0)
	tr.Put("cat", 3)
	tr.Put("", 9)
	if v, ok := tr.Get("cat"); !ok || v != 3 {
		t.Fatalf("Get(cat) = (%v, %v), ожидали (3, true) после перезаписи", v, ok)
	}
	if v, ok := tr.Get("car"); !ok || v != 0 {
		t.Fatalf("Get(car) = (%v, %v): ноль — тоже значение", v, ok)
	}
	if _, ok := tr.Get("ca"); ok {
		t.Fatalf("ca — только префикс, а Get нашёл ключ")
	}
	if _, ok := tr.Get("cats"); ok {
		t.Fatalf("cats не добавляли")
	}
	if v, ok := tr.Get(""); !ok || v != 9 {
		t.Fatalf("пустой ключ: (%v, %v)", v, ok)
	}
	if tr.Len() != 3 {
		t.Fatalf("Len = %d, ожидали 3 (cat, car, пустой)", tr.Len())
	}
}

func TestTriePrefixSorted(t *testing.T) {
	var tr Trie[string]
	words := []string{"go", "gopher", "golang", "gc", "goroutine", "gob", "rust", "g", "Go"}
	for _, w := range words {
		tr.Put(w, strings.ToUpper(w))
	}
	for range 5 {
		got := chkKeys(t, tr.WithPrefix("go"), 100)
		want := []string{"go", "gob", "golang", "gopher", "goroutine"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("WithPrefix(go) = %q, ожидали %q", got, want)
		}
	}
	all := chkKeys(t, tr.WithPrefix(""), 100)
	if !slices.IsSorted(all) || len(all) != len(words) {
		t.Fatalf("WithPrefix(\"\") = %q — ожидали все %d ключей по возрастанию", all, len(words))
	}
	for k, v := range tr.WithPrefix("gc") {
		if k != "gc" || v != "GC" {
			t.Fatalf("WithPrefix(gc) выдал (%q, %q)", k, v)
		}
	}
	if got := chkKeys(t, tr.WithPrefix("java"), 100); len(got) != 0 {
		t.Fatalf("несуществующий префикс дал %q", got)
	}
}

func TestTrieUnicodeAndBreak(t *testing.T) {
	var tr Trie[int]
	for i, w := range []string{"ёж", "еда", "ель", "ехать", "енот", "ж"} {
		tr.Put(w, i)
	}
	if got := chkKeys(t, tr.WithPrefix("е"), 100); !reflect.DeepEqual(got, []string{"еда", "ель", "енот", "ехать"}) {
		t.Fatalf("WithPrefix(е) = %q", got)
	}
	if got := chkKeys(t, tr.WithPrefix("е"), 2); !reflect.DeepEqual(got, []string{"еда", "ель"}) {
		t.Fatalf("break после двух ключей: %q", got)
	}
}
