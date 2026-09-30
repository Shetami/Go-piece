package main

// chkTree строит полное дерево: depth уровней, у каждого узла fan детей.
func chkTree(depth, fan int, next *int) *Node {
	*next++
	n := &Node{Val: *next}
	if depth > 1 {
		for range fan {
			n.Children = append(n.Children, chkTree(depth-1, fan, next))
		}
	}
	return n
}

func chkSumTimeout(t *testing.T, root *Node, limit int, w func(*Node) int) int {
	t.Helper()
	res := make(chan int, 1)
	go func() { res <- Sum(root, limit, w) }()
	select {
	case s := <-res:
		return s
	case <-time.After(3 * time.Second):
		t.Fatal("Sum завис — похоже на дедлок на семафоре в рекурсии")
		return 0
	}
}

func TestSumBasic(t *testing.T) {
	val := func(n *Node) int { return n.Val }
	if s := chkSumTimeout(t, nil, 2, val); s != 0 {
		t.Fatalf("Sum(nil) = %d, ожидали 0", s)
	}
	next := 0
	root := chkTree(4, 3, &next) // 1+3+9+27 = 40 узлов, значения 1..40
	if s := chkSumTimeout(t, root, 4, val); s != 820 {
		t.Fatalf("Sum = %d, ожидали 820", s)
	}
	if s := chkSumTimeout(t, root, 1, val); s != 820 {
		t.Fatalf("Sum с limit=1 = %d, ожидали 820", s)
	}
}

func TestSumLimitAndParallel(t *testing.T) {
	var active, peak atomic.Int32
	slow := func(n *Node) int {
		a := active.Add(1)
		for {
			p := peak.Load()
			if a <= p || peak.CompareAndSwap(p, a) {
				break
			}
		}
		time.Sleep(2 * time.Millisecond)
		active.Add(-1)
		return 1
	}
	next := 0
	root := chkTree(4, 4, &next) // 85 узлов
	if s := chkSumTimeout(t, root, 3, slow); s != 85 {
		t.Fatalf("Sum = %d, ожидали 85", s)
	}
	if p := peak.Load(); p > 3 {
		t.Fatalf("одновременно работало %d вызовов weight при limit=3", p)
	}
	if p := peak.Load(); p < 2 {
		t.Fatal("вызовы weight ни разу не шли параллельно")
	}
}

func TestSumDeepAndWide(t *testing.T) {
	one := func(*Node) int { return 1 }
	chain := &Node{}
	cur := chain
	for range 5000 {
		c := &Node{}
		cur.Children = []*Node{c}
		cur = c
	}
	if s := chkSumTimeout(t, chain, 2, one); s != 5001 {
		t.Fatalf("цепочка: Sum = %d, ожидали 5001", s)
	}
	wide := &Node{}
	for range 5000 {
		wide.Children = append(wide.Children, &Node{Children: []*Node{{}}})
	}
	base := runtime.NumGoroutine()
	var maxG atomic.Int32
	count := func(*Node) int {
		if g := int32(runtime.NumGoroutine()); g > maxG.Load() {
			maxG.Store(g)
		}
		return 1
	}
	if s := chkSumTimeout(t, wide, 4, count); s != 10001 {
		t.Fatalf("широкое дерево: Sum = %d, ожидали 10001", s)
	}
	if extra := int(maxG.Load()) - base; extra > 500 {
		t.Fatalf("при limit=4 одновременно жило %d лишних горутин на 10001 узел — горутина не должна создаваться на каждый узел", extra)
	}
}
