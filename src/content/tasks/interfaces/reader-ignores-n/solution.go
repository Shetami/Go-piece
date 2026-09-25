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
		// Годны только первые n байт; сначала их, потом разбор ошибки:
		// Read вправе вернуть и данные, и io.EOF одновременно.
		b.Write(buf[:n])
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}
	return b.String()
}

func main() {
	fmt.Printf("%q\n", readAll(strings.NewReader("hello, go")))
}
