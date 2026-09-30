package main

func chkRender(t *testing.T, tpl string, vars map[string]string, want string) {
	t.Helper()
	got, err := Render(tpl, vars)
	if err != nil || got != want {
		t.Fatalf("Render(%q) = %q, %v; ожидали %q", tpl, got, err, want)
	}
}

func TestRenderBasic(t *testing.T) {
	v := map[string]string{"name": "Аня", "n": "3", "empty": ""}
	chkRender(t, "Привет, {{name}}! У вас {{ n }} письма.", v, "Привет, Аня! У вас 3 письма.")
	chkRender(t, "{{name}}{{name}}", v, "АняАня")
	chkRender(t, "[{{empty}}]", v, "[]")
	chkRender(t, "без подстановок", v, "без подстановок")
	chkRender(t, "", v, "")
}

func TestRenderLiteralBraces(t *testing.T) {
	v := map[string]string{"x": "1"}
	chkRender(t, `код: \{{x}} = {{x}}`, v, "код: {{x}} = 1")
	chkRender(t, "func() { return }} {x}", v, "func() { return }} {x}")
	chkRender(t, `C:\dir {{x}}`, v, `C:\dir 1`)
}

func TestRenderNoRecursion(t *testing.T) {
	v := map[string]string{"a": "{{b}}", "b": "секрет"}
	chkRender(t, "значение: {{a}}", v, "значение: {{b}}")
}

func TestRenderSyntax(t *testing.T) {
	for _, tpl := range []string{"Привет, {{name", "{{ }}", "{{}}", "ok {{a}} {{b"} {
		got, err := Render(tpl, map[string]string{"name": "x", "a": "1"})
		if !errors.Is(err, ErrSyntax) || got != "" {
			t.Fatalf("Render(%q) = %q, %v; ожидали \"\" и ошибку ErrSyntax", tpl, got, err)
		}
	}
}

func TestRenderMissingAll(t *testing.T) {
	got, err := Render("{{a}} {{ok}} {{b}} {{a}} {{c}}", map[string]string{"ok": "1"})
	if got != "" || !errors.Is(err, ErrMissing) {
		t.Fatalf("Render = %q, %v; ожидали \"\" и ошибку ErrMissing", got, err)
	}
	j, ok := err.(interface{ Unwrap() []error })
	if !ok || len(j.Unwrap()) != 3 {
		t.Fatalf("ожидали errors.Join из трёх ошибок (a, b, c — каждая один раз), получили %v", err)
	}
	for i, name := range []string{"a", "b", "c"} {
		e := j.Unwrap()[i]
		if !errors.Is(e, ErrMissing) || !strings.Contains(e.Error(), name) {
			t.Fatalf("ошибка №%d = %v; ожидали ErrMissing с именем %q (порядок — по первому появлению)", i+1, e, name)
		}
	}
}
