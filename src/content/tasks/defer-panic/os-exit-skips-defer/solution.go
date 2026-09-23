package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

func validate() error { return errors.New("строка 2 пустая") }

// run делает всю работу и возвращает ошибку, а не выходит из программы сам:
// так отложенные вызовы в main успевают отработать.
func run(w io.Writer) error {
	fmt.Fprintln(w, "отчёт: строк 3")
	if err := validate(); err != nil {
		return err
	}
	fmt.Fprintln(w, "отчёт сохранён")
	return nil
}

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	if err := run(w); err != nil {
		fmt.Fprintln(w, "ошибка:", err)
	}
}
