package main

func chkJoin(t *testing.T, root, rel, want string) {
	t.Helper()
	got, err := SafeJoin(root, rel)
	if err != nil || got != want {
		t.Fatalf("SafeJoin(%q, %q) = %q, %v; ожидали %q", root, rel, got, err, want)
	}
}

func chkJoinErr(t *testing.T, root, rel string, target error) {
	t.Helper()
	got, err := SafeJoin(root, rel)
	if !errors.Is(err, target) {
		t.Fatalf("SafeJoin(%q, %q) = %q, %v; ожидали ошибку %v", root, rel, got, err, target)
	}
}

func TestJoinNormal(t *testing.T) {
	chkJoin(t, "/srv/files", "docs/отчёт.pdf", "/srv/files/docs/отчёт.pdf")
	chkJoin(t, "/srv/files", "//a//./b/", "/srv/files/a/b")
	chkJoin(t, "/srv/files", "a/b/../c", "/srv/files/a/c")
	chkJoin(t, "/srv/files", "/etc/passwd", "/srv/files/etc/passwd")
	chkJoin(t, "/srv/files", "", "/srv/files")
	chkJoin(t, "/srv/files", "a/..", "/srv/files")
	chkJoin(t, "/", "a/b", "/a/b")
	chkJoin(t, "/", "", "/")
}

func TestJoinOddNames(t *testing.T) {
	chkJoin(t, "/srv", ".../..a/a..", "/srv/.../..a/a..")
	chkJoin(t, "/srv", ".hidden/./x", "/srv/.hidden/x")
}

func TestJoinTraversal(t *testing.T) {
	chkJoinErr(t, "/srv/files", "../secret", ErrTraversal)
	chkJoinErr(t, "/srv/files", "a/../../files-evil/x", ErrTraversal)
	chkJoinErr(t, "/srv/files", "a/b/../../..", ErrTraversal)
	chkJoinErr(t, "/srv/files", "/../../etc/passwd", ErrTraversal)
	chkJoinErr(t, "/", "..", ErrTraversal)
}

func TestJoinBackslash(t *testing.T) {
	chkJoinErr(t, "/srv/files", `..\..\windows\win.ini`, ErrTraversal)
	chkJoinErr(t, "/srv/files", `a\..\..\x`, ErrTraversal)
	chkJoin(t, "/srv/files", `a\b\..\c`, "/srv/files/a/c")
}

func TestJoinNul(t *testing.T) {
	chkJoinErr(t, "/srv/files", "report.pdf\x00.png", ErrBadPath)
}
