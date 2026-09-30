package main

func chkN(name string, kids ...*Node) *Node { return &Node{Name: name, Children: kids} }

func chkTree() *Node {
	return chkN("repo",
		chkN("cmd", chkN("api", chkN("main.go")), chkN("worker")),
		chkN("node_modules", chkN("left-pad", chkN("index.js"))),
		chkN("internal", chkN("db", chkN("pg.go"), chkN("tx.go"))),
		chkN("go.mod"),
	)
}

func TestWalkPaths(t *testing.T) {
	var paths [][]string
	for p := range Walk(chkTree(), nil) {
		paths = append(paths, p)
	}
	var got []string
	for _, p := range paths {
		got = append(got, strings.Join(p, "/"))
	}
	want := []string{
		"repo", "repo/cmd", "repo/cmd/api", "repo/cmd/api/main.go", "repo/cmd/worker",
		"repo/node_modules", "repo/node_modules/left-pad", "repo/node_modules/left-pad/index.js",
		"repo/internal", "repo/internal/db", "repo/internal/db/pg.go", "repo/internal/db/tx.go", "repo/go.mod",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("пути, сохранённые потребителем:\n%q\nожидали:\n%q\n(если пути перепутаны — отданные слайсы делят один массив)", got, want)
	}
}

func TestWalkSkip(t *testing.T) {
	skipCalls := map[string]int{}
	var got []string
	for p, n := range Walk(chkTree(), func(n *Node) bool {
		skipCalls[n.Name]++
		return n.Name == "node_modules" || strings.HasSuffix(n.Name, ".mod")
	}) {
		if n.Name != p[len(p)-1] {
			t.Fatalf("путь %q не заканчивается узлом %q", p, n.Name)
		}
		got = append(got, n.Name)
	}
	want := []string{"repo", "cmd", "api", "main.go", "worker", "internal", "db", "pg.go", "tx.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("с пропуском node_modules и *.mod: %v, ожидали %v", got, want)
	}
	if skipCalls["left-pad"] != 0 {
		t.Fatal("в поддерево пропущенного узла заходить не нужно — skip для left-pad не должен вызываться")
	}
	for range Walk(nil, nil) {
		t.Fatal("nil-корень — пустой обход")
	}
}

func TestWalkBreakDeep(t *testing.T) {
	visited := 0
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("паника после break: %v — false от yield должен подниматься через всю рекурсию", p)
		}
	}()
	for _, n := range Walk(chkTree(), func(*Node) bool { visited++; return false }) {
		if n.Name == "main.go" {
			break
		}
	}
	if visited != 4 {
		t.Fatalf("после break на main.go обход продолжился: посещено %d узлов, ожидали 4", visited)
	}
}

func TestWalkDeep(t *testing.T) {
	root := chkN("0")
	cur := root
	for i := 1; i < 5000; i++ {
		next := chkN(strconv.Itoa(i))
		cur.Children = []*Node{next}
		cur = next
	}
	n, last := 0, 0
	for p := range Walk(root, nil) {
		n++
		last = len(p)
	}
	if n != 5000 || last != 5000 {
		t.Fatalf("цепочка глубиной 5000: отдано %d узлов, длина последнего пути %d; ожидали 5000 и 5000", n, last)
	}
}
