package main

var chkWantLine int

// chkHere запоминает строку, следующую за вызовом.
func chkHere() { _, _, line, _ := runtime.Caller(1); chkWantLine = line + 1 }

func chkLevel3() {
	chkHere()
	panic("глубоко внутри")
}
func chkLevel2() { chkLevel3() }
func chkLevel1() { chkLevel2() }

func chkNilMap() {
	var m map[string]int
	chkHere()
	m["x"] = 1
}

func chkDivide(a, b int) int {
	chkHere()
	return a / b
}

func chkPanicInDefer() {
	defer func() {
		chkHere()
		panic("вторая, из defer")
	}()
	panic("первая")
}

func chkCheckSite(t *testing.T, name string, fn func(), funcSuffix string, wantVal any) {
	t.Helper()
	loc, v, ok := PanicSite(fn)
	if !ok {
		t.Fatalf("%s: panicked=false, ожидали панику", name)
	}
	if wantVal != nil && v != wantVal {
		t.Fatalf("%s: значение %v, ожидали %v", name, v, wantVal)
	}
	if !strings.HasSuffix(loc.Func, funcSuffix) {
		t.Fatalf("%s: Func = %q, ожидали функцию …%s — место паники, а не вызывающий и не рантайм", name, loc.Func, funcSuffix)
	}
	if loc.Line != chkWantLine || !strings.HasSuffix(loc.File, ".go") {
		t.Fatalf("%s: %s:%d, ожидали строку %d", name, loc.File, loc.Line, chkWantLine)
	}
}

func TestPanicSiteExplicit(t *testing.T) {
	chkCheckSite(t, "panic на третьем уровне", chkLevel1, ".chkLevel3", "глубоко внутри")
}

func TestPanicSiteRuntime(t *testing.T) {
	chkCheckSite(t, "запись в nil-мапу", chkNilMap, ".chkNilMap", nil)
	chkCheckSite(t, "деление на ноль", func() { chkDivide(1, 0) }, ".chkDivide", nil)
}

func TestPanicSiteInDefer(t *testing.T) {
	chkCheckSite(t, "паника в defer во время паники", chkPanicInDefer, ".chkPanicInDefer.func1", "вторая, из defer")
}

func TestPanicSiteNoPanic(t *testing.T) {
	ran := false
	loc, v, ok := PanicSite(func() { ran = true })
	if !ran || ok || v != nil || loc != (Location{}) {
		t.Fatalf("без паники: fn вызвана=%v, %+v, %v, %v; ожидали true и нули", ran, loc, v, ok)
	}
}
