package main

var chkErrNotFound = errors.New("404")

func TestFetchAllBodies(t *testing.T) {
	got, err := FetchAll([]string{"a", "b"}, func(u string) (string, error) { return "body-" + u, nil })
	if err != nil {
		t.Fatalf("ошибок не было, а FetchAll вернул %v", err)
	}
	want := map[string]string{"a": "body-a", "b": "body-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FetchAll = %v, ожидали %v", got, want)
	}
}

func TestFetchAllEmpty(t *testing.T) {
	got, err := FetchAll(nil, func(string) (string, error) { return "", nil })
	if got == nil || len(got) != 0 || err != nil {
		t.Fatalf("на пустом входе ожидали пустую (не nil) карту и nil, получили %v, %v", got, err)
	}
}

func TestFetchAllParallelAndDedup(t *testing.T) {
	// Каждая загрузка ждёт, пока не стартуют все три уникальных:
	// последовательный обход тут не дождётся и получит ошибку.
	var started, calls atomic.Int32
	fetch := func(u string) (string, error) {
		calls.Add(1)
		started.Add(1)
		deadline := time.Now().Add(time.Second)
		for started.Load() < 3 {
			if time.Now().After(deadline) {
				return "", errors.New("не дождались остальных загрузок")
			}
			time.Sleep(time.Millisecond)
		}
		return u, nil
	}
	got, err := FetchAll([]string{"x", "y", "x", "z", "y"}, fetch)
	if err != nil {
		t.Fatalf("загрузки должны идти одновременно, а получили ошибку: %v", err)
	}
	if n := calls.Load(); n != 3 {
		t.Fatalf("fetch вызван %d раз, ожидали 3 — повторяющиеся URL грузятся один раз", n)
	}
	if len(got) != 3 {
		t.Fatalf("ожидали 3 записи в карте, получили %v", got)
	}
}

func TestFetchAllErrorsInInputOrder(t *testing.T) {
	// "c" падает первым по времени, "a" — последним; порядок в ошибке — по входу.
	delay := map[string]time.Duration{"a": 60 * time.Millisecond, "b": 0, "c": 0}
	fetch := func(u string) (string, error) {
		time.Sleep(delay[u])
		if u == "b" {
			return "ok", nil
		}
		return "", chkErrNotFound
	}
	got, err := FetchAll([]string{"a", "b", "c", "a"}, fetch)
	if !errors.Is(err, chkErrNotFound) {
		t.Fatalf("ошибка должна оборачивать исходную (errors.Is), получили %v", err)
	}
	if want := "a: 404\nc: 404"; err.Error() != want {
		t.Fatalf("текст ошибки %q, ожидали %q", err.Error(), want)
	}
	if !reflect.DeepEqual(got, map[string]string{"b": "ok"}) {
		t.Fatalf("в карте должны быть только успешные загрузки, получили %v", got)
	}
}
