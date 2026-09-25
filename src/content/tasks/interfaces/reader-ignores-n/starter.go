package main

import (
	"fmt"
	"io"
	"strings"
)

// readAll читает из r всё до конца кусками по 4 байта.
func readAll(r io.Reader) string {
	var b strings.Builder
	buf := make([]byte, 4)
	for {
		n, err := r.Read(buf)
		b.Write(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		_ = n
	}
	return b.String()
}

func main() {
	fmt.Printf("%q\n", readAll(strings.NewReader("hello, go")))
}
