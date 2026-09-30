package main

// chkWeb — фейковый интернет: страница → ссылки. Нет в карте — ошибка 404.
func chkWeb(pages map[string][]string, calls *sync.Map, active, peak *atomic.Int32) Fetcher {
	return func(ctx context.Context, url string) ([]string, error) {
		if calls != nil {
			n, _ := calls.LoadOrStore(url, new(atomic.Int32))
			n.(*atomic.Int32).Add(1)
		}
		if active != nil {
			n := active.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			active.Add(-1)
		}
		links, ok := pages[url]
		if !ok {
			return nil, errors.New("404 " + url)
		}
		return links, nil
	}
}

func TestCrawlDedupAndErrors(t *testing.T) {
	web := map[string][]string{
		"/":  {"/a", "/b", "/missing"},
		"/a": {"/", "/b", "/c"},
		"/b": {"/a", "/c"},
		"/c": {"/"},
	}
	var calls sync.Map
	got, err := Crawl(context.Background(), "/", 5, 3, chkWeb(web, &calls, nil, nil))
	if err != nil {
		t.Fatalf("Crawl вернул ошибку %v, ожидали nil — ошибки страниц не прерывают обход", err)
	}
	if want := []string{"/", "/a", "/b", "/c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Crawl = %v, ожидали %v", got, want)
	}
	calls.Range(func(k, v any) bool {
		if n := v.(*atomic.Int32).Load(); n != 1 {
			t.Errorf("страница %v загружена %d раз, ожидали один", k, n)
		}
		return true
	})
}

func TestCrawlDepthByShortestPath(t *testing.T) {
	// /c достижима за 1 шаг напрямую и за 2 через /b; /d — сразу за /c.
	web := map[string][]string{
		"/":  {"/b", "/c"},
		"/b": {"/c"},
		"/c": {"/d"},
		"/d": {"/e"},
		"/e": {},
	}
	got, _ := Crawl(context.Background(), "/", 2, 2, chkWeb(web, nil, nil, nil))
	if want := []string{"/", "/b", "/c", "/d"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("maxDepth=2: Crawl = %v, ожидали %v", got, want)
	}
	got, _ = Crawl(context.Background(), "/", 0, 2, chkWeb(web, nil, nil, nil))
	if want := []string{"/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("maxDepth=0: Crawl = %v, ожидали %v", got, want)
	}
}

func TestCrawlWorkerLimit(t *testing.T) {
	web := map[string][]string{"/": nil}
	for i := range 30 {
		p := fmt.Sprintf("/p%d", i)
		web["/"] = append(web["/"], p)
		web[p] = []string{"/"}
	}
	var active, peak atomic.Int32
	got, _ := Crawl(context.Background(), "/", 3, 4, chkWeb(web, nil, &active, &peak))
	if len(got) != 31 {
		t.Fatalf("загружено %d страниц, ожидали 31", len(got))
	}
	if p := peak.Load(); p > 4 || p < 2 {
		t.Fatalf("одновременно шло %d загрузок при workers=4, ожидали от 2 до 4", p)
	}
}

func TestCrawlCancel(t *testing.T) {
	base := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	var started atomic.Int32
	fetch := func(ctx context.Context, url string) ([]string, error) {
		started.Add(1)
		if url == "/" {
			return []string{"/1", "/2", "/3", "/4", "/5", "/6"}, nil
		}
		<-ctx.Done() // «зависший» сервер отвечает только на отмену
		return nil, ctx.Err()
	}
	done := make(chan error, 1)
	go func() { _, err := Crawl(ctx, "/", 3, 2, fetch); done <- err }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("после отмены Crawl вернул %v, ожидали context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("после отмены ctx Crawl не вернулся")
	}
	if n := started.Load(); n > 3 {
		t.Fatalf("после отмены начато %d загрузок, ожидали не больше 3 (корень + workers)", n)
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine() - base; n > 0 {
		t.Fatalf("после Crawl осталось %d лишних горутин", n)
	}
}
