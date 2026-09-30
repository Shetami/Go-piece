package main

func chkBrokenHandler() {
	var m map[string]int
	m["x"]++
}

func chkRun(fn func()) (logs []string, rec any) {
	defer func() { rec = recover() }()
	LogPanic(func(s string) { logs = append(logs, s) }, fn)
	return logs, nil
}

func TestLogPanicNoPanic(t *testing.T) {
	ran := false
	logs, rec := chkRun(func() { ran = true })
	if !ran || len(logs) != 0 || rec != nil {
		t.Fatalf("без паники: fn вызвана=%v, логов %d, паника %v; ожидали true, 0, nil", ran, len(logs), rec)
	}
}

func TestLogPanicString(t *testing.T) {
	logs, rec := chkRun(func() { panic("не нашли конфиг") })
	if rec != "не нашли конфиг" {
		t.Fatalf("паника дальше должна быть с тем же значением, recover() = %#v", rec)
	}
	if len(logs) != 1 || !strings.Contains(logs[0], "не нашли конфиг") {
		t.Fatalf("ожидали ровно одну запись со значением паники, получили %q", logs)
	}
}

func TestLogPanicRuntimeError(t *testing.T) {
	logs, rec := chkRun(chkBrokenHandler)
	if _, ok := rec.(runtime.Error); !ok {
		t.Fatalf("паника рантайма должна остаться runtime.Error, а дальше полетело %T", rec)
	}
	if len(logs) != 1 || !strings.Contains(logs[0], "chkBrokenHandler") {
		t.Fatalf("в логе нужен стек с функцией chkBrokenHandler, где упало; лог: %q", logs)
	}
}

func TestLogPanicNil(t *testing.T) {
	logs, rec := chkRun(func() { panic(nil) })
	if _, ok := rec.(*runtime.PanicNilError); !ok || len(logs) != 1 {
		t.Fatalf("panic(nil): дальше полетело %T, логов %d; ожидали *runtime.PanicNilError и одну запись", rec, len(logs))
	}
}
