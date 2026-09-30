package main

type chkServer struct {
	pages map[string]Page[int]
	fail  map[string]error
	calls []string
}

func (s *chkServer) fetch(ctx context.Context, cursor string) (Page[int], error) {
	s.calls = append(s.calls, cursor)
	if len(s.calls) > 50 {
		return Page[int]{}, errors.New("больше 50 запросов — обход не останавливается")
	}
	if err := s.fail[cursor]; err != nil {
		return Page[int]{}, err
	}
	p, ok := s.pages[cursor]
	if !ok {
		return Page[int]{}, fmt.Errorf("нет страницы %q", cursor)
	}
	return p, nil
}

func chkCollect(seq iter.Seq2[int, error]) (items []int, errs []error) {
	for v, err := range seq {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		items = append(items, v)
	}
	return items, errs
}

func chkBasic() *chkServer {
	return &chkServer{pages: map[string]Page[int]{
		"":   {Items: []int{1, 2}, Next: "p2"},
		"p2": {Items: nil, Next: "p3"},
		"p3": {Items: []int{3}, Next: ""},
	}}
}

func TestAllPages(t *testing.T) {
	s := chkBasic()
	items, errs := chkCollect(All(context.Background(), s.fetch))
	if len(errs) != 0 {
		t.Fatalf("ошибки %v, ожидали без ошибок", errs)
	}
	if !reflect.DeepEqual(items, []int{1, 2, 3}) {
		t.Fatalf("элементы %v, ожидали [1 2 3] — пустая страница с Next не конец", items)
	}
	if !reflect.DeepEqual(s.calls, []string{"", "p2", "p3"}) {
		t.Fatalf("запрошены курсоры %q, ожидали [\"\" p2 p3]", s.calls)
	}
}

func TestAllLazyBreak(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("паника после break: %v — yield вызывают после false", p)
		}
	}()
	s := chkBasic()
	n := 0
	for range All(context.Background(), s.fetch) {
		n++
		if n == 2 {
			break
		}
	}
	if len(s.calls) != 1 {
		t.Fatalf("break на последнем элементе первой страницы, а запросов %d (%q), ожидали 1 — следующую страницу не надо просить заранее", len(s.calls), s.calls)
	}
}

func TestAllTwice(t *testing.T) {
	s := chkBasic()
	seq := All(context.Background(), s.fetch)
	a, _ := chkCollect(seq)
	b, errs := chkCollect(seq)
	if !reflect.DeepEqual(a, []int{1, 2, 3}) || !reflect.DeepEqual(b, []int{1, 2, 3}) || len(errs) != 0 {
		t.Fatalf("два обхода одного результата: %v и %v (ошибки %v), ожидали оба раза [1 2 3] — курсор захвачен один на все обходы?", a, b, errs)
	}
}

func TestAllFetchError(t *testing.T) {
	boom := errors.New("503")
	s := chkBasic()
	s.fail = map[string]error{"p2": boom}
	items, errs := chkCollect(All(context.Background(), s.fetch))
	if !reflect.DeepEqual(items, []int{1, 2}) {
		t.Fatalf("элементы до ошибки %v, ожидали [1 2]", items)
	}
	if len(errs) != 1 {
		t.Fatalf("ошибок %d (%v), ожидали ровно одну, после неё обход заканчивается", len(errs), errs)
	}
	if !errors.Is(errs[0], boom) || !strings.Contains(errs[0].Error(), "p2") {
		t.Fatalf("ошибка %q: ожидали обёртку над исходной (%%w) с курсором p2 в тексте", errs[0])
	}
	if len(s.calls) != 2 {
		t.Fatalf("после ошибки ещё запросы: %q", s.calls)
	}
}

func TestAllCursorLoop(t *testing.T) {
	s := &chkServer{pages: map[string]Page[int]{
		"":  {Items: []int{1}, Next: "a"},
		"a": {Items: []int{2}, Next: "b"},
		"b": {Items: []int{3}, Next: "a"},
	}}
	items, errs := chkCollect(All(context.Background(), s.fetch))
	if !reflect.DeepEqual(items, []int{1, 2, 3}) || len(errs) != 1 || !errors.Is(errs[0], ErrCursorLoop) {
		t.Fatalf("сервер вернул курсор a повторно: элементы %v, ошибки %v; ожидали [1 2 3] и одну ошибку ErrCursorLoop", items, errs)
	}
	if len(s.calls) != 3 {
		t.Fatalf("запросы %q: по курсору a второй раз ходить не надо", s.calls)
	}
}

func TestAllContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := chkBasic()
	var got []error
	n := 0
	for _, err := range All(ctx, s.fetch) {
		if err != nil {
			got = append(got, err)
			continue
		}
		n++
		if n == 2 {
			cancel()
		}
	}
	if len(got) != 1 || !errors.Is(got[0], context.Canceled) {
		t.Fatalf("после отмены ошибки %v, ожидали одну context.Canceled", got)
	}
	if len(s.calls) != 1 {
		t.Fatalf("после отмены контекста запросили %q — перед запросом надо проверить ctx", s.calls)
	}
}
