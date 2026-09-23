package main

import (
	"strconv"
	"strings"
)

// Compress сжимает повторы: "aaabcc" → "a3bc2".
// Одиночный символ пишется без числа. Во входе цифр не бывает.
func Compress(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i := 0; i < len(rs); {
		j := i
		for j < len(rs) && rs[j] == rs[i] {
			j++
		}
		b.WriteRune(rs[i])
		if n := j - i; n > 1 {
			b.WriteString(strconv.Itoa(n))
		}
		i = j
	}
	return b.String()
}
