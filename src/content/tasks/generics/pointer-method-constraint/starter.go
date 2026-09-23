package main

import "fmt"

type Setter interface {
	Set(s string)
}

type Config struct{ Env string }

func (c *Config) Set(s string) { c.Env = s }

// Parse создаёт значение типа T и заполняет его из строки.
func Parse[T Setter](s string) T {
	var v T
	v.Set(s)
	return v
}

func main() {
	cfg := Parse[*Config]("prod")
	fmt.Printf("%+v\n", *cfg)
}
