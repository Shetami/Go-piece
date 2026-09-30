package main

type chkMillis int64

type chkCell struct{ X, Y int }

func TestDijkstraCheaperLonger(t *testing.T) {
	var g Graph[string, chkMillis]
	g.AddEdge("msk", "spb", 100)
	g.AddEdge("msk", "tver", 20)
	g.AddEdge("tver", "novgorod", 30)
	g.AddEdge("novgorod", "spb", 25)
	g.AddEdge("tver", "spb", 90)
	g.AddEdge("spb", "msk", 1)
	path, d, ok := g.ShortestPath("msk", "spb")
	if !ok || d != 75 || !reflect.DeepEqual(path, []string{"msk", "tver", "novgorod", "spb"}) {
		t.Fatalf("msk→spb = %v, %d, %v; ожидали [msk tver novgorod spb], 75 — больше пересадок, но дешевле", path, d, ok)
	}
}

func TestDijkstraEdgeCases(t *testing.T) {
	var g Graph[int, float64]
	g.AddEdge(1, 2, 5)
	g.AddEdge(1, 2, 1.5)
	g.AddEdge(2, 3, 0)
	g.AddEdge(4, 1, 1)
	if p, d, ok := g.ShortestPath(1, 3); !ok || d != 1.5 || !reflect.DeepEqual(p, []int{1, 2, 3}) {
		t.Fatalf("1→3 = %v, %v, %v; ожидали [1 2 3], 1.5 (дешёвое из параллельных рёбер и ребро веса 0)", p, d, ok)
	}
	if p, d, ok := g.ShortestPath(3, 1); ok || p != nil || d != 0 {
		t.Fatalf("3→1 недостижима (граф ориентированный), а получили %v, %v, %v", p, d, ok)
	}
	if p, d, ok := g.ShortestPath(2, 2); !ok || d != 0 || !reflect.DeepEqual(p, []int{2}) {
		t.Fatalf("2→2 = %v, %v, %v; ожидали [2], 0, true", p, d, ok)
	}
	var empty Graph[string, int]
	if _, _, ok := empty.ShortestPath("a", "b"); ok {
		t.Fatalf("в пустом графе нашёлся путь")
	}
}

func TestDijkstraStaleEntries(t *testing.T) {
	// До вершины 9 сначала находится дорогой путь, потом всё более дешёвые.
	var g Graph[int, int]
	for i := 1; i <= 8; i++ {
		g.AddEdge(0, i, i*10)
		g.AddEdge(i, 9, 100-i*12)
	}
	g.AddEdge(9, 10, 1)
	p, d, ok := g.ShortestPath(0, 10)
	if !ok || d != 16 || !reflect.DeepEqual(p, []int{0, 1, 9, 10}) {
		t.Fatalf("0→10 = %v, %d; ожидали [0 1 9 10], 16", p, d)
	}
}

func TestDijkstraGridFast(t *testing.T) {
	const n = 150
	var g Graph[chkCell, int]
	for x := range n {
		for y := range n {
			w := 1 + (x*7+y*13)%5
			if x+1 < n {
				g.AddEdge(chkCell{x, y}, chkCell{x + 1, y}, w)
				g.AddEdge(chkCell{x + 1, y}, chkCell{x, y}, w)
			}
			if y+1 < n {
				g.AddEdge(chkCell{x, y}, chkCell{x, y + 1}, w)
				g.AddEdge(chkCell{x, y + 1}, chkCell{x, y}, w)
			}
		}
	}
	start := time.Now()
	p, d, ok := g.ShortestPath(chkCell{0, 0}, chkCell{n - 1, n - 1})
	if !ok || p[0] != (chkCell{0, 0}) || p[len(p)-1] != (chkCell{n - 1, n - 1}) || d <= 0 {
		t.Fatalf("путь по сетке не найден: ok=%v, d=%d", ok, d)
	}
	sum := 0
	for i := 1; i < len(p); i++ {
		a, b := p[i-1], p[i]
		if abs := (a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y); abs != 1 {
			t.Fatalf("в пути соседние клетки %v и %v не связаны ребром", a, b)
		}
		lo := a
		if b.X < a.X || b.Y < a.Y {
			lo = b
		}
		sum += 1 + (lo.X*7+lo.Y*13)%5
	}
	if sum != d {
		t.Fatalf("вес пути по рёбрам = %d, а ShortestPath сообщил %d", sum, d)
	}
	if el := time.Since(start); el > 2*time.Second {
		t.Fatalf("22 500 вершин обработаны за %v — нужна куча, а не поиск минимума перебором", el)
	}
}
