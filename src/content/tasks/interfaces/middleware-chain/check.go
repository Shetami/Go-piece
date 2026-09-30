package main

type chkNotFound struct{ what string }

func (e *chkNotFound) Error() string { return e.what + " not found" }
func (e *chkNotFound) Status() int   { return 404 }

func chkTrace(log *[]string, name string) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, r *Request) (*Response, error) {
			*log = append(*log, name+">")
			resp, err := next.Serve(ctx, r)
			*log = append(*log, "<"+name)
			return resp, err
		})
	}
}

func chkOK(log *[]string) Handler {
	return HandlerFunc(func(ctx context.Context, r *Request) (*Response, error) {
		*log = append(*log, "h")
		return &Response{Status: 200, Body: "hi " + r.User}, nil
	})
}

func TestChainOrder(t *testing.T) {
	var log []string
	var _ Handler = HandlerFunc(nil)
	h := Chain(chkOK(&log), chkTrace(&log, "a"), chkTrace(&log, "b"), chkTrace(&log, "c"))
	resp, err := h.Serve(context.Background(), &Request{User: "ann"})
	got := strings.Join(log, " ")
	if want := "a> b> c> h <c <b <a"; got != want {
		t.Fatalf("порядок вызовов %q, ожидали %q: первый middleware — внешний", got, want)
	}
	if err != nil || resp == nil || resp.Body != "hi ann" {
		t.Fatalf("ответ %+v, %v", resp, err)
	}
	log = nil
	Chain(chkOK(&log)).Serve(context.Background(), &Request{})
	if strings.Join(log, " ") != "h" {
		t.Fatalf("Chain без middleware: %q, ожидали только обработчик", log)
	}
}

func TestRecover(t *testing.T) {
	panicky := HandlerFunc(func(context.Context, *Request) (*Response, error) { panic("boom") })
	resp, err := Chain(panicky, Recover()).Serve(context.Background(), &Request{})
	if resp != nil || !errors.Is(err, ErrPanic) || !strings.Contains(fmt.Sprint(err), "boom") {
		t.Fatalf("паника строкой: %+v, %v; ожидали nil и ошибку ErrPanic с текстом boom", resp, err)
	}
	io := errors.New("io timeout")
	panicErr := HandlerFunc(func(context.Context, *Request) (*Response, error) { panic(io) })
	_, err = Chain(panicErr, Recover()).Serve(context.Background(), &Request{})
	if !errors.Is(err, ErrPanic) || !errors.Is(err, io) {
		t.Fatalf("паника ошибкой: %v; errors.Is должен находить и ErrPanic, и саму ошибку", err)
	}
	plain := errors.New("plain")
	failing := HandlerFunc(func(context.Context, *Request) (*Response, error) { return nil, plain })
	if _, err = Chain(failing, Recover()).Serve(context.Background(), &Request{}); err != plain {
		t.Fatalf("обычная ошибка через Recover: %v, ожидали её же без изменений", err)
	}
}

func TestErrorsStatus(t *testing.T) {
	failing := HandlerFunc(func(context.Context, *Request) (*Response, error) {
		return nil, fmt.Errorf("load profile: %w", &chkNotFound{"user"})
	})
	resp, err := Chain(failing, Errors()).Serve(context.Background(), &Request{})
	if err != nil || resp == nil || resp.Status != 404 || resp.Body != "load profile: user not found" {
		t.Fatalf("обёрнутая StatusError: %+v, %v; ожидали {404 \"load profile: user not found\"}, nil", resp, err)
	}
	failing = HandlerFunc(func(context.Context, *Request) (*Response, error) { return nil, errors.New("db down") })
	resp, _ = Chain(failing, Errors()).Serve(context.Background(), &Request{})
	if resp == nil || resp.Status != 500 || resp.Body != "db down" {
		t.Fatalf("обычная ошибка: %+v; ожидали {500 \"db down\"}", resp)
	}
	empty := HandlerFunc(func(context.Context, *Request) (*Response, error) { return nil, nil })
	resp, err = Chain(empty, Errors()).Serve(context.Background(), &Request{})
	if err != nil || resp == nil || resp.Status != 500 || resp.Body != "empty response" {
		t.Fatalf("(nil, nil) от обработчика: %+v, %v; ожидали {500 \"empty response\"}", resp, err)
	}
}

func TestFullStack(t *testing.T) {
	var log []string
	h := Chain(chkOK(&log), Errors(), Recover(), RequireUser())
	resp, err := h.Serve(context.Background(), &Request{})
	if err != nil || resp == nil || resp.Status != 401 || len(log) != 0 {
		t.Fatalf("без пользователя: %+v, %v, обработчик вызван=%v; ожидали 401 и невызванный обработчик", resp, err, len(log) > 0)
	}
	var se StatusError
	if _, err := RequireUser()(chkOK(&log)).Serve(context.Background(), &Request{}); !errors.As(err, &se) || err.Error() != "unauthorized" {
		t.Fatalf("RequireUser вернул %v; ожидали StatusError с текстом unauthorized", err)
	}
	panicky := HandlerFunc(func(context.Context, *Request) (*Response, error) { panic("oops") })
	resp, err = Chain(panicky, Errors(), Recover()).Serve(context.Background(), &Request{User: "x"})
	if err != nil || resp == nil || resp.Status != 500 || !strings.Contains(resp.Body, "oops") {
		t.Fatalf("паника за Errors+Recover: %+v, %v; ожидали ответ 500 с текстом паники", resp, err)
	}
}
