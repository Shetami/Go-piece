package main

import (
	"bytes"
	"fmt"
	"io"
)

// process пишет лог, если ему дали, куда писать.
func process(w io.Writer, msg string) {
	if w != nil {
		fmt.Fprintln(w, msg)
	}
}

func main() {
	debug := false

	// Переменная сразу интерфейсного типа: без отладки она остаётся
	// настоящим nil-интерфейсом, а не интерфейсом с nil-указателем внутри.
	var w io.Writer
	if debug {
		w = new(bytes.Buffer)
	}

	process(w, "старт")
	fmt.Println("готово")
}
