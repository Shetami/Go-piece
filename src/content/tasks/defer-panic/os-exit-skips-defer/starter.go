package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

func validate() error { return errors.New("строка 2 пустая") }

func main() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	fmt.Fprintln(w, "отчёт: строк 3")
	if err := validate(); err != nil {
		fmt.Fprintln(w, "ошибка:", err)
		os.Exit(1)
	}
	fmt.Fprintln(w, "отчёт сохранён")
}
