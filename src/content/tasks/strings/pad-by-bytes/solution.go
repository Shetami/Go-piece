package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// pad дополняет s пробелами справа до ширины width символов.
func pad(s string, width int) string {
	// Ширина — в символах, а len считает байты: кириллица занимает по два.
	n := width - utf8.RuneCountInString(s)
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}

func main() {
	rows := [][2]string{{"чай", "90"}, {"coffee", "350"}, {"сахар", "60"}}
	for _, r := range rows {
		fmt.Println(pad(r[0], 10) + "|" + r[1])
	}
}
