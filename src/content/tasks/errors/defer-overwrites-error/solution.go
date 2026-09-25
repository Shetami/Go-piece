package main

import (
	"errors"
	"fmt"
)

type File struct {
	name string
}

func (f *File) Close() error {
	fmt.Println("закрыт", f.name)
	return nil
}

func process(name string) (err error) {
	f := &File{name: name}
	defer func() {
		// Ошибку Close сохраняем, но не затираем ею основную.
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	if name == "bad.csv" {
		return errors.New("битый заголовок")
	}
	return nil
}

func main() {
	fmt.Println(process("ok.csv"))
	fmt.Println(process("bad.csv"))
}
