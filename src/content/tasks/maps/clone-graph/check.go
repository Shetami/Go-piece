package main

// chkIso проверяет, что копия изоморфна оригиналу той же биекцией
// и не делит с ним ни одной вершины.
func chkIso(t *testing.T, orig, cp *Node) {
	t.Helper()
	fwd, back := map[*Node]*Node{}, map[*Node]*Node{}
	var walk func(a, b *Node)
	walk = func(a, b *Node) {
		if (a == nil) != (b == nil) {
			t.Fatalf("nil-ребро не сохранилось")
		}
		if a == nil {
			return
		}
		if a == b {
			t.Fatalf("вершина %q не скопирована — копия ссылается на оригинал", a.Name)
		}
		if m, ok := fwd[a]; ok {
			if m != b {
				t.Fatalf("вершина %q скопирована дважды: общие ссылки в копии разошлись", a.Name)
			}
			return
		}
		if m, ok := back[b]; ok && m != a {
			t.Fatalf("две разные вершины оригинала (%q) склеились в одну копию", a.Name)
		}
		fwd[a], back[b] = b, a
		if a.Name != b.Name || len(a.Edges) != len(b.Edges) {
			t.Fatalf("вершина %q: копия %q с %d рёбрами, ожидали %d", a.Name, b.Name, len(b.Edges), len(a.Edges))
		}
		for i := range a.Edges {
			walk(a.Edges[i], b.Edges[i])
		}
	}
	walk(orig, cp)
}

func chkClone(t *testing.T, root *Node) {
	t.Helper()
	done := make(chan *Node, 1)
	go func() { done <- Clone(root) }()
	select {
	case cp := <-done:
		chkIso(t, root, cp)
	case <-time.After(3 * time.Second):
		t.Fatal("Clone не завершился — бесконечный обход цикла?")
	}
}

func TestCloneDiamond(t *testing.T) {
	d := &Node{Name: "d"}
	b := &Node{Name: "b", Edges: []*Node{d}}
	c := &Node{Name: "c", Edges: []*Node{d, d}}
	chkClone(t, &Node{Name: "a", Edges: []*Node{b, c}})
}

func TestCloneCycles(t *testing.T) {
	a := &Node{Name: "a"}
	b := &Node{Name: "b"}
	a.Edges = []*Node{b, a} // петля
	b.Edges = []*Node{a}    // цикл
	chkClone(t, a)
	cp := Clone(a)
	if cp.Edges[1] != cp || cp.Edges[0].Edges[0] != cp {
		t.Fatal("цикл и петля должны вести обратно в копию корня")
	}
}

func TestCloneSameNames(t *testing.T) {
	x1 := &Node{Name: "x"}
	x2 := &Node{Name: "x"}
	x1.Edges = []*Node{x2}
	chkClone(t, &Node{Name: "root", Edges: []*Node{x1, x2, nil}})
}

func TestCloneNilAndOriginal(t *testing.T) {
	if Clone(nil) != nil {
		t.Fatal("Clone(nil) должен вернуть nil")
	}
	b := &Node{Name: "b"}
	a := &Node{Name: "a", Edges: []*Node{b}}
	cp := Clone(a)
	cp.Edges[0].Name = "ИСПОРЧЕНО"
	cp.Edges[0] = nil
	if a.Edges[0] != b || b.Name != "b" {
		t.Fatal("правка копии изменила оригинал")
	}
}

func TestCloneLongChain(t *testing.T) {
	root := &Node{Name: "0"}
	cur := root
	for i := 1; i < 100000; i++ {
		n := &Node{Name: strconv.Itoa(i)}
		cur.Edges = []*Node{n, root}
		cur = n
	}
	cp := Clone(root)
	n := 0
	for c := cp; c != nil && len(c.Edges) > 0; c = c.Edges[0] {
		if c.Edges[1] != cp {
			t.Fatalf("на глубине %d обратное ребро ведёт не в копию корня", n)
		}
		n++
	}
	if n != 99999 {
		t.Fatalf("длина цепочки в копии %d, ожидали 99999", n)
	}
}
