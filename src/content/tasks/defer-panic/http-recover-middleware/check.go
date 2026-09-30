package main

func chkServe(h http.Handler) (rec *httptest.ResponseRecorder, logs []string, panicked any) {
	rec = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/orders/42", nil)
	mw := Recover(func(s string) { logs = append(logs, s) }, h)
	func() {
		defer func() { panicked = recover() }()
		mw.ServeHTTP(rec, req)
	}()
	return
}

func TestRecoverPassThrough(t *testing.T) {
	rec, logs, p := chkServe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, "ok")
	}))
	if p != nil || rec.Code != 201 || rec.Body.String() != "ok" || len(logs) != 0 {
		t.Fatalf("без паники: код %d, тело %q, логов %d, паника %v; ожидали 201, \"ok\", 0, nil", rec.Code, rec.Body.String(), len(logs), p)
	}
}

func TestRecoverBeforeWrite(t *testing.T) {
	rec, logs, p := chkServe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Trace", "abc")
		panic("нет такого заказа")
	}))
	if p != nil {
		t.Fatalf("паника вылетела из middleware: %v", p)
	}
	if rec.Code != 500 || !strings.Contains(rec.Body.String(), "internal error") {
		t.Fatalf("код %d, тело %q; ожидали 500 и \"internal error\"", rec.Code, rec.Body.String())
	}
	if len(logs) != 1 {
		t.Fatalf("ожидали одну запись в логе, получили %q", logs)
	}
	for _, s := range []string{"нет такого заказа", "POST", "/orders/42"} {
		if !strings.Contains(logs[0], s) {
			t.Fatalf("в логе %q нет %q", logs[0], s)
		}
	}
}

func TestRecoverAfterWrite(t *testing.T) {
	rec, logs, p := chkServe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "частичный ответ")
		panic("упали посреди ответа")
	}))
	if p != nil {
		t.Fatalf("паника вылетела из middleware: %v", p)
	}
	if rec.Code != 200 || rec.Body.String() != "частичный ответ" {
		t.Fatalf("код %d, тело %q; ответ уже начат — дописывать в него нельзя", rec.Code, rec.Body.String())
	}
	if len(logs) != 1 || !strings.Contains(logs[0], "упали посреди ответа") {
		t.Fatalf("паника после записи всё равно должна попасть в лог: %q", logs)
	}
}

func TestRecoverAfterWriteHeader(t *testing.T) {
	rec, _, _ := chkServe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		panic("после заголовка")
	}))
	if rec.Code != 202 || rec.Body.Len() != 0 {
		t.Fatalf("код %d, тело %q; после WriteHeader тело 500 писать нельзя", rec.Code, rec.Body.String())
	}
}

func TestRecoverAbortHandler(t *testing.T) {
	_, logs, p := chkServe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	if p != http.ErrAbortHandler {
		t.Fatalf("http.ErrAbortHandler должен лететь дальше, recover() = %v", p)
	}
	if len(logs) != 0 {
		t.Fatalf("ErrAbortHandler — не ошибка, логировать не надо: %q", logs)
	}
}
