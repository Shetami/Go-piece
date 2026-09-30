package main

func chkAccs(t *testing.T, got, want []Account) {
	t.Helper()
	norm := func(a []Account) string {
		var b strings.Builder
		for _, x := range a {
			fmt.Fprintf(&b, "%s%q; ", x.Name, x.Emails)
		}
		return b.String()
	}
	if norm(got) != norm(want) {
		t.Fatalf("MergeAccounts =\n  %s\nожидали\n  %s", norm(got), norm(want))
	}
}

func TestMergeChain(t *testing.T) {
	in := []Account{
		{"Анна", []string{"a@x", "a2@x"}},
		{"Боб", []string{"b@x"}},
		{"Анна К.", []string{"a3@x", "c@x"}},
		{"Anna", []string{"c@x", "a2@x"}}, // связывает первую и третью
	}
	chkAccs(t, MergeAccounts(in), []Account{
		{"Анна", []string{"a2@x", "a3@x", "a@x", "c@x"}},
		{"Боб", []string{"b@x"}},
	})
}

func TestMergeLateBridge(t *testing.T) {
	// Мост появляется последним: 0 и 1 разные, пока 2 их не свяжет.
	in := []Account{
		{"p", []string{"1"}},
		{"q", []string{"2"}},
		{"r", []string{"3"}},
		{"s", []string{"3", "2", "1"}},
	}
	chkAccs(t, MergeAccounts(in), []Account{{"p", []string{"1", "2", "3"}}})
}

func TestMergeSameNameDifferentPeople(t *testing.T) {
	in := []Account{
		{"Иван", []string{"i1@x"}},
		{"Иван", []string{"i2@x"}},
		{"Пётр", nil},
		{"Иван", []string{"i1@x", "i1@x"}},
	}
	chkAccs(t, MergeAccounts(in), []Account{
		{"Иван", []string{"i1@x"}},
		{"Иван", []string{"i2@x"}},
		{"Пётр", nil},
	})
}

func TestMergeInputUntouched(t *testing.T) {
	e0 := make([]string, 2, 10) // запас ёмкости — соблазн для append
	e0[0], e0[1] = "z@x", "m@x"
	in := []Account{{"A", e0}, {"B", []string{"m@x", "k@x"}}}
	MergeAccounts(in)
	full := e0[:4]
	if in[0].Emails[0] != "z@x" || in[0].Emails[1] != "m@x" || full[2] != "" || full[3] != "" || in[1].Emails[1] != "k@x" {
		t.Fatalf("входные данные изменились: %q (с запасом ёмкости %q), %q", in[0].Emails, full, in[1].Emails)
	}
	if got := MergeAccounts(nil); len(got) != 0 {
		t.Fatalf("пустой вход: %v", got)
	}
}
