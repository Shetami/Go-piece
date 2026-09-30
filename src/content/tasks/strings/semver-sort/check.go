package main

func chkCmp(t *testing.T, a, b string, want int) {
	t.Helper()
	va, err1 := ParseVersion(a)
	vb, err2 := ParseVersion(b)
	if err1 != nil || err2 != nil {
		t.Fatalf("ParseVersion(%q / %q): неожиданная ошибка %v %v", a, b, err1, err2)
	}
	if got := va.Compare(vb); got != want {
		t.Fatalf("Compare(%q, %q) = %d, ожидали %d", a, b, got, want)
	}
}

func TestSemverNumeric(t *testing.T) {
	chkCmp(t, "1.10.0", "1.9.0", 1)
	chkCmp(t, "v2.0.0", "2.0.0", 0)
	chkCmp(t, "0.0.9", "0.0.10", -1)
	chkCmp(t, "1.2.3+build.5", "1.2.3+build.1", 0)
}

func TestSemverPrerelease(t *testing.T) {
	chkCmp(t, "1.0.0-rc.1", "1.0.0", -1)
	chkCmp(t, "1.0.0-alpha", "1.0.0-alpha.1", -1)
	chkCmp(t, "1.0.0-alpha.1", "1.0.0-alpha.beta", -1)
	chkCmp(t, "1.0.0-beta.2", "1.0.0-beta.11", -1)
	chkCmp(t, "1.0.0-rc.1", "1.0.0-beta.11", 1)
	chkCmp(t, "1.0.0-RC", "1.0.0-alpha", -1)
	chkCmp(t, "1.0.0-x-y.1", "1.0.0-x-y.1", 0)
	chkCmp(t, "1.0.0-1.99999999999999999999999", "1.0.0-1.100000000000000000000000", -1)
}

func TestSemverInvalid(t *testing.T) {
	for _, s := range []string{"1.2", "1.2.3.4", "01.2.3", "1.02.3", "1.2.3-", "1.2.3-rc..1", "1.2.3-01", "1.2.3+", "1.2.3-ё", "1.-2.3", "a.b.c", "", "1.2.3 "} {
		if _, err := ParseVersion(s); err == nil {
			t.Fatalf("ParseVersion(%q): ожидали ошибку", s)
		}
	}
	if v, err := ParseVersion("1.2.3-rc.0-x+exp.sha.5114f85"); err != nil || len(v.Pre) != 2 || v.Pre[1] != "0-x" || v.Build != "exp.sha.5114f85" {
		t.Fatalf(`ParseVersion("1.2.3-rc.0-x+exp.sha.5114f85") = %+v, %v`, v, err)
	}
}

func TestSemverSort(t *testing.T) {
	in := []string{"1.0.0", "1.0.0-rc.1", "v0.9.12", "1.0.0-alpha", "0.10.0", "1.0.0+b", "1.0.0-alpha.1", "1.0.0+a", "0.9.2"}
	orig := slices.Clone(in)
	got, err := SortVersions(in)
	want := []string{"0.9.2", "v0.9.12", "0.10.0", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-rc.1", "1.0.0", "1.0.0+b", "1.0.0+a"}
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("SortVersions = %q, %v\nожидали %q (равные по приоритету — в исходном порядке)", got, err, want)
	}
	if !slices.Equal(in, orig) {
		t.Fatalf("SortVersions изменила входной слайс: %q", in)
	}
}

func TestSemverSortErrors(t *testing.T) {
	got, err := SortVersions([]string{"1.0.0", "1.0", "2.0.0", "v1.x.0"})
	if err == nil || got != nil {
		t.Fatalf("SortVersions с кривыми версиями = %q, %v; ожидали nil и ошибку", got, err)
	}
	j, ok := err.(interface{ Unwrap() []error })
	if !ok || len(j.Unwrap()) != 2 {
		t.Fatalf("ожидали errors.Join из двух ошибок — по одной на каждую кривую версию, получили %v", err)
	}
	if !strings.Contains(err.Error(), `"1.0"`) || !strings.Contains(err.Error(), `"v1.x.0"`) {
		t.Fatalf("текст ошибки должен называть обе кривые версии: %v", err)
	}
}
