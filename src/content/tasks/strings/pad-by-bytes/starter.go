package main

import (
	"fmt"
	"strings"
)

// pad дополняет s пробелами справа до ширины width символов.
func pad(s string, width int) string {
	return s + strings.Repeat(" ", width-len(s))
}

func main() {
	rows := [][2]string{{"чай", "90"}, {"coffee", "350"}, {"сахар", "60"}}
	for _, r := range rows {
		fmt.Println(pad(r[0], 10) + "|" + r[1])
	}
}
