package main

var (
	chkErrWrite = errors.New("нет места")
	chkErrClose = errors.New("сбой сброса буфера")
)

type chkWC struct {
	buf    bytes.Buffer
	short  bool
	werr   error
	cerr   error
	closes int
}

func (w *chkWC) Write(p []byte) (int, error) {
	if w.werr != nil {
		return 0, w.werr
	}
	if w.short {
		p = p[:len(p)/2]
	}
	return w.buf.Write(p)
}

func (w *chkWC) Close() error { w.closes++; return w.cerr }

func TestSaveOK(t *testing.T) {
	w := &chkWC{}
	if err := Save(w, []byte("привет")); err != nil || w.buf.String() != "привет" || w.closes != 1 {
		t.Fatalf("успех: err=%v данные=%q закрытий=%d; ожидали nil, «привет», 1", err, w.buf.String(), w.closes)
	}
}

func TestSaveWriteFails(t *testing.T) {
	w := &chkWC{werr: chkErrWrite}
	err := Save(w, []byte("x"))
	if w.closes != 1 {
		t.Fatalf("Close должен вызываться и при ошибке записи, вызван %d раз", w.closes)
	}
	if err == nil || err.Error() != "запись: нет места" || !errors.Is(err, chkErrWrite) {
		t.Fatalf("ожидали «запись: нет места», получили %v", err)
	}
}

func TestSaveCloseFails(t *testing.T) {
	w := &chkWC{cerr: chkErrClose}
	err := Save(w, []byte("x"))
	if err == nil || err.Error() != "закрытие: сбой сброса буфера" || !errors.Is(err, chkErrClose) {
		t.Fatalf("данные не дошли — ошибку Close терять нельзя: ожидали «закрытие: сбой сброса буфера», получили %v", err)
	}
	if w.closes != 1 {
		t.Fatalf("Close вызван %d раз, ожидали ровно 1", w.closes)
	}
}

func TestSaveBothFail(t *testing.T) {
	w := &chkWC{werr: chkErrWrite, cerr: chkErrClose}
	err := Save(w, []byte("x"))
	want := "запись: нет места\nзакрытие: сбой сброса буфера"
	if err == nil || err.Error() != want || !errors.Is(err, chkErrWrite) || !errors.Is(err, chkErrClose) {
		t.Fatalf("упали оба: ожидали %q и обе ошибки через errors.Is, получили %v", want, err)
	}
}

func TestSaveShortWrite(t *testing.T) {
	w := &chkWC{short: true}
	if err := Save(w, []byte("abcd")); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("короткая запись без ошибки: ожидали io.ErrShortWrite, получили %v", err)
	}
}
