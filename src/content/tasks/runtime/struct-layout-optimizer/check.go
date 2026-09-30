package main

type chkMixed struct {
	a bool
	b int64
	c bool
	d int32
	e struct{}
}

type chkTail struct {
	x int32
	y byte
	z struct{}
}

type chkArr struct {
	a [3]byte
	b uint16
	c [0]int64
	d string
	e byte
}

func chkF(name string, size, align uintptr) Field { return Field{name, size, align} }

func chkMixedFields() []Field {
	var s chkMixed
	return []Field{
		chkF("a", unsafe.Sizeof(s.a), unsafe.Alignof(s.a)),
		chkF("b", unsafe.Sizeof(s.b), unsafe.Alignof(s.b)),
		chkF("c", unsafe.Sizeof(s.c), unsafe.Alignof(s.c)),
		chkF("d", unsafe.Sizeof(s.d), unsafe.Alignof(s.d)),
		chkF("e", unsafe.Sizeof(s.e), unsafe.Alignof(s.e)),
	}
}

func TestLayoutMatchesCompiler(t *testing.T) {
	var m chkMixed
	off, size := Layout(chkMixedFields())
	want := []uintptr{unsafe.Offsetof(m.a), unsafe.Offsetof(m.b), unsafe.Offsetof(m.c), unsafe.Offsetof(m.d), unsafe.Offsetof(m.e)}
	if !slices.Equal(off, want) || size != unsafe.Sizeof(m) {
		t.Fatalf("chkMixed: смещения %v, размер %d; компилятор: %v, %d", off, size, want, unsafe.Sizeof(m))
	}

	var tl chkTail
	off, size = Layout([]Field{chkF("x", 4, 4), chkF("y", 1, 1), chkF("z", 0, 1)})
	if size != unsafe.Sizeof(tl) || off[2] != unsafe.Offsetof(tl.z) {
		t.Fatalf("структура с полем нулевого размера в конце: размер %d, компилятор %d", size, unsafe.Sizeof(tl))
	}

	var ar chkArr
	fs := []Field{
		chkF("a", unsafe.Sizeof(ar.a), unsafe.Alignof(ar.a)),
		chkF("b", unsafe.Sizeof(ar.b), unsafe.Alignof(ar.b)),
		chkF("c", unsafe.Sizeof(ar.c), unsafe.Alignof(ar.c)),
		chkF("d", unsafe.Sizeof(ar.d), unsafe.Alignof(ar.d)),
		chkF("e", unsafe.Sizeof(ar.e), unsafe.Alignof(ar.e)),
	}
	off, size = Layout(fs)
	want = []uintptr{unsafe.Offsetof(ar.a), unsafe.Offsetof(ar.b), unsafe.Offsetof(ar.c), unsafe.Offsetof(ar.d), unsafe.Offsetof(ar.e)}
	if !slices.Equal(off, want) || size != unsafe.Sizeof(ar) {
		t.Fatalf("chkArr: смещения %v, размер %d; компилятор: %v, %d", off, size, want, unsafe.Sizeof(ar))
	}
}

func TestLayoutEdge(t *testing.T) {
	if off, size := Layout(nil); len(off) != 0 || size != 0 {
		t.Fatalf("пустая структура: %v, %d", off, size)
	}
	if _, size := Layout([]Field{chkF("z", 0, 1)}); size != unsafe.Sizeof(struct{ z struct{} }{}) {
		t.Fatalf("структура из одного пустого поля: размер %d, ожидали 0", size)
	}
}

type chkMixedBest struct {
	e struct{}
	b int64
	d int32
	a bool
	c bool
}

func TestOptimize(t *testing.T) {
	in := chkMixedFields()
	orig := slices.Clone(in)
	out := Optimize(in)
	if !slices.Equal(in, orig) {
		t.Fatalf("Optimize изменил входной слайс: %v", in)
	}
	var names []string
	for _, f := range out {
		names = append(names, f.Name)
	}
	if got := strings.Join(names, ""); got != "ebdac" {
		t.Fatalf("Optimize: порядок %s, ожидали ebdac (пустое поле первым, дальше по убыванию выравнивания, равные — в исходном порядке)", got)
	}
	if _, size := Layout(out); size != unsafe.Sizeof(chkMixedBest{}) {
		t.Fatalf("размер после Optimize %d, ожидали %d", size, unsafe.Sizeof(chkMixedBest{}))
	}
}
