package main

func TestFieldsParse(t *testing.T) {
	cases := []struct {
		line string
		want []string
	}{
		{`a,b,c`, []string{"a", "b", "c"}},
		{``, []string{""}},
		{`a,,`, []string{"a", "", ""}},
		{`,`, []string{"", ""}},
		{`"x,y",z`, []string{"x,y", "z"}},
		{`"",x`, []string{"", "x"}},
		{`"say ""hi""",1`, []string{`say "hi"`, "1"}},
		{`""""`, []string{`"`}},
		{`id,"ёлка, шар",""`, []string{"id", "ёлка, шар", ""}},
		{`a,"b"`, []string{"a", "b"}},
	}
	for _, c := range cases {
		got, err := AppendFields(nil, c.line)
		if err != nil || !slices.Equal(got, c.want) {
			t.Fatalf("AppendFields(%s) = %q, %v; ожидали %q", c.line, got, err, c.want)
		}
	}
}

func TestFieldsErrors(t *testing.T) {
	for _, line := range []string{`"abc`, `"ab"c,d`, `a"b,c`, `x,"y`, `"a""`, `1,"2" ,3`} {
		dst := []string{"old"}
		got, err := AppendFields(dst, line)
		if !errors.Is(err, ErrSyntax) {
			t.Fatalf("AppendFields(%s): err = %v, ожидали ErrSyntax", line, err)
		}
		if !slices.Equal(got, []string{"old"}) {
			t.Fatalf("AppendFields(%s) при ошибке вернул %q, ожидали dst без изменений", line, got)
		}
	}
}

func TestFieldsAppends(t *testing.T) {
	got, _ := AppendFields([]string{"h"}, "a,b")
	if !slices.Equal(got, []string{"h", "a", "b"}) {
		t.Fatalf("получили %q — поля должны дописываться к dst", got)
	}
}

var chkSinkF []string

func TestFieldsAllocs(t *testing.T) {
	buf := make([]string, 0, 16)
	plain := `42,"Иванов, Иван",ivan@example.com,,"",ok`
	if a := testing.AllocsPerRun(100, func() {
		chkSinkF, _ = AppendFields(buf[:0], plain)
	}); a != 0 {
		t.Fatalf("строка без \"\" при достаточном dst: %.0f выделений, ожидали 0 — поля должны быть подстроками line", a)
	}
	escaped := `1,"он сказал ""да""",3`
	if a := testing.AllocsPerRun(100, func() {
		chkSinkF, _ = AppendFields(buf[:0], escaped)
	}); a > 1 {
		t.Fatalf("строка с одним полем с \"\": %.0f выделений, ожидали не больше 1", a)
	}
}
