package main

func chkPanics(f func()) (p bool) {
	defer func() { p = recover() != nil }()
	f()
	return false
}

func TestBufBasic(t *testing.T) {
	var b Buf
	b.WriteString("привет, ")
	if n, err := b.Write([]byte("мир")); n != 6 || err != nil {
		t.Fatalf("Write = %d, %v; ожидали 6, nil", n, err)
	}
	if b.String() != "привет, мир" || b.Len() != len("привет, мир") {
		t.Fatalf("String = %q, Len = %d", b.String(), b.Len())
	}
}

func TestBufStringStable(t *testing.T) {
	var b Buf
	b.WriteString("abc")
	s1 := b.String()
	b.WriteString("def")
	s2 := b.String()
	b.Reset()
	b.WriteString("XYZXYZ")
	if s1 != "abc" || s2 != "abcdef" {
		t.Fatalf("строки, полученные раньше, изменились: %q, %q — после Reset новая запись затёрла их байты", s1, s2)
	}
	if b.String() != "XYZXYZ" {
		t.Fatalf("после Reset String = %q", b.String())
	}
}

var chkSinkS string

func TestBufStringNoAlloc(t *testing.T) {
	var b Buf
	b.WriteString(strings.Repeat("x", 100))
	if a := testing.AllocsPerRun(100, func() { chkSinkS = b.String() }); a != 0 {
		t.Fatalf("String выделяет память %.0f раз — байты должны не копироваться", a)
	}
}

func TestBufCopyCheck(t *testing.T) {
	var zero Buf
	c := zero
	c.WriteString("копия нулевого — можно")
	zero.WriteString("оригинал")
	if c.String() != "копия нулевого — можно" || zero.String() != "оригинал" {
		t.Fatalf("копия нулевого Buf и оригинал должны работать независимо")
	}

	var a Buf
	a.WriteString("данные")
	cp := a
	if cp.Len() != a.Len() || cp.String() != "данные" {
		t.Fatalf("Len/String на копии должны работать")
	}
	if !chkPanics(func() { cp.WriteString("!") }) {
		t.Fatalf("WriteString в копию непустого Buf не запаниковал")
	}
	if !chkPanics(func() { cp.Write([]byte("!")) }) {
		t.Fatalf("Write в копию непустого Buf не запаниковал")
	}
	if chkPanics(func() { a.WriteString("!") }) {
		t.Fatalf("запись в оригинал не должна паниковать")
	}

	a.Reset()
	cp2 := a
	if chkPanics(func() { cp2.WriteString("ok") }) {
		t.Fatalf("копия Buf после Reset должна быть пригодна к записи")
	}
}
