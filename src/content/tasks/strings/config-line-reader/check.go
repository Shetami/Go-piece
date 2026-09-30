package main

func chkParseCfg(t *testing.T, in string) []Directive {
	t.Helper()
	got, err := ParseConfig(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ParseConfig(%q): неожиданная ошибка %v", in, err)
	}
	return got
}

func chkDirEq(t *testing.T, got []Directive, want []Directive) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("получили %d директив %+v, ожидали %d: %+v", len(got), got, len(want), want)
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.Line != w.Line || g.Key != w.Key || !slices.Equal(g.Args, w.Args) || len(g.Args) != len(w.Args) {
			t.Fatalf("директива %d = %+v (args %q), ожидали %+v (args %q)", i, g, g.Args, w, w.Args)
		}
	}
}

func TestConfigBasic(t *testing.T) {
	in := "# сервер\n\nlisten 0.0.0.0\t8080   # порт\r\n  name \"мой сервис\"  \r\n\n   \nroot /srv\ttitle \"a # не комментарий\" \"\"\nflag"
	chkDirEq(t, chkParseCfg(t, in), []Directive{
		{Line: 3, Key: "listen", Args: []string{"0.0.0.0", "8080"}},
		{Line: 4, Key: "name", Args: []string{"мой сервис"}},
		{Line: 7, Key: "root", Args: []string{"/srv", "title", "a # не комментарий", ""}},
		{Line: 8, Key: "flag"},
	})
}

func TestConfigContinuation(t *testing.T) {
	in := "a 1\nallow 10.0.0.1 \\\n  10.0.0.2 \\\r\n  10.0.0.3\nb 2\nc x\\\n"
	chkDirEq(t, chkParseCfg(t, in), []Directive{
		{Line: 1, Key: "a", Args: []string{"1"}},
		{Line: 2, Key: "allow", Args: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}},
		{Line: 5, Key: "b", Args: []string{"2"}},
		{Line: 6, Key: "c", Args: []string{"x"}},
	})
}

func TestConfigLongLine(t *testing.T) {
	long := strings.Repeat("я", 100_000) // 200 КБ — больше буфера bufio.Scanner по умолчанию
	got, err := ParseConfig(strings.NewReader("key " + long + "\nnext 1\n"))
	if err != nil {
		t.Fatalf("строка в 200 КБ: неожиданная ошибка %v", err)
	}
	if len(got) != 2 || len(got[0].Args) != 1 || got[0].Args[0] != long || got[1].Line != 2 {
		t.Fatalf("длинная строка разобрана неверно: %d директив", len(got))
	}
}

func TestConfigUnclosedQuote(t *testing.T) {
	in := "a 1\n\nmsg \"привет \\\n  мир\nb 2\n"
	got, err := ParseConfig(strings.NewReader(in))
	var le *LineError
	if !errors.As(err, &le) || le.Line != 3 || !errors.Is(err, ErrUnclosedQuote) || got != nil {
		t.Fatalf("ParseConfig(%q) = %v, %v; ожидали nil и *LineError{Line: 3} с ErrUnclosedQuote", in, got, err)
	}
}

type chkBrokenReader struct {
	data string
	err  error
}

func (r *chkBrokenReader) Read(p []byte) (int, error) {
	if r.data == "" {
		return 0, r.err
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func TestConfigReadError(t *testing.T) {
	disk := errors.New("ошибка диска")
	got, err := ParseConfig(&chkBrokenReader{data: "a 1\nb 2\nc", err: disk})
	if !errors.Is(err, disk) || got != nil {
		t.Fatalf("ParseConfig с ломающимся reader = %v, %v; ожидали nil и ошибку чтения", got, err)
	}
}
