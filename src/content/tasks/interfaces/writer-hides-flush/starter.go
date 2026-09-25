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
	var w io.Writer = bufio.NewWriter(os.Stdout)
	report(w)
}
