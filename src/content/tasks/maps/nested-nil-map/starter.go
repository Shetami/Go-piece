package main

import "fmt"

func main() {
	// Кто → к какому ресурсу → какие права.
	acl := map[string]map[string]string{}

	grants := [][3]string{
		{"аня", "отчёты", "чтение"},
		{"аня", "деньги", "запись"},
		{"боря", "отчёты", "запись"},
	}
	for _, g := range grants {
		acl[g[0]][g[1]] = g[2]
	}

	fmt.Println(acl)
}
