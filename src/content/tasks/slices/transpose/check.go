package main

func TestTransposeRect(t *testing.T) {
	m := [][]int{{1, 2, 3}, {4, 5, 6}}
	got := Transpose(m)
	want := [][]int{{1, 4}, {2, 5}, {3, 6}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Transpose(%v) = %v, ожидали %v", m, got, want)
	}
	if !reflect.DeepEqual(m, [][]int{{1, 2, 3}, {4, 5, 6}}) {
		t.Fatalf("исходная матрица изменилась: %v", m)
	}
}

func TestTransposeRow(t *testing.T) {
	got := Transpose([][]int{{7, 8}})
	if !reflect.DeepEqual(got, [][]int{{7}, {8}}) {
		t.Fatalf("Transpose([[7 8]]) = %v, ожидали [[7] [8]]", got)
	}
}

func TestTransposeEmpty(t *testing.T) {
	if got := Transpose(nil); len(got) != 0 {
		t.Fatalf("Transpose(nil) = %v, ожидали пустую", got)
	}
}

func TestTransposeTwice(t *testing.T) {
	m := [][]int{{1, 2}, {3, 4}, {5, 6}}
	if got := Transpose(Transpose(m)); !reflect.DeepEqual(got, m) {
		t.Fatalf("двойное транспонирование дало %v, ожидали %v", got, m)
	}
}

func TestTransposeIndependentRows(t *testing.T) {
	got := Transpose([][]int{{1, 2}, {3, 4}})
	_ = append(got[0], 99)
	if got[1][0] != 2 {
		t.Fatalf("append к первой строке испортил вторую: %v", got)
	}
}
