package main

// chkScanner отдаёт строки через один и тот же буфер — как bufio.Scanner.
func chkScanner(lines ...string) (func() ([]byte, bool), func()) {
	buf := make([]byte, 0, 128)
	i := 0
	next := func() ([]byte, bool) {
		if i >= len(lines) {
			return nil, false
		}
		buf = append(buf[:0], lines[i]...)
		i++
		return buf, true
	}
	scramble := func() {
		b := buf[:cap(buf)]
		for k := range b {
			b[k] = '#'
		}
	}
	return next, scramble
}

func chkStrs(recs []Record) [][]string {
	out := [][]string{}
	for _, r := range recs {
		row := []string{}
		for _, f := range r {
			row = append(row, string(f))
		}
		out = append(out, row)
	}
	return out
}

func TestReadAllFields(t *testing.T) {
	next, scramble := chkScanner("id,name,city", "", "1,Анна,Москва", "2,,", "3,Иван,Казань,")
	recs := ReadAll(next, ',')
	scramble()
	want := [][]string{
		{"id", "name", "city"},
		{"1", "Анна", "Москва"},
		{"2", "", ""},
		{"3", "Иван", "Казань", ""},
	}
	if got := chkStrs(recs); !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadAll = %q\nожидали %q\n(если видны '#' или чужие строки — поля ссылаются на буфер сканера)", got, want)
	}
}

func TestReadAllOwnCopy(t *testing.T) {
	next, _ := chkScanner("aaa;bbb", "ccc;ddd", "eee;fff")
	recs := ReadAll(next, ';')
	if got := chkStrs(recs); !reflect.DeepEqual(got[0], []string{"aaa", "bbb"}) {
		t.Fatalf("первая запись %q — перезаписана следующими строками", got[0])
	}
	if &recs[0][0][0] == &recs[1][0][0] {
		t.Fatalf("две записи делят одну память")
	}
}

func TestReadAllFieldAppend(t *testing.T) {
	next, _ := chkScanner("ab,cd,ef")
	rec := ReadAll(next, ',')[0]
	_ = append(rec[0], 'X', 'Y')
	if string(rec[1]) != "cd" {
		t.Fatalf("append к первому полю затёр второе: %q", rec[1])
	}
	if uintptr(unsafe.Pointer(&rec[1][0]))-uintptr(unsafe.Pointer(&rec[0][0])) != 3 {
		t.Fatalf("поля записи должны быть подслайсами одной копии строки")
	}
}

func TestReadAllAllocs(t *testing.T) {
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "f1,f2,f3,f4,f5,f6,f7,f8"
	}
	allocs := testing.AllocsPerRun(10, func() {
		next, _ := chkScanner(lines...)
		ReadAll(next, ',')
	})
	// 2 на запись (копия строки + слайс полей) + рост результата + сам сканер.
	if allocs > 2*100+20 {
		t.Fatalf("%.0f аллокаций на 100 строк по 8 полей — поля копируются по одному?", allocs)
	}
}
