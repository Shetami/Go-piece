package main

import "strings"

// Concat склеивает строки через sep за одно выделение памяти.
func Concat(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	// Считаем итоговую длину заранее и выделяем память один раз.
	n := len(sep) * (len(parts) - 1)
	for _, p := range parts {
		n += len(p)
	}
	var b strings.Builder
	b.Grow(n)
	for i, p := range parts {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(p)
	}
	// String() отдаёт буфер Builder'а без копирования — второй аллокации нет.
	return b.String()
}
