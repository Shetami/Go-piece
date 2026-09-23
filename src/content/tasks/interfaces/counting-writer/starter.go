package main

import "io"

// CountingWriter передаёт всё записанное в W и считает байты в N.
type CountingWriter struct {
	W io.Writer
	N int64
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	// ваш код
	return 0, nil
}
