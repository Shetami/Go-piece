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

	var buf *bytes.Buffer
	if debug {
		buf = new(bytes.Buffer)
	}

	process(buf, "старт")
	fmt.Println("готово")
}
