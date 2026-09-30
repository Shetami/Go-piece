package main

var (
	_ io.ReadWriteSeeker = (*File)(nil)
	_ io.ReaderAt        = (*File)(nil)
)

func TestFileWriteRead(t *testing.T) {
	f := &File{}
	fmt.Fprintf(f, "hello, %s", "world")
	if pos, _ := f.Seek(0, io.SeekCurrent); pos != 12 {
		t.Fatalf("позиция после записи 12 байт = %d", pos)
	}
	if n, err := f.Read(make([]byte, 4)); n != 0 || err != io.EOF {
		t.Fatalf("Read в конце = %d, %v; ожидали 0, io.EOF", n, err)
	}
	f.Seek(7, io.SeekStart)
	got, err := io.ReadAll(f)
	if err != nil || string(got) != "world" {
		t.Fatalf("чтение с 7: %q, %v", got, err)
	}
	f.Seek(-5, io.SeekEnd)
	f.Write([]byte("WORLD!!"))
	if string(f.Bytes()) != "hello, WORLD!!" {
		t.Fatalf("перезапись хвоста с удлинением: %q", f.Bytes())
	}
}

func TestFileHole(t *testing.T) {
	f := &File{}
	f.Write([]byte("ab"))
	if pos, err := f.Seek(3, io.SeekCurrent); pos != 5 || err != nil {
		t.Fatalf("Seek за конец: %d, %v; позиция за концом допустима", pos, err)
	}
	if size, _ := f.Seek(0, io.SeekEnd); size != 2 {
		t.Fatalf("Seek сам по себе не меняет размер: %d", size)
	}
	f.Seek(5, io.SeekStart)
	f.Write([]byte("z"))
	if got := f.Bytes(); !bytes.Equal(got, []byte("ab\x00\x00\x00z")) {
		t.Fatalf("запись за концом: %q, ожидали \"ab\\x00\\x00\\x00z\"", got)
	}
	b := f.Bytes()
	b[0] = 'X'
	if f.Bytes()[0] != 'a' {
		t.Fatal("Bytes должен возвращать копию")
	}
}

func TestFileSeekErrors(t *testing.T) {
	f := &File{}
	f.Write([]byte("abcdef"))
	f.Seek(2, io.SeekStart)
	if _, err := f.Seek(-3, io.SeekCurrent); err != ErrNegativeOffset {
		t.Fatalf("Seek в минус: %v, ожидали ErrNegativeOffset", err)
	}
	if _, err := f.Seek(0, 42); err != ErrWhence {
		t.Fatalf("неизвестный whence: %v, ожидали ErrWhence", err)
	}
	if pos, _ := f.Seek(0, io.SeekCurrent); pos != 2 {
		t.Fatalf("после ошибочных Seek позиция %d, ожидали прежнюю 2", pos)
	}
}

func TestFileReadAt(t *testing.T) {
	f := &File{}
	f.Write([]byte("0123456789"))
	f.Seek(1, io.SeekStart)
	p := make([]byte, 4)
	if n, err := f.ReadAt(p, 3); n != 4 || err != nil || string(p) != "3456" {
		t.Fatalf("ReadAt(4, 3) = %d, %v, %q", n, err, p)
	}
	if n, err := f.ReadAt(p, 8); n != 2 || err != io.EOF || string(p[:n]) != "89" {
		t.Fatalf("ReadAt у конца = %d, %v; ожидали 2 и io.EOF — меньше len(p) только с ошибкой", n, err)
	}
	if _, err := f.ReadAt(p, -1); err != ErrNegativeOffset {
		t.Fatalf("ReadAt(-1) = %v", err)
	}
	if pos, _ := f.Seek(0, io.SeekCurrent); pos != 1 {
		t.Fatalf("ReadAt сдвинул позицию: %d", pos)
	}
	sec, err := io.ReadAll(io.NewSectionReader(f, 2, 5))
	if err != nil || string(sec) != "23456" {
		t.Fatalf("io.SectionReader поверх File: %q, %v", sec, err)
	}
}
