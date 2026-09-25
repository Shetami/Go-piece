package main

func TestChunkBasic(t *testing.T) {
	got := Chunk([]int{1, 2, 3, 4, 5}, 2)
	want := [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Chunk([1..5], 2) = %v, ожидали %v", got, want)
	}
}

func TestChunkExact(t *testing.T) {
	got := Chunk([]string{"a", "b", "c", "d"}, 2)
	if len(got) != 2 || len(got[1]) != 2 {
		t.Fatalf("Chunk([a b c d], 2) = %q, ожидали два куска по два", got)
	}
}

func TestChunkEmpty(t *testing.T) {
	if got := Chunk([]int{}, 3); len(got) != 0 {
		t.Fatalf("пустой вход: %v, ожидали ноль кусков", got)
	}
}

func TestChunkIndependent(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	ch := Chunk(s, 2)
	_ = append(ch[0], 100)
	if ch[1][0] != 3 || s[2] != 3 {
		t.Fatalf("append к первому куску испортил второй: %v, исходный %v", ch, s)
	}
}

func TestChunkPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("Chunk(s, 0) должен паниковать")
		}
	}()
	Chunk([]int{1}, 0)
}
