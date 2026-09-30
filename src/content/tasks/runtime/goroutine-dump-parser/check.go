package main

const chkDump = `goroutine 1 [running]:
main.main()
	/app/main.go:10 +0x1d

goroutine 7 [chan receive, 3 minutes]:
main.(*Server).worker(0xc000010000, {0x1, 0x2})
	/app/server.go:20 +0x25
created by main.(*Server).Start in goroutine 1
	/app/server.go:15 +0x3e

goroutine 9 [chan receive, 12 minutes]:
main.(*Server).worker(0xc000010000, {0x1, 0x2})
	/app/server.go:20 +0x25
created by main.(*Server).Start in goroutine 1
	/app/server.go:15 +0x3e

goroutine 8 [select, 1 minutes, locked to thread]:
net/http.(*persistConn).writeLoop(0xc0000a2000)
	/usr/local/go/src/net/http/transport.go:2421 +0xf0
created by net/http.(*Transport).dialConn
	/usr/local/go/src/net/http/transport.go:1777 +0x16f1

goroutine 21 [chan receive (nil chan)]:
main.stuck[...](...)
	/app/gen.go:5
main.main.func1()
	/app/main.go:30 +0x1a
created by main.main in goroutine 1
	/app/main.go:29 +0x2b
`

func TestParseStacksSample(t *testing.T) {
	gs, err := ParseStacks(chkDump)
	if err != nil {
		t.Fatalf("ошибка разбора: %v", err)
	}
	want := []Goroutine{
		{1, "running", 0, "main.main", ""},
		{7, "chan receive", 3 * time.Minute, "main.(*Server).worker", "main.(*Server).Start"},
		{9, "chan receive", 12 * time.Minute, "main.(*Server).worker", "main.(*Server).Start"},
		{8, "select", time.Minute, "net/http.(*persistConn).writeLoop", "net/http.(*Transport).dialConn"},
		{21, "chan receive (nil chan)", 0, "main.stuck[...]", "main.main"},
	}
	if len(gs) != len(want) {
		t.Fatalf("разобрано %d горутин, ожидали %d: %+v", len(gs), len(want), gs)
	}
	for i := range want {
		if gs[i] != want[i] {
			t.Fatalf("горутина %d:\n получили %+v\n ожидали  %+v", i, gs[i], want[i])
		}
	}
}

func TestParseStacksBad(t *testing.T) {
	for _, d := range []string{"goroutine x [running]:\nmain.main()", "gorutine 1 [running]:\nmain.main()", "goroutine 1 running:\nmain.main()"} {
		if _, err := ParseStacks(d); err == nil {
			t.Fatalf("некорректный заголовок %q разобран без ошибки", strings.SplitN(d, "\n", 2)[0])
		}
	}
}

func TestGroupStacks(t *testing.T) {
	gs, _ := ParseStacks(chkDump)
	got := GroupStacks(gs)
	if len(got) != 4 {
		t.Fatalf("групп %d, ожидали 4: %+v", len(got), got)
	}
	if got[0].Top != "main.(*Server).worker" || !slices.Equal(got[0].IDs, []int{7, 9}) {
		t.Fatalf("первая группа %+v, ожидали worker с IDs [7 9]", got[0])
	}
	var order []string
	for _, g := range got[1:] {
		order = append(order, g.State)
	}
	if !slices.Equal(order, []string{"chan receive (nil chan)", "running", "select"}) {
		t.Fatalf("группы по одной горутине должны идти по State: %v", order)
	}
}

func chkParked(ch chan int) { <-ch }

func TestParseRealDump(t *testing.T) {
	ch := make(chan int)
	for i := 0; i < 3; i++ {
		go chkParked(ch)
	}
	defer close(ch)
	time.Sleep(10 * time.Millisecond)
	buf := make([]byte, 1<<20)
	gs, err := ParseStacks(string(buf[:runtime.Stack(buf, true)]))
	if err != nil {
		t.Fatalf("настоящий дамп runtime.Stack не разобран: %v", err)
	}
	found := false
	for _, g := range GroupStacks(gs) {
		if strings.HasSuffix(g.Top, ".chkParked") {
			found = true
			if g.State != "chan receive" || len(g.IDs) != 3 {
				t.Fatalf("группа chkParked: %+v, ожидали 3 горутины в chan receive", g)
			}
		}
	}
	if !found {
		t.Fatalf("в настоящем дампе не найдены горутины chkParked")
	}
}
