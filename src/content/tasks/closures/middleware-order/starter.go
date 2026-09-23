package main

import "fmt"

type Handler func(string) string
type Middleware func(Handler) Handler

func tag(name string) Middleware {
	return func(next Handler) Handler {
		return func(req string) string {
			return name + "(" + next(req) + ")"
		}
	}
}

// chain оборачивает h так, чтобы первый middleware в списке оказался снаружи.
func chain(h Handler, ms ...Middleware) Handler {
	for _, m := range ms {
		h = m(h)
	}
	return h
}

func main() {
	h := chain(func(req string) string { return req }, tag("auth"), tag("log"), tag("gzip"))
	fmt.Println(h("запрос"))
}
