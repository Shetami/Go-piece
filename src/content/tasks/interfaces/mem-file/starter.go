package main

import "errors"

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
	// ваши поля
}

func (f *File) Read(p []byte) (int, error) {
	// ваш код
	return 0, nil
}

func (f *File) Write(p []byte) (int, error) {
	// ваш код
	return 0, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	// ваш код
	return 0, nil
}

func (f *File) ReadAt(p []byte, off int64) (int, error) {
	// ваш код
	return 0, nil
}

func (f *File) Bytes() []byte {
	// ваш код
	return nil
}
