package main

func chkMsgs(sizes ...int) [][]byte {
	out := make([][]byte, len(sizes))
	for i, n := range sizes {
		out[i] = bytes.Repeat([]byte{byte('a' + i)}, n)
	}
	return out
}

func chkShape(b [][][]byte) [][]int {
	out := [][]int{}
	for _, batch := range b {
		row := []int{}
		for _, m := range batch {
			row = append(row, len(m))
		}
		out = append(out, row)
	}
	return out
}

func TestBatchesLimits(t *testing.T) {
	cases := []struct {
		sizes           []int
		maxCount, bytes int
		want            [][]int
	}{
		{[]int{1, 1, 1, 1, 1}, 2, 100, [][]int{{1, 1}, {1, 1}, {1}}},
		{[]int{4, 4, 4, 4}, 10, 8, [][]int{{4, 4}, {4, 4}}},     // ровно 8 байт помещаются
		{[]int{3, 3, 3}, 10, 8, [][]int{{3, 3}, {3}}},
		{[]int{0, 0, 0}, 2, 1, [][]int{{0, 0}, {0}}},            // пустые сообщения считаются в maxCount
		{[]int{20, 1, 1}, 5, 10, [][]int{{20}, {1, 1}}},         // большое первым
		{[]int{1, 20, 1}, 5, 10, [][]int{{1}, {20}, {1}}},       // большое в середине
		{[]int{1, 1, 30, 40}, 5, 10, [][]int{{1, 1}, {30}, {40}}}, // два больших подряд
	}
	for _, c := range cases {
		got := chkShape(Batches(chkMsgs(c.sizes...), c.maxCount, c.bytes))
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("размеры %v, maxCount=%d, maxBytes=%d: пачки %v, ожидали %v", c.sizes, c.maxCount, c.bytes, got, c.want)
		}
	}
	if got := Batches(nil, 3, 3); len(got) != 0 {
		t.Fatalf("Batches(nil) = %v, ожидали ноль пачек", got)
	}
}

func TestBatchesNoCopy(t *testing.T) {
	msgs := chkMsgs(1, 1, 1, 1)
	b := Batches(msgs, 2, 100)
	if len(b) != 2 || &b[1][0] != &msgs[2] {
		t.Fatalf("пачки должны быть подслайсами msgs, а не копиями")
	}
	_ = append(b[0], []byte("x"))
	if len(msgs[2]) != 1 || msgs[2][0] != 'c' {
		t.Fatalf("append к первой пачке затёр msgs[2] = %q", msgs[2])
	}
}

func TestBatchesPanics(t *testing.T) {
	for _, lim := range [][2]int{{0, 10}, {10, 0}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("Batches(maxCount=%d, maxBytes=%d): ожидали панику", lim[0], lim[1])
				}
			}()
			Batches(chkMsgs(1), lim[0], lim[1])
		}()
	}
}
