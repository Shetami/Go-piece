package main

// chkNext возвращает "файл:строка" для строки, следующей за вызовом.
func chkNext() string {
	_, file, line, _ := runtime.Caller(1)
	return filepath.Base(file) + ":" + strconv.Itoa(line+1)
}

func chkAssert(l *Logger, msg string) {
	l.Helper()
	l.Log(msg)
}

func chkAssertf(l *Logger, v int) {
	l.Helper()
	l.Logf("v=%d", v)
}

func chkOuter(l *Logger) {
	l.Helper()
	chkAssert(l, "nested")
}

var chkPlainAt string

func chkPlain(l *Logger) {
	chkPlainAt = chkNext()
	l.Log("plain")
}

func chkLastLine(buf *bytes.Buffer) string {
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	return lines[len(lines)-1]
}

func TestCallerDirect(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf)
	at := chkNext()
	l.Log("привет")
	if got, want := buf.String(), at+": привет\n"; got != want {
		t.Fatalf("Log записал %q, ожидали %q", got, want)
	}
	buf.Reset()
	at = chkNext()
	l.Logf("%d+%d", 2, 3)
	if got, want := buf.String(), at+": 2+3\n"; got != want {
		t.Fatalf("Logf записал %q, ожидали %q (место вызова Logf, а не внутренности логгера)", got, want)
	}
}

func TestCallerHelpers(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf)
	at := chkNext()
	chkAssert(l, "a")
	if got, want := chkLastLine(&buf), at+": a"; got != want {
		t.Fatalf("вызов через Helper-функцию: %q, ожидали %q", got, want)
	}
	at = chkNext()
	chkAssertf(l, 7)
	if got, want := chkLastLine(&buf), at+": v=7"; got != want {
		t.Fatalf("Logf через Helper-функцию: %q, ожидали %q", got, want)
	}
	at = chkNext()
	chkOuter(l)
	if got, want := chkLastLine(&buf), at+": nested"; got != want {
		t.Fatalf("две вложенные Helper-функции: %q, ожидали %q", got, want)
	}
	chkPlain(l)
	if got, want := chkLastLine(&buf), chkPlainAt+": plain"; got != want {
		t.Fatalf("функция без Helper: %q, ожидали %q", got, want)
	}
}

func TestCallerConcurrent(t *testing.T) {
	var buf bytes.Buffer
	var mu sync.Mutex
	w := chkLockedWriter{mu: &mu, buf: &buf}
	l := NewLogger(w)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				chkAssert(l, "конкурентно")
				l.Log("напрямую")
			}
		}()
	}
	wg.Wait()
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 800 {
		t.Fatalf("записано %d строк, ожидали 800", len(lines))
	}
	for _, s := range lines {
		if !strings.Contains(s, ".go:") || !(strings.HasSuffix(s, ": конкурентно") || strings.HasSuffix(s, ": напрямую")) {
			t.Fatalf("испорченная строка %q", s)
		}
	}
}

type chkLockedWriter struct {
	mu  *sync.Mutex
	buf *bytes.Buffer
}

func (w chkLockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !bytes.HasSuffix(p, []byte("\n")) || bytes.Count(p, []byte("\n")) != 1 {
		return 0, errors.New("строка должна приходить одним Write")
	}
	return w.buf.Write(p)
}
