package main

func chkQuery(t *testing.T, q string, want map[string][]string) {
	t.Helper()
	got, err := ParseQuery(q)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseQuery(%q) = %q, %v; ожидали %q", q, got, err, want)
	}
}

func TestQueryBasic(t *testing.T) {
	chkQuery(t, "q=%D0%B3%D0%BE+lang&tag=a&tag=b&debug", map[string][]string{
		"q": {"го lang"}, "tag": {"a", "b"}, "debug": {""},
	})
	chkQuery(t, "", map[string][]string{})
	chkQuery(t, "&&a=1&&", map[string][]string{"a": {"1"}})
	chkQuery(t, "a=&=x", map[string][]string{"a": {""}, "": {"x"}})
}

func TestQueryTricky(t *testing.T) {
	chkQuery(t, "expr=1%2B1%3D2&sum=1+1", map[string][]string{"expr": {"1+1=2"}, "sum": {"1 1"}})
	chkQuery(t, "a=b=c&x=%26y", map[string][]string{"a": {"b=c"}, "x": {"&y"}})
	chkQuery(t, "a=1;b=2", map[string][]string{"a": {"1;b=2"}})
	chkQuery(t, "%d0%bf%D1%80=%25", map[string][]string{"пр": {"%"}})
	chkQuery(t, "имя=Аня", map[string][]string{"имя": {"Аня"}})
}

func chkQueryErr(t *testing.T, q string, pos int) {
	t.Helper()
	got, err := ParseQuery(q)
	var qe *QueryError
	if got != nil || !errors.As(err, &qe) || qe.Pos != pos {
		t.Fatalf("ParseQuery(%q) = %q, %v; ожидали nil и *QueryError{Pos: %d}", q, got, err, pos)
	}
}

func TestQueryBadEscape(t *testing.T) {
	chkQueryErr(t, "a=1&b=50%", 8)
	chkQueryErr(t, "a=%4", 2)
	chkQueryErr(t, "a=ok&b=%zz", 7)
	chkQueryErr(t, "a=%G1", 2)
	chkQueryErr(t, "ключ=%x", 9) // позиция в байтах: «ключ» — 8 байт
	chkQueryErr(t, "%2", 0)
}

func TestQueryInvalidUTF8(t *testing.T) {
	chkQueryErr(t, "a=1&name=%D0", 9)
	chkQueryErr(t, "a=1&%FF=1", 4)
	chkQueryErr(t, "a=%D0%B3%D0", 2)
}
