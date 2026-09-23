package main

import (
	"fmt"
	"strconv"
)

func describe(v any) string {
	s := v.(string)
	return "строка длины " + strconv.Itoa(len(s))
}

func main() {
	for _, v := range []any{"го", 42, nil} {
		fmt.Println(describe(v))
	}
}
