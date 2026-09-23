package main

import (
	"errors"
	"fmt"
)

type File struct{ name string }

func (f *File) Write(s string) error { return nil }

func (f *File) Close() error {
	return errors.New(f.name + ": диск переполнен при сбросе буфера")
}

// save пишет данные и сообщает об ошибке записи или закрытия.
//
// Результат именованный: только так defer может поменять то, что
// функция вернёт, уже после return.
func save(f *File) (err error) {
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	err = f.Write("данные")
	return err
}

func main() {
	fmt.Println("ошибка сохранения:", save(&File{name: "report.txt"}))
}
