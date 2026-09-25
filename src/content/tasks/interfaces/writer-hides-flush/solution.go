package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func report(w io.Writer) {
	fmt.Fprintln(w, "заказов: 3")
	fmt.Fprintln(w, "выручка: 1250")
}

func main() {
	// Держим конкретный тип: через io.Writer до Flush не дотянуться.
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	report(w)
}
