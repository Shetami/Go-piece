package main

import (
	"errors"
	"io"
	"slices"
)

var (
	ErrNegativeOffset = errors.New("negative offset")
	ErrWhence         = errors.New("invalid whence")
)

// File — файл в памяти. *File реализует io.ReadWriteSeeker и io.ReaderAt:
//   - Read и Write работают с текущей позиции и сдвигают её;
//   - Write за концом файла дописывает, промежуток между старым концом и
//     позицией заполняется нулями;
//   - Read на позиции >= размера — 0, io.EOF;
//   - Seek: io.SeekStart / io.SeekCurrent / io.SeekEnd; итоговая позиция < 0 —
//     ErrNegativeOffset, неизвестный whence — ErrWhence, в обоих случаях
//     позиция не меняется; позиция за концом допустима;
//   - ReadAt не меняет позицию; прочитано меньше len(p) — вернуть io.EOF;
//     off < 0 — ErrNegativeOffset;
//   - Bytes — копия содержимого.
type File struct {
	data []byte
	pos  int64
}

func (f *File) Read(p []byte) (int, error) {
	if f.pos >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.pos:])
	f.pos += int64(n)
	return n, nil
}

func (f *File) Write(p []byte) (int, error) {
	end := f.pos + int64(len(p))
	if end > int64(len(f.data)) {
		// Дорастить до end: новые байты — нули, «дыра» заполнится сама.
		f.data = append(f.data, make([]byte, end-int64(len(f.data)))...)
	}
	n := copy(f.data[f.pos:], p)
	f.pos += int64(n)
	return n, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	var base int64
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		base = f.pos
	case io.SeekEnd:
		base = int64(len(f.data))
	default:
		return f.pos, ErrWhence
	}
	next := base + offset
	if next < 0 {
		return f.pos, ErrNegativeOffset
	}
	f.pos = next
	return next, nil
}

func (f *File) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, ErrNegativeOffset
	}
	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[off:])
	if n < len(p) {
		// Контракт ReaderAt строже Reader: меньше len(p) — только с ошибкой.
		return n, io.EOF
	}
	return n, nil
}

func (f *File) Bytes() []byte { return slices.Clone(f.data) }
