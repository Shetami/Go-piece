package main

import "io"

// CountingWriter передаёт всё записанное в W и считает байты в N.
type CountingWriter struct {
	W io.Writer
	N int64
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.W.Write(p)
	// Считаем то, что реально записано, а не len(p): при ошибке это может быть меньше.
	c.N += int64(n)
	return n, err
}
