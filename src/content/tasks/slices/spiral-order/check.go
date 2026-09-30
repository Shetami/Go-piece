package main

func chkGrid(r, c int) [][]int {
	m := make([][]int, r)
	for i := range m {
		m[i] = make([]int, c)
		for j := range m[i] {
			m[i][j] = i*c + j + 1
		}
	}
	return m
}

func TestSpiralShapes(t *testing.T) {
	cases := []struct {
		r, c int
		want []int
	}{
		{3, 3, []int{1, 2, 3, 6, 9, 8, 7, 4, 5}},
		{3, 4, []int{1, 2, 3, 4, 8, 12, 11, 10, 9, 5, 6, 7}},
		{4, 3, []int{1, 2, 3, 6, 9, 12, 11, 10, 7, 4, 5, 8}},
		{1, 4, []int{1, 2, 3, 4}},
		{4, 1, []int{1, 2, 3, 4}},
		{1, 1, []int{1}},
		{2, 2, []int{1, 2, 4, 3}},
		{3, 5, []int{1, 2, 3, 4, 5, 10, 15, 14, 13, 12, 11, 6, 7, 8, 9}},
	}
	for _, c := range cases {
		if got := Spiral(chkGrid(c.r, c.c)); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("Spiral(%dx%d) = %v, ожидали %v", c.r, c.c, got, c.want)
		}
	}
}

func TestSpiralEmpty(t *testing.T) {
	if got := Spiral([][]string{}); len(got) != 0 {
		t.Fatalf("Spiral(пустая) = %v", got)
	}
	if got := Spiral([][]string{{}, {}}); len(got) != 0 {
		t.Fatalf("Spiral(две пустые строки) = %v", got)
	}
}

func TestSpiralRagged(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("строки разной длины: ожидали панику")
		}
	}()
	Spiral([][]int{{1, 2}, {3}})
}

func TestSpiralOneAlloc(t *testing.T) {
	m := chkGrid(20, 30)
	orig := chkGrid(20, 30)
	allocs := testing.AllocsPerRun(20, func() { Spiral(m) })
	if allocs > 1 {
		t.Fatalf("Spiral(20x30) делает %.0f аллокаций, ожидали одну — размер результата известен заранее", allocs)
	}
	if !reflect.DeepEqual(m, orig) {
		t.Fatalf("Spiral изменил матрицу")
	}
}
