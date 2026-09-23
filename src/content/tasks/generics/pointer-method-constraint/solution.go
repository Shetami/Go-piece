package main

import "fmt"

type Setter interface {
	Set(s string)
}

type Config struct{ Env string }

func (c *Config) Set(s string) { c.Env = s }

// Parse создаёт значение типа T и заполняет его из строки.
//
// Второй параметр PT — «указатель на T, у которого есть Set». Так мы
// работаем с самим значением T, а метод вызываем через его адрес.
func Parse[T any, PT interface {
	*T
	Setter
}](s string) T {
	var v T
	PT(&v).Set(s)
	return v
}

func main() {
	cfg := Parse[Config]("prod")
	fmt.Printf("%+v\n", cfg)
}
