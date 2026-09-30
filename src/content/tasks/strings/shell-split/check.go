package main

func chkSplit(t *testing.T, line string, want []string) {
	t.Helper()
	got, err := SplitArgs(line)
	if err != nil || len(got) != len(want) || !slices.Equal(got, want) {
		t.Fatalf("SplitArgs(%q) = %q, %v; ожидали %q", line, got, err, want)
	}
}

func TestSplitBasic(t *testing.T) {
	chkSplit(t, "ls  -la\t/tmp\n", []string{"ls", "-la", "/tmp"})
	chkSplit(t, `cp "мой файл.txt" 'a b' dir\ name`, []string{"cp", "мой файл.txt", "a b", "dir name"})
	chkSplit(t, "", nil)
	chkSplit(t, "   \t ", nil)
}

func TestSplitQuotes(t *testing.T) {
	chkSplit(t, `a"b c"'d'`, []string{"ab cd"})
	chkSplit(t, `echo "" '' x`, []string{"echo", "", "", "x"})
	chkSplit(t, `'it''s'`, []string{"its"})
	chkSplit(t, `"say \"hi\"" "c:\\dir" "a\nb"`, []string{`say "hi"`, `c:\dir`, `a\nb`})
	chkSplit(t, `'a\b "c"' "it's"`, []string{`a\b "c"`, "it's"})
	chkSplit(t, `grep -e '#' -e "a # b"`, []string{"grep", "-e", "#", "-e", "a # b"})
}

func TestSplitBackslash(t *testing.T) {
	chkSplit(t, `a\"b \'c\' \\d \x`, []string{`a"b`, "'c'", `\d`, "x"})
	chkSplit(t, `\ `, []string{" "})
}

func chkSplitErr(t *testing.T, line string, pos int) {
	t.Helper()
	got, err := SplitArgs(line)
	var se *SyntaxError
	if got != nil || !errors.As(err, &se) || se.Pos != pos {
		t.Fatalf("SplitArgs(%q) = %q, %v; ожидали nil и *SyntaxError{Pos: %d}", line, got, err, pos)
	}
}

func TestSplitErrors(t *testing.T) {
	chkSplitErr(t, `echo "abc`, 5)
	chkSplitErr(t, `echo 'a"b" c`, 5)
	chkSplitErr(t, `echo "a\"`, 5)
	chkSplitErr(t, `echo abc\`, 8)
	chkSplitErr(t, `привет "мир`, 7) // позиция в символах, а не в байтах
	chkSplitErr(t, `"ok" 'ок' "не ок`, 10)
}
