package main

import (
	"fmt"
	"strconv"
)

func describe(v any) string {
	// Форма с ok не паникует: если внутри не строка, ok == false.
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("не строка: %T", v)
	}
	return "строка длины " + strconv.Itoa(len(s))
}

func main() {
	for _, v := range []any{"го", 42, nil} {
		fmt.Println(describe(v))
	}
}
