package main

type chkRes struct {
	name string
	log  *[]string
	err  error
}

func (r *chkRes) Close() error { *r.log = append(*r.log, "close "+r.name); return r.err }

func chkOpener(name string, log *[]string, openErr, closeErr error) Opener {
	return Opener{Name: name, Open: func() (io.Closer, error) {
		if openErr != nil {
			return nil, openErr
		}
		*log = append(*log, "open "+name)
		return &chkRes{name: name, log: log, err: closeErr}, nil
	}}
}

func TestOpenAllSuccess(t *testing.T) {
	var log []string
	errCache := errors.New("кэш не сбросился")
	res, closeAll, err := OpenAll([]Opener{
		chkOpener("db", &log, nil, nil),
		chkOpener("cache", &log, nil, errCache),
		chkOpener("queue", &log, nil, nil),
	})
	if err != nil || len(res) != 3 || closeAll == nil {
		t.Fatalf("OpenAll: len(res)=%d closeAll=nil:%v err=%v, ожидали 3 ресурса и closeAll", len(res), closeAll == nil, err)
	}
	if r := res[0].(*chkRes); r.name != "db" {
		t.Fatalf("res[0] = %s, ресурсы должны идти в порядке openers", r.name)
	}
	cerr := closeAll()
	want := []string{"open db", "open cache", "open queue", "close queue", "close cache", "close db"}
	if !reflect.DeepEqual(log, want) {
		t.Fatalf("журнал %v,\nожидали %v", log, want)
	}
	if !errors.Is(cerr, errCache) || !strings.Contains(cerr.Error(), "close cache") {
		t.Fatalf("closeAll() = %v, ожидали ошибку \"close cache: ...\"", cerr)
	}
	if again := closeAll(); again != nil || len(log) != 6 {
		t.Fatalf("повторный closeAll: err=%v, журнал %v — ничего не должен закрывать", again, log)
	}
}

func TestOpenAllPartialFailure(t *testing.T) {
	var log []string
	errOpen := errors.New("очередь недоступна")
	errClose := errors.New("db не закрылась")
	res, closeAll, err := OpenAll([]Opener{
		chkOpener("db", &log, nil, errClose),
		chkOpener("cache", &log, nil, nil),
		chkOpener("queue", &log, errOpen, nil),
		chkOpener("never", &log, nil, nil),
	})
	if res != nil || closeAll != nil {
		t.Fatalf("при ошибке ожидали res=nil и closeAll=nil, получили %v, closeAll=nil:%v", res, closeAll == nil)
	}
	want := []string{"open db", "open cache", "close cache", "close db"}
	if !reflect.DeepEqual(log, want) {
		t.Fatalf("журнал %v,\nожидали %v — открытые закрыть в обратном порядке, дальше не открывать", log, want)
	}
	if !errors.Is(err, errOpen) || !errors.Is(err, errClose) {
		t.Fatalf("err = %v, ожидали и ошибку открытия, и ошибку закрытия", err)
	}
	if !strings.Contains(err.Error(), "open queue") || !strings.Contains(err.Error(), "close db") {
		t.Fatalf("err = %q, ожидали имена в формате \"open queue: ...\" и \"close db: ...\"", err)
	}
}

func TestOpenAllPanic(t *testing.T) {
	var log []string
	defer func() {
		if r := recover(); r != "драйвер упал" {
			t.Fatalf("паника должна пролететь с тем же значением, recover() = %v", r)
		}
		want := []string{"open db", "open cache", "close cache", "close db"}
		if !reflect.DeepEqual(log, want) {
			t.Fatalf("при панике журнал %v, ожидали %v", log, want)
		}
	}()
	OpenAll([]Opener{
		chkOpener("db", &log, nil, nil),
		chkOpener("cache", &log, nil, nil),
		{Name: "bad", Open: func() (io.Closer, error) { panic("драйвер упал") }},
	})
	t.Fatal("OpenAll проглотил панику")
}

func TestOpenAllEmpty(t *testing.T) {
	res, closeAll, err := OpenAll(nil)
	if err != nil || len(res) != 0 || closeAll == nil || closeAll() != nil {
		t.Fatalf("пустой список: ожидали пустой res, рабочий closeAll и nil")
	}
}
