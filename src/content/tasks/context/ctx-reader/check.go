package main

// chkSlow — «медленный» источник: каждый Read ждёт gate, потом пишет 'X'.
type chkSlow struct {
	gate  chan struct{}
	calls atomic.Int32
	wrote chan struct{}
}

func (s *chkSlow) Read(p []byte) (int, error) {
	s.calls.Add(1)
	<-s.gate
	for i := range p {
		p[i] = 'X'
	}
	close(s.wrote)
	return len(p), nil
}

func chkRead(r io.Reader, p []byte) (int, error, bool) {
	type res struct {
		n   int
		err error
	}
	ch := make(chan res, 1)
	go func() { n, err := r.Read(p); ch <- res{n, err} }()
	select {
	case x := <-ch:
		return x.n, x.err, true
	case <-time.After(2 * time.Second):
		return 0, nil, false
	}
}

func TestReaderPassThrough(t *testing.T) {
	data := strings.Repeat("данные ", 1000)
	got, err := io.ReadAll(NewReader(context.Background(), strings.NewReader(data)))
	if err != nil || string(got) != data {
		t.Fatalf("ReadAll: err = %v, прочитано %d байт из %d", err, len(got), len(data))
	}
}

func TestReaderEOFWithData(t *testing.T) {
	r := NewReader(context.Background(), iotest.DataErrReader(strings.NewReader("abc")))
	p := make([]byte, 10)
	n, err := r.Read(p)
	if n != 3 || err != io.EOF || string(p[:n]) != "abc" {
		t.Fatalf("Read = %d, %v, %q; ожидали 3, EOF, \"abc\" — данные и EOF вместе", n, err, p[:n])
	}
}

func TestReaderCancelWhileBlocked(t *testing.T) {
	why := errors.New("клиент отключился")
	s := &chkSlow{gate: make(chan struct{}), wrote: make(chan struct{})}
	ctx, cancel := context.WithCancelCause(context.Background())
	r := NewReader(ctx, s)
	go func() { time.Sleep(20 * time.Millisecond); cancel(why) }()
	p := []byte("........")
	n, err, ok := chkRead(r, p)
	if !ok {
		t.Fatalf("Read не прервался отменой контекста — висит вместе с источником")
	}
	if n != 0 || !errors.Is(err, why) {
		t.Fatalf("Read = %d, %v; ожидали 0 и причину отмены", n, err)
	}
	close(s.gate)
	<-s.wrote
	time.Sleep(10 * time.Millisecond)
	if string(p) != "........" {
		t.Fatalf("после возврата Read буфер вызывающего испорчен: %q — источник писал прямо в p", p)
	}
	if _, err, _ := chkRead(r, p); !errors.Is(err, why) || s.calls.Load() != 1 {
		t.Fatalf("следующий Read: err = %v, вызовов источника %d; ожидали причину отмены и без нового вызова", err, s.calls.Load())
	}
}

func TestReaderPreCanceled(t *testing.T) {
	s := &chkSlow{gate: make(chan struct{}), wrote: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for range 20 {
		if _, err, ok := chkRead(NewReader(ctx, s), make([]byte, 4)); !ok || !errors.Is(err, context.Canceled) {
			t.Fatalf("отменённый заранее контекст: err = %v (вернулся: %v), ожидали Canceled", err, ok)
		}
	}
	if s.calls.Load() != 0 {
		t.Fatalf("источник вызван %d раз при отменённом контексте", s.calls.Load())
	}
}
